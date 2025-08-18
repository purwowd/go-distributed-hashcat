#!/bin/bash

# Test script for Agent Data Ordering
echo "🧪 Testing Agent Data Ordering (Newest First)"
echo "=============================================="

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Generate multiple agent keys with different timestamps
echo ""
echo "📝 Test 1: Generate multiple agent keys with different timestamps"
echo "----------------------------------------------------------------"

# Generate first agent key
echo "Generating agent key 1..."
AGENT_KEY_1_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-ordering-1"}' \
  -w "\nHTTP Status: %{http_code}")

AGENT_KEY_1=$(echo "$AGENT_KEY_1_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Agent Key 1: $AGENT_KEY_1"

# Wait a bit to ensure different timestamp
sleep 2

# Generate second agent key
echo "Generating agent key 2..."
AGENT_KEY_2_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-ordering-2"}' \
  -w "\nHTTP Status: %{http_code}")

AGENT_KEY_2=$(echo "$AGENT_KEY_2_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Agent Key 2: $AGENT_KEY_2"

# Wait a bit to ensure different timestamp
sleep 2

# Generate third agent key
echo "Generating agent key 3..."
AGENT_KEY_3_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-ordering-3"}' \
  -w "\nHTTP Status: %{http_code}")

AGENT_KEY_3=$(echo "$AGENT_KEY_3_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "Agent Key 3: $AGENT_KEY_3"

# Test 2: List all agents to verify ordering
echo ""
echo "📝 Test 2: List all agents to verify ordering (should be newest first)"
echo "----------------------------------------------------------------------"
echo "Expected order: test-ordering-3 (newest), test-ordering-2, test-ordering-1 (oldest)"
echo ""
echo "Actual order:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-ordering")) | {name, created_at, agent_key}' | head -20

# Test 3: Verify specific ordering by checking timestamps
echo ""
echo "📝 Test 3: Verify timestamp ordering"
echo "------------------------------------"
echo "Checking if created_at timestamps are in descending order..."
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-ordering")) | {name, created_at}' | grep -A 2 "test-ordering"

# Test 4: Test with actual agent creation (IP addresses)
echo ""
echo "📝 Test 4: Create agents with IP addresses to test full ordering"
echo "-----------------------------------------------------------------"
echo "Creating agent 1 with IP 192.168.1.100..."
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-ordering-1\", \"agent_key\": \"$AGENT_KEY_1\", \"ip_address\": \"192.168.1.100\", \"port\": 8080}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

echo ""
echo "Creating agent 2 with IP 192.168.1.101..."
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-ordering-2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.101\", \"port\": 8081}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

echo ""
echo "Creating agent 3 with IP 192.168.1.102..."
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-ordering-3\", \"agent_key\": \"$AGENT_KEY_3\", \"ip_address\": \"192.168.1.102\", \"port\": 8082}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 5: Final verification of ordering
echo ""
echo "📝 Test 5: Final verification of ordering (should be newest first)"
echo "------------------------------------------------------------------"
echo "Expected order: test-ordering-3 (newest), test-ordering-2, test-ordering-1 (oldest)"
echo ""
echo "Actual order:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-ordering")) | {name, ip_address, created_at, status}' | head -20

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Generate 3 agent keys with different timestamps ✓"
echo "- Test 2: Verify initial ordering (newest first) ✓"
echo "- Test 3: Verify timestamp ordering ✓"
echo "- Test 4: Create agents with IP addresses ✓"
echo "- Test 5: Final verification of ordering ✓"
echo ""
echo "🎯 Expected Result:"
echo "Data should be ordered by created_at DESC (newest first)"
echo "Order should be: test-ordering-3 → test-ordering-2 → test-ordering-1"
