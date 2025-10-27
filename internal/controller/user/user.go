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

	"github.com/rossigee/provider-btcpay/apis/user/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
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
	name := managed.ControllerName(v1alpha1.UserKind)

	// TODO: Fix v2 connection publishers - temporarily using basic setup
	_ = []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.UserGroupVersionKind),
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
		For(&v1alpha1.User{}).
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
	_, ok := mg.(*v1alpha1.User)
	if !ok {
		return nil, errors.New(errNotUser)
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

	return &external{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.BTCPayClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr := mg.(*v1alpha1.User)

	// If we don't have an ID yet, the resource doesn't exist
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	user, err := c.client.GetUser(cr.Status.AtProvider.ID)
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

	// Update status with observed state
	c.updateStatus(cr, user)

	// Check if the resource is up to date
	upToDate := c.isUpToDate(cr, user)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		ConnectionDetails: managed.ConnectionDetails{
			"userId":          []byte(user.ID),
			"email":           []byte(user.Email),
			"name":            []byte(user.Name),
			"isAdministrator": []byte(boolToString(user.IsAdministrator)),
		},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr := mg.(*v1alpha1.User)

	cr.Status.SetConditions(xpv1.Creating())

	req := clients.CreateUserRequest{
		Email:    cr.Spec.ForProvider.Email,
		Password: cr.Spec.ForProvider.Password,
	}

	if cr.Spec.ForProvider.IsAdministrator != nil {
		req.IsAdministrator = *cr.Spec.ForProvider.IsAdministrator
	}
	if cr.Spec.ForProvider.Name != nil {
		req.Name = *cr.Spec.ForProvider.Name
	}
	if cr.Spec.ForProvider.Roles != nil {
		req.Roles = cr.Spec.ForProvider.Roles
	}
	if cr.Spec.ForProvider.Disabled != nil {
		req.Disabled = *cr.Spec.ForProvider.Disabled
	}
	if cr.Spec.ForProvider.RequireEmailConfirmation != nil {
		req.RequireEmailConfirmation = *cr.Spec.ForProvider.RequireEmailConfirmation
	}

	user, err := c.client.CreateUser(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateUser)
	}

	cr.Status.AtProvider.ID = user.ID

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"userId":          []byte(user.ID),
			"email":           []byte(user.Email),
			"name":            []byte(user.Name),
			"isAdministrator": []byte(boolToString(user.IsAdministrator)),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr := mg.(*v1alpha1.User)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New(errUserNotFound)
	}

	req := clients.UpdateUserRequest{}

	if cr.Spec.ForProvider.Name != nil {
		req.Name = *cr.Spec.ForProvider.Name
	}
	if cr.Spec.ForProvider.IsAdministrator != nil {
		req.IsAdministrator = *cr.Spec.ForProvider.IsAdministrator
	}
	if cr.Spec.ForProvider.Roles != nil {
		req.Roles = cr.Spec.ForProvider.Roles
	}
	if cr.Spec.ForProvider.Disabled != nil {
		req.Disabled = *cr.Spec.ForProvider.Disabled
	}

	_, err := c.client.UpdateUser(cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateUser)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr := mg.(*v1alpha1.User)

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalDelete{}, nil // Nothing to delete
	}

	cr.Status.SetConditions(xpv1.Deleting())

	err := c.client.DeleteUser(cr.Status.AtProvider.ID)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteUser)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	// Nothing to disconnect for BTCPay API client
	return nil
}

// isUpToDate checks if the User resource is up to date with the desired state
func (c *external) isUpToDate(cr *v1alpha1.User, user *clients.User) bool {
	if cr.Spec.ForProvider.Email != user.Email {
		return false
	}
	if cr.Spec.ForProvider.Name != nil && *cr.Spec.ForProvider.Name != user.Name {
		return false
	}
	if cr.Spec.ForProvider.IsAdministrator != nil && *cr.Spec.ForProvider.IsAdministrator != user.IsAdministrator {
		return false
	}
	if cr.Spec.ForProvider.Disabled != nil && *cr.Spec.ForProvider.Disabled != user.Disabled {
		return false
	}
	return true
}

// updateStatus updates the user status with observed state
func (c *external) updateStatus(cr *v1alpha1.User, user *clients.User) {
	cr.Status.AtProvider.ID = user.ID
	cr.Status.AtProvider.Email = user.Email
	cr.Status.AtProvider.Name = user.Name
	cr.Status.AtProvider.IsAdministrator = user.IsAdministrator
	cr.Status.AtProvider.Roles = user.Roles
	cr.Status.AtProvider.Disabled = user.Disabled
	cr.Status.AtProvider.EmailConfirmed = user.EmailConfirmed
	cr.Status.AtProvider.RequireEmailConfirmation = user.RequireEmailConfirmation
	cr.Status.AtProvider.LockoutEnabled = user.LockoutEnabled

	if user.LockoutEnd != nil {
		cr.Status.AtProvider.LockoutEnd = &metav1.Time{Time: *user.LockoutEnd}
	}
	if user.CreatedAt != nil {
		cr.Status.AtProvider.CreatedAt = &metav1.Time{Time: *user.CreatedAt}
	}
	if user.LastLogin != nil {
		cr.Status.AtProvider.LastLogin = &metav1.Time{Time: *user.LastLogin}
	}
}

// boolToString converts a boolean to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
