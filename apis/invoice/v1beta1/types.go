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

func (in *Invoice) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return in.Status.GetCondition(ct)
}

func (in *Invoice) SetConditions(c ...xpv1.Condition) {
	in.Status.SetConditions(c...)
}

func (in *Invoice) GetProviderConfigReference() *xpv1.ProviderConfigReference {
	return in.Spec.ProviderConfigReference
}

func (in *Invoice) SetProviderConfigReference(ref *xpv1.ProviderConfigReference) {
	in.Spec.ProviderConfigReference = ref
}

func (in *Invoice) GetWriteConnectionSecretToReference() *xpv1.LocalSecretReference {
	return in.Spec.WriteConnectionSecretToReference
}

func (in *Invoice) SetWriteConnectionSecretToReference(ref *xpv1.LocalSecretReference) {
	in.Spec.WriteConnectionSecretToReference = ref
}

func (in *Invoice) GetManagementPolicies() xpv1.ManagementPolicies {
	return in.Spec.ManagementPolicies
}

func (in *Invoice) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	in.Spec.ManagementPolicies = mp
}

type InvoiceParameters struct {
	StoreRef              StoreReference          `json:"storeRef"`
	Amount                float64                 `json:"amount"`
	Currency              string                  `json:"currency"`
	OrderID               *string                 `json:"orderId,omitempty"`
	NotificationURL       *string                 `json:"notificationURL,omitempty"`
	RedirectURL           *string                 `json:"redirectURL,omitempty"`
	NotificationEmail     *string                 `json:"notificationEmail,omitempty"`
	ItemDesc              *string                 `json:"itemDesc,omitempty"`
	ItemCode              *string                 `json:"itemCode,omitempty"`
	Physical              *bool                   `json:"physical,omitempty"`
	TaxIncluded           *bool                   `json:"taxIncluded,omitempty"`
	BuyerName             *string                 `json:"buyerName,omitempty"`
	BuyerEmail            *string                 `json:"buyerEmail,omitempty"`
	BuyerCountry          *string                 `json:"buyerCountry,omitempty"`
	BuyerZip              *string                 `json:"buyerZip,omitempty"`
	BuyerState            *string                 `json:"buyerState,omitempty"`
	BuyerCity             *string                 `json:"buyerCity,omitempty"`
	BuyerAddress1         *string                 `json:"buyerAddress1,omitempty"`
	BuyerAddress2         *string                 `json:"buyerAddress2,omitempty"`
	BuyerPhone            *string                 `json:"buyerPhone,omitempty"`
	ExtendedNotifications *bool                   `json:"extendedNotifications,omitempty"`
	FullNotifications     *bool                   `json:"fullNotifications,omitempty"`
	Metadata              map[string]string       `json:"metadata,omitempty"`
	CheckoutQueryString   *string                 `json:"checkoutQueryString,omitempty"`
	Receipt               *InvoiceReceiptSettings `json:"receipt,omitempty"`
}

type StoreReference struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type InvoiceReceiptSettings struct {
	Enabled      *bool `json:"enabled,omitempty"`
	ShowQR       *bool `json:"showQR,omitempty"`
	ShowPayments *bool `json:"showPayments,omitempty"`
}

type InvoiceObservation struct {
	ID                                string                 `json:"id,omitempty"`
	StoreID                           string                 `json:"storeId,omitempty"`
	Amount                            float64                `json:"amount,omitempty"`
	Currency                          string                 `json:"currency,omitempty"`
	Type                              string                 `json:"type,omitempty"`
	CheckoutLink                      string                 `json:"checkoutLink,omitempty"`
	Status                            string                 `json:"status,omitempty"`
	AdditionalStatus                  string                 `json:"additionalStatus,omitempty"`
	MonitoringExpiration              *metav1.Time           `json:"monitoringExpiration,omitempty"`
	ExpirationTime                    *metav1.Time           `json:"expirationTime,omitempty"`
	CreatedTime                       *metav1.Time           `json:"createdTime,omitempty"`
	AvailableStatusesForManualMarking []string               `json:"availableStatusesForManualMarking,omitempty"`
	Archived                          bool                   `json:"archived,omitempty"`
	PaymentMethods                    []PaymentMethodDetails `json:"paymentMethods,omitempty"`
}

type PaymentMethodDetails struct {
	PaymentMethod     string           `json:"paymentMethod,omitempty"`
	CryptoCode        string           `json:"cryptoCode,omitempty"`
	Destination       string           `json:"destination,omitempty"`
	PaymentLink       string           `json:"paymentLink,omitempty"`
	Rate              float64          `json:"rate,omitempty"`
	PaymentMethodPaid float64          `json:"paymentMethodPaid,omitempty"`
	TotalPaid         float64          `json:"totalPaid,omitempty"`
	Due               float64          `json:"due,omitempty"`
	Amount            float64          `json:"amount,omitempty"`
	NetworkFee        float64          `json:"networkFee,omitempty"`
	Payments          []PaymentDetails `json:"payments,omitempty"`
}

type PaymentDetails struct {
	ID           string       `json:"id,omitempty"`
	ReceivedDate *metav1.Time `json:"receivedDate,omitempty"`
	Value        float64      `json:"value,omitempty"`
	Fee          float64      `json:"fee,omitempty"`
	Status       string       `json:"status,omitempty"`
	Destination  string       `json:"destination,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,btcpay}
// +genclient
// +genclient:namespaced
// +groupName=store.btcpay.m.crossplane.io

type Invoice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              InvoiceSpec   `json:"spec"`
	Status            InvoiceStatus `json:"status,omitempty"`
}

type InvoiceSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              InvoiceParameters `json:"forProvider"`
}

type InvoiceStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             InvoiceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

type InvoiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Invoice `json:"items"`
}
