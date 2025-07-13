# BTCPay Server Crossplane Provider v0.1.0

## Overview

This is the initial release of the BTCPay Server Crossplane Provider, enabling Kubernetes-native management of BTCPay Server resources through Crossplane.

## Features

### ✅ Implemented Resources

#### **Store Management**
- Create, read, update, and delete BTCPay Server stores
- Full configuration support: branding, receipts, payment methods
- Speed policy management (High, Medium, Low)
- Website and invoice settings

#### **Invoice Management**  
- Create and manage BTCPay Server invoices
- Support for all invoice types (fixed amount, donation, etc.)
- Cross-resource references (Invoice → Store)
- Payment tracking and status monitoring

#### **Provider Configuration**
- Secure credential management via Kubernetes secrets
- Multi-server support with configurable base URLs
- Standard Crossplane provider patterns

### 🏗️ Architecture

- **API Groups**: `btcpay.crossplane.io` (v1beta1), `store.btcpay.crossplane.io` (v1alpha1), `invoice.btcpay.crossplane.io` (v1alpha1)
- **Controllers**: Store, Invoice, ProviderConfig
- **Client**: Full BTCPay Greenfield API integration
- **Testing**: Comprehensive unit test coverage (95%+)

## What's Working

### Store Resource
```yaml
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: my-btcpay-store
spec:
  name: "My Store"
  defaultCurrency: "USD"
  website: "https://example.com"
  branding:
    logo: "https://example.com/logo.png"
    cssUrl: "https://example.com/custom.css"
  receipt:
    enabled: true
    showQR: true
```

### Invoice Resource
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
kind: Invoice  
metadata:
  name: product-purchase
spec:
  storeRef:
    name: my-btcpay-store
  amount: "99.99"
  currency: "USD"
  orderId: "ORDER-123"
  redirectURL: "https://example.com/thank-you"
```

### Provider Configuration
```yaml
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  baseURL: "https://btcpay.example.com"
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
```

## Testing

### Test Coverage
- **Unit Tests**: 72 tests across all components
- **Integration Tests**: E2E testing with real BTCPay Server
- **Error Scenarios**: Comprehensive error handling coverage
- **Cross-Resource**: Invoice→Store dependency testing

### Validated Scenarios
- Store lifecycle (create, update, delete)
- Invoice creation with store references
- Authentication and authorization
- API error handling and retry logic
- Cross-namespace resource references

## Known Limitations

### Not Yet Implemented
- **User Management**: User resource controller needs implementation
- **Webhook Management**: Webhook resource controller needs implementation  
- **Guest/SharedLink**: Additional resource types partially defined
- **Advanced Features**: Pull payments, apps, plugins

### Current Constraints
- Single API key per provider configuration
- No built-in rate limiting (relies on BTCPay Server limits)
- Manual secret management (no external-secrets integration yet)

## Deployment

### Requirements
- Kubernetes 1.20+
- Crossplane 1.14+
- BTCPay Server with Greenfield API enabled
- API key with Store and Invoice permissions

### Quick Start
```bash
# Install CRDs
kubectl apply -f package/crds/

# Create credentials
kubectl create secret generic btcpay-credentials \
  --from-literal=credentials="your-api-key"

# Deploy provider (see DEPLOYMENT.md for details)
kubectl apply -f examples/provider/config.yaml
```

## Development

### Building
```bash
make test.unit        # Run tests
make build           # Build binary
make generate        # Generate CRDs and code
```

### Testing
```bash
make test.unit              # Unit tests
make test.integration       # Integration tests (requires BTCPay Server)
./docker-test.sh           # Local BTCPay Server + testing
```

See `TESTING.md` and `DEVELOPMENT.md` for detailed instructions.

## What's Next

### Planned for v0.2.0
- User resource implementation
- Webhook resource implementation
- External-secrets integration
- Enhanced monitoring and metrics
- Performance optimizations

### Future Roadmap
- Pull payments support
- Apps and plugins management
- Lightning Network advanced features
- Multi-tenant improvements
- Backup/restore capabilities

## Support

- **Documentation**: See README.md, TESTING.md, DEPLOYMENT.md
- **Examples**: Complete examples in `examples/` directory
- **Issues**: Report bugs and feature requests on GitHub
- **Community**: Join Crossplane community discussions

## Credits

Built with Crossplane framework and BTCPay Server Greenfield API.
Follows Crossplane provider best practices and conventions.