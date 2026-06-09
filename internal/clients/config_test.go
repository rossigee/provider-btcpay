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
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/rossigee/provider-btcpay/apis"
	"github.com/rossigee/provider-btcpay/apis/v1beta1"
)

func TestGetConfig(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add core API to scheme: %v", err)
	}
	if err := apis.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add APIs to scheme: %v", err)
	}

	pcObj := getProviderConfig()
	secretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"credentials": []byte(`{"BaseURL":"https://example.com","APIKey":"my-key"}`),
		},
	}

	cases := map[string]struct {
		ctx    context.Context
		c      client.Client
		name   string
		config *Config
		err    error
	}{
		"GetConfigSuccessful": {
			ctx:  context.Background(),
			c:    fake.NewClientBuilder().WithScheme(scheme).WithObjects(pcObj, secretObj).Build(),
			name: "my-config",
			config: &Config{
				BaseURL: "https://example.com",
				APIKey:  "my-key",
			},
		},
		"GetConfigNotFound": {
			ctx:  context.Background(),
			c:    fake.NewClientBuilder().WithScheme(scheme).Build(),
			name: "my-config",
			err:  errors.New(errGetProviderConfig),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GetConfig(tc.ctx, tc.c, tc.name)
			if tc.err != nil {
				if err == nil {
					t.Fatalf("GetConfig() error = %v, wantErr %v", err, tc.err)
				}
			} else if err != nil {
				t.Fatalf("GetConfig() unexpected error: %v", err)
			}
			if tc.config != nil && err == nil {
				if diff := cmp.Diff(tc.config, got); diff != "" {
					t.Fatalf("GetConfig() -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func getProviderConfig() *v1beta1.ProviderConfig {
	return &v1beta1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-config",
		},
		Spec: v1beta1.ProviderConfigSpec{
			BaseURL: ptrs("https://example.com"),
			Credentials: v1beta1.ProviderCredentials{
				Source: xpv2.CredentialsSourceSecret,
				CommonCredentialSelectors: xpv2.CommonCredentialSelectors{
					SecretRef: &xpv2.SecretKeySelector{
						SecretReference: xpv2.SecretReference{
							Name:      "my-secret",
							Namespace: "default",
						},
						Key: "credentials",
					},
				},
			},
		},
	}
}

func ptrs(s string) *string {
	return &s
}
