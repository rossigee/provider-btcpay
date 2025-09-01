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
	}{
		"NetworkTimeout": {
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
				err: errors.Wrap(errors.New("context deadline exceeded"), "cannot get store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return nil, errors.New("context deadline exceeded")
					},
				}
			},
		},
		"UnauthorizedAccess": {
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
				err: errors.Wrap(errors.New("API request failed with status 401: Unauthorized"), "cannot get store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 401: Unauthorized")
					},
				}
			},
		},
		"ForbiddenAccess": {
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
				err: errors.Wrap(errors.New("API request failed with status 403: Forbidden - insufficient permissions"), "cannot get store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 403: Forbidden - insufficient permissions")
					},
				}
			},
		},
		"RateLimitExceeded": {
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
				err: errors.Wrap(errors.New("API request failed with status 429: Too Many Requests"), "cannot get store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					GetStoreFunc: func(storeID string) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 429: Too Many Requests")
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
	}{
		"DuplicateStoreName": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "duplicate-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Existing Store",
							DefaultCurrency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 409: Store name already exists"), "cannot create store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 409: Store name already exists")
					},
				}
			},
		},
		"InvalidCurrency": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "invalid-currency-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Test Store",
							DefaultCurrency: "INVALID",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 400: Invalid currency code"), "cannot create store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 400: Invalid currency code")
					},
				}
			},
		},
		"ServerOverloaded": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "overloaded-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Test Store",
							DefaultCurrency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 503: Service Temporarily Unavailable"), "cannot create store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 503: Service Temporarily Unavailable")
					},
				}
			},
		},
		"QuotaExceeded": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "quota-exceeded-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Test Store",
							DefaultCurrency: "USD",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 402: Payment Required - store limit exceeded"), "cannot create store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					CreateStoreFunc: func(req clients.CreateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 402: Payment Required - store limit exceeded")
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
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.c, got); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestUpdateErrorScenarios(t *testing.T) {
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
		"StoreNotFoundDuringUpdate": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "missing-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Updated Store",
							DefaultCurrency: "EUR",
						},
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "nonexistent-store",
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
		"ConflictDuringUpdate": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "conflict-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Conflicting Name",
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
				err: errors.Wrap(errors.New("API request failed with status 409: Store name conflicts with existing store"), "cannot update store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					UpdateStoreFunc: func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 409: Store name conflicts with existing store")
					},
				}
			},
		},
		"ValidationErrorDuringUpdate": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "validation-error-store",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "", // Empty name should cause validation error
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
				err: errors.Wrap(errors.New("API request failed with status 422: Validation failed - store name cannot be empty"), "cannot update store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					UpdateStoreFunc: func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
						return nil, errors.New("API request failed with status 422: Validation failed - store name cannot be empty")
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
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.u, got); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
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
	}{
		"StoreHasPendingInvoices": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "store-with-invoices",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 409: Cannot delete store with pending invoices"), "cannot delete store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						return errors.New("API request failed with status 409: Cannot delete store with pending invoices")
					},
				}
			},
		},
		"StoreHasPaymentMethods": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "store-with-payment-methods",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 409: Cannot delete store with configured payment methods"), "cannot delete store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						return errors.New("API request failed with status 409: Cannot delete store with configured payment methods")
					},
				}
			},
		},
		"InsufficientPermissions": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unauthorized-delete",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 403: Insufficient permissions to delete store"), "cannot delete store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						return errors.New("API request failed with status 403: Insufficient permissions to delete store")
					},
				}
			},
		},
		"ConcurrentModification": {
			args: args{
				mg: &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name: "concurrent-modification",
					},
					Status: v1alpha1.StoreStatus{
						AtProvider: v1alpha1.StoreObservation{
							ID: "store123",
						},
					},
				},
			},
			want: want{
				err: errors.Wrap(errors.New("API request failed with status 412: Precondition Failed - store was modified"), "cannot delete store"),
			},
			mock: func() *MockBTCPayClient {
				return &MockBTCPayClient{
					DeleteStoreFunc: func(storeID string) error {
						return errors.New("API request failed with status 412: Precondition Failed - store was modified")
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
