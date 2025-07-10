# BTCPay Provider Integration Tests

This directory contains end-to-end integration tests for the BTCPay Crossplane provider.

## Prerequisites

To run these integration tests, you need:

1. A Kubernetes cluster (e.g., kind, minikube, or a real cluster)
2. Crossplane installed in the cluster
3. The BTCPay provider installed
4. A running BTCPay Server instance
5. API credentials configured in a ProviderConfig

## Setup

1. Install Crossplane:
```bash
kubectl create namespace crossplane-system
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

2. Install the BTCPay provider:
```bash
kubectl apply -f package/crds/
kubectl apply -f examples/provider/config.yaml
```

3. Create a secret with your BTCPay API key:
```bash
kubectl create secret generic btcpay-credentials \
  --from-literal=apiKey="YOUR_BTCPAY_API_KEY" \
  -n crossplane-system
```

4. Create a ProviderConfig:
```yaml
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: btcpay-provider-config
spec:
  baseURL: https://your-btcpay-server.com
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
```

## Running the Tests

To run all integration tests:
```bash
go test -tags=integration ./test/e2e/... -v
```

To run specific tests:
```bash
go test -tags=integration ./test/e2e/... -v -run TestStoreLifecycle
```

## Test Coverage

The integration tests cover:

1. **Store Tests** (`store_test.go`):
   - Creating a new store
   - Updating store properties
   - Deleting a store
   - Verifying status updates

2. **Invoice Tests** (`invoice_test.go`):
   - Creating invoices with store references
   - Cross-namespace store references
   - Invoice deletion (archival)
   - Connection details publishing

## Writing New Integration Tests

When adding new integration tests:

1. Use the `+build integration` tag to exclude from unit tests
2. Follow the existing test structure
3. Always clean up resources after tests
4. Use appropriate timeouts for operations
5. Test both success and error scenarios

## Troubleshooting

If tests fail:

1. Check that the BTCPay server is accessible
2. Verify API credentials are correct
3. Ensure CRDs are installed: `kubectl get crds | grep btcpay`
4. Check provider pod logs: `kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-btcpay`
5. Verify ProviderConfig exists: `kubectl get providerconfigs.btcpay.crossplane.io`

## Environment Variables

- `BTCPAY_TEST_SERVER`: Override the BTCPay server URL for tests
- `BTCPAY_TEST_TIMEOUT`: Set custom timeout for operations (default: 5m)
- `SKIP_CLEANUP`: Set to "true" to skip resource cleanup (useful for debugging)