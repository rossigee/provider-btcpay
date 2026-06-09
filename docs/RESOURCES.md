# BTCPay Server Provider Resources

This document provides detailed information about all resources supported by the BTCPay Server provider.

## Table of Contents
- [Store Resource](#store-resource)
- [Invoice Resource](#invoice-resource)
- [Resource Relationships](#resource-relationships)
- [Common Patterns](#common-patterns)

## Store Resource

The `Store` resource represents a BTCPay Server store for managing payment processing.

### API Version
- Group: `store.btcpay.crossplane.io`
- Version: `v1alpha1`
- Kind: `Store`

### Specification

```yaml
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: my-store
spec:
  forProvider:
    # Required: The display name for the store
    name: "My Online Store"
    
    # Required: Default currency for the store
    # Must be a valid currency code (USD, EUR, BTC, etc.)
    defaultCurrency: "USD"
    
    # Optional: Website URL associated with the store
    website: "https://mystore.example.com"
    
    # Optional: Invoice expiration time in seconds (60-86400)
    invoiceExpiration: 900
    
    # Optional: Monitoring expiration time in seconds
    monitoringExpiration: 3600
    
    # Optional: Payment tolerance percentage (0-100)
    paymentTolerance: 0.5
    
    # Optional: Speed policy for confirmation
    # Valid values: High, Medium, Low
    speedPolicy: "Medium"
    
    # Optional: Network fee mode
    # Valid values: Never, Always, MultiplePaymentsOnly
    networkFeeMode: "Never"
    
    # Optional: Other configuration options
    lightningAmountInSatoshi: true
    lightningPrivateRouteHints: true
    onChainWithLnInvoiceFallback: true
    redirectAutomatically: false
    showRecommendedFee: true
    recommendedFeeBlockTarget: 60
    defaultLang: "en"
  
  # Reference to the ProviderConfig
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique store ID assigned by BTCPay
    id: "store-abc123def456"
    
    # Confirmed store name
    name: "My Online Store"
    
    # Store website
    website: "https://mystore.example.com"
    
    # Confirmed default currency
    defaultCurrency: "USD"
    
    # Invoice expiration setting
    invoiceExpiration: 900
    
    # Available payment methods
    paymentMethods:
      - "BTC"
      - "LTC"
      - "LightningNetwork"
    
    # Creation timestamp
    createdAt: "2026-06-09T12:00:00Z"
    
    # Derivation schemes for payment methods
    derivationSchemes:
      BTC: "xpub..."
      LTC: "xpub..."
  
  conditions:
    - type: Ready
      status: "True"
      reason: Available
      message: "Store is ready for use"
    - type: Synced
      status: "True"
      reason: ReconciliationSucceeded
      message: "Latest observed state matches desired state"
```

## Invoice Resource

The `Invoice` resource represents a payment invoice in BTCPay Server.

### API Version
- Group: `invoice.btcpay.crossplane.io`
- Version: `v1alpha1`
- Kind: `Invoice`

### Specification

```yaml
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
kind: Invoice
metadata:
  name: order-001-invoice
spec:
  forProvider:
    # Required: Reference to the store
    storeRef:
      name: my-store
      namespace: default
    
    # Required: Invoice amount
    amount: 100.50
    
    # Required: Currency code
    currency: "USD"
    
    # Optional: Order ID for tracking
    orderID: "ORDER-001"
    
    # Optional: Item description
    itemDesc: "Monthly Subscription"
    
    # Optional: Item code/SKU
    itemCode: "SUB-001"
    
    # Optional: Buyer email address
    buyerEmail: "customer@example.com"
    
    # Optional: Notification URL for webhooks
    notificationURL: "https://api.example.com/webhooks/btcpay"
    
    # Optional: Redirect URL after payment
    redirectURL: "https://example.com/payment-success"
    
    # Optional: Custom metadata
    metadata:
      orderId: "ORDER-001"
      customerId: "CUST-123"
      environment: "production"
    
    # Optional: Other configuration
    physical: false
    taxIncluded: false
    extendedNotifications: false
    fullNotifications: false
    checkoutQueryString: "utm_source=website"
  
  # Reference to the ProviderConfig
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique invoice ID assigned by BTCPay
    id: "invoice-xyz789"
    
    # Store ID
    storeId: "store-abc123def456"
    
    # Invoice amount in the specified currency
    amount: 100.50
    
    # Currency code
    currency: "USD"
    
    # Invoice type
    type: "standard"
    
    # Checkout link for payment
    checkoutLink: "https://btcpay.example.com/i/invoice-xyz789"
    
    # Invoice status
    # Values: new, processing, expired, paid, confirmed, failed, archived
    status: "new"
    
    # Additional status information
    additionalStatus: "paid"
    
    # Whether the invoice is archived
    archived: false
    
    # Payment methods available
    paymentMethods:
      - paymentMethod: "BTC"
        cryptoCode: "BTC"
        destination: "bc1q..."
        rate: "26500.00"
        paid: false
        totalPaid: 0
        due: 0.00378612
    
    # Timestamps
    createdTime: "2026-06-09T12:00:00Z"
    expirationTime: "2026-06-09T12:15:00Z"
    monitoringExpiration: "2026-06-09T13:00:00Z"
  
  conditions:
    - type: Ready
      status: "True"
      reason: Available
      message: "Invoice is ready for payment processing"
    - type: Synced
      status: "True"
      reason: ReconciliationSucceeded
      message: "Latest observed state matches desired state"
```

## Resource Relationships

### Invoice References Store

Invoices use a cross-reference pattern to link to stores:

```yaml
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
kind: Invoice
metadata:
  name: order-002-invoice
spec:
  forProvider:
    storeRef:
      name: my-store  # References a Store in the same namespace
      namespace: default
    amount: 50.00
    currency: "USD"
  providerConfigRef:
    name: default
```

## Common Patterns

### Creating Resources in Different Namespaces

```yaml
---
# Store in default namespace
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: store-prod
  namespace: default
spec:
  forProvider:
    name: "Production Store"
    defaultCurrency: "USD"
  providerConfigRef:
    name: default
---
# Invoice in payment namespace referencing store in default
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
kind: Invoice
metadata:
  name: invoice-001
  namespace: payment
spec:
  forProvider:
    storeRef:
      name: store-prod
      namespace: default  # Cross-namespace reference
    amount: 100.00
    currency: "USD"
  providerConfigRef:
    name: default
```

### Multiple Stores Configuration

```yaml
---
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: store-usd
spec:
  forProvider:
    name: "USD Store"
    defaultCurrency: "USD"
  providerConfigRef:
    name: default
---
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: store-eur
spec:
  forProvider:
    name: "EUR Store"
    defaultCurrency: "EUR"
  providerConfigRef:
    name: default
```

### GitOps Integration

For GitOps workflows, stores and invoices can be version-controlled:

```bash
# Create directory structure
mkdir -p btcpay/stores btcpay/invoices

# Store configurations
cat > btcpay/stores/prod.yaml << EOF
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: prod-store
spec:
  forProvider:
    name: "Production Store"
    defaultCurrency: "USD"
    website: "https://prod.example.com"
  providerConfigRef:
    name: default
EOF

# Sync with Flux/ArgoCD
git add btcpay/
git commit -m "Add BTCPay production store configuration"
git push
```
