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

package store

import (
	"context"

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

	"github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"
	"github.com/rossigee/provider-btcpay/internal/features"
)

const (
	errNotStore      = "managed resource is not a Store custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetPC         = "cannot get ProviderConfig"
	errGetCreds      = "cannot get credentials"
	errNewClient     = "cannot create new BTCPay client"
	errCreateStore   = "cannot create store"
	errUpdateStore   = "cannot update store"
	errDeleteStore   = "cannot delete store"
	errGetStore      = "cannot get store"
	errStoreNotFound = "store not found"
)

// Setup adds a controller that reconciles Store managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.StoreGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1beta1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.StoreGroupVersionKind),
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
		For(&v1alpha1.Store{}).
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
	_, ok := mg.(*v1alpha1.Store)
	if !ok {
		return nil, errors.New(errNotStore)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client := clients.NewClient(*config)

	return &external{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.BTCPayClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*v1alpha1.Store)

	// If we don't have an ID yet, the resource doesn't exist
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	store, err := c.client.GetStore(ctx, cr.Status.AtProvider.ID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetStore)
	}

	if store == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Update status with observed state
	cr.Status.AtProvider.ID = store.ID
	cr.Status.AtProvider.Name = store.Name
	cr.Status.AtProvider.Website = store.Website
	cr.Status.AtProvider.DefaultCurrency = store.DefaultCurrency
	cr.Status.AtProvider.InvoiceExpiration = store.InvoiceExpiration
	cr.Status.AtProvider.PaymentMethods = store.PaymentMethods
	cr.Status.AtProvider.DerivationSchemes = convertDerivationSchemes(store.DerivationSchemes)

	if store.CreatedAt != nil {
		cr.Status.AtProvider.CreatedAt = &metav1.Time{Time: *store.CreatedAt}
	}

	// Check if the resource is up to date
	upToDate := c.isUpToDate(cr, store)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		ConnectionDetails: managed.ConnectionDetails{
			"storeId":         []byte(store.ID),
			"storeName":       []byte(store.Name),
			"defaultCurrency": []byte(store.DefaultCurrency),
		},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*v1alpha1.Store)

	cr.Status.SetConditions(xpv1.Creating())

	req := clients.CreateStoreRequest{
		Name:            cr.Spec.ForProvider.Name,
		DefaultCurrency: cr.Spec.ForProvider.DefaultCurrency,
	}

	if cr.Spec.ForProvider.Website != nil {
		req.Website = *cr.Spec.ForProvider.Website
	}
	if cr.Spec.ForProvider.InvoiceExpiration != nil {
		req.InvoiceExpiration = *cr.Spec.ForProvider.InvoiceExpiration
	}
	if cr.Spec.ForProvider.MonitoringExpiration != nil {
		req.MonitoringExpiration = *cr.Spec.ForProvider.MonitoringExpiration
	}
	if cr.Spec.ForProvider.PaymentTolerance != nil {
		req.PaymentTolerance = *cr.Spec.ForProvider.PaymentTolerance
	}
	if cr.Spec.ForProvider.SpeedPolicy != nil {
		speedPolicy, err := convertSpeedPolicy(*cr.Spec.ForProvider.SpeedPolicy)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, "invalid speedPolicy value")
		}
		req.SpeedPolicy = speedPolicy
	}

	store, err := c.client.CreateStore(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateStore)
	}

	cr.Status.AtProvider.ID = store.ID

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"storeId":         []byte(store.ID),
			"storeName":       []byte(store.Name),
			"defaultCurrency": []byte(store.DefaultCurrency),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*v1alpha1.Store)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New(errStoreNotFound)
	}

	req := clients.UpdateStoreRequest{
		Name:            cr.Spec.ForProvider.Name,
		DefaultCurrency: cr.Spec.ForProvider.DefaultCurrency,
	}

	if cr.Spec.ForProvider.Website != nil {
		req.Website = *cr.Spec.ForProvider.Website
	}
	if cr.Spec.ForProvider.InvoiceExpiration != nil {
		req.InvoiceExpiration = *cr.Spec.ForProvider.InvoiceExpiration
	}
	if cr.Spec.ForProvider.MonitoringExpiration != nil {
		req.MonitoringExpiration = *cr.Spec.ForProvider.MonitoringExpiration
	}
	if cr.Spec.ForProvider.PaymentTolerance != nil {
		req.PaymentTolerance = *cr.Spec.ForProvider.PaymentTolerance
	}
	if cr.Spec.ForProvider.SpeedPolicy != nil {
		speedPolicy, err := convertSpeedPolicy(*cr.Spec.ForProvider.SpeedPolicy)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "invalid speedPolicy value")
		}
		req.SpeedPolicy = speedPolicy
	}

	_, err := c.client.UpdateStore(ctx, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateStore)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*v1alpha1.Store)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil // Nothing to delete
	}

	cr.Status.SetConditions(xpv1.Deleting())

	err := c.client.DeleteStore(ctx, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteStore)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for BTCPay API client
	return nil
}

// isUpToDate checks if the Store resource is up to date with the desired state
func (c *external) isUpToDate(cr *v1alpha1.Store, store *clients.Store) bool {
	if cr.Spec.ForProvider.Name != store.Name {
		return false
	}
	if cr.Spec.ForProvider.DefaultCurrency != store.DefaultCurrency {
		return false
	}
	if cr.Spec.ForProvider.Website != nil && *cr.Spec.ForProvider.Website != store.Website {
		return false
	}
	if cr.Spec.ForProvider.InvoiceExpiration != nil && *cr.Spec.ForProvider.InvoiceExpiration != store.InvoiceExpiration {
		return false
	}
	if cr.Spec.ForProvider.MonitoringExpiration != nil && *cr.Spec.ForProvider.MonitoringExpiration != store.MonitoringExpiration {
		return false
	}
	if cr.Spec.ForProvider.PaymentTolerance != nil && *cr.Spec.ForProvider.PaymentTolerance != store.PaymentTolerance {
		return false
	}
	if cr.Spec.ForProvider.SpeedPolicy != nil {
		speedPolicy, err := convertSpeedPolicy(*cr.Spec.ForProvider.SpeedPolicy)
		if err != nil || speedPolicy != store.SpeedPolicy {
			return false
		}
	}
	return true
}

// convertSpeedPolicy converts string speed policy to integer
func convertSpeedPolicy(policy string) (int, error) {
	switch policy {
	case "High":
		return 1, nil
	case "Medium":
		return 6, nil
	case "Low":
		return 144, nil
	default:
		return 0, errors.New("invalid speedPolicy: must be High, Medium, or Low")
	}
}

// convertDerivationSchemes converts the derivation schemes map
// Only includes string values; non-string values are silently skipped
func convertDerivationSchemes(schemes map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range schemes {
		if str, ok := v.(string); ok {
			result[k] = str
		}
		// Non-string values are silently dropped - this is expected behavior
		// for BTCPay API responses that may contain complex objects
	}
	return result
}
