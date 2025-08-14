#!/bin/bash

# Manual test script for IP Address Validation
echo "🧪 Manual Testing IP Address Validation"
echo "========================================"

# Generate unique test names using timestamp and random number
TIMESTAMP=$(date +%s)
RANDOM_NUM=$((RANDOM % 10000))
AGENT_NAME_1="test-ip-manual-${TIMESTAMP}-${RANDOM_NUM}-1"
AGENT_NAME_2="test-ip-manual-${TIMESTAMP}-${RANDOM_NUM}-2"

echo "Using unique agent names: $AGENT_NAME_1 and $AGENT_NAME_2"

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Generate first agent key
echo ""
echo "📝 Test 1: Generate first agent key"
echo "-----------------------------------"
AGENT_KEY_1_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_1\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Response: $AGENT_KEY_1_RESPONSE"
AGENT_KEY_1=$(echo "$AGENT_KEY_1_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Agent Key 1: $AGENT_KEY_1"

if [ -z "$AGENT_KEY_1" ]; then
    echo "❌ Failed to generate agent key 1"
    kill $SERVER_PID
    exit 1
fi

# Test 2: Generate second agent key
echo ""
echo "📝 Test 2: Generate second agent key"
echo "------------------------------------"
AGENT_KEY_2_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Response: $AGENT_KEY_2_RESPONSE"
AGENT_KEY_2=$(echo "$AGENT_KEY_2_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Agent Key 2: $AGENT_KEY_2"

if [ -z "$AGENT_KEY_2" ]; then
    echo "❌ Failed to generate agent key 2"
    kill $SERVER_PID
    exit 1
fi

# Test 3: Create first agent with IP 192.168.1.200
echo ""
echo "📝 Test 3: Create first agent with IP 192.168.1.200"
echo "----------------------------------------------------"
echo "Expected: Success (201 Created)"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_1\", \"agent_key\": \"$AGENT_KEY_1\", \"ip_address\": \"192.168.1.200\", \"port\": 8080}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 4: Try to create second agent with same IP (should fail)
echo ""
echo "📝 Test 4: Try to create second agent with same IP 192.168.1.200 (should fail)"
echo "---------------------------------------------------------------------------"
echo "Expected: Conflict (409) - IP address already used"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.200\", \"port\": 8081}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 5: Create second agent with different IP (should succeed)
echo ""
echo "📝 Test 5: Create second agent with different IP 192.168.1.201 (should succeed)"
echo "---------------------------------------------------------------------------"
echo "Expected: Success (201 Created)"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME_2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.201\", \"port\": 8081}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 6: Verify final state
echo ""
echo "📝 Test 6: Verify final state - List agents with our test names"
echo "----------------------------------------------------------------"
curl -X GET http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name | contains(\"$AGENT_NAME_1\")) | {name, ip_address, agent_key, status}"
echo ""
curl -X GET http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name | contains(\"$AGENT_NAME_2\")) | {name, ip_address, agent_key, status}"

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Manual test completed!"
echo ""
echo "📋 Expected Results:"
echo "- Test 1: Generate agent key 1 ✓"
echo "- Test 2: Generate agent key 2 ✓"
echo "- Test 3: Create agent 1 with IP 192.168.1.200 ✓ (should succeed)"
echo "- Test 4: Try to create agent 2 with same IP 192.168.1.200 ✓ (should fail with 409)"
echo "- Test 5: Create agent 2 with different IP 192.168.1.201 ✓ (should succeed)"
echo "- Test 6: Verify both agents exist with different IPs ✓"
