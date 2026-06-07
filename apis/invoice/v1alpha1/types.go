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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// InvoiceParameters are the configurable fields of an Invoice.
type InvoiceParameters struct {
	// StoreRef is a reference to the Store resource.
	// +kubebuilder:validation:Required
	StoreRef StoreReference `json:"storeRef"`

	// Amount is the invoice amount in the specified currency.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Amount float64 `json:"amount"`

	// Currency is the currency for the invoice amount.
	// +kubebuilder:validation:Required
	Currency string `json:"currency"`

	// OrderID is an external order identifier.
	// +optional
	OrderID *string `json:"orderId,omitempty"`

	// NotificationURL is the URL to send webhook notifications.
	// +optional
	// +kubebuilder:validation:Pattern="^https://"
	NotificationURL *string `json:"notificationURL,omitempty"`

	// RedirectURL is the URL to redirect after payment completion.
	// +optional
	// +kubebuilder:validation:Pattern="^https://"
	RedirectURL *string `json:"redirectURL,omitempty"`

	// NotificationEmail is the email to send notifications.
	// +optional
	// +kubebuilder:validation:Format=email
	NotificationEmail *string `json:"notificationEmail,omitempty"`

	// ItemDesc is a description of the item being paid for.
	// +optional
	// +kubebuilder:validation:MaxLength=255
	ItemDesc *string `json:"itemDesc,omitempty"`

	// ItemCode is an item code or SKU.
	// +optional
	// +kubebuilder:validation:MaxLength=255
	ItemCode *string `json:"itemCode,omitempty"`

	// Physical indicates if this is for a physical product.
	// +optional
	Physical *bool `json:"physical,omitempty"`

	// TaxIncluded indicates if tax is included in the amount.
	// +optional
	TaxIncluded *bool `json:"taxIncluded,omitempty"`

	// BuyerName is the name of the buyer.
	// +optional
	BuyerName *string `json:"buyerName,omitempty"`

	// BuyerEmail is the email of the buyer.
	// +optional
	BuyerEmail *string `json:"buyerEmail,omitempty"`

	// BuyerCountry is the country of the buyer.
	// +optional
	BuyerCountry *string `json:"buyerCountry,omitempty"`

	// BuyerZip is the ZIP code of the buyer.
	// +optional
	BuyerZip *string `json:"buyerZip,omitempty"`

	// BuyerState is the state of the buyer.
	// +optional
	BuyerState *string `json:"buyerState,omitempty"`

	// BuyerCity is the city of the buyer.
	// +optional
	BuyerCity *string `json:"buyerCity,omitempty"`

	// BuyerAddress1 is the primary address of the buyer.
	// +optional
	BuyerAddress1 *string `json:"buyerAddress1,omitempty"`

	// BuyerAddress2 is the secondary address of the buyer.
	// +optional
	BuyerAddress2 *string `json:"buyerAddress2,omitempty"`

	// BuyerPhone is the phone number of the buyer.
	// +optional
	BuyerPhone *string `json:"buyerPhone,omitempty"`

	// ExtendedNotifications enables extended webhook notifications.
	// +optional
	ExtendedNotifications *bool `json:"extendedNotifications,omitempty"`

	// FullNotifications enables full webhook notifications.
	// +optional
	FullNotifications *bool `json:"fullNotifications,omitempty"`

	// Metadata is additional key-value pairs for the invoice.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`

	// CheckoutQueryString is additional query parameters for the checkout page.
	// +optional
	CheckoutQueryString *string `json:"checkoutQueryString,omitempty"`

	// Receipt configures receipt settings for this invoice.
	// +optional
	Receipt *InvoiceReceiptSettings `json:"receipt,omitempty"`
}

