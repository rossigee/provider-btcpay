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

package paymentmethod

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-btcpay/apis/paymentmethod/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
)

const (
	errNotPaymentMethod      = "managed resource is not a PaymentMethod custom resource"
	errTrackPCUsage          = "cannot track ProviderConfig usage"
	errGetPC                 = "cannot get ProviderConfig"
	errGetCreds              = "cannot get credentials"
	errNewClient             = "cannot create new BTCPay client"
	errCreatePaymentMethod   = "cannot create payment method"
	errUpdatePaymentMethod   = "cannot update payment method"
	errDeletePaymentMethod   = "cannot delete payment method"
	errGetPaymentMethod      = "cannot get payment method"
	errPaymentMethodNotFound = "payment method not found"
	errGetStore              = "cannot get referenced store"
	errStoreNotFound         = "referenced store not found"
	errStoreNotReady         = "referenced store is not ready"
)

// Setup adds a controller that reconciles PaymentMethod managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PaymentMethodKind)

	// TODO: Fix v2 connection publishers - temporarily using basic setup
	_ = []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PaymentMethodGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.PaymentMethod{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube client.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.PaymentMethod)
	if !ok {
		return nil, errors.New(errNotPaymentMethod)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client := clients.NewClient(*cfg)

	return &external{client: client, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.BTCPayClient
	kube   client.Client
}

func (c *external) Disconnect(ctx context.Context) error {
	// No cleanup needed for BTCPay client
	return nil
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.PaymentMethod)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPaymentMethod)
	}

	// Get the external-name annotation to identify the payment method
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Resolve Store reference to get Store ID
	storeID, err := c.resolveStoreReference(ctx, &cr.Spec.ForProvider.StoreRef)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetStore)
	}

	// Parse external name to get crypto code and payment type
	parts := strings.SplitN(externalName, "-", 2)
	if len(parts) != 2 {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, errors.New("invalid external name format, expected 'cryptoCode-paymentType'")
	}
	cryptoCode := parts[0]
	paymentType := parts[1]

	paymentMethod, err := c.client.GetStorePaymentMethod(storeID, cryptoCode, paymentType)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPaymentMethod)
	}

	// Update status with observed values
	cr.Status.AtProvider.PaymentMethodID = paymentMethod.PaymentMethod
	cr.Status.AtProvider.StoreID = storeID
	cr.Status.AtProvider.CryptoCode = paymentMethod.CryptoCode
	cr.Status.AtProvider.PaymentType = paymentMethod.PaymentType
	cr.Status.AtProvider.Enabled = paymentMethod.Enabled
	cr.Status.AtProvider.WalletBalance = &paymentMethod.WalletBalance
	cr.Status.AtProvider.SyncStatus = paymentMethod.SyncStatus
	cr.Status.AtProvider.ErrorMessage = paymentMethod.ErrorMessage

	if paymentMethod.CreatedAt != nil {
		cr.Status.AtProvider.CreatedAt = &metav1.Time{Time: *paymentMethod.CreatedAt}
	}
	if paymentMethod.LastSyncAt != nil {
		cr.Status.AtProvider.LastSyncAt = &metav1.Time{Time: *paymentMethod.LastSyncAt}
	}

	// Payment method is considered up to date if configuration matches
	upToDate := c.isUpToDate(cr, paymentMethod)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.PaymentMethod)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPaymentMethod)
	}

	cr.Status.SetConditions(xpv1.Creating())

	// Resolve Store reference
	storeID, err := c.resolveStoreReference(ctx, &cr.Spec.ForProvider.StoreRef)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGetStore)
	}

	// Build create request
	req := clients.CreateStorePaymentMethodRequest{
		PaymentMethod: fmt.Sprintf("%s-%s", cr.Spec.ForProvider.CryptoCode, cr.Spec.ForProvider.PaymentType),
		CryptoCode:    cr.Spec.ForProvider.CryptoCode,
		PaymentType:   cr.Spec.ForProvider.PaymentType,
		Enabled:       cr.Spec.ForProvider.Enabled != nil && *cr.Spec.ForProvider.Enabled,
	}

	// Set type-specific configuration
	if cr.Spec.ForProvider.OnChain != nil && cr.Spec.ForProvider.PaymentType == "onchain" {
		req.DerivationScheme = cr.Spec.ForProvider.OnChain.DerivationScheme
		if cr.Spec.ForProvider.OnChain.Label != nil {
			req.Label = *cr.Spec.ForProvider.OnChain.Label
		}
	}

	if cr.Spec.ForProvider.Lightning != nil && cr.Spec.ForProvider.PaymentType == "lightning" {
		req.ConnectionString = cr.Spec.ForProvider.Lightning.ConnectionString
	}

	paymentMethod, err := c.client.CreateStorePaymentMethod(storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePaymentMethod)
	}

	// Set external name for future identification
	externalName := fmt.Sprintf("%s-%s", paymentMethod.CryptoCode, paymentMethod.PaymentType)
	meta.SetExternalName(cr, externalName)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.PaymentMethod)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPaymentMethod)
	}

	// Resolve Store reference
	storeID, err := c.resolveStoreReference(ctx, &cr.Spec.ForProvider.StoreRef)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetStore)
	}

	// Build update request
	req := clients.UpdateStorePaymentMethodRequest{
		Enabled: cr.Spec.ForProvider.Enabled,
	}

	// Set type-specific configuration
	if cr.Spec.ForProvider.OnChain != nil && cr.Spec.ForProvider.PaymentType == "onchain" {
		req.DerivationScheme = cr.Spec.ForProvider.OnChain.DerivationScheme
		if cr.Spec.ForProvider.OnChain.Label != nil {
			req.Label = *cr.Spec.ForProvider.OnChain.Label
		}
	}

	if cr.Spec.ForProvider.Lightning != nil && cr.Spec.ForProvider.PaymentType == "lightning" {
		req.ConnectionString = cr.Spec.ForProvider.Lightning.ConnectionString
	}

	_, err = c.client.UpdateStorePaymentMethod(storeID, cr.Spec.ForProvider.CryptoCode, cr.Spec.ForProvider.PaymentType, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePaymentMethod)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.PaymentMethod)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPaymentMethod)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// Resolve Store reference
	storeID, err := c.resolveStoreReference(ctx, &cr.Spec.ForProvider.StoreRef)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errGetStore)
	}

	err = c.client.DeleteStorePaymentMethod(storeID, cr.Spec.ForProvider.CryptoCode, cr.Spec.ForProvider.PaymentType)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeletePaymentMethod)
	}

	return managed.ExternalDelete{}, nil
}

