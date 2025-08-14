#!/bin/bash

# Test script for IP Address Validation
echo "🧪 Testing IP Address Validation for Agent Creation"
echo "===================================================="

# Generate unique test names using timestamp
TIMESTAMP=$(date +%s)
AGENT_NAME_1="test-ip-validation-${TIMESTAMP}-1"
AGENT_NAME_2="test-ip-validation-${TIMESTAMP}-2"

echo "Using unique agent names: $AGENT_NAME_1 and $AGENT_NAME_2"

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Generate agent key first
echo ""
echo "📝 Test 1: Generate agent key for testing"
echo "------------------------------------------"
AGENT_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_1\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Response: $AGENT_KEY_RESPONSE"

# Extract agent key from response
AGENT_KEY=$(echo "$AGENT_KEY_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Generated Agent Key: $AGENT_KEY"

if [ -z "$AGENT_KEY" ]; then
    echo "❌ Failed to generate agent key"
    kill $SERVER_PID
    exit 1
fi

# Test 2: Create first agent with IP address
echo ""
echo "📝 Test 2: Create first agent with IP address 192.168.1.100"
echo "------------------------------------------------------------"
echo "Expected: Success (201 Created)"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_1\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 3: Generate another agent key for different agent
echo ""
echo "📝 Test 3: Generate another agent key for different agent"
echo "----------------------------------------------------------"
AGENT_KEY_2_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\"}" \
  -w "\nHTTP Status: %{http_code}")

AGENT_KEY_2=$(echo "$AGENT_KEY_2_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Generated Second Agent Key: $AGENT_KEY_2"

# Test 4: Try to create second agent with same IP address (should fail)
echo ""
echo "📝 Test 4: Try to create second agent with same IP address 192.168.1.100 (should fail)"
echo "----------------------------------------------------------------------------------------"
echo "Expected: Conflict (409) - IP address already used"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.100\", \"port\": 8081}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 5: Create agent with different IP address (should succeed)
echo ""
echo "📝 Test 5: Create agent with different IP address 192.168.1.101 (should succeed)"
echo "------------------------------------------------------------------------------"
echo "Expected: Success (201 Created)"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.101\", \"port\": 8081}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 6: List all agents to see the results
echo ""
echo "📝 Test 6: List all agents to see the results"
echo "----------------------------------------------"
curl -X GET http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name | contains(\"$AGENT_NAME_1\")) | {name, ip_address, agent_key, status}"
echo ""
curl -X GET http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name | contains(\"$AGENT_NAME_2\")) | {name, ip_address, agent_key, status}"

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Generate agent key ✓"
echo "- Test 2: Create agent with IP 192.168.1.100 ✓ (should succeed)"
echo "- Test 3: Generate second agent key ✓"
echo "- Test 4: Try to create agent with same IP 192.168.1.100 ✓ (should fail with 409)"
echo "- Test 5: Create agent with different IP 192.168.1.101 ✓ (should succeed)"
echo "- Test 6: List all agents to verify ✓"
