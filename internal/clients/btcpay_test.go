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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "404 error",
			err:      &testError{msg: "API request failed with status 404: Not Found"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "API request failed with status 500: Internal Server Error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNewClient(t *testing.T) {
	cfg := Config{
		BaseURL: "https://btcpay.example.com",
		APIKey:  "test-key",
	}

	client := NewClient(cfg)

	if client == nil {
		t.Error("NewClient() returned nil")
		return
	}

	if client.config.BaseURL != cfg.BaseURL {
		t.Errorf("client.config.BaseURL = %v, want %v", client.config.BaseURL, cfg.BaseURL)
	}

	if client.config.APIKey != cfg.APIKey {
		t.Errorf("client.config.APIKey = %v, want %v", client.config.APIKey, cfg.APIKey)
	}

	if client.httpClient == nil {
		t.Error("client.httpClient is nil")
	}

	// Check timeout is set
	if client.httpClient.Timeout == 0 {
		t.Error("client.httpClient.Timeout is not set")
	}
}

// circularReference creates a circular reference that cannot be marshaled to JSON
type circularReference struct {
	Self *circularReference `json:"self"`
}

func TestClient_doRequest_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() *Client
		method      string
		path        string
		body        interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "JSON marshal error with circular reference",
			setupClient: func() *Client {
				return NewClient(Config{
					BaseURL: "https://btcpay.example.com",
					APIKey:  "test-key",
				})
			},
			method: "POST",
			path:   "/test",
			body: func() interface{} {
				cr := &circularReference{}
				cr.Self = cr
				return cr
			}(),
			wantErr:     true,
			errContains: "failed to marshal request body",
		},
		{
			name: "Invalid URL construction",
			setupClient: func() *Client {
				return NewClient(Config{
					BaseURL: "://invalid-url",
					APIKey:  "test-key",
				})
			},
			method:      "GET",
			path:        "/test",
			body:        nil,
			wantErr:     true,
			errContains: "failed to create request",
		},
		{
			name: "HTTP client error - invalid method",
			setupClient: func() *Client {
				return NewClient(Config{
					BaseURL: "https://btcpay.example.com",
					APIKey:  "test-key",
				})
			},
			method:      "INVALID METHOD WITH SPACES",
			path:        "/test",
			body:        nil,
			wantErr:     true,
			errContains: "failed to create request",
		},
		{
			name: "Network error during request execution",
			setupClient: func() *Client {
				client := NewClient(Config{
					BaseURL: "https://nonexistent.invalid.domain.test",
					APIKey:  "test-key",
				})
				// Set a very short timeout to simulate network errors
				client.httpClient.Timeout = 1 * time.Millisecond
				return client
			},
			method:      "GET",
			path:        "/test",
			body:        nil,
			wantErr:     true,
			errContains: "failed to execute request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()

			resp, err := client.doRequest(tt.method, tt.path, tt.body)

			if tt.wantErr {
				if err == nil {
					t.Errorf("doRequest() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("doRequest() error = %v, want error containing %v", err, tt.errContains)
				}
				if resp != nil {
					t.Errorf("doRequest() returned response when error expected")
				}
			} else {
				if err != nil {
					t.Errorf("doRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestClient_doRequest_Success(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "token test-key" {
			t.Errorf("Expected Authorization header 'token test-key', got %v", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got %v", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept header 'application/json', got %v", r.Header.Get("Accept"))
		}

		// Verify request body for POST
		if r.Method == "POST" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body: %v", err)
			}
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				t.Errorf("Failed to unmarshal request body: %v", err)
			}
			if data["test"] != "value" {
				t.Errorf("Expected request body to contain test:value, got %v", data)
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "GET request without body",
			method: "GET",
			path:   "/test",
			body:   nil,
		},
		{
			name:   "POST request with body",
			method: "POST",
			path:   "/test",
			body:   map[string]interface{}{"test": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.doRequest(tt.method, tt.path, tt.body)

			if err != nil {
				t.Errorf("doRequest() unexpected error = %v", err)
				return
			}

			if resp == nil {
				t.Error("doRequest() returned nil response")
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("doRequest() status code = %v, want %v", resp.StatusCode, http.StatusOK)
			}

			// Clean up
			_ = resp.Body.Close()
		})
	}
}

func TestParseResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupResp   func() *http.Response
		target      interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "400 error with empty response body",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader("")),
				}
			},
			target:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "API request failed with status 400:",
		},
		{
			name: "500 error with error message",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Internal server error"}`)),
				}
			},
			target:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "API request failed with status 500:",
		},
		{
			name: "404 error with detailed message",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("Store not found")),
				}
			},
			target:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "API request failed with status 404: Store not found",
		},
		{
			name: "JSON decode error with malformed response",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"invalid": json}`)),
				}
			},
			target:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "failed to decode response",
		},
		{
			name: "StatusNoContent with nil target",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(strings.NewReader("")),
				}
			},
			target:  nil,
			wantErr: false,
		},
		{
			name: "StatusNoContent with non-nil target should not decode",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(strings.NewReader(`{"should": "not be decoded"}`)),
				}
			},
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "Valid JSON response with target",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"key": "value", "number": 42}`)),
				}
			},
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "Valid response with nil target should not decode",
			setupResp: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"key": "value"}`)),
				}
			},
			target:  nil,
			wantErr: false,
		},
		{
			name: "Response body read error simulation",
			setupResp: func() *http.Response {
				// Create a response with a body that will cause read errors
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       &errorReader{},
				}
			},
			target:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "API request failed with status 400:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.setupResp()
			err := parseResponse(resp, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseResponse() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseResponse() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("parseResponse() unexpected error = %v", err)
					return
				}

				// For successful cases with target, verify the data was parsed
				if tt.target != nil && resp.StatusCode == http.StatusOK {
					targetMap, ok := tt.target.(*map[string]interface{})
					if ok && len(*targetMap) == 0 {
						t.Error("parseResponse() should have populated target map but it's empty")
					}
				}
			}
		})
	}
}

// errorReader simulates a body that fails to read
type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (er *errorReader) Close() error {
	return nil
}

func TestParseResponse_TypeSpecific(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		target   interface{}
		validate func(t *testing.T, target interface{})
	}{
		{
			name:   "Parse into Store struct",
			json:   `{"id": "store123", "name": "Test Store", "defaultCurrency": "USD"}`,
			target: &Store{},
			validate: func(t *testing.T, target interface{}) {
				store := target.(*Store)
				if store.ID != "store123" {
					t.Errorf("Expected ID 'store123', got %v", store.ID)
				}
				if store.Name != "Test Store" {
					t.Errorf("Expected Name 'Test Store', got %v", store.Name)
				}
				if store.DefaultCurrency != "USD" {
					t.Errorf("Expected DefaultCurrency 'USD', got %v", store.DefaultCurrency)
				}
			},
		},
		{
			name:   "Parse into Invoice struct",
			json:   `{"id": "inv123", "storeId": "store123", "amount": "100.50", "currency": "USD"}`,
			target: &Invoice{},
			validate: func(t *testing.T, target interface{}) {
				invoice := target.(*Invoice)
				if invoice.ID != "inv123" {
					t.Errorf("Expected ID 'inv123', got %v", invoice.ID)
				}
				if invoice.StoreID != "store123" {
					t.Errorf("Expected StoreID 'store123', got %v", invoice.StoreID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(tt.json)),
			}

			err := parseResponse(resp, tt.target)
			if err != nil {
				t.Errorf("parseResponse() unexpected error = %v", err)
				return
			}

			tt.validate(t, tt.target)
		})
	}
}
