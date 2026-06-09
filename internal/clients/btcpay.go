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

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/rossigee/provider-btcpay/apis/v1beta1"
)

const (
	errNoProviderConfig     = "no providerConfig specified"
	errGetProviderConfig    = "cannot get providerConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal credentials"

	// Default BTCPay Server API path
	defaultAPIPath = "/api/v1"
)

// Config holds the configuration for the BTCPay Server API client
type Config struct {
	BaseURL string
	APIKey  string
}

// Credentials holds the API key for BTCPay Server
type Credentials struct {
	APIKey string `json:"apiKey"`
}

// BTCPayClient defines the interface for BTCPay Server operations
type BTCPayClient interface {
	// Store operations
	GetStore(ctx context.Context, storeID string) (*Store, error)
	ListStores(ctx context.Context) ([]Store, error)
	CreateStore(ctx context.Context, req CreateStoreRequest) (*Store, error)
	UpdateStore(ctx context.Context, storeID string, req UpdateStoreRequest) (*Store, error)
	DeleteStore(ctx context.Context, storeID string) error

	// Invoice operations
	GetInvoice(ctx context.Context, storeID, invoiceID string) (*Invoice, error)
	ListInvoices(ctx context.Context, storeID string) ([]Invoice, error)
	CreateInvoice(ctx context.Context, storeID string, req CreateInvoiceRequest) (*Invoice, error)
	ArchiveInvoice(ctx context.Context, storeID, invoiceID string) error
}

// Client is a BTCPay Server API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new BTCPay Server API client
func NewClient(cfg Config) *Client {
	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetConfig extracts the BTCPay client configuration from a ProviderConfig
func GetConfig(ctx context.Context, c client.Client, pcName string) (*Config, error) {
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, client.ObjectKey{Name: pcName}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, c, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errExtractCredentials)
	}

	creds := &Credentials{}
	if err := json.Unmarshal(data, creds); err != nil {
		return nil, errors.Wrap(err, errUnmarshalCredentials)
	}

	baseURL := ""
	if pc.Spec.BaseURL != nil && *pc.Spec.BaseURL != "" {
		baseURL = *pc.Spec.BaseURL
	}

	if baseURL == "" {
		return nil, errors.New("baseURL is required for BTCPay Server provider")
	}

	// Validate baseURL is a valid URL with https scheme
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "baseURL is not a valid URL")
	}

	if parsedURL.Scheme == "" {
		return nil, errors.New("baseURL must include a scheme (https://)")
	}

	// Allow http for localhost and cluster-local development
	host := strings.Split(parsedURL.Host, ":")[0] // Remove port if present
	isLocalhost := host == "localhost" || host == "127.0.0.1"
	isClusterLocal := strings.HasSuffix(host, ".svc.cluster.local")
	if parsedURL.Scheme != "https" && !isLocalhost && !isClusterLocal {
		return nil, errors.New("baseURL must use https scheme (http only allowed for localhost or cluster-local development)")
	}

	if parsedURL.Host == "" {
		return nil, errors.New("baseURL must include a valid host")
	}

	return &Config{
		BaseURL: baseURL,
		APIKey:  creds.APIKey,
	}, nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s%s", c.config.BaseURL, defaultAPIPath, path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req = req.WithContext(ctx)

	if c.config.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	req.Header.Set("Authorization", "token "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}

	return resp, nil
}

// parseResponse reads and unmarshals the response body
func parseResponse(resp *http.Response, target interface{}) error {
	defer func() {
		// Always drain body to release connection back to pool
		_, _ = io.Copy(io.Discard, resp.Body)
		if err := resp.Body.Close(); err != nil {
			_ = err
		}
	}()

	if resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("API request failed with status 404")
		}
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}

	return nil
}

// Store represents a BTCPay Server store
type Store struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Website              string                 `json:"website,omitempty"`
	DefaultCurrency      string                 `json:"defaultCurrency"`
	InvoiceExpiration    int32                  `json:"invoiceExpiration"`
	MonitoringExpiration int32                  `json:"monitoringExpiration"`
	PaymentTolerance     float64                `json:"paymentTolerance"`
	PaymentMethods       []string               `json:"paymentMethods,omitempty"`
	DerivationSchemes    map[string]interface{} `json:"derivationSchemes,omitempty"`
	SpeedPolicy          int                    `json:"speedPolicy"`
	LightningSettings    map[string]interface{} `json:"lightningSettings,omitempty"`
	OnChainSettings      map[string]interface{} `json:"onChainSettings,omitempty"`
	CreatedAt            *time.Time             `json:"createdAt,omitempty"`
}

// CreateStoreRequest represents a request to create a store
type CreateStoreRequest struct {
	Name                 string                 `json:"name"`
	Website              string                 `json:"website,omitempty"`
	DefaultCurrency      string                 `json:"defaultCurrency"`
	InvoiceExpiration    int32                  `json:"invoiceExpiration,omitempty"`
	MonitoringExpiration int32                  `json:"monitoringExpiration,omitempty"`
	PaymentTolerance     float64                `json:"paymentTolerance,omitempty"`
	SpeedPolicy          int                    `json:"speedPolicy,omitempty"`
	LightningSettings    map[string]interface{} `json:"lightningSettings,omitempty"`
	OnChainSettings      map[string]interface{} `json:"onChainSettings,omitempty"`
}

