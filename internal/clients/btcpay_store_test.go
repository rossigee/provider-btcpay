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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestClient_CreateStore(t *testing.T) {
	type args struct {
		req CreateStoreRequest
	}
	type want struct {
		store *Store
		err   bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				req: CreateStoreRequest{
					Name:            "Test Store",
					DefaultCurrency: "USD",
					Website:         "https://example.com",
				},
			},
			want: want{
				store: &Store{
					ID:              "store123",
					Name:            "Test Store",
					Website:         "https://example.com",
					DefaultCurrency: "USD",
					InvoiceExpiration: 900,
					PaymentTolerance: 0,
					SpeedPolicy:     1, // Medium
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores" {
					t.Errorf("expected /api/v1/stores, got %s", r.URL.Path)
				}

				var req CreateStoreRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}

				resp := Store{
					ID:              "store123",
					Name:            req.Name,
					Website:         req.Website,
					DefaultCurrency: req.DefaultCurrency,
					InvoiceExpiration: 900,
					PaymentTolerance: 0,
					SpeedPolicy:     1,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"ServerError": {
			args: args{
				req: CreateStoreRequest{
					Name:            "Test Store",
					DefaultCurrency: "USD",
				},
			},
			want: want{
				store: nil,
				err:   true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			},
		},
		"InvalidResponse": {
			args: args{
				req: CreateStoreRequest{
					Name:            "Test Store",
					DefaultCurrency: "USD",
				},
			},
			want: want{
				store: nil,
				err:   true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("invalid json"))
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

			got, err := client.CreateStore(tc.args.req)
			if (err != nil) != tc.want.err {
				t.Errorf("CreateStore() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.store, got); diff != "" {
				t.Errorf("CreateStore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_GetStore(t *testing.T) {
	type args struct {
		storeID string
	}
	type want struct {
		store *Store
		err   bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID: "store123",
			},
			want: want{
				store: &Store{
					ID:              "store123",
					Name:            "Test Store",
					Website:         "https://example.com",
					DefaultCurrency: "USD",
					InvoiceExpiration: 900,
					PaymentTolerance: 0,
					SpeedPolicy:     1,
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123" {
					t.Errorf("expected /api/v1/stores/store123, got %s", r.URL.Path)
				}

				resp := Store{
					ID:              "store123",
					Name:            "Test Store",
					Website:         "https://example.com",
					DefaultCurrency: "USD",
					InvoiceExpiration: 900,
					PaymentTolerance: 0,
					SpeedPolicy:     1,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"NotFound": {
			args: args{
				storeID: "nonexistent",
			},
			want: want{
				store: nil,
				err:   true,
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

			got, err := client.GetStore(tc.args.storeID)
			if (err != nil) != tc.want.err {
				t.Errorf("GetStore() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.store, got); diff != "" {
				t.Errorf("GetStore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_UpdateStore(t *testing.T) {
	type args struct {
		storeID string
		req     UpdateStoreRequest
	}
	type want struct {
		store *Store
		err   bool
	}

	cases := map[string]struct {
		args args
		want want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"Success": {
			args: args{
				storeID: "store123",
				req: UpdateStoreRequest{
					Name:            "Updated Store",
					Website:         "https://updated.com",
					InvoiceExpiration: 1800,
					SpeedPolicy:     2,
				},
			},
			want: want{
				store: &Store{
					ID:              "store123",
					Name:            "Updated Store",
					Website:         "https://updated.com",
					DefaultCurrency: "USD",
					InvoiceExpiration: 1800,
					SpeedPolicy:     2,
				},
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123" {
					t.Errorf("expected /api/v1/stores/store123, got %s", r.URL.Path)
				}

				var req UpdateStoreRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}

				resp := Store{
					ID:              "store123",
					Name:            req.Name,
					Website:         req.Website,
					DefaultCurrency: "USD",
					InvoiceExpiration: req.InvoiceExpiration,
					SpeedPolicy:     req.SpeedPolicy,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
		},
		"NotFound": {
			args: args{
				storeID: "nonexistent",
				req: UpdateStoreRequest{
					Name: "Updated Store",
				},
			},
			want: want{
				store: nil,
				err:   true,
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

			got, err := client.UpdateStore(tc.args.storeID, tc.args.req)
			if (err != nil) != tc.want.err {
				t.Errorf("UpdateStore() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.store, got); diff != "" {
				t.Errorf("UpdateStore() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_DeleteStore(t *testing.T) {
	type args struct {
		storeID string
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
				storeID: "store123",
			},
			want: want{
				err: false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/stores/store123" {
					t.Errorf("expected /api/v1/stores/store123, got %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		"NotFound": {
			args: args{
				storeID: "nonexistent",
			},
			want: want{
				err: true,
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

			err := client.DeleteStore(tc.args.storeID)
			if (err != nil) != tc.want.err {
				t.Errorf("DeleteStore() error = %v, wantErr %v", err, tc.want.err)
			}
		})
	}
}

func TestClient_doRequest(t *testing.T) {
	type args struct {
		method string
		path   string
		body   interface{}
	}
	type want struct {
		statusCode int
		err        bool
	}

	cases := map[string]struct {
		args    args
		want    want
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		"AuthHeaderSet": {
			args: args{
				method: http.MethodGet,
				path:   "/test",
				body:   nil,
			},
			want: want{
				statusCode: http.StatusOK,
				err:        false,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "token test-key" {
					t.Errorf("expected Authorization header 'token test-key', got %s", auth)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		"Timeout": {
			args: args{
				method: http.MethodGet,
				path:   "/timeout",
				body:   nil,
			},
			want: want{
				statusCode: 0,
				err:        true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Simulate timeout by sleeping longer than client timeout
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			// Create client with short timeout for timeout test
			client := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})
			if name == "Timeout" {
				client.httpClient.Timeout = 100 * time.Millisecond
			}

			resp, err := client.doRequest(tc.args.method, tc.args.path, tc.args.body)
			if (err != nil) != tc.want.err {
				t.Errorf("doRequest() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if resp != nil && resp.StatusCode != tc.want.statusCode {
				t.Errorf("doRequest() statusCode = %v, want %v", resp.StatusCode, tc.want.statusCode)
			}
		})
	}
}

func TestClient_StoreAPIErrorScenarios(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		handler  func(w http.ResponseWriter, r *http.Request)
		testFunc func(client *Client) error
		wantErr  bool
	}{
		{
			name:   "GetStore - Server Error",
			method: "GetStore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			},
			testFunc: func(client *Client) error {
				_, err := client.GetStore("test-store")
				return err
			},
			wantErr: true,
		},
		{
			name:   "GetStore - Not Found",
			method: "GetStore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`Store not found`))
			},
			testFunc: func(client *Client) error {
				_, err := client.GetStore("nonexistent-store")
				return err
			},
			wantErr: true,
		},
		{
			name:   "CreateStore - Validation Error",
			method: "CreateStore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid store name"}`))
			},
			testFunc: func(client *Client) error {
				_, err := client.CreateStore(CreateStoreRequest{
					Name:            "",
					DefaultCurrency: "USD",
				})
				return err
			},
			wantErr: true,
		},
		{
			name:   "UpdateStore - Not Found",
			method: "UpdateStore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`Store not found`))
			},
			testFunc: func(client *Client) error {
				_, err := client.UpdateStore("nonexistent-store", UpdateStoreRequest{
					Name: "Updated Store",
				})
				return err
			},
			wantErr: true,
		},
		{
			name:   "DeleteStore - Server Error",
			method: "DeleteStore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Cannot delete store with pending invoices"}`))
			},
			testFunc: func(client *Client) error {
				return client.DeleteStore("store-with-invoices")
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

			if tt.wantErr {
				if err == nil {
					t.Errorf("%s expected error but got none", tt.method)
				}
			} else {
				if err != nil {
					t.Errorf("%s unexpected error = %v", tt.method, err)
				}
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}