# provider-btcpay

[![Build Status](https://github.com/crossplane-contrib/provider-btcpay/workflows/CI/badge.svg)](https://github.com/crossplane-contrib/provider-btcpay/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/crossplane-contrib/provider-btcpay)](https://goreportcard.com/report/github.com/crossplane-contrib/provider-btcpay)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A [Crossplane](https://crossplane.io/) provider for managing [BTCPay Server](https://btcpayserver.org/) resources through the BTCPay Greenfield API.

## Overview

This provider enables you to manage BTCPay Server resources using Kubernetes Custom Resources. It supports:

- **Store Management**: Create, update, and delete BTCPay stores
- **Invoice Management**: Create and manage invoices for payment processing
- **Declarative Configuration**: Manage payment infrastructure as code
- **GitOps Workflows**: Integrate BTCPay management with your GitOps pipeline
- **Cross-references**: Reference stores from invoices using Kubernetes native patterns

## Getting Started

### Prerequisites

- Kubernetes cluster with Crossplane installed
- BTCPay Server instance with API access enabled
- BTCPay API key with appropriate permissions

### Installation

1. Install the provider:

```bash
kubectl crossplane install provider ghcr.io/crossplane-contrib/provider-btcpay:v0.1.0
```

2. Create a secret with your BTCPay API credentials:

```bash
kubectl create secret generic btcpay-credentials \
  --from-literal=credentials='{"apiKey":"YOUR_BTCPAY_API_KEY"}' \
  -n crossplane-system
```

3. Configure the provider:

```yaml
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  baseURL: https://your-btcpay-server.com
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
```

## Usage

### Create a Store

```yaml
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: my-store
spec:
  forProvider:
    name: "My Store"
    defaultCurrency: "USD"
    website: "https://mystore.com"
    speedPolicy: "Medium"
  providerConfigRef:
    name: default
```

### Create an Invoice

```yaml
apiVersion: invoice.btcpay.crossplane.io/v1alpha1
kind: Invoice
metadata:
  name: my-invoice
spec:
  forProvider:
    storeRef:
      name: my-store
    amount: 100.50
    currency: "USD"
    orderID: "ORDER-001"
    itemDesc: "Product Purchase"
    buyerEmail: "customer@example.com"
    metadata:
      customerID: "12345"
      productSKU: "PROD-001"
  providerConfigRef:
    name: default
```

## Resource Types

### Store

The Store resource represents a BTCPay store. Key parameters:

- `name` - Store display name
- `defaultCurrency` - Default currency code (e.g., USD, EUR, BTC)
- `website` - Store website URL
- `speedPolicy` - Transaction speed policy (High, Medium, Low)

### Invoice

The Invoice resource represents a BTCPay invoice. Key parameters:

- `storeRef` - Reference to the Store resource
- `amount` - Invoice amount
- `currency` - Currency code
- `orderID` - Optional order identifier
- `itemDesc` - Item description
- `buyerEmail` - Customer email address
- `metadata` - Custom metadata key-value pairs

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Crossplane CLI

### Building

```bash
make build
```

### Testing

Run unit tests:
```bash
make test.unit
```

Run integration tests (requires a running BTCPay server):
```bash
INTEGRATION_TESTS=true make test.integration
```

### Running Locally

For development, you can run the provider locally:

```bash
make run
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

provider-btcpay is under the Apache 2.0 license.