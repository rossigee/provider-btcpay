# Provider Configuration Guide

This guide walks through all configuration options for the BTCPay Server provider.

## Table of Contents
- [Prerequisites](#prerequisites)
- [API Key Setup](#api-key-setup)
- [Provider Installation](#provider-installation)
- [ProviderConfig Setup](#providerconfig-setup)
- [Self-Hosted BTCPay](#self-hosted-btcpay)
- [Troubleshooting](#troubleshooting)

## Prerequisites

1. A Kubernetes cluster with Crossplane installed
2. A BTCPay Server instance (cloud hosted or self-hosted)
3. A BTCPay API key with appropriate permissions

## API Key Setup

### Obtaining an API Key

1. Log into your BTCPay Server instance
2. Navigate to Account Settings → API keys
3. Click "Create a new API key"
4. Give your key a descriptive name (e.g., "crossplane-provider")
5. Grant permissions for Stores and Invoices management
6. Generate and securely store the API key

### Required Permissions

Your API key needs the following permissions:
- Store management: can view and modify stores
- Invoice management: can view and create invoices

## Provider Installation

### Option 1: Using Crossplane CLI

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-btcpay:latest
```

### Option 2: Using Kubernetes Manifest

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-btcpay
spec:
  package: ghcr.io/rossigee/provider-btcpay:latest
  # Optional: specify a specific version
  # package: ghcr.io/rossigee/provider-btcpay:v0.3.0
```

Apply the manifest:
```bash
kubectl apply -f provider-btcpay.yaml
```

### Verify Installation

Check that the provider is healthy:
```bash
kubectl get providers.pkg.crossplane.io
```

## ProviderConfig Setup

### 1. Create the Credentials Secret

The provider expects credentials in JSON format:

```bash
# Create the secret directly
kubectl create secret generic btcpay-credentials \
  --from-literal=credentials='{"apiKey":"YOUR_BTCPAY_API_KEY"}' \
  -n crossplane-system

# Or create from a file
echo '{"apiKey":"YOUR_BTCPAY_API_KEY"}' > credentials.json
kubectl create secret generic btcpay-credentials \
  --from-file=credentials=credentials.json \
  -n crossplane-system
rm credentials.json  # Clean up
```

### 2. Create the ProviderConfig

Create a file named `provider-config.yaml`:

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

Apply the configuration:
```bash
kubectl apply -f provider-config.yaml
```

### 3. Verify ProviderConfig

```bash
kubectl get providerconfigs.btcpay.crossplane.io
kubectl describe providerconfig.btcpay.crossplane.io default
```

## Self-Hosted BTCPay

If you're using a self-hosted BTCPay instance:

```yaml
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: self-hosted
spec:
  baseURL: https://btcpay.internal.company.com
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
```

### Multiple ProviderConfigs

You can create multiple ProviderConfigs for different BTCPay instances:

```yaml
# Cloud hosted instance
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: btcpay-cloud
spec:
  baseURL: https://mainnet.demo.btcpayserver.org
  credentials:
    source: Secret
    secretRef:
      name: cloud-credentials
      namespace: crossplane-system
      key: credentials
---
# Self-hosted instance
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: btcpay-internal
spec:
  baseURL: https://btcpay.internal.company.com
  credentials:
    source: Secret
    secretRef:
      name: internal-credentials
      namespace: crossplane-system
      key: credentials
```

Then reference the specific config in your resources:

```yaml
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: my-store
spec:
  providerConfigRef:
    name: btcpay-internal  # Use specific config
  forProvider:
    name: "My Store"
    defaultCurrency: "USD"
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   ```
   cannot get credentials: cannot extract credentials
   ```
   - Verify the secret exists and contains valid JSON
   - Check the secret namespace matches the ProviderConfig

2. **API Permission Errors**
   ```
   API request failed with status 403: Forbidden
   ```
   - Ensure your API key has sufficient permissions
   - Verify the API key is not expired or revoked

3. **Connection Errors**
   ```
   failed to send http request
   ```
   - Verify the baseURL is correct and accessible
   - Check network connectivity from the provider pod

### Debug Commands

```bash
# Check provider logs
kubectl logs -n crossplane-system deployment/provider-btcpay-*

# Verify secret contents (be careful with sensitive data)
kubectl get secret btcpay-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d

# Check provider config status
kubectl describe providerconfig.btcpay.crossplane.io default

# List all BTCPay resources
kubectl get stores.store.btcpay.crossplane.io
kubectl get invoices.invoice.btcpay.crossplane.io
```
