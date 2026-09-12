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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebhookParameters are the configurable fields of a Webhook.
type WebhookParameters struct {
	// StoreRef is a reference to the Store resource.
	// +kubebuilder:validation:Required
	StoreRef StoreReference `json:"storeRef"`

	// URL is the webhook endpoint URL.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https://"
	// +kubebuilder:validation:MaxLength=2048
	URL string `json:"url"`

	// Enabled indicates if the webhook is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// AutomaticRedelivery enables automatic redelivery of failed webhooks.
	// +optional
	AutomaticRedelivery *bool `json:"automaticRedelivery,omitempty"`

	// AuthorizedEvents is a list of events that trigger the webhook.
	// If empty, all events are subscribed.
	// +optional
	AuthorizedEvents []string `json:"authorizedEvents,omitempty"`

	// Secret is the webhook secret for signature verification.
	// +optional
	Secret *string `json:"secret,omitempty"`
}

// StoreReference references a Store resource.
type StoreReference struct {
	// Name of the Store resource.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the Store resource. If not specified, defaults to the Webhook's namespace.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// WebhookObservation are the observable fields of a Webhook.
type WebhookObservation struct {
	// ID is the unique identifier of the webhook in BTCPay Server.
	ID string `json:"id,omitempty"`

	// StoreID is the ID of the store this webhook belongs to.
	StoreID string `json:"storeId,omitempty"`

	// URL is the webhook endpoint URL.
	URL string `json:"url,omitempty"`

	// Enabled indicates if the webhook is enabled.
	Enabled bool `json:"enabled,omitempty"`

	// AutomaticRedelivery indicates if automatic redelivery is enabled.
	AutomaticRedelivery bool `json:"automaticRedelivery,omitempty"`

	// AuthorizedEvents is the list of subscribed events.
	AuthorizedEvents []string `json:"authorizedEvents,omitempty"`

	// Secret is the webhook secret.
	Secret string `json:"secret,omitempty"`
}

// A WebhookSpec defines the desired state of a Webhook.
type WebhookSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              WebhookParameters `json:"forProvider"`
}

// A WebhookStatus represents the observed state of a Webhook.
type WebhookStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 WebhookObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Webhook is a managed resource that represents a BTCPay Server webhook.
// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:name="WEBHOOK-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".spec.forProvider.url"
// +kubebuilder:printcolumn:name="ENABLED",type="boolean",JSONPath=".spec.forProvider.enabled"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec"`
	Status WebhookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebhookList contains a list of Webhook
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}
