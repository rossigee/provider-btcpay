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

func (in *Store) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return in.Status.GetCondition(ct)
}

func (in *Store) SetConditions(c ...xpv1.Condition) {
	in.Status.SetConditions(c...)
}

func (in *Store) GetProviderConfigReference() *xpv1.ProviderConfigReference {
	return in.Spec.ProviderConfigReference
}

func (in *Store) SetProviderConfigReference(ref *xpv1.ProviderConfigReference) {
	in.Spec.ProviderConfigReference = ref
}

func (in *Store) GetWriteConnectionSecretToReference() *xpv1.LocalSecretReference {
	return in.Spec.WriteConnectionSecretToReference
}

func (in *Store) SetWriteConnectionSecretToReference(ref *xpv1.LocalSecretReference) {
	in.Spec.WriteConnectionSecretToReference = ref
}

func (in *Store) GetManagementPolicies() xpv1.ManagementPolicies {
	return in.Spec.ManagementPolicies
}

func (in *Store) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	in.Spec.ManagementPolicies = mp
}

// StoreParameters are the configurable fields of a Store.
type StoreParameters struct {
	// Name is the display name of the store in BTCPay Server.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// DefaultCurrency is the default currency for the store.
	// +kubebuilder:validation:Required
	DefaultCurrency string `json:"defaultCurrency"`

	// Website is the website URL associated with the store.
	// +optional
	Website *string `json:"website,omitempty"`

	// InvoiceExpiration is the default invoice expiration time in seconds.
	// +optional
	InvoiceExpiration *int32 `json:"invoiceExpiration,omitempty"`

	// PaymentTolerance is the payment tolerance percentage for underpayments.
	// +optional
	PaymentTolerance *float64 `json:"paymentTolerance,omitempty"`

	// NetworkFeeMode determines how network fees are handled.
	// +optional
	NetworkFeeMode *string `json:"networkFeeMode,omitempty"`

	// SpeedPolicy determines the confirmation speed policy.
	// +optional
	SpeedPolicy *string `json:"speedPolicy,omitempty"`

	// LightningAmountInSatoshi enables Lightning amounts to be displayed in Satoshi.
	// +optional
	LightningAmountInSatoshi *bool `json:"lightningAmountInSatoshi,omitempty"`

	// LightningPrivateRouteHints enables private route hints for Lightning invoices.
	// +optional
	LightningPrivateRouteHints *bool `json:"lightningPrivateRouteHints,omitempty"`

	// OnChainWithLnInvoiceFallback enables on-chain fallback for Lightning invoices.
	// +optional
	OnChainWithLnInvoiceFallback *bool `json:"onChainWithLnInvoiceFallback,omitempty"`

	// RedirectAutomatically enables automatic redirect after payment.
	// +optional
	RedirectAutomatically *bool `json:"redirectAutomatically,omitempty"`

	// ShowRecommendedFee shows recommended fee on checkout.
	// +optional
	ShowRecommendedFee *bool `json:"showRecommendedFee,omitempty"`

	// RecommendedFeeBlockTarget is the block target for recommended fee calculation.
	// +optional
	RecommendedFeeBlockTarget *int32 `json:"recommendedFeeBlockTarget,omitempty"`

	// DefaultLang is the default language for the store.
	// +optional
	DefaultLang *string `json:"defaultLang,omitempty"`

	// MonitoringExpiration is the time in seconds after which invoices expire for monitoring.
	// +optional
	MonitoringExpiration *int32 `json:"monitoringExpiration,omitempty"`

	// CheckoutType determines the checkout experience.
	// +optional
	CheckoutType *string `json:"checkoutType,omitempty"`

	// Receipt configures receipt settings.
	// +optional
	Receipt *ReceiptSettings `json:"receipt,omitempty"`

	// Branding configures store branding.
	// +optional
	Branding *BrandingSettings `json:"branding,omitempty"`
}

// ReceiptSettings configures receipt generation and delivery.
type ReceiptSettings struct {
	// Enabled determines if receipts are generated.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// ShowQR shows QR code on receipts.
	// +optional
	ShowQR *bool `json:"showQR,omitempty"`
	// ShowPayments shows payment details on receipts.
	// +optional
	ShowPayments *bool `json:"showPayments,omitempty"`
}

// BrandingSettings configures store appearance and branding.
type BrandingSettings struct {
	// LogoURL is the URL of the store logo.
	// +optional
	LogoURL *string `json:"logoUrl,omitempty"`
	// CSS is custom CSS for the store.
	// +optional
	CSS *string `json:"css,omitempty"`
	// HtmlTitle is the HTML title for store pages.
	// +optional
	HtmlTitle *string `json:"htmlTitle,omitempty"`
}

// StoreObservation are the observable fields of a Store.
type StoreObservation struct {
	// ID is the unique identifier of the store in BTCPay Server.
	ID string `json:"id,omitempty"`
	// Name is the current name of the store.
	Name string `json:"name,omitempty"`
	// Website is the website URL associated with the store.
	Website string `json:"website,omitempty"`
	// DefaultCurrency is the default currency for the store.
	DefaultCurrency string `json:"defaultCurrency,omitempty"`
	// InvoiceExpiration is the current invoice expiration time.
	InvoiceExpiration int32 `json:"invoiceExpiration,omitempty"`
	// PaymentMethods are the enabled payment methods for the store.
	PaymentMethods []string `json:"paymentMethods,omitempty"`
	// CreatedAt is the timestamp when the store was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
	// DerivationSchemes are the configured derivation schemes.
	DerivationSchemes map[string]string `json:"derivationSchemes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
// +genclient
// +genclient:namespaced
// +groupName=store.btcpay.m.crossplane.io

// A Store is a managed resource that represents a BTCPay Server store.
type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StoreSpec   `json:"spec"`
	Status            StoreStatus `json:"status,omitempty"`
}

type StoreSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              StoreParameters `json:"forProvider"`
}

type StoreStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             StoreObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// StoreList contains a list of Store.
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Store `json:"items"`
}
