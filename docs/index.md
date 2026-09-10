# Provider BTCPay Server Documentation

A Crossplane provider for managing BTCPay Server resources.

## Quick Links

- [Configuration](configuration.md) — Authentication and connection setup
- [Development](development.md) — Building, testing, and contributing

## Resource Documentation

### Payment Resources

| Resource | API Group | Description |
|----------|-----------|-------------|
| [Store](resources/store.md) | `store.btcpay.crossplane.io/v1alpha1` | BTCPay Server store |
| [Invoice](resources/invoice.md) | `invoice.btcpay.crossplane.io/v1alpha1` | Payment invoices |

## Resource Relationships

- **Invoice** references **Store** via `spec.storeRef`
