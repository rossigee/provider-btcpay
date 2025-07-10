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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_CreateInvoice(t *testing.T) {
	type args struct {
		storeID string
		req     CreateInvoiceRequest
	}
	type want struct {
		invoice *Invoice
		err     bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID: "store123",
				req: CreateInvoiceRequest{
					Amount:   100.50,
					Currency: "USD",
					OrderID:  "order-123",
					ItemDesc: "Test Product",
					BuyerEmail: "test@example.com",
					NotificationURL: "https://example.com/webhook",
					Metadata: map[string]interface{}{
						"customer_id": "cust123",
						"order_type":  "digital",
					},
				},
			},
			want: want{
				invoice: &Invoice{
					ID:       "inv123",
					StoreID:  "store123",
					Amount:   "100.50",
					Currency: "USD",
					Status:   "New",
					OrderID:  "order-123",
					CheckoutLink: "https://btcpay.example.com/i/inv123",
					Metadata: map[string]interface{}{
						"customer_id": "cust123",
						"order_type":  "digital",
					},
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123/invoices" {
					t.Errorf("expected /api/v1/stores/store123/invoices, got %s", r.URL.Path)
				}

				var req CreateInvoiceRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}

				resp := Invoice{
					ID:       "inv123",
					StoreID:  "store123",
					Amount:   fmt.Sprintf("%.2f", req.Amount),
					Currency: req.Currency,
					Status:   "New",
					OrderID:  req.OrderID,
					CheckoutLink: "https://btcpay.example.com/i/inv123",
					Metadata: req.Metadata,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"MinimalInvoice": {
			args: args{
				storeID: "store123",
				req: CreateInvoiceRequest{
					Amount:   10.0,
					Currency: "BTC",
				},
			},
			want: want{
				invoice: &Invoice{
					ID:       "inv124",
					StoreID:  "store123",
					Amount:   "10.00",
					Currency: "BTC",
					Status:   "New",
					CheckoutLink: "https://btcpay.example.com/i/inv124",
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				var req CreateInvoiceRequest
				json.NewDecoder(r.Body).Decode(&req)

				resp := Invoice{
					ID:       "inv124",
					StoreID:  "store123",
					Amount:   fmt.Sprintf("%.2f", req.Amount),
					Currency: req.Currency,
					Status:   "New",
					CheckoutLink: "https://btcpay.example.com/i/inv124",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"InvalidStore": {
			args: args{
				storeID: "invalid",
				req: CreateInvoiceRequest{
					Amount:   50.0,
					Currency: "USD",
				},
			},
			want: want{
				invoice: nil,
				err:     true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "The specified store was not found",
				})
			},
		},
		"ValidationError": {
			args: args{
				storeID: "store123",
				req: CreateInvoiceRequest{
					Amount:   -10.0, // Invalid negative amount
					Currency: "USD",
				},
			},
			want: want{
				invoice: nil,
				err:     true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Amount must be positive",
				})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			got, err := client.CreateInvoice(tc.args.storeID, tc.args.req)
			if (err != nil) != tc.want.err {
				t.Errorf("CreateInvoice() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			
			if diff := cmp.Diff(tc.want.invoice, got); diff != "" {
				t.Errorf("CreateInvoice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_GetInvoice(t *testing.T) {
	type args struct {
		storeID   string
		invoiceID string
	}
	type want struct {
		invoice *Invoice
		err     bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID:   "store123",
				invoiceID: "inv123",
			},
			want: want{
				invoice: &Invoice{
					ID:       "inv123",
					StoreID:  "store123",
					Amount:   "100.50",
					Currency: "USD",
					Status:   "Paid",
					OrderID:  "order-123",
					CheckoutLink: "https://btcpay.example.com/i/inv123",
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123/invoices/inv123" {
					t.Errorf("expected /api/v1/stores/store123/invoices/inv123, got %s", r.URL.Path)
				}

				resp := Invoice{
					ID:       "inv123",
					StoreID:  "store123",
					Amount:   "100.50",
					Currency: "USD",
					Status:   "Paid",
					OrderID:  "order-123",
					CheckoutLink: "https://btcpay.example.com/i/inv123",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"NotFound": {
			args: args{
				storeID:   "store123",
				invoiceID: "nonexistent",
			},
			want: want{
				invoice: nil,
				err:     true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "The specified invoice was not found",
				})
			},
		},
		"Expired": {
			args: args{
				storeID:   "store123",
				invoiceID: "inv125",
			},
			want: want{
				invoice: &Invoice{
					ID:       "inv125",
					StoreID:  "store123",
					Amount:   "50.00",
					Currency: "USD",
					Status:   "Expired",
					AdditionalStatus: "paidPartial",
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Invoice{
					ID:       "inv125",
					StoreID:  "store123",
					Amount:   "50.00",
					Currency: "USD",
					Status:   "Expired",
					AdditionalStatus: "paidPartial",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			got, err := client.GetInvoice(tc.args.storeID, tc.args.invoiceID)
			if (err != nil) != tc.want.err {
				t.Errorf("GetInvoice() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.invoice, got); diff != "" {
				t.Errorf("GetInvoice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ArchiveInvoice(t *testing.T) {
	type args struct {
		storeID   string
		invoiceID string
	}
	type want struct {
		err bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID:   "store123",
				invoiceID: "inv123",
			},
			want: want{
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123/invoices/inv123" {
					t.Errorf("expected /api/v1/stores/store123/invoices/inv123, got %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		"NotFound": {
			args: args{
				storeID:   "store123",
				invoiceID: "nonexistent",
			},
			want: want{
				err: true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "The specified invoice was not found",
				})
			},
		},
		"Forbidden": {
			args: args{
				storeID:   "store123",
				invoiceID: "inv123",
			},
			want: want{
				err: true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Cannot archive a paid invoice",
				})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			err := client.ArchiveInvoice(tc.args.storeID, tc.args.invoiceID)
			if (err != nil) != tc.want.err {
				t.Errorf("ArchiveInvoice() error = %v, wantErr %v", err, tc.want.err)
			}
		})
	}
}

func TestClient_ListStores(t *testing.T) {
	type want struct {
		stores []Store
		err    bool
	}

	cases := map[string]struct {
		want    want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			want: want{
				stores: []Store{
					{
						ID:              "store1",
						Name:            "Store One",
						DefaultCurrency: "USD",
						SpeedPolicy:     1,
					},
					{
						ID:              "store2",
						Name:            "Store Two",
						DefaultCurrency: "EUR",
						SpeedPolicy:     2,
					},
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores" {
					t.Errorf("expected /api/v1/stores, got %s", r.URL.Path)
				}

				resp := []Store{
					{
						ID:              "store1",
						Name:            "Store One",
						DefaultCurrency: "USD",
						SpeedPolicy:     1,
					},
					{
						ID:              "store2",
						Name:            "Store Two",
						DefaultCurrency: "EUR",
						SpeedPolicy:     2,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"EmptyList": {
			want: want{
				stores: []Store{},
				err:    false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := []Store{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"ServerError": {
			want: want{
				stores: nil,
				err:    true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			got, err := client.ListStores()
			if (err != nil) != tc.want.err {
				t.Errorf("ListStores() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.stores, got); diff != "" {
				t.Errorf("ListStores() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ListInvoices(t *testing.T) {
	type args struct {
		storeID string
	}
	type want struct {
		invoices []Invoice
		err      bool
	}

	cases := map[string]struct {
		args    args
		want    want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID: "store123",
			},
			want: want{
				invoices: []Invoice{
					{
						ID:       "inv1",
						StoreID:  "store123",
						Amount:   "50.00",
						Currency: "USD",
						Status:   "Paid",
					},
					{
						ID:       "inv2",
						StoreID:  "store123",
						Amount:   "75.00",
						Currency: "USD",
						Status:   "New",
					},
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123/invoices" {
					t.Errorf("expected /api/v1/stores/store123/invoices, got %s", r.URL.Path)
				}

				resp := []Invoice{
					{
						ID:       "inv1",
						StoreID:  "store123",
						Amount:   "50.00",
						Currency: "USD",
						Status:   "Paid",
					},
					{
						ID:       "inv2",
						StoreID:  "store123",
						Amount:   "75.00",
						Currency: "USD",
						Status:   "New",
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"EmptyList": {
			args: args{
				storeID: "store123",
			},
			want: want{
				invoices: []Invoice{},
				err:      false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := []Invoice{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"StoreNotFound": {
			args: args{
				storeID: "nonexistent",
			},
			want: want{
				invoices: nil,
				err:      true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "The specified store was not found",
				})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			got, err := client.ListInvoices(tc.args.storeID)
			if (err != nil) != tc.want.err {
				t.Errorf("ListInvoices() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.invoices, got); diff != "" {
				t.Errorf("ListInvoices() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_InvoiceAPIErrorScenarios(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		handler  func(w http.ResponseWriter, r *http.Request)
		testFunc func(client *Client) error
		wantErr  bool
	}{
		{
			name:   "GetInvoice - Server Error",
			method: "GetInvoice",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Database connection failed"}`))
			},
			testFunc: func(client *Client) error {
				_, err := client.GetInvoice("test-store", "test-invoice")
				return err
			},
			wantErr: true,
		},
		{
			name:   "CreateInvoice - Validation Error",
			method: "CreateInvoice",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid amount: must be positive"}`))
			},
			testFunc: func(client *Client) error {
				_, err := client.CreateInvoice("test-store", CreateInvoiceRequest{
					Amount:   -100.0,
					Currency: "USD",
				})
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})

			err := tt.testFunc(client)

			if tt.wantErr && err == nil {
				t.Errorf("%s expected error but got none", tt.method)
			} else if !tt.wantErr && err != nil {
				t.Errorf("%s unexpected error = %v", tt.method, err)
			}
		})
	}
}
