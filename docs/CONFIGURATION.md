# Provider Configuration Guide

This guide walks through all configuration options for the BTCPay Server provider.

## Table of Contents
- [Prerequisites](#prerequisites)
- [API Key Setup](#api-key-setup)
- [Provider Installation](#provider-installation)
- [ProviderConfig Setup](#providerconfig-setup)
- [Self-Hosted BTCPay Server](#self-hosted-btcpay-server)
- [Troubleshooting](#troubleshooting)

## Prerequisites

1. A Kubernetes cluster with Crossplane installed
2. A BTCPay Server instance with Greenfield API enabled
3. A BTCPay Server API key with appropriate permissions

## API Key Setup

### Obtaining an API Key

1. Log into your BTCPay Server instance
2. Navigate to Account → API Keys
3. Click "Create API Key"
4. Give your key a descriptive name (e.g., "crossplane-provider")
5. Select appropriate permissions for the resources you want to manage

### Required Permissions

Your API key needs the following permissions:
- Store management: btcpay.store.canmodifystoresettings, btcpay.store.cancreateinvoice
- User management: btcpay.user.canmanageusers (if managing users)
- Webhook management: btcpay.store.canmodifystoresettings (if managing webhooks)

## Provider Installation

### Option 1: Using Crossplane CLI

```bash
kubectl crossplane install provider rossigee/provider-btcpay:latest
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
  # package: ghcr.io/rossigee/provider-btcpay:v0.4.0
```

Apply the manifest:
```bash
kubectl apply -f provider-plausible.yaml
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
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
  # For BTCPay Server, specify your instance URL
  baseURL: https://btcpay.example.com
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

## Self-Hosted BTCPay Server

If you're using a self-hosted BTCPay Server instance:

```yaml
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: self-hosted
spec:
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
  # Required for self-hosted instances
  baseURL: https://btcpay.internal.company.com
```

### Multiple ProviderConfigs

You can create multiple ProviderConfigs for different BTCPay Server instances:

```yaml
# Primary instance
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: btcpay-primary
spec:
  credentials:
    source: Secret
    secretRef:
      name: primary-credentials
      namespace: crossplane-system
      key: credentials
  baseURL: https://btcpay.example.com
---
# Secondary instance
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: btcpay-secondary
spec:
  credentials:
    source: Secret
    secretRef:
      name: secondary-credentials
      namespace: crossplane-system
      key: credentials
  baseURL: https://btcpay.secondary.com
```

Then reference the specific config in your resources:

```yaml
apiVersion: store.btcpay.crossplane.io/v1beta1
kind: Store
metadata:
  name: my-store
spec:
  providerConfigRef:
    name: btcpay-primary  # Use specific config
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
   - Ensure your API key has Site Provisioning API access
   - Contact Plausible support if needed

3. **Connection Errors**
   ```
   failed to send http request
   ```
   - For self-hosted: verify the baseURL is correct
   - Check network connectivity from the provider pod

### Debug Commands

```bash
# Check provider logs
kubectl logs -n crossplane-system deployment/provider-btcpay-*

# Verify secret contents (be careful with sensitive data)
kubectl get secret btcpay-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d

# Check provider config status
kubectl describe providerconfig.btcpay.crossplane.io default

# List all BTCPay Server resources
kubectl get stores.store.btcpay.crossplane.io
kubectl get invoices.invoice.btcpay.crossplane.io
kubectl get paymentmethods.paymentmethod.btcpay.crossplane.io
kubectl get users.user.btcpay.crossplane.io
kubectl get webhooks.webhook.btcpay.crossplane.io
```