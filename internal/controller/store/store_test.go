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
	"errors"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"github.com/rossigee/provider-btcpay/internal/clients"
)

func TestObserve(t *testing.T) {
	ctx := context.Background()

	storeID := "test-store-id"

	cases := map[string]struct {
		cr      *storev1alpha1.Store
		mock    *MockBTCPayClient
		want    managed.ExternalObservation
		wantErr bool
	}{
		"StoreNotFound": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-store",
					UID:               types.UID("test-uid"),
					Generation:        1,
					CreationTimestamp: metav1.Now(),
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name: "test-store",
					},
				},
			},
			mock: &MockBTCPayClient{
				GetStoreFunc: func(ctx context.Context, storeID string) (*clients.Store, error) {
					return nil, errors.New("status 404")
				},
			},
			want: managed.ExternalObservation{
				ResourceExists: false,
			},
			wantErr: false,
		},
		"StoreFoundUpToDate": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-store",
					UID:               types.UID("test-uid"),
					Generation:        1,
					CreationTimestamp: metav1.Now(),
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:              "test-store",
						DefaultCurrency:   "USD",
						InvoiceExpiration: int32Ptr(3600),
					},
				},
				Status: storev1alpha1.StoreStatus{
					AtProvider: storev1alpha1.StoreObservation{
						ID: storeID,
					},
				},
			},
			mock: &MockBTCPayClient{
				GetStoreFunc: func(ctx context.Context, sid string) (*clients.Store, error) {
					if sid != storeID {
						t.Errorf("expected store ID %s, got %s", storeID, sid)
					}
					return &clients.Store{
						ID:                storeID,
						Name:              "test-store",
						DefaultCurrency:   "USD",
						InvoiceExpiration: 3600,
						PaymentMethods:    []string{"BTC", "Lightning"},
					}, nil
				},
			},
			want: managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: true,
			},
			wantErr: false,
		},
		"StoreFoundNeedsUpdate": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-store",
					UID:               types.UID("test-uid"),
					Generation:        1,
					CreationTimestamp: metav1.Now(),
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name: "updated-store",
					},
				},
				Status: storev1alpha1.StoreStatus{
					AtProvider: storev1alpha1.StoreObservation{
						ID: storeID,
					},
				},
			},
			mock: &MockBTCPayClient{
				GetStoreFunc: func(ctx context.Context, sid string) (*clients.Store, error) {
					return &clients.Store{
						ID:                storeID,
						Name:              "old-store",
						DefaultCurrency:   "USD",
						InvoiceExpiration: 3600,
						PaymentMethods:    []string{"BTC"},
					}, nil
				},
			},
			want: managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: false,
			},
			wantErr: false,
		},
		"NoStoreIDYet": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-store",
					UID:               types.UID("test-uid"),
					Generation:        1,
					CreationTimestamp: metav1.Now(),
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name: "test-store",
					},
				},
			},
			mock: &MockBTCPayClient{},
			want: managed.ExternalObservation{
				ResourceExists: false,
			},
			wantErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := &external{client: tc.mock}
			got, err := ext.Observe(ctx, tc.cr)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got.ResourceExists != tc.want.ResourceExists {
				t.Errorf("ResourceExists: got %v, want %v", got.ResourceExists, tc.want.ResourceExists)
			}
			if got.ResourceUpToDate != tc.want.ResourceUpToDate {
				t.Errorf("ResourceUpToDate: got %v, want %v", got.ResourceUpToDate, tc.want.ResourceUpToDate)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		cr      *storev1alpha1.Store
		mock    *MockBTCPayClient
		wantErr bool
	}{
		"CreateSuccess": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:            "test-store",
						DefaultCurrency: "USD",
					},
				},
			},
			mock: &MockBTCPayClient{
				CreateStoreFunc: func(ctx context.Context, req clients.CreateStoreRequest) (*clients.Store, error) {
					if req.Name != "test-store" {
						t.Errorf("expected name 'test-store', got %s", req.Name)
					}
					return &clients.Store{
						ID:              "new-store-id",
						Name:            req.Name,
						DefaultCurrency: req.DefaultCurrency,
					}, nil
				},
			},
			wantErr: false,
		},
		"CreateFailure": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name: "test-store",
					},
				},
			},
			mock: &MockBTCPayClient{
				CreateStoreFunc: func(ctx context.Context, req clients.CreateStoreRequest) (*clients.Store, error) {
					return nil, errors.New("api error")
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := &external{client: tc.mock}
			_, err := ext.Create(ctx, tc.cr)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tc.wantErr && tc.cr.Status.AtProvider.ID == "" {
				t.Error("expected store ID to be set in status")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		cr      *storev1alpha1.Store
		mock    *MockBTCPayClient
		wantErr bool
	}{
		"UpdateSuccess": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:            "updated-store",
						DefaultCurrency: "EUR",
					},
				},
				Status: storev1alpha1.StoreStatus{
					AtProvider: storev1alpha1.StoreObservation{
						ID: "store-123",
					},
				},
			},
			mock: &MockBTCPayClient{
				UpdateStoreFunc: func(ctx context.Context, storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
					if storeID != "store-123" {
						t.Errorf("expected store ID 'store-123', got %s", storeID)
					}
					return &clients.Store{
						ID:              storeID,
						Name:            req.Name,
						DefaultCurrency: req.DefaultCurrency,
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := &external{client: tc.mock}
			_, err := ext.Update(ctx, tc.cr)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		cr      *storev1alpha1.Store
		mock    *MockBTCPayClient
		wantErr bool
	}{
		"DeleteSuccess": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
				Status: storev1alpha1.StoreStatus{
					AtProvider: storev1alpha1.StoreObservation{
						ID: "store-123",
					},
				},
			},
			mock: &MockBTCPayClient{
				DeleteStoreFunc: func(ctx context.Context, storeID string) error {
					if storeID != "store-123" {
						t.Errorf("expected store ID 'store-123', got %s", storeID)
					}
					return nil
				},
			},
			wantErr: false,
		},
		"DeleteNotFound": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
				Status: storev1alpha1.StoreStatus{
					AtProvider: storev1alpha1.StoreObservation{
						ID: "store-123",
					},
				},
			},
			mock: &MockBTCPayClient{
				DeleteStoreFunc: func(ctx context.Context, storeID string) error {
					return errors.New("status 404")
				},
			},
			wantErr: false,
		},
		"NoStoreID": {
			cr: &storev1alpha1.Store{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-store",
				},
			},
			mock:    &MockBTCPayClient{},
			wantErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := &external{client: tc.mock}
			_, err := ext.Delete(ctx, tc.cr)

			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	store := &clients.Store{
		ID:                "test-id",
		Name:              "test-store",
		DefaultCurrency:   "USD",
		InvoiceExpiration: 3600,
	}

	cases := map[string]struct {
		cr   *storev1alpha1.Store
		want bool
	}{
		"UpToDate": {
			cr: &storev1alpha1.Store{
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:              "test-store",
						DefaultCurrency:   "USD",
						InvoiceExpiration: int32Ptr(3600),
					},
				},
			},
			want: true,
		},
		"NameChanged": {
			cr: &storev1alpha1.Store{
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:              "different-name",
						DefaultCurrency:   "USD",
						InvoiceExpiration: int32Ptr(3600),
					},
				},
			},
			want: false,
		},
		"CurrencyChanged": {
			cr: &storev1alpha1.Store{
				Spec: storev1alpha1.StoreSpec{
					ForProvider: storev1alpha1.StoreParameters{
						Name:              "test-store",
						DefaultCurrency:   "EUR",
						InvoiceExpiration: int32Ptr(3600),
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := &external{}
			got := ext.isUpToDate(tc.cr, store)
			if got != tc.want {
				t.Errorf("isUpToDate: got %v, want %v", got, tc.want)
			}
		})
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

var _ resource.Managed = &storev1alpha1.Store{}

func TestStoreImplementsManagedInterface(t *testing.T) {
	var _ resource.Managed = &storev1alpha1.Store{}
}
