# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# BTCPay Server Crossplane Provider

## Architecture Overview

This is a **Crossplane managed resource provider** for BTCPay Server integration, following standard Crossplane patterns:

- **Core Resources**: Store and Invoice management with full CRUD operations
- **External Client Pattern**: BTCPay Greenfield API abstraction with interface-based design
- **Cross-Resource References**: Invoices reference Stores using Kubernetes-native `spec.storeRef`
- **Provider Configuration**: Authentication via ProviderConfig with Kubernetes secret references

**Key Directory Structure**:
- `apis/` - CRD definitions (Store, Invoice, User, Webhook, etc.)
- `internal/clients/` - BTCPay Greenfield API client implementation
- `internal/controller/` - Crossplane managed resource controllers
- `examples/` - Complete usage examples and production setups
- `test/e2e/` - Integration tests requiring real BTCPay server

## Development Commands

### Essential Build Commands
```bash
# Code generation (ALWAYS run after API changes)
make generate

# Build and test
make build
make test.unit
make test.integration  # Requires BTCPay server + k8s cluster

# Local development
make run              # Run provider out-of-cluster
make install-crds     # Install CRDs into cluster

# Packaging
make docker-build
make xpkg-build       # Build Crossplane package
```

### Test Commands
```bash
make test.unit           # Unit tests only (fastest)
make test.integration    # Full integration tests
make test.coverage       # Tests with coverage report
make test.all           # All tests
go.test.unit.smart      # Smart unit testing (only packages with tests)
```

## Critical Implementation Patterns

### Standard Crossplane Resource Controller
All controllers follow this pattern:
```go
func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error)
func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error)
func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error)
func (c *external) Delete(ctx context.Context, mg resource.Managed) error
```

### Cross-Resource References
Invoices reference Stores using Crossplane's standard pattern:
```yaml
spec:
  storeRef:
    name: my-store-resource
    namespace: default
```

### Status Conditions
Use Crossplane's standard conditions:
- `xpv1.Available()` - Resource ready
- `xpv1.Creating()` - Resource being created
- `xpv1.Deleting()` - Resource being deleted

### Error Handling
- Always wrap errors with context using `errors.Wrap()`
- Detect 404s to determine resource existence
- Handle BTCPay API rate limits and failures gracefully

## BTCPay Client Usage

The BTCPay client (`internal/clients/btcpay.go`) provides abstracted access to BTCPay Greenfield API:
- **Stores**: CRUD operations, branding, receipt customization
- **Invoices**: Creation, status monitoring, payment tracking
- **Authentication**: API key-based authentication via ProviderConfig

## Testing Requirements

### Unit Tests
- Mock external dependencies (BTCPay client)
- Use table-driven tests for multiple scenarios
- Test error conditions and edge cases
- Located in controller packages

### Integration Tests
- Require real Kubernetes cluster and BTCPay Server instance
- Test complete resource lifecycle
- Use environment variables for configuration
- Located in `test/e2e/`

## API Design Conventions

### Field Validation
Use kubebuilder validation tags extensively:
```go
// +kubebuilder:validation:Required
// +kubebuilder:validation:Enum=BTC;LTC
// +kubebuilder:validation:Pattern="^[a-zA-Z0-9_-]+$"
```

### Optional Fields
Use pointer types for optional fields:
```go
DefaultCurrency    *string `json:"defaultCurrency,omitempty"`
EnableDiscounts    *bool   `json:"enableDiscounts,omitempty"`
```

### Status Reporting
Include observed state in status:
```go
type StoreObservation struct {
    ID              string `json:"id,omitempty"`
    CreatedAt       string `json:"createdAt,omitempty"`
    PaymentMethods  []string `json:"paymentMethods,omitempty"`
}
```

## Current Implementation Status

**✅ Complete**: Store and Invoice resources with full functionality
**⚠️ Partial**: User, Webhook, Guest, SharedLink resources (APIs defined, controllers not implemented)

When implementing new resources:
1. Define CRD in `apis/<resource>/v1alpha1/`
2. Generate with `make generate`
3. Implement controller in `internal/controller/<resource>/`
4. Add BTCPay client methods if needed
5. Create examples in `examples/<resource>/`
6. Add integration tests in `test/e2e/`

## Known Issues

- Some documentation still references "Plausible" instead of "BTCPay" (template artifacts)
- Makefile contains `XPKGS = provider-plausible` instead of `provider-btcpay`
- Additional resource controllers need implementation beyond API definitions