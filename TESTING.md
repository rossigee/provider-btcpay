# Testing the BTCPay Server Crossplane Provider

This guide covers multiple ways to test the BTCPay provider locally and in different environments.

## Testing Options

### 1. Unit Tests (Fastest)
Already passing - comprehensive mock-based testing:
```bash
make test.unit
```

### 2. Local Development Testing
Test the provider locally without full Kubernetes deployment.

### 3. Full Integration Testing
Complete end-to-end testing with Kubernetes cluster + BTCPay Server.

## Option 1: Quick Local Development Testing

### Prerequisites
- Go 1.21+
- A BTCPay Server instance (can be remote)
- BTCPay Server API key

### Setup
1. **Get BTCPay Server access**:
   - Use existing instance: `https://your-btcpay-server.example.com` (from config)
   - Or run locally with Docker:
     ```bash
     docker run -d -p 14142:14142 btcpayserver/btcpayserver:latest
     ```

2. **Get API credentials**:
   - Log into BTCPay Server → Account Settings → Access Tokens
   - Create new token with Store and Invoice permissions
   - Save the API key

3. **Test the client directly**:
   ```bash
   cd internal/clients
   go test -v -run TestClient_ListStores
   ```

### Manual Client Testing
Create a simple test file to verify BTCPay connectivity:

```go
// test_manual.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/crossplane-contrib/provider-btcpay/internal/clients"
)

func main() {
    apiKey := os.Getenv("BTCPAY_API_KEY")
    baseURL := os.Getenv("BTCPAY_BASE_URL") // e.g., https://your-btcpay-server.example.com
    
    if apiKey == "" || baseURL == "" {
        log.Fatal("Set BTCPAY_API_KEY and BTCPAY_BASE_URL environment variables")
    }
    
    client := clients.NewBTCPayClient(baseURL, apiKey)
    
    // Test listing stores
    stores, err := client.ListStores(context.Background())
    if err != nil {
        log.Fatalf("Failed to list stores: %v", err)
    }
    
    fmt.Printf("Found %d stores:\n", len(stores))
    for _, store := range stores {
        fmt.Printf("- %s (%s)\n", store.Name, store.ID)
    }
}
```

Run with:
```bash
export BTCPAY_API_KEY="your-api-key"
export BTCPAY_BASE_URL="https://your-btcpay-server.example.com"
go run test_manual.go
```

## Option 2: Full Integration Testing

### Prerequisites
- Docker
- kubectl
- kind (for local cluster)

### Setup Local Kubernetes Cluster

1. **Create cluster**:
   ```bash
   kind create cluster --name btcpay-test
   kubectl cluster-info
   ```

2. **Install Crossplane**:
   ```bash
   kubectl create namespace crossplane-system
   helm repo add crossplane-stable https://charts.crossplane.io/stable
   helm repo update
   helm install crossplane crossplane-stable/crossplane \
     --namespace crossplane-system \
     --wait
   ```

3. **Build and load provider image**:
   ```bash
   # Build the provider
   make docker-build
   
   # Load into kind cluster
   kind load docker-image provider-btcpay:v0.1.0-dev --name btcpay-test
   ```

4. **Install provider CRDs**:
   ```bash
   kubectl apply -f package/crds/
   ```

5. **Create credentials secret**:
   ```bash
   kubectl create secret generic btcpay-credentials \
     --from-literal=credentials="YOUR_API_KEY" \
     -n crossplane-system
   ```

6. **Create ProviderConfig**:
   ```bash
   kubectl apply -f examples/provider/config.yaml
   ```

### Test with Example Resources

1. **Create a test store**:
   ```bash
   kubectl apply -f examples/store/basic-store.yaml
   ```

2. **Check store status**:
   ```bash
   kubectl get stores -o wide
   kubectl describe store basic-store
   ```

3. **Create a test invoice**:
   ```bash
   kubectl apply -f examples/invoice/product-purchase.yaml
   ```

4. **Monitor resources**:
   ```bash
   # Watch all BTCPay resources
   kubectl get stores,invoices -w
   
   # Check provider logs
   kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-btcpay -f
   ```

### Run Integration Tests

```bash
# Set environment variables
export BTCPAY_TEST_SERVER="https://your-btcpay-server.example.com"
export BTCPAY_TEST_TIMEOUT="10m"

# Run integration tests
make test.integration
```

## Option 3: Testing Against Public BTCPay Server

If you have access to a public BTCPay Server instance:

1. **Configure for your instance**:
   ```bash
   # Update examples/provider/config.yaml
   vim examples/provider/config.yaml
   # Change baseURL to your BTCPay Server
   ```

2. **Get API credentials**:
   - Access your BTCPay Server admin panel
   - Go to Account Settings → Access Tokens
   - Create token with appropriate permissions
   - Update secret:
   ```bash
   kubectl create secret generic btcpay-credentials \
     --from-literal=credentials="your-actual-api-key" \
     -n crossplane-system \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

## Option 4: Using BTCPay Server Docker for Local Testing

### Run BTCPay Server Locally

```bash
# Run BTCPay Server with Docker
docker run -d \
  --name btcpay-test \
  -p 14142:80 \
  -e BTCPAY_NETWORK=regtest \
  -e BTCPAY_BIND=0.0.0.0:80 \
  btcpayserver/btcpayserver:latest
```

Wait for startup, then:
1. Access http://localhost:14142
2. Complete setup wizard
3. Create admin account
4. Generate API key in Account Settings
5. Update config to use `http://host.docker.internal:14142` (or localhost)

## Troubleshooting

### Common Issues

1. **Provider not starting**:
   ```bash
   kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-btcpay
   ```

2. **CRDs not found**:
   ```bash
   kubectl get crds | grep btcpay
   kubectl apply -f package/crds/
   ```

3. **Authentication errors**:
   ```bash
   kubectl get secrets -n crossplane-system
   kubectl describe secret btcpay-credentials -n crossplane-system
   ```

4. **Store/Invoice stuck in Creating**:
   ```bash
   kubectl describe store your-store-name
   kubectl get events --sort-by=.metadata.creationTimestamp
   ```

### Debug Commands

```bash
# Check all BTCPay resources
kubectl get stores,invoices,providerconfigs.btcpay.crossplane.io -A

# View provider configuration
kubectl get providerconfigs.btcpay.crossplane.io -o yaml

# Monitor provider logs in real-time
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-btcpay -f

# Check Crossplane status
kubectl get providers.pkg.crossplane.io
```

## Cleanup

```bash
# Delete test resources
kubectl delete stores,invoices --all

# Remove provider
kubectl delete -f package/crds/

# Delete cluster
kind delete cluster --name btcpay-test
```

## Next Steps

Once basic testing works:
1. Test cross-resource dependencies (Invoice → Store references)
2. Test error scenarios (invalid configs, network issues)
3. Test updates and deletions
4. Performance testing with multiple resources
5. Integration with CI/CD pipelines