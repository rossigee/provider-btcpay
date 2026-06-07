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
	"time"

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

func TestObserveComprehensive(t *testing.T) {
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
	}{
		"InvoiceDoesNotExist": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							// No ID set - resource doesn't exist
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
		},
		"InvoiceExistsAndUpToDate": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "store123"},
							Amount:   100.50,
							Currency: "USD",
							OrderID:  testStringPtr("order-123"),
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
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) (*clients.Invoice, error) {
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						if invoiceID != "invoice123" {
							t.Errorf("Expected invoiceID 'invoice123', got %v", invoiceID)
						}
						return &clients.Invoice{
							ID:       "invoice123",
							StoreID:  "store123",
							Amount:   "100.50",
							Currency: "USD",
							OrderID:  "order-123",
							Status:   "New",
						}, nil
					},
				}
			},
		},
		"InvoiceExistsAndUpToDateAlways": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "store123"},
							Amount:   200.00,
							Currency: "EUR",
							OrderID:  testStringPtr("updated-order"),
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
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true, // Invoices are always up-to-date since they're immutable
				},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) (*clients.Invoice, error) {
						return &clients.Invoice{
							ID:       "invoice123",
							StoreID:  "store123",
							Amount:   "100.50", // Different from spec but still up-to-date
							Currency: "USD",    // Different from spec but still up-to-date
							OrderID:  "order-123",
							Status:   "New",
						}, nil
					},
				}
			},
		},
		"GetInvoiceError": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
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
				err: errors.Wrap(errors.New("API request failed with status 500: Internal Server Error"), "cannot get invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 500: Internal Server Error")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create a mock kube client that returns a Store
			kube := &test.MockClient{
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if store, ok := obj.(*storev1alpha1.Store); ok {
						store.Status.AtProvider.ID = "store123"
						return nil
					}
					return errors.New("store not found")
				},
			}

			e := external{
				client: tc.mock(),
				kube:   kube,
			}

			got, err := e.Observe(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreateComprehensive(t *testing.T) {
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
	}{
		"CreateSuccess": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef:          v1alpha1.StoreReference{Name: "store123"},
							Amount:            150.75,
							Currency:          "USD",
							OrderID:           testStringPtr("order-456"),
							NotificationURL:   testStringPtr("https://webhook.example.com"),
							RedirectURL:       testStringPtr("https://redirect.example.com"),
							NotificationEmail: testStringPtr("notify@example.com"),
							ItemDesc:          testStringPtr("Test Item"),
							ItemCode:          testStringPtr("ITEM-001"),
							Physical:          testBoolPtr(true),
							TaxIncluded:       testBoolPtr(false),
							BuyerEmail:        testStringPtr("buyer@example.com"),
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(ctx context.Context, storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						// Verify request parameters
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						if req.Amount != 150.75 {
							t.Errorf("Expected Amount 150.75, got %v", req.Amount)
						}
						if req.Currency != "USD" {
							t.Errorf("Expected Currency 'USD', got %v", req.Currency)
						}
						if req.OrderID != "order-456" {
							t.Errorf("Expected OrderID 'order-456', got %v", req.OrderID)
						}
						if req.NotificationURL != "https://webhook.example.com" {
							t.Errorf("Expected NotificationURL 'https://webhook.example.com', got %v", req.NotificationURL)
						}
						if req.RedirectURL != "https://redirect.example.com" {
							t.Errorf("Expected RedirectURL 'https://redirect.example.com', got %v", req.RedirectURL)
						}
						if req.NotificationEmail != "notify@example.com" {
							t.Errorf("Expected NotificationEmail 'notify@example.com', got %v", req.NotificationEmail)
						}
						if req.ItemDesc != "Test Item" {
							t.Errorf("Expected ItemDesc 'Test Item', got %v", req.ItemDesc)
						}
						if req.ItemCode != "ITEM-001" {
							t.Errorf("Expected ItemCode 'ITEM-001', got %v", req.ItemCode)
						}
						if req.Physical != true {
							t.Errorf("Expected Physical true, got %v", req.Physical)
						}
						if req.TaxIncluded != false {
							t.Errorf("Expected TaxIncluded false, got %v", req.TaxIncluded)
						}
						if req.BuyerEmail != "buyer@example.com" {
							t.Errorf("Expected BuyerEmail 'buyer@example.com', got %v", req.BuyerEmail)
						}

						return &clients.Invoice{
							ID:              "new-invoice-789",
							StoreID:         storeID,
							Amount:          "150.75",
							Currency:        req.Currency,
							OrderID:         req.OrderID,
							NotificationURL: req.NotificationURL,
							RedirectURL:     req.RedirectURL,
							Status:          "New",
							CheckoutLink:    "https://btcpay.example.com/checkout/new-invoice-789",
							CreatedTime:     &time.Time{},
						}, nil
					},
				}
			},
		},
		"CreateWithMinimalFields": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "minimal-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "store123"},
							Amount:   50.00,
							Currency: "BTC",
							// No optional fields
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(ctx context.Context, storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						// Verify optional fields are empty
						if req.OrderID != "" {
							t.Errorf("Expected OrderID to be empty, got %v", req.OrderID)
						}
						if req.NotificationURL != "" {
							t.Errorf("Expected NotificationURL to be empty, got %v", req.NotificationURL)
						}
						if req.RedirectURL != "" {
							t.Errorf("Expected RedirectURL to be empty, got %v", req.RedirectURL)
						}
						if req.ItemDesc != "" {
							t.Errorf("Expected ItemDesc to be empty, got %v", req.ItemDesc)
						}

						return &clients.Invoice{
							ID:           "minimal-invoice-456",
							StoreID:      storeID,
							Amount:       "50.00",
							Currency:     req.Currency,
							Status:       "New",
							CheckoutLink: "https://btcpay.example.com/checkout/minimal-invoice-456",
						}, nil
					},
				}
			},
		},
		"CreateError": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "error-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "store123"},
							Amount:   -10.00, // Invalid amount
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Invalid invoice amount"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(ctx context.Context, storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return nil, errors.New("API request failed with status 400: Invalid invoice amount")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create a mock kube client that returns a Store
			kube := &test.MockClient{
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if store, ok := obj.(*storev1alpha1.Store); ok {
						store.Status.AtProvider.ID = "store123"
						return nil
					}
					return errors.New("store not found")
				},
			}

			e := external{
				client: tc.mock(),
				kube:   kube,
			}

			got, err := e.Create(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.c, got); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateComprehensive(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		u   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		args args
		want want
		mock func() *MockBTCPayClient
	}{
		"UpdateNoOp": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "store123"},
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
				u: managed.ExternalUpdate{}, // Update is a no-op for invoices
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.mock()}

			got, err := e.Update(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.u, got); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDeleteComprehensive(t *testing.T) {
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
	}{
		"DeleteSuccess": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
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
				err: nil,
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) error {
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						if invoiceID != "invoice123" {
							t.Errorf("Expected invoiceID 'invoice123', got %v", invoiceID)
						}
						return nil
					},
				}
			},
		},
		"DeleteNotFoundIgnored": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
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
				err: nil, // 404 errors are ignored in Delete
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) error {
						return errors.New("API request failed with status 404: Invoice not found")
					},
				}
			},
		},
		"DeleteServerError": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-invoice",
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
				err: errors.Wrap(errors.New("API request failed with status 500: Internal Server Error"), "cannot delete invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(ctx context.Context, storeID, invoiceID string) error {
						return errors.New("API request failed with status 500: Internal Server Error")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create a mock kube client that returns a Store
			kube := &test.MockClient{
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					if store, ok := obj.(*storev1alpha1.Store); ok {
						store.Status.AtProvider.ID = "store123"
						return nil
					}
					return errors.New("store not found")
				},
			}

			e := external{
				client: tc.mock(),
				kube:   kube,
			}

			_, err := e.Delete(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

// Helper functions
func testStringPtr(s string) *string {
	return &s
}

func testBoolPtr(b bool) *bool {
	return &b
}
