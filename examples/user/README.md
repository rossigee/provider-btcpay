# BTCPay Server User Examples

This directory contains example Kubernetes manifests for managing BTCPay Server users using the Crossplane BTCPay provider.

## Examples

### basic-user.yaml
Creates an administrator user with full permissions:
- Email: `admin@example.com`
- Administrator privileges: `true`
- Server admin role
- Email confirmation disabled for quick setup

### regular-user.yaml
Creates a regular store owner user:
- Email: `storeowner@example.com`
- Store owner role (no server admin)
- Email confirmation required
- Limited to store management permissions

## Usage

1. First, ensure you have a BTCPay ProviderConfig configured with valid API credentials
2. Apply the user manifest:
   ```bash
   kubectl apply -f basic-user.yaml
   ```
3. Check the user status:
   ```bash
   kubectl get users
   kubectl describe user btcpay-admin-user
   ```

## Security Notes

- Change the default passwords in these examples before using in production
- Use Kubernetes secrets for sensitive information like passwords
- Consider using BTCPay Server's built-in user invite functionality for better security
- Administrator users have full server access - use sparingly

## BTCPay User Roles

- `ServerAdmin`: Full server administration permissions
- `StoreOwner`: Can manage assigned stores only
- `Guest`: Read-only access to assigned stores
- Custom roles can be defined in BTCPay Server configuration