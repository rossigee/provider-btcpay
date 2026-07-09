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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`

	// BaseURL is the base URL of the BTCPay Server API instance.
	// This should be the full URL to your BTCPay Server instance
	// (e.g., https://btcpay.example.com)
	// +optional
	BaseURL *string `json:"baseURL,omitempty"`
}

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=Secret
	Source xpv1.CredentialsSource `json:"source"`

	xpv1.CommonCredentialSelectors `json:",inline"`
}

// CommonCredentialSelectors mirrors xpv1 for test usage.
type CommonCredentialSelectors = xpv1.CommonCredentialSelectors

// SecretKeySelector mirrors xpv1 for test usage.
type SecretKeySelector = xpv1.SecretKeySelector

// SecretReference mirrors xpv1 for test usage.
type SecretReference = xpv1.SecretReference

// A ProviderConfigStatus reflects the observed state of a ProviderConfig.
type ProviderConfigStatus struct {
	xpv1.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures a BTCPay provider.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,btcpay}
// +kubebuilder:storageversion
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig.
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,btcpay}
// +kubebuilder:storageversion
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ProviderConfigReference xpv1.ProviderConfigReference `json:"providerConfigRef"`
	ResourceReference       xpv1.TypedReference          `json:"resourceRef"`
}

// GetProviderConfigReference returns the provider config reference.
func (p *ProviderConfigUsage) GetProviderConfigReference() xpv1.ProviderConfigReference {
	return p.ProviderConfigReference
}

// SetProviderConfigReference sets the provider config reference.
func (p *ProviderConfigUsage) SetProviderConfigReference(r xpv1.ProviderConfigReference) {
	p.ProviderConfigReference = r
}

// GetResourceReference returns the resource reference.
func (p *ProviderConfigUsage) GetResourceReference() xpv1.TypedReference {
	return p.ResourceReference
}

// SetResourceReference sets the resource reference.
func (p *ProviderConfigUsage) SetResourceReference(r xpv1.TypedReference) {
	p.ResourceReference = r
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}

// ProviderConfigUsageGroupVersionKind is the GroupVersionKind for ProviderConfigUsage
var ProviderConfigUsageGroupVersionKind = schema.GroupVersionKind{
	Group:   Group,
	Version: Version,
	Kind:    "ProviderConfigUsage",
}

// ProviderConfigUsageListGroupVersionKind is the GroupVersionKind for ProviderConfigUsageList
var ProviderConfigUsageListGroupVersionKind = schema.GroupVersionKind{
	Group:   Group,
	Version: Version,
	Kind:    "ProviderConfigUsageList",
}
