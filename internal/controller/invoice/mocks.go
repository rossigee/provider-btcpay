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
	"github.com/rossigee/provider-btcpay/internal/clients"
)

// MockBTCPayClient is a mock implementation of the BTCPayClient interface for testing
type MockBTCPayClient struct {
	// Store operations (needed for interface compliance)
	GetStoreFunc    func(storeID string) (*clients.Store, error)
	ListStoresFunc  func() ([]clients.Store, error)
	CreateStoreFunc func(req clients.CreateStoreRequest) (*clients.Store, error)
	UpdateStoreFunc func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error)
	DeleteStoreFunc func(storeID string) error

	// Invoice operations
	GetInvoiceFunc     func(storeID, invoiceID string) (*clients.Invoice, error)
	ListInvoicesFunc   func(storeID string) ([]clients.Invoice, error)
	CreateInvoiceFunc  func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error)
	ArchiveInvoiceFunc func(storeID, invoiceID string) error
}

// Store operations (for interface compliance)
func (m *MockBTCPayClient) GetStore(storeID string) (*clients.Store, error) {
	if m.GetStoreFunc != nil {
		return m.GetStoreFunc(storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListStores() ([]clients.Store, error) {
	if m.ListStoresFunc != nil {
		return m.ListStoresFunc()
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateStore(req clients.CreateStoreRequest) (*clients.Store, error) {
	if m.CreateStoreFunc != nil {
		return m.CreateStoreFunc(req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) UpdateStore(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error) {
	if m.UpdateStoreFunc != nil {
		return m.UpdateStoreFunc(storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) DeleteStore(storeID string) error {
	if m.DeleteStoreFunc != nil {
		return m.DeleteStoreFunc(storeID)
	}
	return nil
}

// Invoice operations
func (m *MockBTCPayClient) GetInvoice(storeID, invoiceID string) (*clients.Invoice, error) {
	if m.GetInvoiceFunc != nil {
		return m.GetInvoiceFunc(storeID, invoiceID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListInvoices(storeID string) ([]clients.Invoice, error) {
	if m.ListInvoicesFunc != nil {
		return m.ListInvoicesFunc(storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateInvoice(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error) {
	if m.CreateInvoiceFunc != nil {
		return m.CreateInvoiceFunc(storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ArchiveInvoice(storeID, invoiceID string) error {
	if m.ArchiveInvoiceFunc != nil {
		return m.ArchiveInvoiceFunc(storeID, invoiceID)
	}
	return nil
}
