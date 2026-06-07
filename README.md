# provider-btcpay

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-btcpay/ci.yml?branch=master)][build]
[![Version](https://img.shields.io/github/v/release/rossigee/provider-btcpay)][releases]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[build]: https://github.com/rossigee/provider-btcpay/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-btcpay/releases

A [Crossplane](https://crossplane.io/) provider for managing [BTCPay Server](https://btcpayserver.org/) resources through the BTCPay Greenfield API.

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-btcpay:v0.2.0`

## Overview

This provider enables you to manage BTCPay Server resources using Kubernetes Custom Resources.

## Features

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

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-btcpay:v0.2.0
```

### Configuration

Create a secret with your BTCPay credentials:

```bash
kubectl create secret generic btcpay-credentials \
  --from-literal=credentials='{"apiKey":"YOUR_BTCPAY_API_KEY"}' \
  -n crossplane-system
```

Create the ProviderConfig:

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
  providerConfigRef:
    name: default
```

## Resource Types

| Resource | API Version | Description |
|----------|-------------|-------------|
| Store | `store.btcpay.crossplane.io/v1alpha1` | BTCPay store configuration |
| Invoice | `invoice.btcpay.crossplane.io/v1alpha1` | Payment invoices |

## Development

```bash
# Build the provider
make build

# Run tests
make test

# Lint code
make lint

# Generate CRDs
make generate
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

provider-btcpay is under the Apache 2.0 license.