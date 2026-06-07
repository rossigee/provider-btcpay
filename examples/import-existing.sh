#!/bin/bash
# Script to import existing BTCPay Server stores into Crossplane
#
# This script demonstrates how to import existing BTCPay stores
# that were created outside of Crossplane into Crossplane management.
#
# Usage:
#   ./import-existing.sh <btcpay-base-url> <api-key> <store-id-1> [<store-id-2> ...]
#
# Example:
#   ./import-existing.sh https://your-btcpay-server.example.com "ak_prod_abc123def456" "store-uuid-1" "store-uuid-2"

set -e

if [ $# -lt 3 ]; then
    echo "Usage: $0 <btcpay-base-url> <api-key> <store-id-1> [<store-id-2> ...]"
    echo ""
    echo "Example:"
    echo "  $0 https://your-btcpay-server.example.com 'ak_prod_abc123' 'abc123'"
    exit 1
fi

BTCPAY_BASE_URL="$1"
API_KEY="$2"
shift 2
STORE_IDS=("$@")

echo "Importing BTCPay stores into Crossplane..."
echo "BTCPay Server: $BTCPAY_BASE_URL"
echo "Stores to import: ${STORE_IDS[*]}"
echo ""

# Verify kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl."
    exit 1
fi

# Create provider config with credentials
echo "Setting up BTCPay credentials..."
kubectl create secret generic btcpay-credentials \
    --from-literal=credentials="$API_KEY" \
    -n crossplane-system \
    --dry-run=client -o yaml | kubectl apply -f -

echo "Creating ProviderConfig..."
cat <<EOF | kubectl apply -f -
apiVersion: btcpay.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: import-config
spec:
  baseURL: "$BTCPAY_BASE_URL"
  credentials:
    source: Secret
    secretRef:
      name: btcpay-credentials
      namespace: crossplane-system
      key: credentials
EOF

echo "Importing stores..."
for store_id in "${STORE_IDS[@]}"; do
    store_name=$(echo "$store_id" | tr '/' '-' | tr '_' '-')

    echo "Importing store: $store_id as $store_name"

    cat <<EOF | kubectl apply -f -
apiVersion: store.btcpay.crossplane.io/v1alpha1
kind: Store
metadata:
  name: $store_name
  annotations:
    crossplane.io/external-name: "$store_id"
spec:
  forProvider:
    name: "Imported Store - $store_name"
    defaultCurrency: "USD"
  providerConfigRef:
    name: import-config
EOF
done

echo ""
echo "Import process completed. Checking store status..."
kubectl get stores -o wide

echo ""
echo "Use 'kubectl describe store <store-name>' to check individual store status."
