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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/rossigee/provider-btcpay/apis/v1beta1"
)

type mockManagedResource struct {
	metav1.ObjectMeta
	spec mockManagedResourceSpec
}

type mockManagedResourceSpec struct {
	xpv1.ResourceSpec
}

func (m *mockManagedResource) GetProviderConfigReference() *xpv1.Reference {
	return m.spec.ProviderConfigReference
}

func (m *mockManagedResource) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return m.spec.WriteConnectionSecretToReference
}

func (m *mockManagedResource) SetProviderConfigReference(r *xpv1.Reference) {
	m.spec.ProviderConfigReference = r
}

func (m *mockManagedResource) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	m.spec.WriteConnectionSecretToReference = r
}

func (m *mockManagedResource) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return xpv1.Condition{}
}

func (m *mockManagedResource) SetConditions(c ...xpv1.Condition) {}

func (m *mockManagedResource) GetDeletionPolicy() xpv1.DeletionPolicy {
	return m.spec.DeletionPolicy
}

func (m *mockManagedResource) SetDeletionPolicy(p xpv1.DeletionPolicy) {
	m.spec.DeletionPolicy = p
}

func (m *mockManagedResource) GetManagementPolicies() xpv1.ManagementPolicies {
	return m.spec.ManagementPolicies
}

func (m *mockManagedResource) SetManagementPolicies(p xpv1.ManagementPolicies) {
	m.spec.ManagementPolicies = p
}

func (m *mockManagedResource) DeepCopyObject() runtime.Object {
	return &mockManagedResource{
		ObjectMeta: *m.DeepCopy(),
		spec:       m.spec,
	}
}

