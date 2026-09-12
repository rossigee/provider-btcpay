# Store

**API Version**: `store.btcpay.m.crossplane.io/v1beta1`

The `Store` resource represents a BTCPay Server store for managing payment processing.

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `forProvider.name` | string | yes | Display name for the store |
| `forProvider.defaultCurrency` | string | yes | Default currency (USD, EUR, BTC, etc.) |
| `forProvider.website` | string | no | Website URL (must start with https://) |
| `forProvider.invoiceExpiration` | int | no | Invoice expiration in seconds (60-86400) |
| `forProvider.monitoringExpiration` | int | no | Monitoring expiration in seconds (600-86400) |
| `forProvider.paymentTolerance` | float | no | Payment tolerance percentage (0-100) |
| `forProvider.speedPolicy` | string | no | Speed policy (High, Medium, Low) |
| `forProvider.networkFeeMode` | string | no | Network fee mode (Never, Always, MultiplePaymentsOnly) |
| `forProvider.lightningAmountInSatoshi` | bool | no | Lightning amount in satoshi |
| `forProvider.lightningPrivateRouteHints` | bool | no | Lightning private route hints |
| `forProvider.onChainWithLnInvoiceFallback` | bool | no | On-chain with LN invoice fallback |
| `forProvider.redirectAutomatically` | bool | no | Redirect automatically |
| `forProvider.showRecommendedFee` | bool | no | Show recommended fee |
| `forProvider.recommendedFeeBlockTarget` | int | no | Recommended fee block target (1-1008) |
| `forProvider.defaultLang` | string | no | Default language |
| `forProvider.checkoutType` | string | no | Checkout type (V1, V2) |
| `forProvider.receipt` | object | no | Receipt settings |
| `forProvider.receipt.enabled` | bool | no | Enable receipt generation |
| `forProvider.receipt.showQR` | bool | no | Show QR code on receipts |
| `forProvider.receipt.showPayments` | bool | no | Show payment details on receipts |
| `forProvider.branding` | object | no | Branding settings |
| `forProvider.branding.logoUrl` | string | no | URL to store logo |
| `forProvider.branding.css` | string | no | Custom CSS |
| `forProvider.branding.htmlTitle` | string | no | HTML title for store pages |

## Example

```yaml
apiVersion: store.btcpay.m.crossplane.io/v1beta1
kind: Store
metadata:
  name: my-store
  namespace: production
spec:
  forProvider:
    name: "My Online Store"
    defaultCurrency: "USD"
    website: "https://mystore.example.com"
    invoiceExpiration: 900
    monitoringExpiration: 3600
    speedPolicy: "Medium"
    networkFeeMode: "Never"
  providerConfigRef:
    name: default
```

## Behavior

- **Create**: Creates a new BTCPay Server store
- **Update**: Updates store configuration
- **Delete**: Deletes the store

## Status Fields

- `status.atProvider.id` — Unique store ID
- `status.atProvider.name` — Confirmed store name
- `status.atProvider.website` — Store website
- `status.atProvider.defaultCurrency` — Default currency
- `status.atProvider.invoiceExpiration` — Current invoice expiration time
- `status.atProvider.paymentMethods` — Available payment methods
- `status.atProvider.createdAt` — Creation timestamp
- `status.atProvider.derivationSchemes` — Derivation schemes
