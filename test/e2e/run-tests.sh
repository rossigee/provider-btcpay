#!/bin/bash
# Script to run BTCPay provider integration tests

# Exit on error
set -e

echo "BTCPay Provider Integration Tests"
echo "================================="
echo ""
echo "Prerequisites:"
echo "- Kubernetes cluster running"
echo "- Crossplane installed"
echo "- BTCPay provider installed with CRDs"
echo "- BTCPay server accessible"
echo "- ProviderConfig created"
echo ""

# Check if integration tests should run
if [ "$INTEGRATION_TESTS" != "true" ] && [ "$RUN_INTEGRATION_TESTS" != "true" ]; then
    echo "Skipping integration tests. Set INTEGRATION_TESTS=true to run."
    exit 0
fi

echo "Running integration tests..."
echo ""

# Change to the provider directory
cd "$(dirname "$0")/../.."

# Run the integration tests
go test -tags=integration -v ./test/e2e/... "$@"