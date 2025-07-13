# BTCPay Server Crossplane Provider - Deployment Guide

## Quick Start Deployment

### Prerequisites
- Kubernetes cluster with Crossplane installed
- BTCPay Server instance with Greenfield API access
- API key with Store and Invoice permissions

### 1. Install CRDs

```bash
# Install all CRDs
kubectl apply -f package/crds/
```

### 2. Create Provider Configuration

```bash
# Create namespace for provider
kubectl create namespace btcpay-system

# Create API credentials secret
kubectl create secret generic btcpay-credentials \
  --from-literal=credentials="YOUR_BTCPAY_API_KEY" \
  -n btcpay-system

# Apply provider config
kubectl apply -f - <<EOF
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  baseURL: "https://your-btcpay-server.com"
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: btcpay-system
      key: credentials
EOF
```

### 3. Run Provider

#### Option A: In-Cluster Deployment
```bash
# Build and load provider image (if you have Docker access)
docker build -t provider-btcpay:latest .
kind load docker-image provider-btcpay:latest  # If using kind

# Create deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: provider-btcpay
  namespace: btcpay-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: provider-btcpay
  template:
    metadata:
      labels:
        app: provider-btcpay
    spec:
      containers:
      - name: provider
        image: provider-btcpay:latest
        args:
        - --debug
        - --sync=1h
        - --poll=1m
        ports:
        - containerPort: 8080
          name: metrics
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          limits:
            memory: "256Mi"
            cpu: "100m"
          requests:
            memory: "128Mi"
            cpu: "50m"
EOF
```

#### Option B: Local Development
```bash
# Run provider locally (requires kubeconfig access)
./bin/linux_amd64/provider --debug
```

### 4. Test with Sample Resources

```bash
# Create a test store
kubectl apply -f examples/store/basic-store.yaml

# Check store status
kubectl get stores -o wide
kubectl describe store basic-store

# Create a test invoice
kubectl apply -f examples/invoice/product-purchase.yaml

# Check invoice status
kubectl get invoices -o wide
kubectl describe invoice product-purchase
```

### 5. Monitor Provider

```bash
# Check provider logs
kubectl logs -n btcpay-system deployment/provider-btcpay -f

# Check resource status
kubectl get stores,invoices,providerconfigs.btcpay.crossplane.io

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

## Troubleshooting

### Common Issues

1. **Provider not starting**
   ```bash
   kubectl logs -n btcpay-system deployment/provider-btcpay
   ```

2. **Authentication errors**
   ```bash
   kubectl describe secret btcpay-credentials -n btcpay-system
   kubectl get providerconfigs.btcpay.crossplane.io -o yaml
   ```

3. **Resources stuck in Creating**
   ```bash
   kubectl describe store <store-name>
   kubectl get events --field-selector involvedObject.name=<store-name>
   ```

4. **BTCPay Server connectivity**
   ```bash
   # Test from within cluster
   kubectl run debug --image=curlimages/curl -it --rm -- \
     curl -H "Authorization: token YOUR_API_KEY" \
     https://your-btcpay-server.com/api/v1/stores
   ```

### Configuration Validation

```bash
# Validate CRDs are installed
kubectl get crds | grep btcpay

# Expected output:
# invoices.invoice.btcpay.crossplane.io
# stores.store.btcpay.crossplane.io  
# providerconfigs.btcpay.crossplane.io
# providerconfigusages.btcpay.crossplane.io

# Check provider configuration
kubectl get providerconfigs.btcpay.crossplane.io -o yaml

# Verify secret exists and has correct key
kubectl get secret btcpay-credentials -n btcpay-system -o yaml
```

## Production Considerations

### Security
- Use dedicated service account with minimal permissions
- Store API keys in external secret management system
- Enable network policies to restrict provider access
- Use TLS for BTCPay Server communication

### Monitoring
- Enable Prometheus metrics on port 8080
- Set up alerts for provider health and resource failures
- Monitor BTCPay Server API rate limits

### High Availability
- Run multiple provider replicas
- Use leader election (enabled by default)
- Implement proper resource limits and health checks

### Backup
- Backup ProviderConfig and secrets
- Document BTCPay Server configuration
- Keep API key rotation procedures

## Integration Examples

See `examples/` directory for:
- Basic store setup
- Lightning-enabled stores  
- Product purchase invoices
- Complete deployment configurations