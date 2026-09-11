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

package sharedlink

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	sharedlinkv1alpha1 "github.com/rossigee/provider-btcpay/apis/sharedlink/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotSharedLink    = "managed resource is not a SharedLink custom resource"
	errTrackPCUsage     = "cannot track ProviderConfig usage"
	errGetCreds         = "cannot get credentials"
	errCreateSharedLink = "cannot create shared link"
	errDeleteSharedLink = "cannot delete shared link"
	errGetSharedLink    = "cannot get shared link"
	errGetStore         = "cannot get referenced store"
	errStoreNotReady    = "referenced store is not ready"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(sharedlinkv1alpha1.SharedLinkGroupKind.String())
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
	r := managed.NewReconciler(mgr, resource.ManagedKind(sharedlinkv1alpha1.SharedLinkGroupVersionKind), opts...)
	return ctrl.NewControllerManagedBy(mgr).Named(name).WithOptions(o.ForControllerRuntime()).WithEventFilter(resource.DesiredStateChanged()).For(&sharedlinkv1alpha1.SharedLink{}).Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*sharedlinkv1alpha1.SharedLink)
	if !ok {
		return nil, errors.New(errNotSharedLink)
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
	return &external{client: client, kube: c.kube}, nil
}

type external struct {
	client clients.BTCPayClient
	kube   client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*sharedlinkv1alpha1.SharedLink)
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	link, err := c.client.GetSharedLink(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetSharedLink)
	}
	if link == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	cr.Status.AtProvider.ID = link.ID
	cr.Status.AtProvider.StoreID = link.StoreID
	cr.Status.AtProvider.Amount = link.Amount
	cr.Status.AtProvider.Currency = link.Currency
	cr.Status.AtProvider.Description = link.Description
	cr.Status.AtProvider.URL = link.URL
	cr.Status.SetConditions(xpv2.Available())
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: c.isUpToDate(cr, link)}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*sharedlinkv1alpha1.SharedLink)
	cr.Status.SetConditions(xpv2.Creating())
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	req := clients.CreateSharedLinkRequest{}
	if cr.Spec.ForProvider.Amount != nil {
		req.Amount = cr.Spec.ForProvider.Amount
	}
	if cr.Spec.ForProvider.Currency != nil {
		req.Currency = cr.Spec.ForProvider.Currency
	}
	if cr.Spec.ForProvider.Description != nil {
		req.Description = cr.Spec.ForProvider.Description
	}
	link, err := c.client.CreateSharedLink(ctx, storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateSharedLink)
	}
	if cr.Annotations == nil {
		cr.Annotations = make(map[string]string)
	}
	cr.Annotations["crossplane.io/external-name"] = link.ID
	cr.Status.AtProvider.ID = link.ID
	cr.Status.AtProvider.URL = link.URL
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*sharedlinkv1alpha1.SharedLink)
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil
	}
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, err
	}
	cr.Status.SetConditions(xpv2.Deleting())
	err = c.client.DeleteSharedLink(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteSharedLink)
	}
	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error { return nil }

func (c *external) getStoreID(ctx context.Context, cr *sharedlinkv1alpha1.SharedLink) (string, error) {
	namespace := cr.Namespace
	if cr.Spec.ForProvider.StoreRef.Namespace != nil {
		namespace = *cr.Spec.ForProvider.StoreRef.Namespace
	}
	store := &storev1alpha1.Store{}
	if err := c.kube.Get(ctx, client.ObjectKey{Name: cr.Spec.ForProvider.StoreRef.Name, Namespace: namespace}, store); err != nil {
		return "", errors.Wrap(err, errGetStore)
	}
	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errStoreNotReady)
	}
	return store.Status.AtProvider.ID, nil
}

func (c *external) isUpToDate(cr *sharedlinkv1alpha1.SharedLink, link *clients.SharedLink) bool {
	if cr.Spec.ForProvider.Amount != nil && *cr.Spec.ForProvider.Amount != link.Amount {
		return false
	}
	if cr.Spec.ForProvider.Currency != nil && *cr.Spec.ForProvider.Currency != link.Currency {
		return false
	}
	if cr.Spec.ForProvider.Description != nil && *cr.Spec.ForProvider.Description != link.Description {
		return false
	}
	return true
}
