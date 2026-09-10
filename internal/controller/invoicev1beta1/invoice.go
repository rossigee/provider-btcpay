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

package invoicev1beta1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	invoicev1beta1 "github.com/rossigee/provider-btcpay/apis/invoice/v1beta1"
	storev1beta1 "github.com/rossigee/provider-btcpay/apis/store/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotInvoice      = "managed resource is not an Invoice custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetCreds        = "cannot get credentials"
	errCreateInvoice   = "cannot create invoice"
	errDeleteInvoice   = "cannot delete invoice"
	errGetInvoice      = "cannot get invoice"
	errGetStore        = "cannot get referenced store"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(invoicev1beta1.InvoiceGroupKind.String())

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnector(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(invoicev1beta1.InvoiceGroupVersionKind),
		opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&invoicev1beta1.Invoice{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*invoicev1beta1.Invoice)
	if !ok {
		return nil, errors.New(errNotInvoice)
	}

	if err := c.usage.Track(ctx, cr); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pcName := ""
	if cr.Spec.ProviderConfigReference != nil {
		pcName = cr.Spec.ProviderConfigReference.Name
	}
	config, err := clients.GetConfig(ctx, c.kube, pcName)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client := clients.NewClient(*config)

	return &external{kube: c.kube, client: client}, nil
}

type external struct {
	kube  client.Client
	client clients.BTCPayClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*invoicev1beta1.Invoice)

	var externalName string
	if cr.Annotations != nil {
		externalName = cr.Annotations["crossplane.io/external-name"]
	}

	invoiceID := cr.Status.AtProvider.ID
	if invoiceID == "" && externalName != "" && externalName != cr.Name {
		invoiceID = externalName
	}

	if invoiceID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	invoice, err := c.client.GetInvoice(ctx, invoiceID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetInvoice)
	}

	cr.Status.AtProvider.ID = invoice.ID
	cr.Status.AtProvider.StoreID = invoice.StoreID
	cr.Status.AtProvider.Amount = invoice.Amount
	cr.Status.AtProvider.Currency = invoice.Currency
	cr.Status.AtProvider.Status = invoice.Status
	cr.Status.AtProvider.CheckoutLink = invoice.CheckoutLink

	cr.Status.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: false,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*invoicev1beta1.Invoice)

	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	invoice, err := c.client.CreateInvoice(ctx, storeID, &clients.CreateInvoiceParams{
		Amount:             cr.Spec.ForProvider.Amount,
		Currency:           cr.Spec.ForProvider.Currency,
		OrderID:            cr.Spec.ForProvider.OrderID,
		NotificationURL:    cr.Spec.ForProvider.NotificationURL,
		RedirectURL:        cr.Spec.ForProvider.RedirectURL,
		NotificationEmail:  cr.Spec.ForProvider.NotificationEmail,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateInvoice)
	}

	cr.Status.AtProvider.ID = invoice.ID
	cr.Status.AtProvider.StoreID = storeID
	cr.Status.AtProvider.CheckoutLink = invoice.CheckoutLink
	cr.Status.SetConditions(xpv2.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*invoicev1beta1.Invoice)

	_, err := c.client.UpdateInvoice(ctx, cr.Status.AtProvider.ID, &clients.UpdateInvoiceParams{
		OrderID: cr.Spec.ForProvider.OrderID,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update invoice")
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr := mg.(*invoicev1beta1.Invoice)

	if cr.Status.AtProvider.ID == "" {
		return nil
	}

	err := c.client.DeleteInvoice(ctx, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return errors.Wrap(err, errDeleteInvoice)
	}

	cr.Status.SetConditions(xpv2.Deleting())

	return nil
}

func (c *external) getStoreID(ctx context.Context, cr *invoicev1beta1.Invoice) (string, error) {
	storeName := cr.Spec.ForProvider.StoreRef.Name
	storeNamespace := cr.Spec.ForProvider.StoreRef.Namespace
	if storeNamespace == nil {
		ns := "default"
		storeNamespace = &ns
	}

	store := &storev1beta1.Store{}
	err := c.kube.Get(ctx, client.ObjectKey{
		Name:      storeName,
		Namespace: *storeNamespace,
	}, store)
	if err != nil {
		return "", errors.Wrap(err, errGetStore)
	}

	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errGetStore)
	}

	return store.Status.AtProvider.ID, nil
}
