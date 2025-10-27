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

// StoreReference represents a reference to a Store resource.
type StoreReference struct {
	// Name of the referenced Store.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the referenced Store.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// WebhookParameters are the configurable fields of a Webhook.
type WebhookParameters struct {
	// StoreRef is a reference to the Store resource.
	// +kubebuilder:validation:Required
	StoreRef StoreReference `json:"storeRef"`

	// URL is the webhook endpoint URL.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=uri
	URL string `json:"url"`

	// Enabled determines if the webhook is active.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// AutomaticRedelivery enables automatic redelivery on failed attempts.
	// +optional
	AutomaticRedelivery *bool `json:"automaticRedelivery,omitempty"`

	// Secret is the webhook secret for signature validation.
	// +optional
	Secret *string `json:"secret,omitempty"`

	// AuthorizedEvents are the events that trigger this webhook.
	// +kubebuilder:validation:Enum=InvoiceCreated;InvoiceReceivedPayment;InvoicePaymentSettled;InvoiceExpired;InvoiceInvalid;InvoiceProcessing
	// +optional
	AuthorizedEvents []string `json:"authorizedEvents,omitempty"`

	// PaymentMethod filters events by payment method.
	// +optional
	PaymentMethod *string `json:"paymentMethod,omitempty"`
}

// WebhookObservation are the observable fields of a Webhook.
type WebhookObservation struct {
	// ID is the unique identifier of the webhook in BTCPay Server.
	ID string `json:"id,omitempty"`

	// StoreID is the store identifier this webhook belongs to.
	StoreID string `json:"storeId,omitempty"`

	// URL is the current webhook endpoint URL.
	URL string `json:"url,omitempty"`

	// Enabled indicates if the webhook is currently active.
	Enabled bool `json:"enabled,omitempty"`

	// AutomaticRedelivery indicates if automatic redelivery is enabled.
	AutomaticRedelivery bool `json:"automaticRedelivery,omitempty"`

	// AuthorizedEvents are the events that trigger this webhook.
	AuthorizedEvents []string `json:"authorizedEvents,omitempty"`

	// PaymentMethod indicates the payment method filter.
	PaymentMethod string `json:"paymentMethod,omitempty"`

	// CreatedAt is the webhook creation timestamp.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// LastDeliveryAt is the timestamp of the last webhook delivery attempt.
	LastDeliveryAt *metav1.Time `json:"lastDeliveryAt,omitempty"`

	// LastDeliveryErrorMessage contains the error message from the last failed delivery.
	LastDeliveryErrorMessage string `json:"lastDeliveryErrorMessage,omitempty"`

	// DeliveryCount is the total number of delivery attempts.
	DeliveryCount int32 `json:"deliveryCount,omitempty"`

	// SuccessfulDeliveryCount is the number of successful deliveries.
	SuccessfulDeliveryCount int32 `json:"successfulDeliveryCount,omitempty"`
}

// A WebhookSpec defines the desired state of a Webhook.
type WebhookSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       WebhookParameters `json:"forProvider"`
}

// A WebhookStatus represents the observed state of a Webhook.
type WebhookStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          WebhookObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,btcpay}

// A Webhook is a managed resource that represents a BTCPay Server webhook.
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec"`
	Status WebhookStatus `json:"status,omitempty"`
}

// GetCondition of this Webhook.
func (w *Webhook) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return w.Status.GetCondition(ct)
}

// GetManagementPolicies of this Webhook.
func (w *Webhook) GetManagementPolicies() xpv1.ManagementPolicies {
	return w.Spec.ManagementPolicies
}

// SetManagementPolicies of this Webhook.
func (w *Webhook) SetManagementPolicies(m xpv1.ManagementPolicies) {
	w.Spec.ManagementPolicies = m
}

// SetConditions of this Webhook.
func (w *Webhook) SetConditions(c ...xpv1.Condition) {
	w.Status.SetConditions(c...)
}

// +kubebuilder:object:root=true

// WebhookList contains a list of Webhook.
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}
