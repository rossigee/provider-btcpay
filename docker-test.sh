#!/bin/bash
set -e

echo "🚀 Setting up BTCPay Server + Provider testing environment"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BTCPAY_PORT=14142
PROVIDER_IMAGE="provider-btcpay:test"

echo -e "${YELLOW}Step 1: Building provider image...${NC}"
make docker-build
docker tag provider-btcpay:v0.1.0-dev $PROVIDER_IMAGE

echo -e "${YELLOW}Step 2: Starting BTCPay Server...${NC}"
docker run -d \
  --name btcpay-test \
  -p $BTCPAY_PORT:80 \
  -e BTCPAY_NETWORK=regtest \
  -e BTCPAY_BIND=0.0.0.0:80 \
  -e BTCPAY_EXPLORERURL=http://explorer:3000 \
  btcpayserver/btcpayserver:latest

echo -e "${GREEN}BTCPay Server starting on http://localhost:$BTCPAY_PORT${NC}"
echo "Waiting for BTCPay Server to be ready..."

# Wait for BTCPay to be ready
timeout=60
counter=0
while [ $counter -lt $timeout ]; do
  if curl -s http://localhost:$BTCPAY_PORT/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ BTCPay Server is ready!${NC}"
    break
  fi
  sleep 2
  counter=$((counter + 2))
  echo -n "."
done

if [ $counter -ge $timeout ]; then
  echo -e "${RED}❌ BTCPay Server failed to start within $timeout seconds${NC}"
  exit 1
fi

echo -e "${YELLOW}Step 3: BTCPay Server Setup Instructions${NC}"
echo "1. Open http://localhost:$BTCPAY_PORT in your browser"
echo "2. Complete the setup wizard (create admin account)"
echo "3. Go to Account Settings → Access Tokens"
echo "4. Create a new API token with Store and Invoice permissions"
echo "5. Copy the API key and run:"
echo "   export BTCPAY_API_KEY='your-api-key'"
echo "   export BTCPAY_BASE_URL='http://localhost:$BTCPAY_PORT'"
echo "   go run test_btcpay_client.go"

echo -e "${YELLOW}Step 4: Testing with provider binary${NC}"
echo "To test the provider binary directly:"
echo "1. Create a config file with your API key"
echo "2. Run: ./bin/provider --config config.yaml"

echo -e "${YELLOW}Cleanup commands:${NC}"
echo "docker stop btcpay-test && docker rm btcpay-test"

echo -e "${GREEN}🎉 Test environment ready!${NC}"