# BTCPay Server Webhook Examples

This directory contains example Kubernetes manifests for managing BTCPay Server webhooks using the Crossplane BTCPay provider.

## Examples

### invoice-webhook.yaml
Creates a webhook for invoice-related events:
- Listens to key invoice events (created, payment received, settled, expired)
- Includes automatic redelivery for reliability
- Filtered to BTC payments only
- Uses webhook secret for security

### all-events-webhook.yaml
Creates a comprehensive webhook for all invoice events:
- Listens to all available invoice events including processing and invalid states
- Disabled automatic redelivery (manual control)
- No payment method filter (all payment types)
- Useful for comprehensive monitoring

## Usage

1. Ensure you have a BTCPay Store resource created first:
   ```bash
   kubectl apply -f ../store/basic-store.yaml
   ```

2. Apply the webhook manifest:
   ```bash
   kubectl apply -f invoice-webhook.yaml
   ```

3. Check the webhook status:
   ```bash
   kubectl get webhooks
   kubectl describe webhook invoice-webhook
   ```

## Store References

Webhooks must reference an existing BTCPay Store resource using `storeRef`:
```yaml
storeRef:
  name: my-bitcoin-store  # Must match an existing Store resource
```

## Event Types

BTCPay Server supports these webhook events:
- `InvoiceCreated`: New invoice created
- `InvoiceReceivedPayment`: Payment received (partial or full)
- `InvoicePaymentSettled`: Payment confirmed and settled
- `InvoiceExpired`: Invoice expired without payment
- `InvoiceInvalid`: Invoice marked as invalid
- `InvoiceProcessing`: Invoice being processed

## Security Configuration

- **webhook secret**: Used for HMAC signature verification
- **URL**: Must be publicly accessible HTTPS endpoint
- **automatic redelivery**: Retries failed webhook deliveries

## Monitoring

Check webhook delivery status in the BTCPay Server dashboard or via the webhook resource status:
```bash
kubectl get webhook invoice-webhook -o yaml
```

The status will show delivery counts, last delivery time, and any error messages.