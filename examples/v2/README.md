# BTCPay Provider v2 Examples

This directory contains examples demonstrating the v2 Crossplane Provider for BTCPay Server with **namespace isolation** and **modern resource management**.

## Key v2 Features

### 🏗️ **Dual-Scope Architecture**
- **v1alpha1**: Cluster-scoped resources (backward compatible)
- **v1beta1**: Namespaced resources with `.m.` API groups (v2 native)

### 🎯 **Namespace Isolation**
- Resources scoped to specific namespaces
- Cross-resource references within the same namespace
- Multi-tenant deployments with proper isolation

### 🔧 **Enhanced Resource Management**
- Management policies for fine-grained control
- Connection secrets in target namespaces
- Improved RBAC and security model

## API Groups

### v2 Namespaced APIs (Recommended)
```yaml
apiVersion: store.btcpay.m.crossplane.io/v1beta1
apiVersion: invoice.btcpay.m.crossplane.io/v1beta1
```

### v1 Cluster-Scoped APIs (Legacy Support)
```yaml
apiVersion: store.btcpay.crossplane.io/v1alpha1
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
```

## Example Files

### 📁 `namespace-setup.yaml`
Complete multi-tenant namespace setup including:
- Dedicated namespaces for production/staging
- RBAC configuration
- Network policies for isolation
- Shared ProviderConfig

### 📁 `store-namespaced.yaml`
v2 Store resource demonstrating:
- Namespace-scoped deployment
- Enhanced configuration options
- Connection secret management
- Management policies

### 📁 `invoice-namespaced.yaml`
v2 Invoice resource showing:
- Cross-resource references within namespace
- Comprehensive invoice configuration
- Customer and item details
- Webhook and notification setup

## Deployment Guide

### 1. Setup Namespaces and RBAC
```bash
kubectl apply -f namespace-setup.yaml
```

### 2. Deploy Store Resources
```bash
kubectl apply -f store-namespaced.yaml
```

### 3. Create Invoices
```bash
kubectl apply -f invoice-namespaced.yaml
```

### 4. Verify Resources
```bash
# List stores in production namespace
kubectl get stores -n btcpay-production

# List invoices in production namespace
kubectl get invoices -n btcpay-production

# Check cross-resource references
kubectl describe invoice customer-order-001 -n btcpay-production
```

## Migration from v1

### Backward Compatibility
- ✅ v1alpha1 resources continue to work unchanged
- ✅ No breaking changes for existing deployments
- ✅ Gradual migration supported

### Migration Strategy
1. **Install v2 Provider**: Deploy provider with dual-scope support
2. **Create Namespaces**: Set up target namespaces
3. **Create v2 Resources**: Deploy new resources using v1beta1 APIs
4. **Validate**: Ensure both v1 and v2 resources work together
5. **Migrate Gradually**: Move resources to v2 APIs over time

### Key Differences
| Feature | v1 (cluster-scoped) | v2 (namespaced) |
|---------|-------------------|-----------------|
| **API Group** | `store.btcpay.crossplane.io` | `store.btcpay.m.crossplane.io` |
| **Scope** | Cluster | Namespace |
| **Cross-References** | Cluster-wide | Within namespace |
| **RBAC** | Cluster-level | Namespace-level |
| **Isolation** | None | Full namespace isolation |

## Security Considerations

### Network Policies
- Isolate namespaces from each other
- Allow egress to BTCPay Server only
- Restrict cross-namespace communication

### RBAC
- Namespace-scoped permissions
- Principle of least privilege
- Separate roles for different environments

### Secret Management
- Connection secrets in target namespaces
- Provider credentials in crossplane-system
- Proper secret rotation policies

## Troubleshooting

### Resource Not Found
```bash
# Check if resource exists in correct namespace
kubectl get stores -A
kubectl get invoices -A
```

### Cross-Reference Issues
```bash
# Verify store exists in same namespace as invoice
kubectl get store production-store -n btcpay-production
kubectl describe invoice customer-order-001 -n btcpay-production
```

### Provider Logs
```bash
# Check provider logs for v2-specific issues
kubectl logs -n crossplane-system deployment/provider-btcpay
```

## Benefits of v2 Architecture

### 🔒 **Enhanced Security**
- Namespace isolation prevents resource conflicts
- Fine-grained RBAC control
- Network policies for traffic isolation

### 🏢 **Multi-Tenancy**
- Multiple teams/environments per cluster
- Isolated resource management
- Separate configuration per namespace

### 📈 **Scalability**
- Better resource organization
- Reduced cluster-wide resource conflicts
- Improved performance with scoped operations

### 🛡️ **Compliance**
- Clear data boundaries
- Audit trails per namespace
- Regulated environment support

## Next Steps

1. **Complete Runtime Migration**: Update to crossplane-runtime v2
2. **Add MRDs**: Implement Managed Resource Definitions
3. **Enhanced Examples**: Add complex multi-resource scenarios
4. **Documentation**: Complete migration guide and best practices