// resolveStoreReference resolves a store reference to get the BTCPay Store ID
func (c *external) resolveStoreReference(ctx context.Context, storeRef *v1alpha1.StoreReference) (string, error) {
	// Get the referenced Store resource
	store := &storev1alpha1.Store{}
	storeKey := client.ObjectKey{
		Name:      storeRef.Name,
		Namespace: "",
	}
	if storeRef.Namespace != nil {
		storeKey.Namespace = *storeRef.Namespace
	}

	if err := c.kube.Get(ctx, storeKey, store); err != nil {
		return "", errors.Wrap(err, errStoreNotFound)
	}

	// Check if the store is ready
	ready := false
	for _, condition := range store.Status.Conditions {
		if condition.Type == xpv1.TypeReady && condition.Status == corev1.ConditionTrue {
			ready = true
			break
		}
	}
	if !ready {
		return "", errors.New(errStoreNotReady)
	}

	// Get the Store ID from the status
	storeID := store.Status.AtProvider.ID
	if storeID == "" {
		return "", errors.New("store has no ID in status")
	}

	return storeID, nil
}

// isUpToDate checks if the payment method is up to date with the desired configuration
func (c *external) isUpToDate(cr *v1alpha1.PaymentMethod, observed *clients.StorePaymentMethod) bool {
	// Check enabled state
	desired := cr.Spec.ForProvider.Enabled != nil && *cr.Spec.ForProvider.Enabled
	if observed.Enabled != desired {
		return false
	}

	// Check crypto code and payment type (these shouldn't change)
	if observed.CryptoCode != cr.Spec.ForProvider.CryptoCode {
		return false
	}
	if observed.PaymentType != cr.Spec.ForProvider.PaymentType {
		return false
	}

	// Configuration details are complex to compare due to different storage formats
	// For now, assume always up to date if basic fields match
	return true
}
