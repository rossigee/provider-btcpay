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

	"github.com/rossigee/provider-btcpay/internal/clients"
)

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

	GetWebhookFunc    func(ctx context.Context, storeID, webhookID string) (*clients.Webhook, error)
	ListWebhooksFunc  func(ctx context.Context, storeID string) ([]clients.Webhook, error)
	CreateWebhookFunc func(ctx context.Context, storeID string, req clients.CreateWebhookRequest) (*clients.Webhook, error)
	UpdateWebhookFunc func(ctx context.Context, storeID, webhookID string, req clients.UpdateWebhookRequest) (*clients.Webhook, error)
	DeleteWebhookFunc func(ctx context.Context, storeID, webhookID string) error

	GetUserFunc    func(ctx context.Context, userID string) (*clients.User, error)
	ListUsersFunc  func(ctx context.Context) ([]clients.User, error)
	CreateUserFunc func(ctx context.Context, req clients.CreateUserRequest) (*clients.User, error)
	UpdateUserFunc func(ctx context.Context, userID string, req clients.UpdateUserRequest) (*clients.User, error)
	DeleteUserFunc func(ctx context.Context, userID string) error

	GetApiKeyFunc    func(ctx context.Context, apiKeyID string) (*clients.ApiKey, error)
	ListApiKeysFunc  func(ctx context.Context) ([]clients.ApiKey, error)
	CreateApiKeyFunc func(ctx context.Context, req clients.CreateApiKeyRequest) (*clients.ApiKey, error)
	DeleteApiKeyFunc func(ctx context.Context, apiKeyID string) error

	GetGuestFunc    func(ctx context.Context, storeID, guestID string) (*clients.Guest, error)
	ListGuestsFunc  func(ctx context.Context, storeID string) ([]clients.Guest, error)
	CreateGuestFunc func(ctx context.Context, storeID string, req clients.CreateGuestRequest) (*clients.Guest, error)
	DeleteGuestFunc func(ctx context.Context, storeID, guestID string) error

	GetSharedLinkFunc    func(ctx context.Context, storeID, linkID string) (*clients.SharedLink, error)
	ListSharedLinksFunc  func(ctx context.Context, storeID string) ([]clients.SharedLink, error)
	CreateSharedLinkFunc func(ctx context.Context, storeID string, req clients.CreateSharedLinkRequest) (*clients.SharedLink, error)
	DeleteSharedLinkFunc func(ctx context.Context, storeID, linkID string) error
}

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
func (m *MockBTCPayClient) GetWebhook(ctx context.Context, storeID, webhookID string) (*clients.Webhook, error) {
	if m.GetWebhookFunc != nil {
		return m.GetWebhookFunc(ctx, storeID, webhookID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) ListWebhooks(ctx context.Context, storeID string) ([]clients.Webhook, error) {
	if m.ListWebhooksFunc != nil {
		return m.ListWebhooksFunc(ctx, storeID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) CreateWebhook(ctx context.Context, storeID string, req clients.CreateWebhookRequest) (*clients.Webhook, error) {
	if m.CreateWebhookFunc != nil {
		return m.CreateWebhookFunc(ctx, storeID, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) UpdateWebhook(ctx context.Context, storeID, webhookID string, req clients.UpdateWebhookRequest) (*clients.Webhook, error) {
	if m.UpdateWebhookFunc != nil {
		return m.UpdateWebhookFunc(ctx, storeID, webhookID, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) DeleteWebhook(ctx context.Context, storeID, webhookID string) error {
	if m.DeleteWebhookFunc != nil {
		return m.DeleteWebhookFunc(ctx, storeID, webhookID)
	}
	return nil
}
func (m *MockBTCPayClient) GetUser(ctx context.Context, userID string) (*clients.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) ListUsers(ctx context.Context) ([]clients.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}
func (m *MockBTCPayClient) CreateUser(ctx context.Context, req clients.CreateUserRequest) (*clients.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) UpdateUser(ctx context.Context, userID string, req clients.UpdateUserRequest) (*clients.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, userID, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) DeleteUser(ctx context.Context, userID string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, userID)
	}
	return nil
}
func (m *MockBTCPayClient) GetApiKey(ctx context.Context, apiKeyID string) (*clients.ApiKey, error) {
	if m.GetApiKeyFunc != nil {
		return m.GetApiKeyFunc(ctx, apiKeyID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) ListApiKeys(ctx context.Context) ([]clients.ApiKey, error) {
	if m.ListApiKeysFunc != nil {
		return m.ListApiKeysFunc(ctx)
	}
	return nil, nil
}
func (m *MockBTCPayClient) CreateApiKey(ctx context.Context, req clients.CreateApiKeyRequest) (*clients.ApiKey, error) {
	if m.CreateApiKeyFunc != nil {
		return m.CreateApiKeyFunc(ctx, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) DeleteApiKey(ctx context.Context, apiKeyID string) error {
	if m.DeleteApiKeyFunc != nil {
		return m.DeleteApiKeyFunc(ctx, apiKeyID)
	}
	return nil
}
func (m *MockBTCPayClient) GetGuest(ctx context.Context, storeID, guestID string) (*clients.Guest, error) {
	if m.GetGuestFunc != nil {
		return m.GetGuestFunc(ctx, storeID, guestID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) ListGuests(ctx context.Context, storeID string) ([]clients.Guest, error) {
	if m.ListGuestsFunc != nil {
		return m.ListGuestsFunc(ctx, storeID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) CreateGuest(ctx context.Context, storeID string, req clients.CreateGuestRequest) (*clients.Guest, error) {
	if m.CreateGuestFunc != nil {
		return m.CreateGuestFunc(ctx, storeID, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) DeleteGuest(ctx context.Context, storeID, guestID string) error {
	if m.DeleteGuestFunc != nil {
		return m.DeleteGuestFunc(ctx, storeID, guestID)
	}
	return nil
}
func (m *MockBTCPayClient) GetSharedLink(ctx context.Context, storeID, linkID string) (*clients.SharedLink, error) {
	if m.GetSharedLinkFunc != nil {
		return m.GetSharedLinkFunc(ctx, storeID, linkID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) ListSharedLinks(ctx context.Context, storeID string) ([]clients.SharedLink, error) {
	if m.ListSharedLinksFunc != nil {
		return m.ListSharedLinksFunc(ctx, storeID)
	}
	return nil, nil
}
func (m *MockBTCPayClient) CreateSharedLink(ctx context.Context, storeID string, req clients.CreateSharedLinkRequest) (*clients.SharedLink, error) {
	if m.CreateSharedLinkFunc != nil {
		return m.CreateSharedLinkFunc(ctx, storeID, req)
	}
	return nil, nil
}
func (m *MockBTCPayClient) DeleteSharedLink(ctx context.Context, storeID, linkID string) error {
	if m.DeleteSharedLinkFunc != nil {
		return m.DeleteSharedLinkFunc(ctx, storeID, linkID)
	}
	return nil
}
