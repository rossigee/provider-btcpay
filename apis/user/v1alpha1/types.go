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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserParameters are the configurable fields of a User.
type UserParameters struct {
	// Email is the user's email address.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=email
	Email string `json:"email"`

	// Name is the user's display name.
	// +optional
	Name *string `json:"name,omitempty"`

	// IsAdministrator indicates if the user is an administrator.
	// +optional
	IsAdministrator *bool `json:"isAdministrator,omitempty"`

	// EmailConfirmed indicates if the email is confirmed.
	// +optional
	EmailConfirmed *bool `json:"emailConfirmed,omitempty"`

	// Approved indicates if the user is approved.
	// +optional
	Approved *bool `json:"approved,omitempty"`

	// Disabled indicates if the user is disabled.
	// +optional
	Disabled *bool `json:"disabled,omitempty"`
}

// UserObservation are the observable fields of a User.
type UserObservation struct {
	// ID is the unique identifier of the user in BTCPay Server.
	ID string `json:"id,omitempty"`

	// Email is the user's email address.
	Email string `json:"email,omitempty"`

	// Name is the user's display name.
	Name string `json:"name,omitempty"`

	// IsAdministrator indicates if the user is an administrator.
	IsAdministrator bool `json:"isAdministrator,omitempty"`

	// EmailConfirmed indicates if the email is confirmed.
	EmailConfirmed bool `json:"emailConfirmed,omitempty"`

	// Approved indicates if the user is approved.
	Approved bool `json:"approved,omitempty"`

	// Disabled indicates if the user is disabled.
	Disabled bool `json:"disabled,omitempty"`

	// Created is when the user was created.
	Created *metav1.Time `json:"created,omitempty"`
}

// A UserSpec defines the desired state of a User.
type UserSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              UserParameters `json:"forProvider"`
}

// A UserStatus represents the observed state of a User.
type UserStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".spec.forProvider.email"
// +kubebuilder:printcolumn:name="ADMIN",type="boolean",JSONPath=".spec.forProvider.isAdministrator"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}
