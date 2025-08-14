#!/bin/bash

# Simple test script for Port Default Feature
echo "🧪 Testing Port Default Feature (Simple Test)"
echo "=============================================="

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Get existing agent key from database
echo ""
echo "📝 Test 1: Get existing agent key from database"
echo "-----------------------------------------------"
AGENT_KEY_RESPONSE=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[0].agent_key' | tr -d '"')
echo "Using existing Agent Key: $AGENT_KEY_RESPONSE"

if [ -z "$AGENT_KEY_RESPONSE" ] || [ "$AGENT_KEY_RESPONSE" = "null" ]; then
    echo "❌ No existing agent key found"
    kill $SERVER_PID
    exit 1
fi

# Test 2: Create agent with empty port (should default to 8080)
echo ""
echo "📝 Test 2: Create agent with empty port (should default to 8080)"
echo "----------------------------------------------------------------"
echo "Expected: Success with port 8080"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-port-simple-$(date +%s)\", \"agent_key\": \"$AGENT_KEY_RESPONSE\", \"ip_address\": \"192.168.1.200\"}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 3: Verify agent has port 8080
echo ""
echo "📝 Test 3: Verify agent has port 8080"
echo "--------------------------------------"
echo "Checking if agent has port 8080:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-port-simple")) | {name, ip_address, port, status}'

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "🎯 Expected Result:"
echo "Agent should be created with port 8080 when port is not specified"
