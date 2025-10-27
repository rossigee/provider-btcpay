# BTCPay Server Crossplane Provider v0.4.0

## Overview

This release significantly expands the BTCPay Server Crossplane Provider with new resources, API stability improvements, and enhanced functionality for comprehensive BTCPay Server management.

## What's New in v0.4.0

### ✅ New Implemented Resources

#### **PaymentMethod Resource**
- Full CRUD operations for store payment method configurations
- Support for Lightning Network, On-Chain BTC, and altcoins
- Payment method settings and preferences management

#### **User Resource**
- User account management with role-based access control
- Create, update, and manage BTCPay Server users
- Permission assignment and user lifecycle management

#### **Webhook Resource**
- Webhook configuration for real-time event notifications
- Support for invoice events, store events, and payment notifications
- Automatic webhook URL management and secret handling

### 🔄 API Stability Improvements

#### **v1beta1 API Versions**
- **Store v1beta1**: Promoted from v1alpha1 with backward compatibility
- **Invoice v1beta1**: Enhanced invoice management with additional fields
- **ProviderConfig v1beta1**: Stable provider configuration API

#### **Namespace-scoped Resources**
- Support for namespaced deployments
- Cross-namespace resource references
- Improved multi-tenant isolation

### 🛠️ Technical Enhancements

- **Controller Refactoring**: Unified controller architecture supporting multiple API versions
- **Build System Updates**: Improved Makefile and CI/CD reliability
- **Documentation Corrections**: Fixed all "Plausible" references to "BTCPay"
- **Generated Code Updates**: Consistent managed resource generation

## Complete Resource Support

### Core Resources (✅ Fully Implemented)

#### **Store Management**
```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: my-btcpay-store
spec:
  name: "My Store"
  defaultCurrency: "USD"
  website: "https://example.com"
  speedPolicy: "MediumSpeed"
  paymentMethodCriteria:
    - paymentMethod: "BTC_LightningNetwork"
      value: "0.001"
```

#### **Invoice Management**
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: product-purchase
spec:
  storeRef:
    name: my-btcpay-store
  amount: "99.99"
  currency: "USD"
  type: "Standard"
  checkout:
    redirectURL: "https://example.com/thank-you"
    redirectAutomatically: true
```

#### **PaymentMethod Management**
```yaml
apiVersion: paymentmethod.btcpay.crossplane.io/v1alpha1
kind: PaymentMethod
metadata:
  name: lightning-payment
spec:
  storeRef:
    name: my-btcpay-store
  paymentMethod: "BTC_LightningNetwork"
  enabled: true
  settings:
    lightningDescriptionTemplate: "Payment for {StoreName}"
```

#### **User Management**
```yaml
apiVersion: user.btcpay.crossplane.io/v1alpha1
kind: User
metadata:
  name: merchant-user
spec:
  email: "merchant@example.com"
  isAdministrator: false
  roles:
    - StoreOwner
```

#### **Webhook Management**
```yaml
apiVersion: webhook.btcpay.crossplane.io/v1alpha1
kind: Webhook
metadata:
  name: invoice-webhook
spec:
  storeRef:
    name: my-btcpay-store
  url: "https://api.example.com/webhooks/btcpay"
  events:
    - InvoiceCreated
    - InvoicePaymentSettled
  enabled: true
```

## Testing & Quality

### Enhanced Test Coverage
- **Unit Tests**: 150+ tests across all controllers
- **Integration Tests**: E2E validation for all resources
- **Error Scenarios**: Comprehensive failure mode testing
- **Cross-Resource Testing**: Dependency and reference validation

### Validated Scenarios
- Complete resource lifecycles (CRUD operations)
- Cross-resource dependencies and references
- Authentication and authorization flows
- Rate limiting and retry logic
- Multi-namespace deployments

## Migration Guide

### From v0.2.x to v0.4.0

#### API Version Updates
- `store.btcpay.crossplane.io/v1alpha1` → `store.btcpay.crossplane.io/v1beta1`
- `invoice.btcpay.crossplane.io/v1alpha1` → `invoice.btcpay.crossplane.io/v1beta1`
- Backward compatibility maintained for existing resources

#### New Resources
- Add PaymentMethod, User, and Webhook resources as needed
- Update ProviderConfig if using new authentication features

## Deployment

### Requirements
- Kubernetes 1.20+
- Crossplane 1.14+
- BTCPay Server with Greenfield API enabled
- API key with appropriate permissions for managed resources

### Quick Start
```bash
# Install updated CRDs
kubectl apply -f package/crds/

# Update existing credentials if needed
kubectl apply -f examples/provider/secret.yaml.tmpl

# Deploy provider
kubectl apply -f examples/provider/config.yaml

# Create resources using v1beta1 APIs
kubectl apply -f examples/store/basic-store.yaml
kubectl apply -f examples/invoice/product-purchase.yaml
```

## Development

### Building & Testing
```bash
make generate        # Generate updated CRDs and code
make test.unit       # Run comprehensive unit tests
make test.integration # Run integration tests
make build          # Build provider binary
make docker-build   # Build Docker image
```

## Known Limitations

### Future Enhancements
- **Pull Payments**: Advanced payment flow support
- **Apps & Plugins**: BTCPay Server extension management
- **Lightning Features**: Advanced Lightning Network operations
- **External Secrets**: Automated secret management integration

## What's Next

### Planned for v0.5.0
- Pull payments resource implementation
- Apps and plugins management
- Enhanced monitoring and metrics
- Performance optimizations
- Backup/restore capabilities

## Support

- **Documentation**: Updated README.md, RESOURCES.md, DEVELOPMENT.md
- **Examples**: Complete examples for all resources in `examples/` directory
- **Issues**: Report bugs and feature requests on GitHub
- **Community**: Join Crossplane and BTCPay Server communities

## Credits

Built with Crossplane framework and BTCPay Server Greenfield API.
Special thanks to contributors and the open-source community.