// StoreReference references a Store resource.
type StoreReference struct {
	// Name of the Store resource.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the Store resource. If not specified, defaults to the Invoice's namespace.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// InvoiceReceiptSettings configures receipt generation for an invoice.
type InvoiceReceiptSettings struct {
	// Enabled determines if a receipt should be generated.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// ShowQR shows QR code on the receipt.
	// +optional
	ShowQR *bool `json:"showQR,omitempty"`

	// ShowPayments shows payment details on the receipt.
	// +optional
	ShowPayments *bool `json:"showPayments,omitempty"`
}

// InvoiceObservation are the observable fields of an Invoice.
type InvoiceObservation struct {
	// ID is the unique identifier of the invoice in BTCPay Server.
	ID string `json:"id,omitempty"`

	// StoreID is the ID of the store this invoice belongs to.
	StoreID string `json:"storeId,omitempty"`

	// Amount is the invoice amount.
	Amount float64 `json:"amount,omitempty"`

	// Currency is the invoice currency.
	Currency string `json:"currency,omitempty"`

	// Type is the invoice type.
	Type string `json:"type,omitempty"`

	// CheckoutLink is the URL for the checkout page.
	CheckoutLink string `json:"checkoutLink,omitempty"`

	// Status is the current status of the invoice.
	Status string `json:"status,omitempty"`

	// AdditionalStatus provides additional status information.
	AdditionalStatus string `json:"additionalStatus,omitempty"`

	// MonitoringExpiration is when monitoring expires.
	MonitoringExpiration *metav1.Time `json:"monitoringExpiration,omitempty"`

	// ExpirationTime is when the invoice expires.
	ExpirationTime *metav1.Time `json:"expirationTime,omitempty"`

	// CreatedTime is when the invoice was created.
	CreatedTime *metav1.Time `json:"createdTime,omitempty"`

	// AvailableStatusesForManualMarking lists statuses that can be manually set.
	AvailableStatusesForManualMarking []string `json:"availableStatusesForManualMarking,omitempty"`

	// Archived indicates if the invoice is archived.
	Archived bool `json:"archived,omitempty"`

	// PaymentMethods are the available payment methods for this invoice.
	PaymentMethods []PaymentMethodDetails `json:"paymentMethods,omitempty"`
}

// PaymentMethodDetails provides details about a payment method.
type PaymentMethodDetails struct {
	// PaymentMethod is the payment method identifier (e.g., "BTC", "BTC-LightningNetwork").
	PaymentMethod string `json:"paymentMethod,omitempty"`

	// CryptoCode is the cryptocurrency code (e.g., "BTC").
	CryptoCode string `json:"cryptoCode,omitempty"`

	// Destination is the payment destination (address or Lightning invoice).
	Destination string `json:"destination,omitempty"`

	// PaymentLink is a link to initiate payment.
	PaymentLink string `json:"paymentLink,omitempty"`

	// Rate is the exchange rate used.
	Rate float64 `json:"rate,omitempty"`

	// PaymentMethodPaid is the amount paid in this payment method.
	PaymentMethodPaid float64 `json:"paymentMethodPaid,omitempty"`

	// TotalPaid is the total amount paid.
	TotalPaid float64 `json:"totalPaid,omitempty"`

	// Due is the amount still due.
	Due float64 `json:"due,omitempty"`

	// Amount is the amount for this payment method.
	Amount float64 `json:"amount,omitempty"`

	// NetworkFee is the network fee for this payment method.
	NetworkFee float64 `json:"networkFee,omitempty"`

	// Payments are the individual payments made.
	Payments []PaymentDetails `json:"payments,omitempty"`
}

// PaymentDetails provides details about an individual payment.
type PaymentDetails struct {
	// ID is the payment identifier.
	ID string `json:"id,omitempty"`

	// ReceivedDate is when the payment was received.
	ReceivedDate *metav1.Time `json:"receivedDate,omitempty"`

	// Value is the payment value.
	Value float64 `json:"value,omitempty"`

	// Fee is the payment fee.
	Fee float64 `json:"fee,omitempty"`

	// Status is the payment status.
	Status string `json:"status,omitempty"`

	// Destination is the payment destination.
	Destination string `json:"destination,omitempty"`
}

// An InvoiceSpec defines the desired state of an Invoice.
type InvoiceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       InvoiceParameters `json:"forProvider"`
}

// An InvoiceStatus represents the observed state of an Invoice.
type InvoiceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          InvoiceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Invoice is a managed resource that represents a BTCPay Server invoice.
// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:name="INVOICE-ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="AMOUNT",type="string",JSONPath=".spec.forProvider.amount"
// +kubebuilder:printcolumn:name="CURRENCY",type="string",JSONPath=".spec.forProvider.currency"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,btcpay}
type Invoice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InvoiceSpec   `json:"spec"`
	Status InvoiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InvoiceList contains a list of Invoice
type InvoiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Invoice `json:"items"`
}

// ReceiptSettings contains receipt configuration for an invoice.
type ReceiptSettings struct {
	// Enabled indicates whether receipts are enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// ShowQR indicates whether to show QR code on receipts.
	// +optional
	ShowQR *bool `json:"showQR,omitempty"`

	// ShowPayments indicates whether to show payments on receipts.
	// +optional
	ShowPayments *bool `json:"showPayments,omitempty"`
}

// CheckoutSettings contains checkout configuration for an invoice.
type CheckoutSettings struct {
	// SpeedPolicy defines the speed policy for the checkout.
	// +optional
	SpeedPolicy *string `json:"speedPolicy,omitempty"`

	// PaymentMethods are the enabled payment methods.
	// +optional
	PaymentMethods []string `json:"paymentMethods,omitempty"`

	// DefaultPaymentMethod is the default payment method.
	// +optional
	DefaultPaymentMethod *string `json:"defaultPaymentMethod,omitempty"`

	// ExpirationMinutes is the expiration time in minutes.
	// +optional
	ExpirationMinutes *int32 `json:"expirationMinutes,omitempty"`

	// MonitoringExpiration is the monitoring expiration time.
	// +optional
	MonitoringExpiration *int32 `json:"monitoringExpiration,omitempty"`

	// PaymentTolerance is the payment tolerance percentage.
	// +optional
	PaymentTolerance *float64 `json:"paymentTolerance,omitempty"`

	// RedirectAutomatically indicates whether to redirect automatically.
	// +optional
	RedirectAutomatically *bool `json:"redirectAutomatically,omitempty"`

	// RequiresRefundEmail indicates whether refund email is required.
	// +optional
	RequiresRefundEmail *bool `json:"requiresRefundEmail,omitempty"`

	// CheckoutType defines the checkout type.
	// +optional
	CheckoutType *string `json:"checkoutType,omitempty"`
}
