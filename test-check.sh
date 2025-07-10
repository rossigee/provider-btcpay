#!/bin/bash

# Test script to verify the test setup is working correctly

echo "Checking test setup for provider-btcpay..."
echo "========================================="

# Check if all test files exist
echo "Test files found:"
find . -name "*_test.go" -type f | grep -v vendor | grep -v .work | sort

echo ""
echo "Packages with tests:"
go list ./... | xargs -I {} sh -c 'if ls {}/...*_test.go 2>/dev/null | grep -q .; then echo {}; fi' 2>/dev/null | sort

echo ""
echo "Running tests on specific packages:"
echo "- Testing internal/clients..."
go test -v ./internal/clients -count=1 | tail -5

echo ""
echo "- Testing internal/controller/store..."
go test -v ./internal/controller/store -count=1 | tail -5

echo ""
echo "- Testing internal/controller/invoice..."
go test -v ./internal/controller/invoice -count=1 | tail -5

echo ""
echo "Test setup verification complete!"