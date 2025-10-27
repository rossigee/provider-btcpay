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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"github.com/rossigee/provider-btcpay/apis/webhook/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
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
	name := managed.ControllerName(v1alpha1.WebhookKind)

	// TODO: Fix v2 connection publishers - temporarily using basic setup
	_ = []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.WebhookGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			// TODO: Fix v2 usage tracker - temporarily removed
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))
	// TODO: Fix v2 connection publishers - managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Webhook{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube client.Client
	// TODO: Fix v2 usage tracker - usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return nil, errors.New(errNotWebhook)
	}

	// TODO: Fix v2 usage tracking
	// if err := c.usage.Track(ctx, mg); err != nil {
	//	return nil, errors.Wrap(err, errTrackPCUsage)
	// }

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
	cr := mg.(*v1alpha1.Webhook)

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

	webhook, err := c.client.GetWebhook(storeID, cr.Status.AtProvider.ID)
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

	// Update status with observed state
	c.updateStatus(cr, webhook)

	// Check if the resource is up to date
	upToDate := c.isUpToDate(cr, webhook)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		ConnectionDetails: managed.ConnectionDetails{
			"webhookId": []byte(webhook.ID),
			"storeId":   []byte(webhook.StoreID),
			"url":       []byte(webhook.URL),
			"enabled":   []byte(boolToString(webhook.Enabled)),
		},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*v1alpha1.Webhook)

	cr.Status.SetConditions(xpv1.Creating())

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	req := clients.CreateWebhookRequest{
		URL: cr.Spec.ForProvider.URL,
	}

	if cr.Spec.ForProvider.Enabled != nil {
		req.Enabled = *cr.Spec.ForProvider.Enabled
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil {
		req.AutomaticRedelivery = *cr.Spec.ForProvider.AutomaticRedelivery
	}
	if cr.Spec.ForProvider.Secret != nil {
		req.Secret = *cr.Spec.ForProvider.Secret
	}
	if cr.Spec.ForProvider.AuthorizedEvents != nil {
		req.AuthorizedEvents = cr.Spec.ForProvider.AuthorizedEvents
	}
	if cr.Spec.ForProvider.PaymentMethod != nil {
		req.PaymentMethod = *cr.Spec.ForProvider.PaymentMethod
	}

	webhook, err := c.client.CreateWebhook(storeID, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateWebhook)
	}

	cr.Status.AtProvider.ID = webhook.ID

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"webhookId": []byte(webhook.ID),
			"storeId":   []byte(webhook.StoreID),
			"url":       []byte(webhook.URL),
			"enabled":   []byte(boolToString(webhook.Enabled)),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*v1alpha1.Webhook)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New(errWebhookNotFound)
	}

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	req := clients.UpdateWebhookRequest{
		URL: cr.Spec.ForProvider.URL,
	}

	if cr.Spec.ForProvider.Enabled != nil {
		req.Enabled = *cr.Spec.ForProvider.Enabled
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil {
		req.AutomaticRedelivery = *cr.Spec.ForProvider.AutomaticRedelivery
	}
	if cr.Spec.ForProvider.Secret != nil {
		req.Secret = *cr.Spec.ForProvider.Secret
	}
	if cr.Spec.ForProvider.AuthorizedEvents != nil {
		req.AuthorizedEvents = cr.Spec.ForProvider.AuthorizedEvents
	}
	if cr.Spec.ForProvider.PaymentMethod != nil {
		req.PaymentMethod = *cr.Spec.ForProvider.PaymentMethod
	}

	_, err = c.client.UpdateWebhook(storeID, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateWebhook)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*v1alpha1.Webhook)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil // Nothing to delete
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// Get the store ID from the referenced store
	storeID, err := c.getStoreID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	err = c.client.DeleteWebhook(storeID, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteWebhook)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for BTCPay API client
	return nil
}

// getStoreID retrieves the store ID from the referenced Store resource
func (c *external) getStoreID(ctx context.Context, cr *v1alpha1.Webhook) (string, error) {
	store := &storev1alpha1.Store{}
	key := client.ObjectKey{Name: cr.Spec.ForProvider.StoreRef.Name}

	if cr.Spec.ForProvider.StoreRef.Namespace != nil {
		key.Namespace = *cr.Spec.ForProvider.StoreRef.Namespace
	}

	if err := c.kube.Get(ctx, key, store); err != nil {
		return "", errors.Wrap(err, errGetStore)
	}

	if store.Status.AtProvider.ID == "" {
		return "", errors.New(errStoreNotReady)
	}

	return store.Status.AtProvider.ID, nil
}

// isUpToDate checks if the Webhook resource is up to date with the desired state
func (c *external) isUpToDate(cr *v1alpha1.Webhook, webhook *clients.Webhook) bool {
	if cr.Spec.ForProvider.URL != webhook.URL {
		return false
	}
	if cr.Spec.ForProvider.Enabled != nil && *cr.Spec.ForProvider.Enabled != webhook.Enabled {
		return false
	}
	if cr.Spec.ForProvider.AutomaticRedelivery != nil && *cr.Spec.ForProvider.AutomaticRedelivery != webhook.AutomaticRedelivery {
		return false
	}
	if cr.Spec.ForProvider.PaymentMethod != nil && *cr.Spec.ForProvider.PaymentMethod != webhook.PaymentMethod {
		return false
	}
	return true
}

// updateStatus updates the webhook status with observed state
func (c *external) updateStatus(cr *v1alpha1.Webhook, webhook *clients.Webhook) {
	cr.Status.AtProvider.ID = webhook.ID
	cr.Status.AtProvider.StoreID = webhook.StoreID
	cr.Status.AtProvider.URL = webhook.URL
	cr.Status.AtProvider.Enabled = webhook.Enabled
	cr.Status.AtProvider.AutomaticRedelivery = webhook.AutomaticRedelivery
	cr.Status.AtProvider.AuthorizedEvents = webhook.AuthorizedEvents
	cr.Status.AtProvider.PaymentMethod = webhook.PaymentMethod
	cr.Status.AtProvider.LastDeliveryErrorMessage = webhook.LastDeliveryErrorMessage
	cr.Status.AtProvider.DeliveryCount = webhook.DeliveryCount
	cr.Status.AtProvider.SuccessfulDeliveryCount = webhook.SuccessfulDeliveryCount

	if webhook.CreatedAt != nil {
		cr.Status.AtProvider.CreatedAt = &metav1.Time{Time: *webhook.CreatedAt}
	}
	if webhook.LastDeliveryAt != nil {
		cr.Status.AtProvider.LastDeliveryAt = &metav1.Time{Time: *webhook.LastDeliveryAt}
	}
}

// boolToString converts a boolean to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
