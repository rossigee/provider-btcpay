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

package webhook

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
	webhookv1beta1 "github.com/rossigee/provider-btcpay/apis/webhook/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotWebhook      = "managed resource is not a Webhook custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errNewClient       = "cannot create new BTCPay client"
	errCreateWebhook   = "cannot create webhook"
	errUpdateWebhook   = "cannot update webhook"
	errDeleteWebhook   = "cannot delete webhook"
	errGetWebhook      = "cannot get webhook"
	errWebhookNotFound = "webhook not found"
	errGetStore        = "cannot get referenced store"
	errStoreNotFound   = "referenced store not found"
	errStoreNotReady   = "referenced store is not ready"
)

// Setup adds a controller that reconciles Webhook managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(webhookv1beta1.WebhookGroupKind.String())

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
		resource.ManagedKind(webhookv1beta1.WebhookGroupVersionKind),
		opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&webhookv1beta1.Webhook{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*webhookv1beta1.Webhook)
	if !ok {
		return nil, errors.New(errNotWebhook)
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

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.BTCPayClient
	kube   client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*webhookv1beta1.Webhook)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	webhook, err := c.client.GetWebhook(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetWebhook)
	}

	if webhook == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	c.updateStatus(cr, webhook)
	cr.Status.SetConditions(xpv2.Available())

	upToDate := c.isUpToDate(cr, webhook)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*webhookv1beta1.Webhook)

	cr.Status.SetConditions(xpv2.Creating())

	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	req := clients.CreateWebhookRequest{
		URL: cr.Spec.ForProvider.URL,
	}

	if cr.Spec.ForProvider.Enabled != nil {
		req.Enabled = cr.Spec.ForProvider.Enabled
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil {
		req.AutomaticRedelivery = cr.Spec.ForProvider.AutomaticRedelivery
	}
	if cr.Spec.ForProvider.AuthorizedEvents != nil {
		req.AuthorizedEvents = cr.Spec.ForProvider.AuthorizedEvents
	}
	if cr.Spec.ForProvider.Secret != nil {
		req.Secret = cr.Spec.ForProvider.Secret
	}

	webhook, err := c.client.CreateWebhook(ctx, storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateWebhook)
	}

	if cr.Annotations == nil {
		cr.Annotations = make(map[string]string)
	}
	cr.Annotations["crossplane.io/external-name"] = webhook.ID
	c.updateStatus(cr, webhook)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*webhookv1beta1.Webhook)

	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	req := clients.UpdateWebhookRequest{
		URL: cr.Spec.ForProvider.URL,
	}

	if cr.Spec.ForProvider.Enabled != nil {
		req.Enabled = cr.Spec.ForProvider.Enabled
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil {
		req.AutomaticRedelivery = cr.Spec.ForProvider.AutomaticRedelivery
	}
	if cr.Spec.ForProvider.AuthorizedEvents != nil {
		req.AuthorizedEvents = cr.Spec.ForProvider.AuthorizedEvents
	}
	if cr.Spec.ForProvider.Secret != nil {
		req.Secret = cr.Spec.ForProvider.Secret
	}

	webhook, err := c.client.UpdateWebhook(ctx, storeID, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateWebhook)
	}

	c.updateStatus(cr, webhook)

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*webhookv1beta1.Webhook)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil
	}

	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		if errors.Cause(err).Error() == errStoreNotFound {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, err
	}

	cr.Status.SetConditions(xpv2.Deleting())

	err = c.client.DeleteWebhook(ctx, storeID, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteWebhook)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

func (c *external) getStoreID(ctx context.Context, cr *webhookv1beta1.Webhook) (string, error) {
	storeRef := cr.Spec.ForProvider.StoreRef
	namespace := cr.Namespace
	if storeRef.Namespace != nil {
		namespace = *storeRef.Namespace
	}

	store := &storev1beta1.Store{}
	key := client.ObjectKey{Name: storeRef.Name, Namespace: namespace}
	if err := c.kube.Get(ctx, key, store); err != nil {
		return "", errors.Wrap(err, errGetStore)
	}

	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errStoreNotReady)
	}

	return store.Status.AtProvider.ID, nil
}

func (c *external) updateStatus(cr *webhookv1beta1.Webhook, webhook *clients.Webhook) {
	cr.Status.AtProvider.ID = webhook.ID
	cr.Status.AtProvider.StoreID = webhook.StoreID
	cr.Status.AtProvider.URL = webhook.URL
	cr.Status.AtProvider.Enabled = webhook.Enabled
	cr.Status.AtProvider.AutomaticRedelivery = webhook.AutomaticRedelivery
	cr.Status.AtProvider.AuthorizedEvents = webhook.AuthorizedEvents
	cr.Status.AtProvider.Secret = webhook.Secret
}

func (c *external) isUpToDate(cr *webhookv1beta1.Webhook, webhook *clients.Webhook) bool {
	if cr.Spec.ForProvider.URL != webhook.URL {
		return false
	}
	if cr.Spec.ForProvider.Enabled != nil && *cr.Spec.ForProvider.Enabled != webhook.Enabled {
		return false
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil && *cr.Spec.ForProvider.AutomaticRedelivery != webhook.AutomaticRedelivery {
		return false
	}
	if cr.Spec.ForProvider.Secret != nil && *cr.Spec.ForProvider.Secret != webhook.Secret {
		return false
	}
	if !equalStringSlices(cr.Spec.ForProvider.AuthorizedEvents, webhook.AuthorizedEvents) {
		return false
	}
	return true
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	// Order matters for simplicity; BTCPay returns same order
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