func (m *mockManagedResource) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func TestGetConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		kube client.Client
		mg   resource.Managed
	}
	type want struct {
		config *Config
		err    bool
	}

	providerConfigName := "test-provider-config"
	namespace := "crossplane-system"
	secretName := "btcpay-credentials"
	secretKey := "credentials"

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				ctx: context.Background(),
				kube: fake.NewClientBuilder().WithScheme(func() *runtime.Scheme {
					scheme := runtime.NewScheme()
					_ = v1beta1.AddToScheme(scheme)
					_ = corev1.AddToScheme(scheme)
					return scheme
				}()).WithObjects(
					&v1beta1.ProviderConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: providerConfigName,
						},
						Spec: v1beta1.ProviderConfigSpec{
							BaseURL: stringPtr("https://btcpay.example.com"),
							Credentials: v1beta1.ProviderCredentials{
								Source: xpv1.CredentialsSourceSecret,
								CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
									SecretRef: &xpv1.SecretKeySelector{
										SecretReference: xpv1.SecretReference{
											Name:      secretName,
											Namespace: namespace,
										},
										Key: secretKey,
									},
								},
							},
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
						Data: map[string][]byte{
							secretKey: []byte(`{"apiKey": "test-api-key-123"}`),
						},
					},
				).Build(),
				mg: &mockManagedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource",
						Namespace: namespace,
					},
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: providerConfigName,
							},
						},
					},
				},
			},
			want: want{
				config: &Config{
					BaseURL: "https://btcpay.example.com",
					APIKey:  "test-api-key-123",
				},
				err: false,
			},
		},
		"SuccessWithDefaultBaseURL": {
			args: args{
				ctx: context.Background(),
				kube: fake.NewClientBuilder().WithScheme(func() *runtime.Scheme {
					scheme := runtime.NewScheme()
					_ = v1beta1.AddToScheme(scheme)
					_ = corev1.AddToScheme(scheme)
					return scheme
				}()).WithObjects(
					&v1beta1.ProviderConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: providerConfigName,
						},
						Spec: v1beta1.ProviderConfigSpec{
							// BaseURL not specified, should use from credentials
							Credentials: v1beta1.ProviderCredentials{
								Source: xpv1.CredentialsSourceSecret,
								CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
									SecretRef: &xpv1.SecretKeySelector{
										SecretReference: xpv1.SecretReference{
											Name:      secretName,
											Namespace: namespace,
										},
										Key: secretKey,
									},
								},
							},
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
						Data: map[string][]byte{
							secretKey: []byte(`{"base_url": "https://btcpay.default.com", "apiKey": "test-api-key-456"}`),
						},
					},
				).Build(),
				mg: &mockManagedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource",
						Namespace: namespace,
					},
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: providerConfigName,
							},
						},
					},
				},
			},
			want: want{
				config: &Config{
					BaseURL: "https://btcpay.default.com",
					APIKey:  "test-api-key-456",
				},
				err: false,
			},
		},
		"NoProviderConfigRef": {
			args: args{
				ctx:  context.Background(),
				kube: fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build(),
				mg: &mockManagedResource{
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							// No ProviderConfigReference
						},
					},
				},
			},
			want: want{
				config: nil,
				err:    true,
			},
		},
		"ProviderConfigNotFound": {
			args: args{
				ctx:  context.Background(),
				kube: fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build(),
				mg: &mockManagedResource{
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "nonexistent",
							},
						},
					},
				},
			},
			want: want{
				config: nil,
				err:    true,
			},
		},
		"SecretNotFound": {
			args: args{
				ctx: context.Background(),
				kube: fake.NewClientBuilder().WithScheme(func() *runtime.Scheme {
					scheme := runtime.NewScheme()
					_ = v1beta1.AddToScheme(scheme)
					_ = corev1.AddToScheme(scheme)
					return scheme
				}()).WithObjects(
					&v1beta1.ProviderConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: providerConfigName,
						},
						Spec: v1beta1.ProviderConfigSpec{
							Credentials: v1beta1.ProviderCredentials{
								Source: xpv1.CredentialsSourceSecret,
								CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
									SecretRef: &xpv1.SecretKeySelector{
										SecretReference: xpv1.SecretReference{
											Name:      "nonexistent-secret",
											Namespace: namespace,
										},
										Key: secretKey,
									},
								},
							},
						},
					},
				).Build(),
				mg: &mockManagedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource",
						Namespace: namespace,
					},
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: providerConfigName,
							},
						},
					},
				},
			},
			want: want{
				config: nil,
				err:    true,
			},
		},
		"InvalidCredentialsJSON": {
			args: args{
				ctx: context.Background(),
				kube: fake.NewClientBuilder().WithScheme(func() *runtime.Scheme {
					scheme := runtime.NewScheme()
					_ = v1beta1.AddToScheme(scheme)
					_ = corev1.AddToScheme(scheme)
					return scheme
				}()).WithObjects(
					&v1beta1.ProviderConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: providerConfigName,
						},
						Spec: v1beta1.ProviderConfigSpec{
							Credentials: v1beta1.ProviderCredentials{
								Source: xpv1.CredentialsSourceSecret,
								CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
									SecretRef: &xpv1.SecretKeySelector{
										SecretReference: xpv1.SecretReference{
											Name:      secretName,
											Namespace: namespace,
										},
										Key: secretKey,
									},
								},
							},
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
						Data: map[string][]byte{
							secretKey: []byte(`invalid json`),
						},
					},
				).Build(),
				mg: &mockManagedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource",
						Namespace: namespace,
					},
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: providerConfigName,
							},
						},
					},
				},
			},
			want: want{
				config: nil,
				err:    true,
			},
		},
		"MissingAPIKey": {
			args: args{
				ctx: context.Background(),
				kube: fake.NewClientBuilder().WithScheme(func() *runtime.Scheme {
					scheme := runtime.NewScheme()
					_ = v1beta1.AddToScheme(scheme)
					_ = corev1.AddToScheme(scheme)
					return scheme
				}()).WithObjects(
					&v1beta1.ProviderConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: providerConfigName,
						},
						Spec: v1beta1.ProviderConfigSpec{
							BaseURL: stringPtr("https://btcpay.example.com"),
							Credentials: v1beta1.ProviderCredentials{
								Source: xpv1.CredentialsSourceSecret,
								CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
									SecretRef: &xpv1.SecretKeySelector{
										SecretReference: xpv1.SecretReference{
											Name:      secretName,
											Namespace: namespace,
										},
										Key: secretKey,
									},
								},
							},
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: namespace,
						},
						Data: map[string][]byte{
							secretKey: []byte(`{"base_url": "https://btcpay.example.com"}`), // Missing apiKey
						},
					},
				).Build(),
				mg: &mockManagedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource",
						Namespace: namespace,
					},
					spec: mockManagedResourceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: providerConfigName,
							},
						},
					},
				},
			},
			want: want{
				config: nil,
				err:    true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			got, err := GetConfig(tc.args.ctx, tc.args.kube, tc.args.mg)
			if (err != nil) != tc.want.err {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.config, got); diff != "" {
				t.Errorf("GetConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetConfigErrorMessages(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = v1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Test specific error messages
	tests := []struct {
		name    string
		kube    client.Client
		mg      resource.Managed
		wantErr string
	}{
		{
			name: "NoProviderConfigRef",
			kube: fake.NewClientBuilder().WithScheme(scheme).Build(),
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{},
				},
			},
			wantErr: "cannot get providerConfig",
		},
		{
			name: "MissingAPIKey",
			kube: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&v1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-config"},
					Spec: v1beta1.ProviderConfigSpec{
						Credentials: v1beta1.ProviderCredentials{
							Source: xpv1.CredentialsSourceSecret,
							CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
								SecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name:      "test-secret",
										Namespace: "default",
									},
									Key: "creds",
								},
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"creds": []byte(`{}`),
					},
				},
			).Build(),
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{Name: "test-config"},
					},
				},
			},
			wantErr: "cannot unmarshal credentials",
		},
		{
			name: "MissingBaseURLEverywhere",
			kube: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&v1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-config"},
					Spec: v1beta1.ProviderConfigSpec{
						// No BaseURL in spec
						Credentials: v1beta1.ProviderCredentials{
							Source: xpv1.CredentialsSourceSecret,
							CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
								SecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name:      "test-secret",
										Namespace: "default",
									},
									Key: "creds",
								},
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"creds": []byte(`{"apiKey": "test-key"}`), // No base_url in credentials either
					},
				},
			).Build(),
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{Name: "test-config"},
					},
				},
			},
			wantErr: "baseURL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetConfig(ctx, tt.kube, tt.mg)
			if err == nil {
				t.Error("GetConfig() expected error, got nil")
				return
			}
			if !errors.Is(err, errors.New(tt.wantErr)) && err.Error() != tt.wantErr {
				// Check if error contains the expected message
				if !containsError(err, tt.wantErr) {
					t.Errorf("GetConfig() error = %v, want error containing %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestGetConfig_AdditionalEdgeCases(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = v1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name    string
		objects []client.Object
		mg      resource.Managed
		want    *Config
		wantErr bool
	}{
		{
			name: "EmptySecretData",
			objects: []client.Object{
				&v1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-config"},
					Spec: v1beta1.ProviderConfigSpec{
						BaseURL: stringPtr("https://test.com"),
						Credentials: v1beta1.ProviderCredentials{
							Source: xpv1.CredentialsSourceSecret,
							CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
								SecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name: "test-secret", Namespace: "default",
									},
									Key: "creds",
								},
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
					Data:       map[string][]byte{}, // Empty data
				},
			},
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{Name: "test-config"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "SecretKeyNotFound",
			objects: []client.Object{
				&v1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-config"},
					Spec: v1beta1.ProviderConfigSpec{
						BaseURL: stringPtr("https://test.com"),
						Credentials: v1beta1.ProviderCredentials{
							Source: xpv1.CredentialsSourceSecret,
							CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
								SecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name: "test-secret", Namespace: "default",
									},
									Key: "missing-key", // Key doesn't exist
								},
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
					Data: map[string][]byte{
						"other-key": []byte(`{"apiKey": "test"}`),
					},
				},
			},
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{Name: "test-config"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BaseURLFromCredentialsOverridesSpec",
			objects: []client.Object{
				&v1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "test-config"},
					Spec: v1beta1.ProviderConfigSpec{
						BaseURL: stringPtr("https://spec.com"), // This should be overridden
						Credentials: v1beta1.ProviderCredentials{
							Source: xpv1.CredentialsSourceSecret,
							CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
								SecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name: "test-secret", Namespace: "default",
									},
									Key: "creds",
								},
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
					Data: map[string][]byte{
						"creds": []byte(`{"base_url": "https://credentials.com", "apiKey": "test-key"}`),
					},
				},
			},
			mg: &mockManagedResource{
				spec: mockManagedResourceSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{Name: "test-config"},
					},
				},
			},
			want: &Config{
				BaseURL: "https://credentials.com", // Should use credentials value
				APIKey:  "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.objects...).Build()

			got, err := GetConfig(ctx, kube, tt.mg)

			if tt.wantErr {
				if err == nil {
					t.Error("GetConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetConfig() unexpected error = %v", err)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func containsError(err error, substr string) bool {
	if err == nil {
		return false
	}
	return errors.Cause(err).Error() == substr
}
