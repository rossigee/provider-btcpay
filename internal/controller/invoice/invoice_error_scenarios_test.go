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

package invoice

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
)

func TestObserveErrorScenarios(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args args
		want want
		mock func() *MockBTCPayClient
		kube func() client.Client
	}{
		"StoreNotFoundInKubernetes": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "nonexistent-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("stores.store.btcpay.crossplane.io \"nonexistent-store\" not found"), "cannot get referenced store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("stores.store.btcpay.crossplane.io \"nonexistent-store\" not found")
					},
				}
			},
		},
		"StoreNotReady": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "unready-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.New("referenced store is not ready"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							// Store exists but has no ID (not ready)
							store.Status.AtProvider.ID = ""
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InvoiceExpired": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "expired-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 410: Invoice has expired"), "cannot get invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 410: Invoice has expired")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"NetworkTimeoutDuringObserve": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "timeout-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("context deadline exceeded"), "cannot get invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return nil, errors.New("context deadline exceeded")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"UnauthorizedAccessToInvoice": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unauthorized-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 403: Access denied to invoice"), "cannot get invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 403: Access denied to invoice")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				client: tc.mock(),
				kube:   tc.kube(),
			}

			got, err := e.Observe(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.o, got); diff != "" {
					t.Errorf("Observe(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestCreateErrorScenarios(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		c   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		args args
		want want
		mock func() *MockBTCPayClient
		kube func() client.Client
	}{
		"InvalidStoreForInvoice": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "invalid-store-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 404: Store not found"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 404: Store not found")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InvalidAmountValidation": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "invalid-amount-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   -50.00, // Negative amount
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Invoice amount must be positive"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 400: Invoice amount must be positive")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"UnsupportedCurrency": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unsupported-currency-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "FAKE", // Unsupported currency
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Unsupported currency"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 400: Unsupported currency")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"StoreInMaintenanceMode": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "maintenance-store-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 503: Store is in maintenance mode"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 503: Store is in maintenance mode")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"RateLimitExceededOnCreate": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rate-limited-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "test-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 429: Rate limit exceeded for invoice creation"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 429: Rate limit exceeded for invoice creation")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				client: tc.mock(),
				kube:   tc.kube(),
			}

			got, err := e.Create(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.c, got); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteErrorScenarios(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
		mock func() *MockBTCPayClient
		kube func() client.Client
	}{
		"InvoiceAlreadyPaid": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "paid-invoice",
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 409: Cannot archive paid invoice"), "cannot delete invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
						return errors.New("API request failed with status 409: Cannot archive paid invoice")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InvoiceBeingProcessed": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "processing-invoice",
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 409: Invoice is currently being processed"), "cannot delete invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
						return errors.New("API request failed with status 409: Invoice is currently being processed")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InsufficientPermissionsToArchive": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unauthorized-archive",
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 403: Insufficient permissions to archive invoice"), "cannot delete invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
						return errors.New("API request failed with status 403: Insufficient permissions to archive invoice")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"ServerErrorDuringArchive": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "server-error-archive",
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice123",
							StoreID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 500: Internal server error during archive"), "cannot delete invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
						return errors.New("API request failed with status 500: Internal server error during archive")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				client: tc.mock(),
				kube:   tc.kube(),
			}

			_, err := e.Delete(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
