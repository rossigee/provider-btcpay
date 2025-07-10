//go:build integration
// +build integration

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

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	apisv1beta1 "github.com/crossplane-contrib/provider-btcpay/apis/v1beta1"
	invoicev1alpha1 "github.com/crossplane-contrib/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/crossplane-contrib/provider-btcpay/apis/store/v1alpha1"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestMain(m *testing.M) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		// Still run the tests but they will be skipped individually
		os.Exit(m.Run())
	}

	// Setup test environment
	ctx, cancel = context.WithCancel(context.Background())
	
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../package/crds",
		},
		ErrorIfCRDPathMissing: true,
		// Set UseExistingCluster to true if you want to run against
		// an existing cluster instead of creating a test environment
		UseExistingCluster: getBoolEnv("USE_EXISTING_CLUSTER", false),
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		panic(err)
	}

	// Add schemes
	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := xpv1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := apisv1beta1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := storev1alpha1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := invoicev1alpha1.AddToScheme(s); err != nil {
		panic(err)
	}

	// Create client
	k8sClient, err = client.New(cfg, client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cancel()
	if !getBoolEnv("SKIP_CLEANUP", false) {
		if err := testEnv.Stop(); err != nil {
			panic(err)
		}
	}

	os.Exit(code)
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// getClient returns the test Kubernetes client
func getClient() client.Client {
	return k8sClient
}

// getTimeout returns the timeout for operations
func getTimeout() time.Duration {
	if timeout := os.Getenv("BTCPAY_TEST_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			return d
		}
	}
	return 5 * time.Minute
}