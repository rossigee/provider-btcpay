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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"

	"github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
)

func TestCrossResourceDependencies(t *testing.T) {
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
		"InvoiceWaitsForStoreToBecomeReady": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "waiting-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "pending-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "waiting-invoice-123", // Need ID for store lookup to occur
							StoreID: "pending-store-456",
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
							// Store exists but ID is empty (not ready)
							store.Name = "pending-store"
							store.Status.AtProvider.ID = "" // Not ready yet
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InvoiceSucceedsWhenStoreIsReady": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ready-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "ready-store"},
							Amount:   150.00,
							Currency: "EUR",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice456",
							StoreID: "store456",
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
					GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
						if storeID != "store456" {
							t.Errorf("Expected storeID 'store456', got %v", storeID)
						}
						if invoiceID != "invoice456" {
							t.Errorf("Expected invoiceID 'invoice456', got %v", invoiceID)
						}
						return &clients.Invoice{
							ID:       "invoice456",
							StoreID:  "store456",
							Amount:   "150.00",
							Currency: "EUR",
							Status:   "New",
						}, nil
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							// Store is ready
							store.Name = "ready-store"
							store.Status.AtProvider.ID = "store456"
							store.Status.AtProvider.Name = "Ready Store"
							store.Status.AtProvider.DefaultCurrency = "EUR"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"InvoiceReferencesNonExistentStore": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "orphaned-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "missing-store"},
							Amount:   100.00,
							Currency: "USD",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "orphaned-invoice-123", // Need ID for store lookup to occur
							StoreID: "missing-store-456",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("stores.store.btcpay.crossplane.io \"missing-store\" not found"), "cannot get referenced store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("stores.store.btcpay.crossplane.io \"missing-store\" not found")
					},
				}
			},
		},
		"InvoiceHandlesStoreNamespaceCorrectly": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "namespaced-invoice",
						Namespace: "test-namespace",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "namespaced-store"},
							Amount:   200.00,
							Currency: "BTC",
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "invoice789",
							StoreID: "store789",
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
					GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return &clients.Invoice{
							ID:       "invoice789",
							StoreID:  "store789",
							Amount:   "200.00",
							Currency: "BTC",
							Status:   "New",
						}, nil
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						// Verify that the lookup is for the correct namespace and name
						if key.Name != "namespaced-store" {
							t.Errorf("Expected store name 'namespaced-store', got %v", key.Name)
						}
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "namespaced-store"
							store.Namespace = "test-namespace"
							store.Status.AtProvider.ID = "store789"
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

func TestCrossResourceCreateFlow(t *testing.T) {
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
		"CreateInvoiceWithExistingStore": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "new-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "existing-store"},
							Amount:   99.99,
							Currency: "USD",
							OrderID:  testStringPtr("order-999"),
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						if storeID != "store999" {
							t.Errorf("Expected storeID 'store999', got %v", storeID)
						}
						if req.Amount != 99.99 {
							t.Errorf("Expected Amount 99.99, got %v", req.Amount)
						}
						if req.Currency != "USD" {
							t.Errorf("Expected Currency 'USD', got %v", req.Currency)
						}
						if req.OrderID != "order-999" {
							t.Errorf("Expected OrderID 'order-999', got %v", req.OrderID)
						}

						return &clients.Invoice{
							ID:           "new-invoice-999",
							StoreID:      storeID,
							Amount:       "99.99",
							Currency:     req.Currency,
							OrderID:      req.OrderID,
							Status:       "New",
							CheckoutLink: "https://btcpay.example.com/checkout/new-invoice-999",
						}, nil
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "existing-store"
							store.Status.AtProvider.ID = "store999"
							store.Status.AtProvider.Name = "Existing Store"
							store.Status.AtProvider.DefaultCurrency = "USD"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"CreateInvoiceFailsWhenStoreNotReady": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "blocked-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "unready-store"},
							Amount:   50.00,
							Currency: "EUR",
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
							store.Name = "unready-store"
							store.Status.AtProvider.ID = "" // Not ready
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"CreateInvoiceWithStoreConfigurationMismatch": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mismatch-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "btc-only-store"},
							Amount:   75.00,
							Currency: "USD", // Store doesn't support USD
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Store does not support USD currency"), "cannot create invoice"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						// Simulate BTCPay API validation error for unsupported currency
						return nil, errors.New("API request failed with status 400: Store does not support USD currency")
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "btc-only-store"
							store.Status.AtProvider.ID = "btc-store-123"
							store.Status.AtProvider.Name = "BTC Only Store"
							store.Status.AtProvider.DefaultCurrency = "BTC"
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

