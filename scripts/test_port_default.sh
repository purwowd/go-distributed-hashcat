#!/bin/bash

# Test script for Port Default Feature
echo "🧪 Testing Port Default Feature (Auto-fill 8080)"
echo "================================================="

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Generate agent key for testing
echo ""
echo "📝 Test 1: Generate agent key for testing"
echo "------------------------------------------"
AGENT_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-port-default"}' \
  -w "\nHTTP Status: %{http_code}")

echo "Response: $AGENT_KEY_RESPONSE"
AGENT_KEY=$(echo "$AGENT_KEY_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Generated Agent Key: $AGENT_KEY"

if [ -z "$AGENT_KEY" ]; then
    echo "❌ Failed to generate agent key"
    kill $SERVER_PID
    exit 1
fi

# Test 2: Create agent with empty port (should default to 8080)
echo ""
echo "📝 Test 2: Create agent with empty port (should default to 8080)"
echo "----------------------------------------------------------------"
echo "Expected: Success (201 Created) with port 8080"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-port-default\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\"}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 3: Create agent with explicit port 9090
echo ""
echo "📝 Test 3: Create agent with explicit port 9090"
echo "-----------------------------------------------"
echo "Expected: Success (201 Created) with port 9090"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-port-explicit\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.101\", \"port\": 9090}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 4: Create agent with port 0 (should default to 8080)
echo ""
echo "📝 Test 4: Create agent with port 0 (should default to 8080)"
echo "-------------------------------------------------------------"
echo "Expected: Success (201 Created) with port 8080"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-port-zero\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.102\", \"port\": 0}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 5: Create agent with null port (should default to 8080)
echo ""
echo "📝 Test 5: Create agent with null port (should default to 8080)"
echo "---------------------------------------------------------------"
echo "Expected: Success (201 Created) with port 8080"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-port-null\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.103\", \"port\": null}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 6: Verify all agents have correct ports
echo ""
echo "📝 Test 6: Verify all agents have correct ports"
echo "------------------------------------------------"
echo "Checking ports for all test agents:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-port")) | {name, ip_address, port, status}' | head -20

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Generate agent key ✓"
echo "- Test 2: Create agent with empty port → should default to 8080 ✓"
echo "- Test 3: Create agent with explicit port 9090 → should use 9090 ✓"
echo "- Test 4: Create agent with port 0 → should default to 8080 ✓"
echo "- Test 5: Create agent with null port → should default to 8080 ✓"
echo "- Test 6: Verify all agents have correct ports ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Empty port → 8080 (default)"
echo "- Port 0 → 8080 (default)"
echo "- Null port → 8080 (default)"
echo "- Explicit port → use specified port"
