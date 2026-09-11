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

type SharedLinkParameters struct {
	StoreRef    StoreReference `json:"storeRef"`
	Amount      *float64       `json:"amount,omitempty"`
	Currency    *string        `json:"currency,omitempty"`
	Description *string        `json:"description,omitempty"`
}

type StoreReference struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type SharedLinkObservation struct {
	ID          string  `json:"id,omitempty"`
	StoreID     string  `json:"storeId,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type SharedLinkSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              SharedLinkParameters `json:"forProvider"`
}

type SharedLinkStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 SharedLinkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
type SharedLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SharedLinkSpec   `json:"spec"`
	Status            SharedLinkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type SharedLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SharedLink `json:"items"`
}
