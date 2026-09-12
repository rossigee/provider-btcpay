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

// ApiKeyParameters are the configurable fields of an ApiKey.
type ApiKeyParameters struct {
	// Label is a human-readable label for the API key.
	// +kubebuilder:validation:Required
	Label string `json:"label"`

	// Permissions is a list of permissions granted to the API key.
	// +kubebuilder:validation:Required
	Permissions []string `json:"permissions"`
}

// ApiKeyObservation are the observable fields of an ApiKey.
type ApiKeyObservation struct {
	// ID is the unique identifier of the API key.
	ID string `json:"id,omitempty"`

	// ApiKey is the actual API key value (only returned on creation).
	// +optional
	ApiKey string `json:"apiKey,omitempty"`

	// Label is the label of the API key.
	Label string `json:"label,omitempty"`

	// Permissions is the list of permissions.
	Permissions []string `json:"permissions,omitempty"`
}

// A ApiKeySpec defines the desired state of a ApiKey.
type ApiKeySpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              ApiKeyParameters `json:"forProvider"`
}

// A ApiKeyStatus represents the observed state of a ApiKey.
type ApiKeyStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 ApiKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
// +kubebuilder:printcolumn:name="LABEL",type="string",JSONPath=".spec.forProvider.label"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type ApiKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiKeySpec   `json:"spec"`
	Status ApiKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ApiKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiKey `json:"items"`
}
