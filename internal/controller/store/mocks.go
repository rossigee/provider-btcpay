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
	"github.com/rossigee/provider-btcpay/internal/clients"
)

// MockBTCPayClient is a mock implementation of the BTCPayClient interface for testing
type MockBTCPayClient struct {
	GetStoreFunc    func(storeID string) (*clients.Store, error)
	ListStoresFunc  func() ([]clients.Store, error)
	CreateStoreFunc func(req clients.CreateStoreRequest) (*clients.Store, error)
	UpdateStoreFunc func(storeID string, req clients.UpdateStoreRequest) (*clients.Store, error)
	DeleteStoreFunc func(storeID string) error

	GetInvoiceFunc     func(storeID, invoiceID string) (*clients.Invoice, error)
	ListInvoicesFunc   func(storeID string) ([]clients.Invoice, error)
	CreateInvoiceFunc  func(storeID string, req clients.CreateInvoiceRequest) (*clients.Invoice, error)
	ArchiveInvoiceFunc func(storeID, invoiceID string) error

	// User operations
	GetUserFunc    func(userID string) (*clients.User, error)
	ListUsersFunc  func() ([]clients.User, error)
	CreateUserFunc func(req clients.CreateUserRequest) (*clients.User, error)
	UpdateUserFunc func(userID string, req clients.UpdateUserRequest) (*clients.User, error)
	DeleteUserFunc func(userID string) error

	// Webhook operations
	GetWebhookFunc    func(storeID, webhookID string) (*clients.Webhook, error)
	ListWebhooksFunc  func(storeID string) ([]clients.Webhook, error)
	CreateWebhookFunc func(storeID string, req clients.CreateWebhookRequest) (*clients.Webhook, error)
	UpdateWebhookFunc func(storeID, webhookID string, req clients.UpdateWebhookRequest) (*clients.Webhook, error)
	DeleteWebhookFunc func(storeID, webhookID string) error

	// Payment Method operations
	GetStorePaymentMethodFunc    func(storeID, cryptoCode, paymentType string) (*clients.StorePaymentMethod, error)
	ListStorePaymentMethodsFunc  func(storeID string) ([]clients.StorePaymentMethod, error)
	CreateStorePaymentMethodFunc func(storeID string, req clients.CreateStorePaymentMethodRequest) (*clients.StorePaymentMethod, error)
	UpdateStorePaymentMethodFunc func(storeID, cryptoCode, paymentType string, req clients.UpdateStorePaymentMethodRequest) (*clients.StorePaymentMethod, error)
	DeleteStorePaymentMethodFunc func(storeID, cryptoCode, paymentType string) error
}

// Store operations
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

// User operations
func (m *MockBTCPayClient) GetUser(userID string) (*clients.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(userID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListUsers() ([]clients.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc()
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateUser(req clients.CreateUserRequest) (*clients.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) UpdateUser(userID string, req clients.UpdateUserRequest) (*clients.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(userID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) DeleteUser(userID string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(userID)
	}
	return nil
}

// Webhook operations
func (m *MockBTCPayClient) GetWebhook(storeID, webhookID string) (*clients.Webhook, error) {
	if m.GetWebhookFunc != nil {
		return m.GetWebhookFunc(storeID, webhookID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListWebhooks(storeID string) ([]clients.Webhook, error) {
	if m.ListWebhooksFunc != nil {
		return m.ListWebhooksFunc(storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateWebhook(storeID string, req clients.CreateWebhookRequest) (*clients.Webhook, error) {
	if m.CreateWebhookFunc != nil {
		return m.CreateWebhookFunc(storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) UpdateWebhook(storeID, webhookID string, req clients.UpdateWebhookRequest) (*clients.Webhook, error) {
	if m.UpdateWebhookFunc != nil {
		return m.UpdateWebhookFunc(storeID, webhookID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) DeleteWebhook(storeID, webhookID string) error {
	if m.DeleteWebhookFunc != nil {
		return m.DeleteWebhookFunc(storeID, webhookID)
	}
	return nil
}

// Payment Method operations
func (m *MockBTCPayClient) GetStorePaymentMethod(storeID, cryptoCode, paymentType string) (*clients.StorePaymentMethod, error) {
	if m.GetStorePaymentMethodFunc != nil {
		return m.GetStorePaymentMethodFunc(storeID, cryptoCode, paymentType)
	}
	return nil, nil
}

func (m *MockBTCPayClient) ListStorePaymentMethods(storeID string) ([]clients.StorePaymentMethod, error) {
	if m.ListStorePaymentMethodsFunc != nil {
		return m.ListStorePaymentMethodsFunc(storeID)
	}
	return nil, nil
}

func (m *MockBTCPayClient) CreateStorePaymentMethod(storeID string, req clients.CreateStorePaymentMethodRequest) (*clients.StorePaymentMethod, error) {
	if m.CreateStorePaymentMethodFunc != nil {
		return m.CreateStorePaymentMethodFunc(storeID, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) UpdateStorePaymentMethod(storeID, cryptoCode, paymentType string, req clients.UpdateStorePaymentMethodRequest) (*clients.StorePaymentMethod, error) {
	if m.UpdateStorePaymentMethodFunc != nil {
		return m.UpdateStorePaymentMethodFunc(storeID, cryptoCode, paymentType, req)
	}
	return nil, nil
}

func (m *MockBTCPayClient) DeleteStorePaymentMethod(storeID, cryptoCode, paymentType string) error {
	if m.DeleteStorePaymentMethodFunc != nil {
		return m.DeleteStorePaymentMethodFunc(storeID, cryptoCode, paymentType)
	}
	return nil
}
