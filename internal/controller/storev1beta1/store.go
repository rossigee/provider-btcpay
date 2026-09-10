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

package storev1beta1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	storev1beta1 "github.com/rossigee/provider-btcpay/apis/store/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotStore      = "managed resource is not a Store custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetCreds      = "cannot get credentials"
	errCreateStore   = "cannot create store"
	errUpdateStore   = "cannot update store"
	errDeleteStore   = "cannot delete store"
	errGetStore      = "cannot get store"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(storev1beta1.StoreGroupKind.String())

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
		resource.ManagedKind(storev1beta1.StoreGroupVersionKind),
		opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&storev1beta1.Store{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*storev1beta1.Store)
	if !ok {
		return nil, errors.New(errNotStore)
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

	return &external{client: client}, nil
}

type external struct {
	client clients.BTCPayClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*storev1beta1.Store)

	var externalName string
	if cr.Annotations != nil {
		externalName = cr.Annotations["crossplane.io/external-name"]
	}

	storeID := cr.Status.AtProvider.ID
	if storeID == "" && externalName != "" && externalName != cr.Name {
		storeID = externalName
	}

	if storeID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	store, err := c.client.GetStore(ctx, storeID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetStore)
	}

	cr.Status.AtProvider.ID = store.ID
	cr.Status.AtProvider.Name = store.Name
	cr.Status.AtProvider.DefaultCurrency = store.DefaultCurrency

	cr.Status.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: false,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*storev1beta1.Store)

	store, err := c.client.CreateStore(ctx, &clients.CreateStoreParams{
		Name:              cr.Spec.ForProvider.Name,
		DefaultCurrency:   cr.Spec.ForProvider.DefaultCurrency,
		Website:           cr.Spec.ForProvider.Website,
		InvoiceExpiration: cr.Spec.ForProvider.InvoiceExpiration,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateStore)
	}

	cr.Status.AtProvider.ID = store.ID
	cr.Status.SetConditions(xpv2.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*storev1beta1.Store)

	_, err := c.client.UpdateStore(ctx, cr.Status.AtProvider.ID, &clients.UpdateStoreParams{
		Name:              cr.Spec.ForProvider.Name,
		DefaultCurrency:   cr.Spec.ForProvider.DefaultCurrency,
		Website:           cr.Spec.ForProvider.Website,
		InvoiceExpiration: cr.Spec.ForProvider.InvoiceExpiration,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateStore)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr := mg.(*storev1beta1.Store)

	if cr.Status.AtProvider.ID == "" {
		return nil
	}

	err := c.client.DeleteStore(ctx, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return errors.Wrap(err, errDeleteStore)
	}

	cr.Status.SetConditions(xpv2.Deleting())

	return nil
}
