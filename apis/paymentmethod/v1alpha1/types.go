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

// OnChainConfig configures on-chain payment method settings.
type OnChainConfig struct {
	// DerivationScheme is the derivation scheme for the wallet.
	// +kubebuilder:validation:Required
	DerivationScheme string `json:"derivationScheme"`

	// Label is a human-readable label for the payment method.
	// +optional
	Label *string `json:"label,omitempty"`

	// AccountKeyPath is the account key path for HD wallets.
	// +optional
	AccountKeyPath *string `json:"accountKeyPath,omitempty"`
}

// LightningConfig configures Lightning Network payment method settings.
type LightningConfig struct {
	// ConnectionString is the Lightning Network connection string.
	// +kubebuilder:validation:Required
	ConnectionString string `json:"connectionString"`

	// DisableBoltcard disables Boltcard support.
	// +optional
	DisableBoltcard *bool `json:"disableBoltcard,omitempty"`

	// InternalNodeRef references an internal Lightning node.
	// +optional
	InternalNodeRef *string `json:"internalNodeRef,omitempty"`
}

// PaymentMethodParameters are the configurable fields of a PaymentMethod.
type PaymentMethodParameters struct {
	// StoreRef is a reference to the Store resource.
	// +kubebuilder:validation:Required
	StoreRef StoreReference `json:"storeRef"`

	// CryptoCode is the cryptocurrency code (BTC, LTC, etc.).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=BTC;LTC;BCH;BTG;DASH;DOGE;GRS;MONA;ETH;BNB;USDT;USDC;DAI
	CryptoCode string `json:"cryptoCode"`

	// PaymentType determines if this is onchain or lightning.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=onchain;lightning
	PaymentType string `json:"paymentType"`

	// Enabled determines if the payment method is active.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// OnChain configuration for on-chain payment methods.
	// +optional
	OnChain *OnChainConfig `json:"onChain,omitempty"`

	// Lightning configuration for Lightning Network payment methods.
	// +optional
	Lightning *LightningConfig `json:"lightning,omitempty"`
}

// PaymentMethodObservation are the observable fields of a PaymentMethod.
type PaymentMethodObservation struct {
	// PaymentMethodID is the unique identifier of the payment method.
	PaymentMethodID string `json:"paymentMethodId,omitempty"`

	// StoreID is the store identifier this payment method belongs to.
	StoreID string `json:"storeId,omitempty"`

	// CryptoCode is the cryptocurrency code.
	CryptoCode string `json:"cryptoCode,omitempty"`

	// PaymentType indicates if this is onchain or lightning.
	PaymentType string `json:"paymentType,omitempty"`

	// Enabled indicates if the payment method is currently active.
	Enabled bool `json:"enabled,omitempty"`

	// WalletBalance is the current wallet balance (for on-chain).
	WalletBalance *string `json:"walletBalance,omitempty"`

	// DerivationScheme is the current derivation scheme (for on-chain).
	DerivationScheme string `json:"derivationScheme,omitempty"`

	// ConnectionInfo contains Lightning Network connection details.
	ConnectionInfo string `json:"connectionInfo,omitempty"`

	// CreatedAt is the payment method creation timestamp.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// LastSyncAt is the timestamp of the last synchronization.
	LastSyncAt *metav1.Time `json:"lastSyncAt,omitempty"`

	// SyncStatus indicates the synchronization status.
	SyncStatus string `json:"syncStatus,omitempty"`

	// ErrorMessage contains any error from the last sync.
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// A PaymentMethodSpec defines the desired state of a PaymentMethod.
type PaymentMethodSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PaymentMethodParameters `json:"forProvider"`
}

// A PaymentMethodStatus represents the observed state of a PaymentMethod.
type PaymentMethodStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PaymentMethodObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,btcpay}

// A PaymentMethod is a managed resource that represents a BTCPay Server payment method.
type PaymentMethod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaymentMethodSpec   `json:"spec"`
	Status PaymentMethodStatus `json:"status,omitempty"`
}

// GetCondition of this PaymentMethod.
func (pm *PaymentMethod) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return pm.Status.GetCondition(ct)
}

// GetManagementPolicies of this PaymentMethod.
func (pm *PaymentMethod) GetManagementPolicies() xpv1.ManagementPolicies {
	return pm.Spec.ManagementPolicies
}

// SetManagementPolicies of this PaymentMethod.
func (pm *PaymentMethod) SetManagementPolicies(m xpv1.ManagementPolicies) {
	pm.Spec.ManagementPolicies = m
}

// SetConditions of this PaymentMethod.
func (pm *PaymentMethod) SetConditions(c ...xpv1.Condition) {
	pm.Status.SetConditions(c...)
}

// +kubebuilder:object:root=true

// PaymentMethodList contains a list of PaymentMethod.
type PaymentMethodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PaymentMethod `json:"items"`
}
