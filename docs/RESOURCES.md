# BTCPay Server Provider Resources

This document provides detailed information about all resources supported by the BTCPay Server provider.

## Table of Contents
- [Store Resource](#store-resource)
- [Invoice Resource](#invoice-resource)
- [PaymentMethod Resource](#paymentmethod-resource)
- [User Resource](#user-resource)
- [Webhook Resource](#webhook-resource)
- [Resource Relationships](#resource-relationships)
- [Common Patterns](#common-patterns)

## Store Resource

The `Store` resource represents a BTCPay Server store for managing payments and invoices.

### API Version
- Group: `store.btcpay.crossplane.io`
- Version: `v1beta1`
- Kind: `Store`

### Specification

```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: my-store
spec:
  # Required: The name of the store
  name: "My Store"

  # Optional: Default currency for the store
  # Must be a valid currency code (e.g., "USD", "EUR", "BTC")
  defaultCurrency: "USD"

  # Optional: Website URL associated with the store
  website: "https://example.com"

  # Optional: Invoice payment speed policy
  # Must be one of: "HighSpeed", "MediumSpeed", "LowSpeed", "LowMediumSpeed"
  speedPolicy: "MediumSpeed"

  # Optional: Payment method criteria
  # List of payment method requirements
  paymentMethodCriteria:
    - paymentMethod: "BTC_LightningNetwork"
      value: "0.001"  # Minimum amount in store currency

  # Optional: Network fee settings
  networkFeeMode: "MultiplePaymentsOnly"  # or "Always", "Never"

  # Optional: Lightning settings
  lightningDescriptionTemplate: "Payment for {StoreName}"

  # Reference to the ProviderConfig
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique ID assigned by BTCPay Server
    id: "store-abc123"

    # Store name
    name: "My Store"

    # Default currency
    defaultCurrency: "USD"

    # Website URL
    website: "https://example.com"

    # Speed policy
    speedPolicy: "MediumSpeed"

    # Payment methods available
    paymentMethods: ["BTC", "BTC_LightningNetwork"]

    # Creation timestamp
    createdAt: "2023-01-01T00:00:00Z"

  # Crossplane resource conditions
  conditions:
  - type: Ready
    status: "True"
    reason: Available
    lastTransitionTime: "2023-01-01T00:00:00Z"
```

### Examples

#### Basic Store
```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: company-store
spec:
  name: "Company Store"
  defaultCurrency: "USD"
  website: "https://company.com"
  providerConfigRef:
    name: default
```

#### Store with Lightning Network
```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: lightning-store
spec:
  name: "Lightning Store"
  defaultCurrency: "USD"
  speedPolicy: "HighSpeed"
  paymentMethodCriteria:
    - paymentMethod: "BTC_LightningNetwork"
      value: "0.001"
  lightningDescriptionTemplate: "Payment to {StoreName}"
  providerConfigRef:
    name: default
```

#### Multi-Currency Store
```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: multi-currency-store
spec:
  name: "Multi-Currency Store"
  defaultCurrency: "USD"
  speedPolicy: "MediumSpeed"
  paymentMethodCriteria:
    - paymentMethod: "BTC"
      value: "0.0001"
    - paymentMethod: "BTC_LightningNetwork"
      value: "0.001"
  providerConfigRef:
    name: default
```

### Important Notes

1. **Store Names**: Must be unique within your BTCPay Server instance
2. **Currency Support**: Default currency affects all invoice calculations
3. **Speed Policy**: Affects fee estimation for different payment speeds
4. **Payment Methods**: Criteria define minimum amounts for each payment method

## Invoice Resource

The `Invoice` resource represents a BTCPay Server invoice for payment processing.

### API Version
- Group: `invoice.btcpay.crossplane.io`
- Version: `v1beta1`
- Kind: `Invoice`

### Specification

```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: product-purchase
spec:
  # Store association (required)
  storeRef:
    name: my-store
    namespace: default  # Optional, defaults to same namespace

  # Required: Invoice amount
  amount: "99.99"

  # Required: Currency code
  currency: "USD"

  # Optional: Invoice type
  # Must be one of: "Standard", "TopUp"
  type: "Standard"

  # Optional: Checkout settings
  checkout:
    # Redirect URL after payment
    redirectURL: "https://example.com/thank-you"
    # Automatically redirect after payment
    redirectAutomatically: true
    # Custom checkout expiration (in minutes)
    expirationMinutes: 60

  # Optional: Order information
  orderId: "ORDER-12345"
  itemDesc: "Product Purchase"

  # Optional: Buyer information
  buyer:
    name: "John Doe"
    email: "john@example.com"
    address1: "123 Main St"
    address2: "Apt 4B"
    city: "New York"
    state: "NY"
    zip: "10001"
    country: "US"

  # Provider config reference
  providerConfigRef:
    name: default
```

### Status Fields

```yaml
status:
  atProvider:
    # The unique ID assigned by BTCPay Server
    id: "invoice-xyz789"

    # Invoice status
    # Possible values: New, Processing, Paid, Invalid, Expired
    status: "New"

    # Amount and currency
    amount: "99.99"
    currency: "USD"

    # Checkout URL for payment
    checkoutLink: "https://btcpay.example.com/i/invoice-xyz789"

    # Payment details
    payments: []

    # Creation and expiration timestamps
    createdTime: 1640995200
    expirationTime: 1641002400

    # Monitoring timestamp
    monitoringExpiration: 1641006000

  conditions:
  - type: Ready
    status: "True"
    reason: Available
```

### Examples

#### Basic Invoice
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: product-purchase
spec:
  storeRef:
    name: my-store
  amount: "99.99"
  currency: "USD"
  orderId: "ORDER-12345"
  itemDesc: "Premium Product"
  providerConfigRef:
    name: default
```

#### Invoice with Buyer Information
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: customer-invoice
spec:
  storeRef:
    name: my-store
  amount: "149.99"
  currency: "USD"
  orderId: "INV-2024-001"
  itemDesc: "Annual Subscription"
  buyer:
    name: "Jane Smith"
    email: "jane@example.com"
    address1: "456 Oak Ave"
    city: "San Francisco"
    state: "CA"
    zip: "94105"
    country: "US"
  checkout:
    redirectURL: "https://example.com/thank-you"
    redirectAutomatically: true
  providerConfigRef:
    name: default
```

#### Top-Up Invoice
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: wallet-topup
spec:
  storeRef:
    name: my-store
  amount: "0.01"
  currency: "BTC"
  type: "TopUp"
  itemDesc: "Wallet Top-up"
  providerConfigRef:
    name: default
```

#### Invoice with Custom Expiration
```yaml
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: urgent-payment
spec:
  storeRef:
    name: my-store
  amount: "499.99"
  currency: "USD"
  itemDesc: "Express Service"
  checkout:
    expirationMinutes: 15  # Only 15 minutes to pay
    redirectURL: "https://example.com/payment-complete"
    redirectAutomatically: false
  providerConfigRef:
    name: default
```

### Important Notes

1. **Store Reference**: Invoices must reference an existing Store resource
2. **Amount & Currency**: Amount must be positive, currency must be supported
3. **Invoice Types**: "Standard" for fixed amounts, "TopUp" for variable amounts
4. **Expiration**: Default is 15 minutes, can be customized up to store limits
5. **Order IDs**: Must be unique within a store if provided
6. **Buyer Information**: Optional but recommended for record keeping

## PaymentMethod Resource

The `PaymentMethod` resource configures payment methods for a BTCPay Server store.

### API Version
- Group: `paymentmethod.btcpay.crossplane.io`
- Version: `v1alpha1`
- Kind: `PaymentMethod`

### Specification

```yaml
apiVersion: paymentmethod.btcpay.crossplane.io/v1alpha1
kind: PaymentMethod
metadata:
  name: lightning-payment
spec:
  # Store association (required)
  storeRef:
    name: my-store

  # Payment method identifier
  paymentMethod: "BTC_LightningNetwork"

  # Enable/disable this payment method
  enabled: true

  # Method-specific settings
  settings:
    lightningDescriptionTemplate: "Payment for {StoreName}"
```

## User Resource

The `User` resource manages BTCPay Server user accounts.

### API Version
- Group: `user.btcpay.crossplane.io`
- Version: `v1alpha1`
- Kind: `User`

### Specification

```yaml
apiVersion: user.btcpay.crossplane.io/v1alpha1
kind: User
metadata:
  name: merchant-user
spec:
  # User email (required)
  email: "merchant@example.com"

  # Administrator status
  isAdministrator: false

  # User roles
  roles:
    - StoreOwner
```

## Webhook Resource

The `Webhook` resource configures webhooks for event notifications.

### API Version
- Group: `webhook.btcpay.crossplane.io`
- Version: `v1alpha1`
- Kind: `Webhook`

### Specification

```yaml
apiVersion: webhook.btcpay.crossplane.io/v1alpha1
kind: Webhook
metadata:
  name: invoice-webhook
spec:
  # Store association (required)
  storeRef:
    name: my-store

  # Webhook URL
  url: "https://api.example.com/webhooks/btcpay"

  # Events to listen for
  events:
    - InvoiceCreated
    - InvoicePaymentSettled
    - InvoiceExpired

  # Webhook enabled status
  enabled: true
```

## Resource Relationships

### Store → Invoice Relationship

Invoices depend on Stores. This relationship is established through:

```yaml
storeRef:
  name: my-store
  namespace: default
```

Using store references ensures that:
- The store exists before creating invoices
- Proper dependency management in Kubernetes
- Automatic handling of store lifecycle changes

### Store → PaymentMethod Relationship

PaymentMethods are associated with Stores:

```yaml
storeRef:
  name: my-store
```

### Deletion Behavior

- Deleting a Store does NOT automatically delete its Invoices or PaymentMethods
- Invoices and PaymentMethods must be explicitly deleted before removing a Store
- Use Kubernetes finalizers or owner references for cascading deletion if needed

## Common Patterns

### Complete Store Setup

```yaml
# 1. Create the store
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: ecommerce-store
  labels:
    app: ecommerce
    environment: production
spec:
  name: "E-commerce Store"
  defaultCurrency: "USD"
  website: "https://shop.example.com"
  speedPolicy: "MediumSpeed"
  providerConfigRef:
    name: default
---
# 2. Configure payment methods
apiVersion: paymentmethod.btcpay.crossplane.io/v1alpha1
kind: PaymentMethod
metadata:
  name: lightning-payment
  labels:
    app: ecommerce
spec:
  storeRef:
    name: ecommerce-store
  paymentMethod: "BTC_LightningNetwork"
  enabled: true
  settings:
    lightningDescriptionTemplate: "Payment for {StoreName}"
  providerConfigRef:
    name: default
---
# 3. Set up webhooks for order processing
apiVersion: webhook.btcpay.crossplane.io/v1alpha1
kind: Webhook
metadata:
  name: order-webhook
  labels:
    app: ecommerce
spec:
  storeRef:
    name: ecommerce-store
  url: "https://api.example.com/webhooks/btcpay"
  events:
    - InvoiceCreated
    - InvoicePaymentSettled
    - InvoiceExpired
  enabled: true
  providerConfigRef:
    name: default
```

### Multi-Environment Setup

```yaml
# Development store
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: app-dev-store
  labels:
    environment: development
spec:
  name: "Dev Store"
  defaultCurrency: "USD"
  speedPolicy: "LowSpeed"
  providerConfigRef:
    name: dev-btcpay
---
# Production store
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: app-prod-store
  labels:
    environment: production
spec:
  name: "Production Store"
  defaultCurrency: "USD"
  speedPolicy: "HighSpeed"
  paymentMethodCriteria:
    - paymentMethod: "BTC_LightningNetwork"
      value: "0.001"
  providerConfigRef:
    name: prod-btcpay
---
# Development invoice example
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: dev-test-invoice
  labels:
    environment: development
spec:
  storeRef:
    name: app-dev-store
  amount: "10.00"
  currency: "USD"
  itemDesc: "Development Test"
  providerConfigRef:
    name: dev-btcpay
---
# Production invoice example
apiVersion: invoice.btcpay.crossplane.io/v1beta1
kind: Invoice
metadata:
  name: prod-customer-invoice
  labels:
    environment: production
spec:
  storeRef:
    name: app-prod-store
  amount: "99.99"
  currency: "USD"
  itemDesc: "Premium Service"
  orderId: "PROD-2024-001"
  buyer:
    email: "customer@example.com"
  checkout:
    redirectURL: "https://example.com/thank-you"
  providerConfigRef:
    name: prod-btcpay
```