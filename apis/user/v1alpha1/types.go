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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// UserParameters are the configurable fields of a User.
type UserParameters struct {
	// Email is the user's email address.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=email
	Email string `json:"email"`

	// Password is the user's password.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=8
	Password string `json:"password"`

	// IsAdministrator determines if the user has administrator privileges.
	// +optional
	IsAdministrator *bool `json:"isAdministrator,omitempty"`

	// Name is the user's display name.
	// +optional
	Name *string `json:"name,omitempty"`

	// Roles are the user's roles within BTCPay Server.
	// +optional
	Roles []string `json:"roles,omitempty"`

	// Disabled determines if the user account is disabled.
	// +optional
	Disabled *bool `json:"disabled,omitempty"`

	// RequireEmailConfirmation determines if email confirmation is required.
	// +optional
	RequireEmailConfirmation *bool `json:"requireEmailConfirmation,omitempty"`
}

// UserObservation are the observable fields of a User.
type UserObservation struct {
	// ID is the unique identifier of the user in BTCPay Server.
	ID string `json:"id,omitempty"`

	// Email is the current email address of the user.
	Email string `json:"email,omitempty"`

	// Name is the current display name of the user.
	Name string `json:"name,omitempty"`

	// IsAdministrator indicates if the user has administrator privileges.
	IsAdministrator bool `json:"isAdministrator,omitempty"`

	// Roles are the current user roles.
	Roles []string `json:"roles,omitempty"`

	// Disabled indicates if the user account is disabled.
	Disabled bool `json:"disabled,omitempty"`

	// EmailConfirmed indicates if the user's email has been confirmed.
	EmailConfirmed bool `json:"emailConfirmed,omitempty"`

	// RequireEmailConfirmation indicates if email confirmation is required.
	RequireEmailConfirmation bool `json:"requireEmailConfirmation,omitempty"`

	// LockoutEnabled indicates if account lockout is enabled.
	LockoutEnabled bool `json:"lockoutEnabled,omitempty"`

	// LockoutEnd is the time when account lockout ends.
	LockoutEnd *metav1.Time `json:"lockoutEnd,omitempty"`

	// CreatedAt is the user creation timestamp.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// LastLogin is the timestamp of the user's last login.
	LastLogin *metav1.Time `json:"lastLogin,omitempty"`
}

// A UserSpec defines the desired state of a User.
type UserSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserParameters `json:"forProvider"`
}

// A UserStatus represents the observed state of a User.
type UserStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,btcpay}

// A User is a managed resource that represents a BTCPay Server user.
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec"`
	Status UserStatus `json:"status,omitempty"`
}

// GetCondition of this User.
func (u *User) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return u.Status.GetCondition(ct)
}

// GetManagementPolicies of this User.
func (u *User) GetManagementPolicies() xpv1.ManagementPolicies {
	return u.Spec.ManagementPolicies
}

// SetManagementPolicies of this User.
func (u *User) SetManagementPolicies(m xpv1.ManagementPolicies) {
	u.Spec.ManagementPolicies = m
}

// SetConditions of this User.
func (u *User) SetConditions(c ...xpv1.Condition) {
	u.Status.SetConditions(c...)
}

// +kubebuilder:object:root=true

// UserList contains a list of User.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}
