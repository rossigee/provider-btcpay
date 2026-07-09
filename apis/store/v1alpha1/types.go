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

// StoreParameters are the configurable fields of a Store.
type StoreParameters struct {
	// Name is the display name of the store in BTCPay Server.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	Name string `json:"name"`

	// DefaultCurrency is the default currency for the store.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=USD;EUR;BTC;GBP;JPY;CAD;AUD;CHF;CNY;SEK;NZD;MXN;SGD;HKD;NOK;TRY;RUB;ZAR;BRL;MYR;BGN;CZK;DKK;HRK;HUF;ISK;PLN;RON;THB;ILS;CLP;PHP;AED;COP;EGP;HUF;IDR;INR;KRW;KWD;LKR;MAD;NPR;PKR;QAR;SAR;UAH;UYU;VND;ZMW
	DefaultCurrency string `json:"defaultCurrency"`

	// Website is the website URL associated with the store.
	// +optional
	// +kubebuilder:validation:Pattern="^https://"
	// +kubebuilder:validation:MaxLength=2048
	Website *string `json:"website,omitempty"`

	// InvoiceExpiration is the default invoice expiration time in seconds.
	// Default is 900 seconds (15 minutes).
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=86400
	// +optional
	InvoiceExpiration *int32 `json:"invoiceExpiration,omitempty"`

	// PaymentTolerance is the payment tolerance percentage for underpayments.
	// Default is 0.0 (no tolerance).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +optional
	PaymentTolerance *float64 `json:"paymentTolerance,omitempty"`

	// NetworkFeeMode determines how network fees are handled.
	// +kubebuilder:validation:Enum=Never;Always;MultiplePaymentsOnly
	// +optional
	NetworkFeeMode *string `json:"networkFeeMode,omitempty"`

	// SpeedPolicy determines the confirmation speed policy.
	// +kubebuilder:validation:Enum=High;Medium;Low
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
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1008
	// +optional
	RecommendedFeeBlockTarget *int32 `json:"recommendedFeeBlockTarget,omitempty"`

	// DefaultLang is the default language for the store.
	// +optional
	DefaultLang *string `json:"defaultLang,omitempty"`

	// MonitoringExpiration is the time in seconds after which invoices expire for monitoring.
	// +kubebuilder:validation:Minimum=600
	// +kubebuilder:validation:Maximum=86400
	// +optional
	MonitoringExpiration *int32 `json:"monitoringExpiration,omitempty"`

	// CheckoutType determines the checkout experience.
	// +kubebuilder:validation:Enum=V1;V2
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

// A StoreSpec defines the desired state of a Store.
type StoreSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              StoreParameters `json:"forProvider"`
}

// A StoreStatus represents the observed state of a Store.
type StoreStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 StoreObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Store is a managed resource that represents a BTCPay Server store.
// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:name="STORE-NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="STORE-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="CURRENCY",type="string",JSONPath=".spec.forProvider.defaultCurrency"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreSpec   `json:"spec"`
	Status StoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StoreList contains a list of Store
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Store `json:"items"`
}
