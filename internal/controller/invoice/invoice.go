/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package invoice

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"
	"github.com/rossigee/provider-btcpay/internal/features"
)

const (
	errNotInvoice      = "managed resource is not an Invoice custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errNewClient       = "cannot create new BTCPay client"
	errCreateInvoice   = "cannot create invoice"
	errDeleteInvoice   = "cannot delete invoice"
	errGetInvoice      = "cannot get invoice"
	errInvoiceNotFound = "invoice not found"
	errGetStore        = "cannot get referenced store"
	errStoreNotFound   = "referenced store not found"
	errStoreNotReady   = "referenced store is not ready"
)

// Setup adds a controller that reconciles Invoice managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.InvoiceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1beta1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.InvoiceGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Invoice{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Invoice)
	if !ok {
		return nil, errors.New(errNotInvoice)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client := clients.NewClient(*config)

	return &external{client: client, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.BTCPayClient
	kube   client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*v1alpha1.Invoice)

	// If we don't have an ID yet, the resource doesn't exist
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	invoice, err := c.client.GetInvoice(storeID, cr.Status.AtProvider.ID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetInvoice)
	}

	if invoice == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Update status with observed state
	c.updateStatus(cr, invoice)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true, // Invoices are typically immutable after creation
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*v1alpha1.Invoice)

	cr.Status.SetConditions(xpv1.Creating())

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	req := clients.CreateInvoiceRequest{
		Amount:   cr.Spec.ForProvider.Amount,
		Currency: cr.Spec.ForProvider.Currency,
	}

	if cr.Spec.ForProvider.OrderID != nil {
		req.OrderID = *cr.Spec.ForProvider.OrderID
	}
	if cr.Spec.ForProvider.NotificationURL != nil {
		req.NotificationURL = *cr.Spec.ForProvider.NotificationURL
	}
	if cr.Spec.ForProvider.RedirectURL != nil {
		req.RedirectURL = *cr.Spec.ForProvider.RedirectURL
	}
	if cr.Spec.ForProvider.NotificationEmail != nil {
		req.NotificationEmail = *cr.Spec.ForProvider.NotificationEmail
	}
	if cr.Spec.ForProvider.ItemDesc != nil {
		req.ItemDesc = *cr.Spec.ForProvider.ItemDesc
	}
	if cr.Spec.ForProvider.ItemCode != nil {
		req.ItemCode = *cr.Spec.ForProvider.ItemCode
	}
	if cr.Spec.ForProvider.Physical != nil {
		req.Physical = *cr.Spec.ForProvider.Physical
	}
	if cr.Spec.ForProvider.TaxIncluded != nil {
		req.TaxIncluded = *cr.Spec.ForProvider.TaxIncluded
	}
	if cr.Spec.ForProvider.BuyerEmail != nil {
		req.BuyerEmail = *cr.Spec.ForProvider.BuyerEmail
	}
	if cr.Spec.ForProvider.ExtendedNotifications != nil {
		req.ExtendedNotifications = *cr.Spec.ForProvider.ExtendedNotifications
	}
	if cr.Spec.ForProvider.FullNotifications != nil {
		req.FullNotifications = *cr.Spec.ForProvider.FullNotifications
	}
	if cr.Spec.ForProvider.Metadata != nil {
		req.Metadata = convertMetadata(cr.Spec.ForProvider.Metadata)
	}
	if cr.Spec.ForProvider.CheckoutQueryString != nil {
		req.CheckoutQueryString = *cr.Spec.ForProvider.CheckoutQueryString
	}

	invoice, err := c.client.CreateInvoice(storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateInvoice)
	}

	cr.Status.AtProvider.ID = invoice.ID
	c.updateStatus(cr, invoice)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Invoices are typically immutable after creation in BTCPay
	// We only observe their status changes
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*v1alpha1.Invoice)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil // Nothing to delete
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	// Archive the invoice (BTCPay doesn't allow hard deletion)
	err = c.client.ArchiveInvoice(storeID, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteInvoice)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for BTCPay API client
	return nil
}

// getStoreID retrieves the store ID from the referenced Store resource
func (c *external) getStoreID(ctx context.Context, cr *v1alpha1.Invoice) (string, error) {
	store := &storev1alpha1.Store{}
	key := client.ObjectKey{Name: cr.Spec.ForProvider.StoreRef.Name}

	if err := c.kube.Get(ctx, key, store); err != nil {
		return "", errors.Wrap(err, errGetStore)
	}

	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errStoreNotReady)
	}

	return store.Status.AtProvider.ID, nil
}

// updateStatus updates the invoice status with observed state
func (c *external) updateStatus(cr *v1alpha1.Invoice, invoice *clients.Invoice) {
	cr.Status.AtProvider.ID = invoice.ID
	cr.Status.AtProvider.StoreID = invoice.StoreID

	// Convert amount from string to float64
	if amount, err := strconv.ParseFloat(invoice.Amount, 64); err == nil {
		cr.Status.AtProvider.Amount = amount
	}
	cr.Status.AtProvider.Currency = invoice.Currency
	cr.Status.AtProvider.Type = invoice.Type
	cr.Status.AtProvider.CheckoutLink = invoice.CheckoutLink
	cr.Status.AtProvider.Status = invoice.Status
	cr.Status.AtProvider.AdditionalStatus = invoice.AdditionalStatus
	cr.Status.AtProvider.Archived = invoice.Archived

	if invoice.MonitoringExpiration != nil {
		cr.Status.AtProvider.MonitoringExpiration = &metav1.Time{Time: *invoice.MonitoringExpiration}
	}
	if invoice.ExpirationTime != nil {
		cr.Status.AtProvider.ExpirationTime = &metav1.Time{Time: *invoice.ExpirationTime}
	}
	if invoice.CreatedTime != nil {
		cr.Status.AtProvider.CreatedTime = &metav1.Time{Time: *invoice.CreatedTime}
	}

	// Convert payment methods
	if len(invoice.PaymentMethods) > 0 {
		cr.Status.AtProvider.PaymentMethods = convertPaymentMethods(invoice.PaymentMethods)
	}
}

// convertMetadata converts string map to interface map
func convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// convertPaymentMethods converts client payment methods to API types
func convertPaymentMethods(methods []clients.PaymentMethod) []v1alpha1.PaymentMethodDetails {
	result := make([]v1alpha1.PaymentMethodDetails, len(methods))
	for i, method := range methods {
		result[i] = v1alpha1.PaymentMethodDetails{
			PaymentMethod:     method.PaymentMethod,
			CryptoCode:        method.CryptoCode,
			Destination:       method.Destination,
			PaymentLink:       method.PaymentLink,
			Rate:              method.Rate,
			PaymentMethodPaid: method.PaymentMethodPaid,
			TotalPaid:         method.TotalPaid,
			Due:               method.Due,
			Amount:            method.Amount,
			NetworkFee:        method.NetworkFee,
			Payments:          convertPayments(method.Payments),
		}
	}
	return result
}

// convertPayments converts client payments to API types
func convertPayments(payments []clients.Payment) []v1alpha1.PaymentDetails {
	result := make([]v1alpha1.PaymentDetails, len(payments))
	for i, payment := range payments {
		result[i] = v1alpha1.PaymentDetails{
			ID:          payment.ID,
			Value:       payment.Value,
			Fee:         payment.Fee,
			Status:      payment.Status,
			Destination: payment.Destination,
		}
		if payment.ReceivedDate != nil {
			result[i].ReceivedDate = &metav1.Time{Time: *payment.ReceivedDate}
		}
	}
	return result
}
