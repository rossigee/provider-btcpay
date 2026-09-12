# Invoice

**API Version**: `invoice.btcpay.m.crossplane.io/v1beta1`

The `Invoice` resource represents a payment invoice in BTCPay Server.

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `forProvider.storeRef` | object | yes | Reference to the Store |
| `forProvider.storeRef.name` | string | yes | Name of the Store |
| `forProvider.storeRef.namespace` | string | no | Namespace of the Store (defaults to Invoice's namespace) |
| `forProvider.amount` | float | yes | Invoice amount |
| `forProvider.currency` | string | yes | Currency code |
| `forProvider.orderID` | string | no | Order ID for tracking |
| `forProvider.itemDesc` | string | no | Item description |
| `forProvider.itemCode` | string | no | Item code/SKU |
| `forProvider.buyerEmail` | string | no | Buyer email address |
| `forProvider.buyerName` | string | no | Buyer name |
| `forProvider.buyerCountry` | string | no | Buyer country |
| `forProvider.buyerZip` | string | no | Buyer ZIP code |
| `forProvider.buyerState` | string | no | Buyer state |
| `forProvider.buyerCity` | string | no | Buyer city |
| `forProvider.buyerAddress1` | string | no | Buyer primary address |
| `forProvider.buyerAddress2` | string | no | Buyer secondary address |
| `forProvider.buyerPhone` | string | no | Buyer phone number |
| `forProvider.notificationURL` | string | no | Webhook notification URL (must be https://) |
| `forProvider.notificationEmail` | string | no | Notification email address |
| `forProvider.redirectURL` | string | no | Redirect URL after payment (must be https://) |
| `forProvider.metadata` | object | no | Custom metadata |
| `forProvider.physical` | bool | no | Physical goods flag |
| `forProvider.taxIncluded` | bool | no | Tax included flag |
| `forProvider.extendedNotifications` | bool | no | Extended notifications |
| `forProvider.fullNotifications` | bool | no | Full notifications |
| `forProvider.checkoutQueryString` | string | no | Checkout query string |
| `forProvider.receipt` | object | no | Receipt settings |
| `forProvider.receipt.enabled` | bool | no | Enable receipt generation |
| `forProvider.receipt.showQR` | bool | no | Show QR code on receipt |
| `forProvider.receipt.showPayments` | bool | no | Show payment details on receipt |

## Example

```yaml
apiVersion: invoice.btcpay.m.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: order-001-invoice
  namespace: production
spec:
  forProvider:
    storeRef:
      name: my-store
      namespace: production
    amount: 100.50
    currency: "USD"
    orderID: "ORDER-001"
    itemDesc: "Monthly Subscription"
    itemCode: "SUB-001"
    buyerEmail: "customer@example.com"
    notificationURL: "https://api.example.com/webhooks/btcpay"
    redirectURL: "https://example.com/payment-success"
    metadata:
      orderId: "ORDER-001"
      customerId: "CUST-123"
  providerConfigRef:
    name: default
```

## Behavior

- **Create**: Creates a new invoice
- **Update**: Updates invoice metadata and settings
- **Delete**: Cancels the invoice

## Status Fields

- `status.atProvider.id` — Unique invoice ID
- `status.atProvider.storeId` — Store ID
- `status.atProvider.amount` — Invoice amount
- `status.atProvider.currency` — Currency code
- `status.atProvider.type` — Invoice type
- `status.atProvider.checkoutLink` — Payment checkout link
- `status.atProvider.status` — Invoice status
- `status.atProvider.additionalStatus` — Additional status information
- `status.atProvider.monitoringExpiration` — Monitoring expiration timestamp
- `status.atProvider.expirationTime` — Invoice expiration timestamp
- `status.atProvider.createdTime` — Creation timestamp
- `status.atProvider.availableStatusesForManualMarking` — Statuses that can be manually set
- `status.atProvider.archived` — Whether invoice is archived
- `status.atProvider.paymentMethods` — Available payment methods with details

## Resource Relationships

Invoices reference Stores using `spec.storeRef`:
```yaml
spec:
  storeRef:
    name: my-store
    namespace: production
```
