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

package apikey

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	apikeyv1beta1 "github.com/rossigee/provider-btcpay/apis/apikey/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotApiKey      = "managed resource is not a ApiKey custom resource"
	errTrackPCUsage   = "cannot track ProviderConfig usage"
	errGetPC          = "cannot get ProviderConfig"
	errGetCreds       = "cannot get credentials"
	errNewClient      = "cannot create new BTCPay client"
	errCreateApiKey   = "cannot create API key"
	errDeleteApiKey   = "cannot delete API key"
	errGetApiKey      = "cannot get API key"
	errApiKeyNotFound = "API key not found"
)

// Setup adds a controller that reconciles ApiKey managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(apikeyv1beta1.ApiKeyGroupKind.String())

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
		resource.ManagedKind(apikeyv1beta1.ApiKeyGroupVersionKind),
		opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&apikeyv1beta1.ApiKey{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*apikeyv1beta1.ApiKey)
	if !ok {
		return nil, errors.New(errNotApiKey)
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
	cr := mg.(*apikeyv1beta1.ApiKey)

	var externalName string
	if cr.Annotations != nil {
		externalName = cr.Annotations["crossplane.io/external-name"]
	}

	apiKeyID := cr.Status.AtProvider.ID
	if apiKeyID == "" && externalName != "" && externalName != cr.Name {
		apiKeyID = externalName
	}

	if apiKeyID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	apiKey, err := c.client.GetApiKey(ctx, apiKeyID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetApiKey)
	}

	if apiKey == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider.ID = apiKey.ID
	cr.Status.AtProvider.Label = apiKey.Label
	cr.Status.AtProvider.Permissions = apiKey.Permissions

	cr.Status.SetConditions(xpv2.Available())

	upToDate := c.isUpToDate(cr, apiKey)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*apikeyv1beta1.ApiKey)

	cr.Status.SetConditions(xpv2.Creating())

	req := clients.CreateApiKeyRequest{
		Label:       cr.Spec.ForProvider.Label,
		Permissions: cr.Spec.ForProvider.Permissions,
	}

	apiKey, err := c.client.CreateApiKey(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateApiKey)
	}

	if cr.Annotations == nil {
		cr.Annotations = make(map[string]string)
	}
	cr.Annotations["crossplane.io/external-name"] = apiKey.ID
	cr.Status.AtProvider.ID = apiKey.ID
	cr.Status.AtProvider.Label = apiKey.Label
	cr.Status.AtProvider.Permissions = apiKey.Permissions
	cr.Status.AtProvider.ApiKey = apiKey.ApiKey

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"apiKey": []byte(apiKey.ApiKey),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// API keys are immutable after creation (only label/permissions can be changed via recreate)
	// For now, treat as up-to-date if observed
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*apikeyv1beta1.ApiKey)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil
	}

	cr.Status.SetConditions(xpv2.Deleting())

	err := c.client.DeleteApiKey(ctx, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteApiKey)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

func (c *external) isUpToDate(cr *apikeyv1beta1.ApiKey, apiKey *clients.ApiKey) bool {
	if cr.Spec.ForProvider.Label != apiKey.Label {
		return false
	}
	if !equalStringSlices(cr.Spec.ForProvider.Permissions, apiKey.Permissions) {
		return false
	}
	return true
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
