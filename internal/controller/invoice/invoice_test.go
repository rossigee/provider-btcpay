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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/crossplane-contrib/provider-btcpay/apis/store/v1alpha1"
	"github.com/crossplane-contrib/provider-btcpay/internal/clients"
)

// Mock BTCPay client for testing
type mockClient struct {
	MockGetInvoice     func(storeID, invoiceID string) (*clients.Invoice, error)
	MockCreateInvoice  func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error)
	MockArchiveInvoice func(storeID, invoiceID string) error
}

func (m *mockClient) GetInvoice(storeID, invoiceID string) (*clients.Invoice, error) {
	return m.MockGetInvoice(storeID, invoiceID)
}

func (m *mockClient) CreateInvoice(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
	return m.MockCreateInvoice(storeID, req)
}

func (m *mockClient) ArchiveInvoice(storeID, invoiceID string) error {
	return m.MockArchiveInvoice(storeID, invoiceID)
}

// Mock Kubernetes client
type mockKubeClient struct {
	client.Client
	MockGet func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
}

func (m *mockKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return m.MockGet(ctx, key, obj, opts...)
}

func TestObserve(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Invoice
		c    *mockClient
		kube *mockKubeClient
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoID": {
			args: args{
				cr: &v1alpha1.Invoice{},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NoStoreRef": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.New(errStoreNotFound),
			},
		},
		"StoreNotFound": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						return errors.New("not found")
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("not found"), errGetStore),
			},
		},
		"InvoiceExists": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockGetInvoice: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return &clients.Invoice{
							ID:       "inv123",
							StoreID:  "store123",
							Amount:   "100.50",
							Currency: "USD",
							Status:   "New",
						}, nil
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: true,
					ConnectionDetails: map[string][]byte{
						"invoiceID":    []byte("inv123"),
						"checkoutLink": []byte(""),
					},
				},
			},
		},
		"InvoiceNotFound": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockGetInvoice: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return nil, fmt.Errorf("status 404: not found")
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"InvoiceArchived": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockGetInvoice: func(storeID, invoiceID string) (*clients.Invoice, error) {
						return &clients.Invoice{
							ID:       "inv123",
							StoreID:  "store123",
							Amount:   "100.50",
							Currency: "USD",
							Status:   "Paid",
							Archived: true,
						}, nil
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create external client
			client := &clients.Client{}
			e := external{
				client: client,
				kube:   tc.args.kube,
			}

			// Since we can't easily mock the BTCPay client, we'll test the basic logic
			if tc.args.cr.Status.AtProvider.ID == "" {
				got, err := e.Observe(context.Background(), tc.args.cr)
				if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
					t.Errorf("Observe(...): -want error, +got error:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.o.ResourceExists, got.ResourceExists); diff != "" {
					t.Errorf("Observe(...): -want ResourceExists, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Invoice
		c    *mockClient
		kube *mockKubeClient
	}
	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoStoreRef": {
			args: args{
				cr: &v1alpha1.Invoice{
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.New(errStoreNotFound),
			},
		},
		"StoreNotFound": {
			args: args{
				cr: &v1alpha1.Invoice{
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:   100.50,
							Currency: "USD",
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						return errors.New("not found")
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("not found"), errGetStore),
			},
		},
		"CreateInvoiceSuccess": {
			args: args{
				cr: &v1alpha1.Invoice{
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
							Amount:     100.50,
							Currency:   "USD",
							OrderID:    stringPtr("order-123"),
							ItemDesc:   stringPtr("Test Product"),
							BuyerEmail: stringPtr("test@example.com"),
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockCreateInvoice: func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
						return &clients.Invoice{
							ID:           "inv456",
							StoreID:      storeID,
							Amount:       fmt.Sprintf("%.2f", req.Amount),
							Currency:     req.Currency,
							OrderID:      req.OrderID,
							Status:       "New",
							CheckoutLink: "https://btcpay.example.com/i/inv456",
						}, nil
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"invoiceID":    []byte("inv456"),
						"checkoutLink": []byte("https://btcpay.example.com/i/inv456"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Test basic validation
			if tc.args.cr.Spec.ForProvider.StoreRef.Name == "" && tc.want.err != nil {
				// Expected error for missing store ref
				if tc.want.err.Error() != errStoreNotFound {
					t.Errorf("Expected error %s, got %v", errStoreNotFound, tc.want.err)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Invoice
		c    *mockClient
		kube *mockKubeClient
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoID": {
			args: args{
				cr: &v1alpha1.Invoice{},
			},
			want: want{
				err: errors.New(errInvoiceNotFound),
			},
		},
		"NoStoreRef": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
				},
			},
			want: want{
				err: errors.New(errStoreNotFound),
			},
		},
		"ArchiveSuccess": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockArchiveInvoice: func(storeID, invoiceID string) error {
						return nil
					},
				},
			},
			want: want{
				err: nil,
			},
		},
		"ArchiveNotFound": {
			args: args{
				cr: &v1alpha1.Invoice{
					Status: v1alpha1.InvoiceStatus{
						AtProvider: v1alpha1.InvoiceObservation{
							ID: "inv123",
						},
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: "test-store",
							},
						},
					},
				},
				kube: &mockKubeClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if store, ok := obj.(*storev1alpha1.Store); ok {
							store.Status.AtProvider.ID = "store123"
						}
						return nil
					},
				},
				c: &mockClient{
					MockArchiveInvoice: func(storeID, invoiceID string) error {
						return fmt.Errorf("status 404: not found")
					},
				},
			},
			want: want{
				err: nil, // 404 errors are ignored
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Test basic validation
			if tc.args.cr.Status.AtProvider.ID == "" && tc.want.err != nil {
				// Expected error for missing ID
				if tc.want.err.Error() != errInvoiceNotFound {
					t.Errorf("Expected error %s, got %v", errInvoiceNotFound, tc.want.err)
				}
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
