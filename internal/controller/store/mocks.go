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

	"github.com/rossigee/provider-btcpay/internal/clients"
)

// MockBTCPayClient is a mock implementation of the BTCPayClient interface for testing
type MockBTCPayClient struct {
	GetStoreFunc    func(ctx context.Context, storeID string) (*clients.Store, error)
	ListStoresFunc  func(ctx context.Context) ([]clients.Store, error)
	CreateStoreFunc func(ctx context.Context, req clients.CreateStoreRequest) (*clients.Store, error)
	UpdateStoreFunc func(ctx context.Context, storeID string, req clients.UpdateStoreRequest) (*clients.Store, error)
	DeleteStoreFunc func(ctx context.Context, storeID string) error

	GetInvoiceFunc     func(ctx context.Context, storeID, invoiceID string) (*clients.Invoice, error)
	ListInvoicesFunc   func(ctx context.Context, storeID string) ([]clients.Invoice, error)
	CreateInvoiceFunc  func(ctx context.Context, storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error)
	ArchiveInvoiceFunc func(ctx context.Context, storeID, invoiceID string) error
}

// Store operations
func (m *MockBTCPayClient) GetStore(ctx context.Context, storeID string) (*clients.Store, error) {
	if m.GetStoreFunc != nil {
		return m.GetStoreFunc(ctx, storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListStores(ctx context.Context) ([]clients.Store, error) {
	if m.ListStoresFunc != nil {
		return m.ListStoresFunc(ctx)
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateStore(ctx context.Context, req clients.CreateStoreRequest) (*clients.Store, error) {
	if m.CreateStoreFunc != nil {
		return m.CreateStoreFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) UpdateStore(ctx context.Context, storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
	if m.UpdateStoreFunc != nil {
		return m.UpdateStoreFunc(ctx, storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) DeleteStore(ctx context.Context, storeID string) error {
	if m.DeleteStoreFunc != nil {
		return m.DeleteStoreFunc(ctx, storeID)
	}
	return nil
}

// Invoice operations
func (m *MockBTCPayClient) GetInvoice(ctx context.Context, storeID, invoiceID string) (*clients.Invoice, error) {
	if m.GetInvoiceFunc != nil {
		return m.GetInvoiceFunc(ctx, storeID, invoiceID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListInvoices(ctx context.Context, storeID string) ([]clients.Invoice, error) {
	if m.ListInvoicesFunc != nil {
		return m.ListInvoicesFunc(ctx, storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateInvoice(ctx context.Context, storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
	if m.CreateInvoiceFunc != nil {
		return m.CreateInvoiceFunc(ctx, storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ArchiveInvoice(ctx context.Context, storeID, invoiceID string) error {
	if m.ArchiveInvoiceFunc != nil {
		return m.ArchiveInvoiceFunc(ctx, storeID, invoiceID)
	}
	return nil
}
