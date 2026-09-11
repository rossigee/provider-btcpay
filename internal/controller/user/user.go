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

package user

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"

	userv1alpha1 "github.com/rossigee/provider-btcpay/apis/user/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-btcpay/apis/v1beta1"
	"github.com/rossigee/provider-btcpay/internal/clients"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotUser      = "managed resource is not a User custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"
	errNewClient    = "cannot create new BTCPay client"
	errCreateUser   = "cannot create user"
	errUpdateUser   = "cannot update user"
	errDeleteUser   = "cannot delete user"
	errGetUser      = "cannot get user"
	errUserNotFound = "user not found"
)

// Setup adds a controller that reconciles User managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(userv1alpha1.UserGroupKind.String())

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
		resource.ManagedKind(userv1alpha1.UserGroupVersionKind),
		opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&userv1alpha1.User{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage *resource.ProviderConfigUsageTracker
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*userv1alpha1.User)
	if !ok {
		return nil, errors.New(errNotUser)
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
	cr := mg.(*userv1alpha1.User)

	var externalName string
	if cr.Annotations != nil {
		externalName = cr.Annotations["crossplane.io/external-name"]
	}

	userID := cr.Status.AtProvider.ID
	if userID == "" && externalName != "" && externalName != cr.Name {
		userID = externalName
	}

	if userID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	user, err := c.client.GetUser(ctx, userID)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetUser)
	}

	if user == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider.ID = user.ID
	cr.Status.AtProvider.Email = user.Email
	cr.Status.AtProvider.Name = user.Name
	cr.Status.AtProvider.IsAdministrator = user.IsAdministrator
	cr.Status.AtProvider.EmailConfirmed = user.EmailConfirmed
	cr.Status.AtProvider.Approved = user.Approved
	cr.Status.AtProvider.Disabled = user.Disabled

	cr.Status.SetConditions(xpv2.Available())

	upToDate := c.isUpToDate(cr, user)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*userv1alpha1.User)

	cr.Status.SetConditions(xpv2.Creating())

	req := clients.CreateUserRequest{
		Email: cr.Spec.ForProvider.Email,
	}

	if cr.Spec.ForProvider.Name != nil {
		req.Name = cr.Spec.ForProvider.Name
	}
	if cr.Spec.ForProvider.IsAdministrator != nil {
		req.IsAdministrator = cr.Spec.ForProvider.IsAdministrator
	}

	user, err := c.client.CreateUser(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateUser)
	}

	if cr.Annotations == nil {
		cr.Annotations = make(map[string]string)
	}
	cr.Annotations["crossplane.io/external-name"] = user.ID
	cr.Status.AtProvider.ID = user.ID
	if user.Created != "" {
		if t, err := parseTime(user.Created); err == nil {
			cr.Status.AtProvider.Created = t
		}
	}

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*userv1alpha1.User)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New(errUserNotFound)
	}

	req := clients.UpdateUserRequest{}

	if cr.Spec.ForProvider.Name != nil {
		req.Name = cr.Spec.ForProvider.Name
	}
	if cr.Spec.ForProvider.IsAdministrator != nil {
		req.IsAdministrator = cr.Spec.ForProvider.IsAdministrator
	}
	if cr.Spec.ForProvider.EmailConfirmed != nil {
		req.EmailConfirmed = cr.Spec.ForProvider.EmailConfirmed
	}
	if cr.Spec.ForProvider.Approved != nil {
		req.Approved = cr.Spec.ForProvider.Approved
	}
	if cr.Spec.ForProvider.Disabled != nil {
		req.Disabled = cr.Spec.ForProvider.Disabled
	}

	_, err := c.client.UpdateUser(ctx, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateUser)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*userv1alpha1.User)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil
	}

	cr.Status.SetConditions(xpv2.Deleting())

	err := c.client.DeleteUser(ctx, cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteUser)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

func (c *external) isUpToDate(cr *userv1alpha1.User, user *clients.User) bool {
	if cr.Spec.ForProvider.Email != user.Email {
		return false
	}
	if cr.Spec.ForProvider.Name != nil && *cr.Spec.ForProvider.Name != user.Name {
		return false
	}
	if cr.Spec.ForProvider.IsAdministrator != nil && *cr.Spec.ForProvider.IsAdministrator != user.IsAdministrator {
		return false
	}
	if cr.Spec.ForProvider.EmailConfirmed != nil && *cr.Spec.ForProvider.EmailConfirmed != user.EmailConfirmed {
		return false
	}
	if cr.Spec.ForProvider.Approved != nil && *cr.Spec.ForProvider.Approved != user.Approved {
		return false
	}
	if cr.Spec.ForProvider.Disabled != nil && *cr.Spec.ForProvider.Disabled != user.Disabled {
		return false
	}
	return true
}

func parseTime(s string) (*metav1.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	mt := metav1.NewTime(t)
	return &mt, nil
}