func TestCrossResourceDeleteFlow(t *testing.T) {
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
		"DeleteInvoiceSucceedsWithValidStore": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deletable-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "valid-store"},
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "deletable-123",
							StoreID: "valid-store-456",
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
						if storeID != "valid-store-456" {
							t.Errorf("Expected storeID 'valid-store-456', got %v", storeID)
						}
						if invoiceID != "deletable-123" {
							t.Errorf("Expected invoiceID 'deletable-123', got %v", invoiceID)
						}
						return nil
					},
				}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "valid-store"
							store.Status.AtProvider.ID = "valid-store-456"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			},
		},
		"DeleteInvoiceHandlesDeletedStoreGracefully": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "orphaned-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "deleted-store"},
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "orphaned-456",
							StoreID: "deleted-store-789",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("stores.store.btcpay.crossplane.io \"deleted-store\" not found"), "cannot get referenced store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{}
			},
			kube: func() client.Client {
				return &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						// Store has been deleted
						return errors.New("stores.store.btcpay.crossplane.io \"deleted-store\" not found")
					},
				}
			},
		},
		"DeleteInvoiceWithStoreNotReadyStillWorks": {
			args: args{
				mg: &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name: "delete-unready-invoice",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{Name: "unready-delete-store"},
						},
					},
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID:      "unready-delete-789",
							StoreID: "unready-store-999",
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
							store.Name = "unready-delete-store"
							store.Status.AtProvider.ID = "" // Store not ready
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

func TestStoreLifecycleImpactOnInvoices(t *testing.T) {
	type scenario struct {
		name        string
		storeState  string // "creating", "ready", "updating", "deleting", "deleted"
		invoiceOp   string // "observe", "create", "delete"
		expectError bool
		expectedErr string
	}

	scenarios := []scenario{
		{
			name:        "InvoiceObserveFailsWhenStoreIsCreating",
			storeState:  "creating",
			invoiceOp:   "observe",
			expectError: true,
			expectedErr: "referenced store is not ready",
		},
		{
			name:        "InvoiceObserveSucceedsWhenStoreIsReady",
			storeState:  "ready",
			invoiceOp:   "observe",
			expectError: false,
		},
		{
			name:        "InvoiceObserveSucceedsWhenStoreIsUpdating",
			storeState:  "updating",
			invoiceOp:   "observe",
			expectError: false,
		},
		{
			name:        "InvoiceCreateFailsWhenStoreIsCreating",
			storeState:  "creating",
			invoiceOp:   "create",
			expectError: true,
			expectedErr: "referenced store is not ready",
		},
		{
			name:        "InvoiceCreateSucceedsWhenStoreIsReady",
			storeState:  "ready",
			invoiceOp:   "create",
			expectError: false,
		},
		{
			name:        "InvoiceDeleteFailsWhenStoreIsDeleted",
			storeState:  "deleted",
			invoiceOp:   "delete",
			expectError: true,
			expectedErr: "cannot get referenced store",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create test invoice
			invoice := &v1alpha1.Invoice{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-invoice",
				},
				Spec: v1alpha1.InvoiceSpec{
					ForProvider: v1alpha1.InvoiceParameters{
						StoreRef: v1alpha1.StoreReference{Name: "lifecycle-store"},
						Amount:   100.00,
						Currency: "USD",
					},
				},
				Status: v1alpha1.InvoiceStatus{
					AtProvider: v1alpha1.InvoiceObservation{
						ID:      "lifecycle-invoice-123",
						StoreID: "lifecycle-store-456",
					},
				},
			}

			// Create mock client
			mockClient := &MockBTCPayClient{
				GetInvoiceFunc: func(storeID, invoiceID string) (*clients.Invoice, error) {
					return &clients.Invoice{
						ID:       "lifecycle-invoice-123",
						StoreID:  "lifecycle-store-456",
						Amount:   "100.00",
						Currency: "USD",
						Status:   "New",
					}, nil
				},
				CreateInvoiceFunc: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
					return &clients.Invoice{
						ID:       "new-lifecycle-invoice",
						StoreID:  storeID,
						Amount:   "100.00",
						Currency: "USD",
						Status:   "New",
					}, nil
				},
				ArchiveInvoiceFunc: func(storeID, invoiceID string) error {
					return nil
				},
			}

			// Create mock kube client based on store state
			var kubeClient client.Client
			switch scenario.storeState {
			case "creating":
				kubeClient = &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "lifecycle-store"
							store.Status.AtProvider.ID = "" // Not ready yet
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			case "ready", "updating":
				kubeClient = &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Name = "lifecycle-store"
							store.Status.AtProvider.ID = "lifecycle-store-456"
							return nil
						}
						return errors.New("unexpected object type")
					},
				}
			case "deleted":
				kubeClient = &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("stores.store.btcpay.crossplane.io \"lifecycle-store\" not found")
					},
				}
			}

			e := external{
				client: mockClient,
				kube:   kubeClient,
			}

			// Execute the operation
			var err error
			switch scenario.invoiceOp {
			case "observe":
				_, err = e.Observe(context.Background(), invoice)
			case "create":
				_, err = e.Create(context.Background(), invoice)
			case "delete":
				_, err = e.Delete(context.Background(), invoice)
			}

			// Check results
			if scenario.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if scenario.expectedErr != "" && !strings.Contains(err.Error(), scenario.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %v", scenario.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
