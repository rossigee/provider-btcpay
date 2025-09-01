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

package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
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
		"StoreDoesNotExist": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
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
		"StoreExistsAndUpToDate": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Test Store",
							DefaultCurrency: "USD",
							Website:         testStringPtr("https://test.com"),
							SpeedPolicy:     testStringPtr("Medium"),
						},
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
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
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						return &clients.Store{
							ID:              "store123",
							Name:            "Test Store",
							DefaultCurrency: "USD",
							Website:         "https://test.com",
							SpeedPolicy:     6, // Medium
						}, nil
					},
				}
			},
		},
		"StoreExistsButOutdated": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Updated Store Name",
							DefaultCurrency: "EUR",
							Website:         testStringPtr("https://updated.com"),
							SpeedPolicy:     testStringPtr("High"),
						},
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return &clients.Store{
							ID:              "store123",
							Name:            "Test Store", // Different from spec
							DefaultCurrency: "USD",        // Different from spec
							Website:         "https://test.com",
							SpeedPolicy:     6, // Medium
						}, nil
					},
				}
			},
		},
		"GetStoreError": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 500: Internal Server Error"), "cannot get store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 500: Internal Server Error")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.mock()}

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
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "New Test Store",
							DefaultCurrency: "USD",
							Website:         testStringPtr("https://newstore.com"),
							SpeedPolicy:     testStringPtr("High"),
						},
					},
				},
			},
			want: want{
				c: managed.ExternalCreation{},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						// Verify request parameters
						if req.Name != "New Test Store" {
							t.Errorf("Expected Name 'New Test Store', got %v", req.Name)
						}
						if req.DefaultCurrency != "USD" {
							t.Errorf("Expected DefaultCurrency 'USD', got %v", req.DefaultCurrency)
						}
						if req.Website != "https://newstore.com" {
							t.Errorf("Expected Website 'https://newstore.com', got %v", req.Website)
						}
						if req.SpeedPolicy != 1 { // High = 1
							t.Errorf("Expected SpeedPolicy 1 (High), got %v", req.SpeedPolicy)
						}

						return &clients.Store{
							ID:              "new-store-123",
							Name:            req.Name,
							DefaultCurrency: req.DefaultCurrency,
							Website:         req.Website,
							SpeedPolicy:     req.SpeedPolicy,
						}, nil
					},
				}
			},
		},
		"CreateWithMinimalFields": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "minimal-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Minimal Store",
							DefaultCurrency: "BTC",
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
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						// Verify optional fields are empty
						if req.Website != "" {
							t.Errorf("Expected Website to be empty, got %v", req.Website)
						}
						if req.SpeedPolicy != 0 {
							t.Errorf("Expected SpeedPolicy to be 0, got %v", req.SpeedPolicy)
						}

						return &clients.Store{
							ID:              "minimal-store-456",
							Name:            req.Name,
							DefaultCurrency: req.DefaultCurrency,
						}, nil
					},
				}
			},
		},
		"CreateError": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "error-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Error Store",
							DefaultCurrency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Invalid store name"), "cannot create store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 400: Invalid store name")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.mock()}

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
		"UpdateSuccess": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Updated Store Name",
							DefaultCurrency: "EUR",
							Website:         testStringPtr("https://updated.com"),
							SpeedPolicy:     testStringPtr("Low"),
						},
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				u: managed.ExternalUpdate{},
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					UpdateStoreFunc: func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
						// Verify parameters
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						if req.Name != "Updated Store Name" {
							t.Errorf("Expected Name 'Updated Store Name', got %v", req.Name)
						}
						if req.DefaultCurrency != "EUR" {
							t.Errorf("Expected DefaultCurrency 'EUR', got %v", req.DefaultCurrency)
						}
						if req.Website != "https://updated.com" {
							t.Errorf("Expected Website 'https://updated.com', got %v", req.Website)
						}
						if req.SpeedPolicy != 144 { // Low = 144
							t.Errorf("Expected SpeedPolicy 144 (Low), got %v", req.SpeedPolicy)
						}

						return &clients.Store{
							ID:              storeID,
							Name:            req.Name,
							DefaultCurrency: req.DefaultCurrency,
							Website:         req.Website,
							SpeedPolicy:     req.SpeedPolicy,
						}, nil
					},
				}
			},
		},
		"UpdateError": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Updated Store",
							DefaultCurrency: "USD",
						},
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 404: Store not found"), "cannot update store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					UpdateStoreFunc: func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 404: Store not found")
					},
				}
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
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: nil,
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						if storeID != "store123" {
							t.Errorf("Expected storeID 'store123', got %v", storeID)
						}
						return nil
					},
				}
			},
		},
		"DeleteError": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-store",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 500: Cannot delete store with pending invoices"), "cannot delete store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						return errors.New("API request failed with status 500: Cannot delete store with pending invoices")
					},
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.mock()}

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