// UpdateStoreRequest represents a request to update a store
type UpdateStoreRequest struct {
	Name                 string                 `json:"name,omitempty"`
	Website              string                 `json:"website,omitempty"`
	DefaultCurrency      string                 `json:"defaultCurrency,omitempty"`
	InvoiceExpiration    int32                  `json:"invoiceExpiration,omitempty"`
	MonitoringExpiration int32                  `json:"monitoringExpiration,omitempty"`
	PaymentTolerance     float64                `json:"paymentTolerance,omitempty"`
	SpeedPolicy          int                    `json:"speedPolicy,omitempty"`
	LightningSettings    map[string]interface{} `json:"lightningSettings,omitempty"`
	OnChainSettings      map[string]interface{} `json:"onChainSettings,omitempty"`
}

// Invoice represents a BTCPay Server invoice
type Invoice struct {
	ID                   string                 `json:"id"`
	StoreID              string                 `json:"storeId"`
	Amount               string                 `json:"amount"`
	Currency             string                 `json:"currency"`
	Type                 string                 `json:"type"`
	CheckoutLink         string                 `json:"checkoutLink"`
	Status               string                 `json:"status"`
	AdditionalStatus     string                 `json:"additionalStatus"`
	MonitoringExpiration *time.Time             `json:"monitoringExpiration,omitempty"`
	ExpirationTime       *time.Time             `json:"expirationTime,omitempty"`
	CreatedTime          *time.Time             `json:"createdTime,omitempty"`
	OrderID              string                 `json:"orderId,omitempty"`
	NotificationURL      string                 `json:"notificationURL,omitempty"`
	RedirectURL          string                 `json:"redirectURL,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	PaymentMethods       []PaymentMethod        `json:"paymentMethods,omitempty"`
	Archived             bool                   `json:"archived"`
}

// PaymentMethod represents a payment method for an invoice
type PaymentMethod struct {
	PaymentMethod     string    `json:"paymentMethod"`
	CryptoCode        string    `json:"cryptoCode"`
	Destination       string    `json:"destination"`
	PaymentLink       string    `json:"paymentLink"`
	Rate              float64   `json:"rate"`
	PaymentMethodPaid float64   `json:"paymentMethodPaid"`
	TotalPaid         float64   `json:"totalPaid"`
	Due               float64   `json:"due"`
	Amount            float64   `json:"amount"`
	NetworkFee        float64   `json:"networkFee"`
	Payments          []Payment `json:"payments,omitempty"`
}

// Payment represents an individual payment
type Payment struct {
	ID           string     `json:"id"`
	ReceivedDate *time.Time `json:"receivedDate,omitempty"`
	Value        float64    `json:"value"`
	Fee          float64    `json:"fee"`
	Status       string     `json:"status"`
	Destination  string     `json:"destination"`
}

// CreateInvoiceRequest represents a request to create an invoice
type CreateInvoiceRequest struct {
	Amount                float64                `json:"amount"`
	Currency              string                 `json:"currency"`
	OrderID               string                 `json:"orderId,omitempty"`
	NotificationURL       string                 `json:"notificationURL,omitempty"`
	RedirectURL           string                 `json:"redirectURL,omitempty"`
	NotificationEmail     string                 `json:"notificationEmail,omitempty"`
	ItemDesc              string                 `json:"itemDesc,omitempty"`
	ItemCode              string                 `json:"itemCode,omitempty"`
	Physical              bool                   `json:"physical,omitempty"`
	TaxIncluded           bool                   `json:"taxIncluded,omitempty"`
	BuyerEmail            string                 `json:"buyerEmail,omitempty"`
	ExtendedNotifications bool                   `json:"extendedNotifications,omitempty"`
	FullNotifications     bool                   `json:"fullNotifications,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	CheckoutQueryString   string                 `json:"checkoutQueryString,omitempty"`
}

// GetStore retrieves a store by ID
func (c *Client) GetStore(ctx context.Context, storeID string) (*Store, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/stores/%s", storeID), nil)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := parseResponse(resp, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

// ListStores retrieves all stores
func (c *Client) ListStores(ctx context.Context) ([]Store, error) {
	resp, err := c.doRequest(ctx, "GET", "/stores", nil)
	if err != nil {
		return nil, err
	}

	var stores []Store
	if err := parseResponse(resp, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// CreateStore creates a new store
func (c *Client) CreateStore(ctx context.Context, req CreateStoreRequest) (*Store, error) {
	resp, err := c.doRequest(ctx, "POST", "/stores", req)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := parseResponse(resp, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

// UpdateStore updates an existing store
func (c *Client) UpdateStore(ctx context.Context, storeID string, req UpdateStoreRequest) (*Store, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/stores/%s", storeID), req)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := parseResponse(resp, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

// DeleteStore deletes a store
func (c *Client) DeleteStore(ctx context.Context, storeID string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/stores/%s", storeID), nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// GetInvoice retrieves an invoice by ID
func (c *Client) GetInvoice(ctx context.Context, storeID, invoiceID string) (*Invoice, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/stores/%s/invoices/%s", storeID, invoiceID), nil)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := parseResponse(resp, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// ListInvoices retrieves all invoices for a store
func (c *Client) ListInvoices(ctx context.Context, storeID string) ([]Invoice, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/stores/%s/invoices", storeID), nil)
	if err != nil {
		return nil, err
	}

	var invoices []Invoice
	if err := parseResponse(resp, &invoices); err != nil {
		return nil, err
	}

	return invoices, nil
}

// CreateInvoice creates a new invoice
func (c *Client) CreateInvoice(ctx context.Context, storeID string, req CreateInvoiceRequest) (*Invoice, error) {
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/stores/%s/invoices", storeID), req)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := parseResponse(resp, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// ArchiveInvoice archives an invoice (soft delete)
func (c *Client) ArchiveInvoice(ctx context.Context, storeID, invoiceID string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/stores/%s/invoices/%s", storeID, invoiceID), nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// IsNotFound returns true if the error indicates the resource was not found
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "status 404")
}
