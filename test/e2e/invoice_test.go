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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
)

// TestInvoiceLifecycle tests the full lifecycle of a BTCPay Invoice resource.
func TestInvoiceLifecycle(t *testing.T) {
	cases := map[string]struct {
		reason string
		test   func(t *testing.T, c client.Client)
	}{
		"CreateAndDeleteInvoice": {
			reason: "Should be able to create and delete a BTCPay invoice",
			test: func(t *testing.T, c client.Client) {
				ctx := context.Background()

				// First create a store for the invoice
				store := &storev1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-store-for-invoice",
						Namespace: "default",
					},
					Spec: storev1alpha1.StoreSpec{
						ForProvider: storev1alpha1.StoreParameters{
							Name:            "Store for Invoice Test",
							DefaultCurrency: "USD",
						},
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "btcpay-provider-config",
							},
						},
					},
				}

				if err := c.Create(ctx, store); err != nil {
					t.Fatalf("Cannot create store: %v", err)
				}

				// Wait for the store to be ready
				if err := waitForCondition(ctx, c, store, xpv1.TypeReady, xpv1.ConditionTrue, 5*time.Minute); err != nil {
					t.Fatalf("Store did not become ready: %v", err)
				}

				// Create an invoice
				invoice := &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-invoice",
						Namespace: "default",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: store.Name,
							},
							Amount:      100.50,
							Currency:    "USD",
							OrderID:     stringPtr("TEST-ORDER-001"),
							ItemDesc:    stringPtr("Integration Test Product"),
							BuyerEmail:  stringPtr("test@example.com"),
							Metadata: map[string]string{
								"test":        "true",
								"environment": "integration",
							},
						},
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "btcpay-provider-config",
							},
						},
					},
				}

				// Create the invoice
				if err := c.Create(ctx, invoice); err != nil {
					t.Fatalf("Cannot create invoice: %v", err)
				}

				// Wait for the invoice to be ready
				if err := waitForCondition(ctx, c, invoice, xpv1.TypeReady, xpv1.ConditionTrue, 2*time.Minute); err != nil {
					t.Fatalf("Invoice did not become ready: %v", err)
				}

				// Verify the invoice has an ID assigned
				if err := c.Get(ctx, types.NamespacedName{Name: invoice.Name, Namespace: invoice.Namespace}, invoice); err != nil {
					t.Fatalf("Cannot get invoice: %v", err)
				}

				if invoice.Status.AtProvider.ID == "" {
					t.Error("Invoice ID was not set")
				}

				if invoice.Status.AtProvider.CheckoutLink == "" {
					t.Error("Invoice checkout link was not set")
				}

				if invoice.Status.AtProvider.Status == "" {
					t.Error("Invoice status was not set")
				}

				// Verify connection details were published
				if invoice.Status.AtProvider.CheckoutLink == "" {
					t.Error("Checkout link was not published to connection details")
				}

				// Delete the invoice
				if err := c.Delete(ctx, invoice); err != nil {
					t.Fatalf("Cannot delete invoice: %v", err)
				}

				// Wait for deletion
				if err := waitForDeletion(ctx, c, invoice, 2*time.Minute); err != nil {
					t.Fatalf("Invoice was not deleted: %v", err)
				}

				// Clean up the store
				if err := c.Delete(ctx, store); err != nil {
					t.Fatalf("Cannot delete store: %v", err)
				}
			},
		},
		"InvoiceWithCrossReference": {
			reason: "Should resolve cross-references between Invoice and Store",
			test: func(t *testing.T, c client.Client) {
				ctx := context.Background()

				// Create store in a different namespace (assuming it exists)
				// In a real test, we would create the namespace first
				store := &storev1alpha1.Store{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cross-ref-store",
						Namespace: "default", // Use default for now
					},
					Spec: storev1alpha1.StoreSpec{
						ForProvider: storev1alpha1.StoreParameters{
							Name:            "Cross Reference Store",
							DefaultCurrency: "EUR",
						},
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "btcpay-provider-config",
							},
						},
					},
				}

				if err := c.Create(ctx, store); err != nil {
					t.Fatalf("Cannot create store: %v", err)
				}

				// Wait for the store to be ready
				if err := waitForCondition(ctx, c, store, xpv1.TypeReady, xpv1.ConditionTrue, 5*time.Minute); err != nil {
					t.Fatalf("Store did not become ready: %v", err)
				}

				// Create invoice with cross-namespace reference
				invoice := &v1alpha1.Invoice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cross-ref-invoice",
						Namespace: "default",
					},
					Spec: v1alpha1.InvoiceSpec{
						ForProvider: v1alpha1.InvoiceParameters{
							StoreRef: v1alpha1.StoreReference{
								Name: store.Name,
								// Namespace is optional when in the same namespace
							},
							Amount:   50.00,
							Currency: "EUR",
						},
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "btcpay-provider-config",
							},
						},
					},
				}

				// Create the invoice
				if err := c.Create(ctx, invoice); err != nil {
					t.Fatalf("Cannot create invoice: %v", err)
				}

				// Wait for the invoice to be ready
				if err := waitForCondition(ctx, c, invoice, xpv1.TypeReady, xpv1.ConditionTrue, 2*time.Minute); err != nil {
					t.Fatalf("Invoice did not become ready: %v", err)
				}

				// Verify the cross-reference was resolved
				if err := c.Get(ctx, types.NamespacedName{Name: invoice.Name, Namespace: invoice.Namespace}, invoice); err != nil {
					t.Fatalf("Cannot get invoice: %v", err)
				}

				if invoice.Status.AtProvider.StoreID != store.Status.AtProvider.ID {
					t.Errorf("Invoice store ID mismatch: got %s, want %s", invoice.Status.AtProvider.StoreID, store.Status.AtProvider.ID)
				}

				// Clean up
				if err := c.Delete(ctx, invoice); err != nil {
					t.Fatalf("Cannot delete invoice: %v", err)
				}
				if err := c.Delete(ctx, store); err != nil {
					t.Fatalf("Cannot delete store: %v", err)
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