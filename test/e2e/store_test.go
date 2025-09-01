//go:build integration
// +build integration

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

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
)

// TestStoreLifecycle tests the full lifecycle of a BTCPay Store resource.
func TestStoreLifecycle(t *testing.T) {
	cases := map[string]struct {
		reason string
		test   func(t *testing.T, c client.Client)
	}{
		"CreateAndDeleteStore": {
			reason: "Should be able to create and delete a BTCPay store",
			test: func(t *testing.T, c client.Client) {
				ctx := context.Background()

				// Create a test store
				store := &v1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-store",
						Namespace: "default",
					},
					Spec: v1alpha1.StoreSpec{
						ForProvider: v1alpha1.StoreParameters{
							Name:            "Integration Test Store",
							DefaultCurrency: "USD",
							Website:         stringPtr("https://integration-test.example.com"),
							SpeedPolicy:     stringPtr("Medium"),
						},
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "btcpay-provider-config",
							},
						},
					},
				}

				// Create the store
				if err := c.Create(ctx, store); err != nil {
					t.Fatalf("Cannot create store: %v", err)
				}

				// Wait for the store to be ready
				if err := waitForCondition(ctx, c, store, xpv1.TypeReady, xpv1.ConditionTrue, 5*time.Minute); err != nil {
					t.Fatalf("Store did not become ready: %v", err)
				}

				// Verify the store has an ID assigned
				if err := c.Get(ctx, types.NamespacedName{Name: store.Name, Namespace: store.Namespace}, store); err != nil {
					t.Fatalf("Cannot get store: %v", err)
				}

				if store.Status.AtProvider.ID == "" {
					t.Error("Store ID was not set")
				}

				// Update the store
				store.Spec.ForProvider.Name = "Updated Integration Test Store"
				store.Spec.ForProvider.SpeedPolicy = stringPtr("High")
				if err := c.Update(ctx, store); err != nil {
					t.Fatalf("Cannot update store: %v", err)
				}

				// Wait for the update to be reflected
				time.Sleep(10 * time.Second)

				// Verify the update
				if err := c.Get(ctx, types.NamespacedName{Name: store.Name, Namespace: store.Namespace}, store); err != nil {
					t.Fatalf("Cannot get updated store: %v", err)
				}

				if store.Status.AtProvider.Name != "Updated Integration Test Store" {
					t.Errorf("Store name was not updated: got %s, want %s", store.Status.AtProvider.Name, "Updated Integration Test Store")
				}

				// Delete the store
				if err := c.Delete(ctx, store); err != nil {
					t.Fatalf("Cannot delete store: %v", err)
				}

				// Wait for deletion
				if err := waitForDeletion(ctx, c, store, 5*time.Minute); err != nil {
					t.Fatalf("Store was not deleted: %v", err)
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Skip if not running integration tests
			if os.Getenv("INTEGRATION_TESTS") != "true" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
				t.Skipf("Skipping integration test %s: set INTEGRATION_TESTS=true to run", name)
				return
			}
			
			// Use the test client from suite setup
			tc.test(t, getClient())
		})
	}
}

