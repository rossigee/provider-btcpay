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

package guest

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	guestv1beta1 "github.com/rossigee/provider-btcpay/apis/guest/v1beta1"
	storev1beta1 "github.com/rossigee/provider-btcpay/apis/store/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotGuest      = "managed resource is not a Guest custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetCreds      = "cannot get credentials"
	errCreateGuest   = "cannot create guest"
	errDeleteGuest   = "cannot delete guest"
	errGetGuest      = "cannot get guest"
	errGetStore      = "cannot get referenced store"
	errStoreNotReady = "referenced store is not ready"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(guestv1beta1.GuestGroupKind.String())
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
	r := managed.NewReconciler(mgr, resource.ManagedKind(guestv1beta1.GuestGroupVersionKind), opts...)
	return ctrl.NewControllerManagedBy(mgr).Named(name).WithOptions(o.ForControllerRuntime()).WithEventFilter(resource.DesiredStateChanged()).For(&guestv1beta1.Guest{}).Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*guestv1beta1.Guest)
	if !ok {
		return nil, errors.New(errNotGuest)
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
	cr := mg.(*guestv1beta1.Guest)
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	guest, err := c.client.GetGuest(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetGuest)
	}
	if guest == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	cr.Status.AtProvider.ID = guest.ID
	cr.Status.AtProvider.StoreID = guest.StoreID
	cr.Status.AtProvider.Email = guest.Email
	cr.Status.AtProvider.Name = guest.Name
	cr.Status.SetConditions(xpv2.Available())
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: c.isUpToDate(cr, guest)}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*guestv1beta1.Guest)
	cr.Status.SetConditions(xpv2.Creating())
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	req := clients.CreateGuestRequest{Email: cr.Spec.ForProvider.Email}
	if cr.Spec.ForProvider.Name != nil {
		req.Name = cr.Spec.ForProvider.Name
	}
	guest, err := c.client.CreateGuest(ctx, storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateGuest)
	}
	if cr.Annotations == nil {
		cr.Annotations = make(map[string]string)
	}
	cr.Annotations["crossplane.io/external-name"] = guest.ID
	cr.Status.AtProvider.ID = guest.ID
	cr.Status.AtProvider.StoreID = guest.StoreID
	cr.Status.AtProvider.Email = guest.Email
	cr.Status.AtProvider.Name = guest.Name
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*guestv1beta1.Guest)
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil
	}
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, err
	}
	cr.Status.SetConditions(xpv2.Deleting())
	err = c.client.DeleteGuest(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteGuest)
	}
	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error { return nil }

func (c *external) getStoreID(ctx context.Context, cr *guestv1beta1.Guest) (string, error) {
	namespace := cr.Namespace
	if cr.Spec.ForProvider.StoreRef.Namespace != nil {
		namespace = *cr.Spec.ForProvider.StoreRef.Namespace
	}
	store := &storev1beta1.Store{}
	if err := c.kube.Get(ctx, client.ObjectKey{Name: cr.Spec.ForProvider.StoreRef.Name, Namespace: namespace}, store); err != nil {
		return "", errors.Wrap(err, errGetStore)
	}
	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errStoreNotReady)
	}
	return store.Status.AtProvider.ID, nil
}

func (c *external) isUpToDate(cr *guestv1beta1.Guest, guest *clients.Guest) bool {
	if cr.Spec.ForProvider.Email != guest.Email {
		return false
	}
	if cr.Spec.ForProvider.Name != nil && *cr.Spec.ForProvider.Name != guest.Name {
		return false
	}
	return true
}
