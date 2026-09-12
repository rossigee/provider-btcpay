# Provider BTCPay Server Documentation

A Crossplane provider for managing BTCPay Server resources.

## Quick Links

- [Configuration](configuration.md) — Authentication and connection setup
- [Development](development.md) — Building, testing, and contributing

## Resource Documentation

### Payment Resources

| Resource | API Group | Description |
|----------|-----------|-------------|
| [Store](resources/store.md) | `store.btcpay.m.crossplane.io/v1beta1` | BTCPay Server store |
| [Invoice](resources/invoice.md) | `invoice.btcpay.m.crossplane.io/v1beta1` | Payment invoices |
| ApiKey | `apikey.btcpay.m.crossplane.io/v1beta1` | API keys |
| Guest | `guest.btcpay.m.crossplane.io/v1beta1` | Guest checkout access |
| SharedLink | `sharedlink.btcpay.m.crossplane.io/v1beta1` | Shared payment links |
| User | `user.btcpay.m.crossplane.io/v1beta1` | Server users |
| Webhook | `webhook.btcpay.m.crossplane.io/v1beta1` | Event webhooks |
| ProviderConfig | `btcpay.m.crossplane.io/v1beta1` | Provider credentials (cluster-scoped) |

## Resource Relationships

- **Invoice** references **Store** via `spec.storeRef`

## API Coverage Gaps

BTCPay Greenfield API surface not yet modeled: pull payments/payouts, refunds, lightning-network node info and address management, payment-request templates, store payment-method configuration (on-chain/lightning setup), notifications, and server version/info endpoints.
