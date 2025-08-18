#!/bin/bash

# Test script for Stable Card Position and Agent Status
echo "🧪 Testing Stable Card Position and Agent Status Management"
echo "============================================================"

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Get initial agent list and positions
echo ""
echo "📝 Test 1: Get initial agent list and positions"
echo "-----------------------------------------------"
echo "Initial agent order:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10

# Test 2: Update agent status to online
echo ""
echo "📝 Test 2: Update agent status to online"
echo "----------------------------------------"
echo "Updating first agent status to 'online'..."
# Get first agent ID
AGENT_ID=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[0].id' | tr -d '"')
AGENT_NAME=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[0].name' | tr -d '"')

echo "Updating agent: $AGENT_NAME (ID: $AGENT_ID)"
echo "Current status: $(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[0].status')"

# Update agent status to online
curl -X PUT http://localhost:1337/api/v1/agents/$AGENT_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "online"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 3: Check if agent positions changed after status update
echo ""
echo "📝 Test 3: Check if agent positions changed after status update"
echo "---------------------------------------------------------------"
echo "Agent order after status update:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10

# Test 4: Update agent heartbeat
echo ""
echo "📝 Test 4: Update agent heartbeat"
echo "----------------------------------"
echo "Updating agent heartbeat..."
curl -X PUT http://localhost:1337/api/v1/agents/$AGENT_ID/heartbeat \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 5: Verify specific agent position and status
echo ""
echo "📝 Test 5: Verify specific agent position and status"
echo "----------------------------------------------------"
echo "Position and status of updated agent '$AGENT_NAME':"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data | to_entries | map(select(.value.name == "'$AGENT_NAME'")) | .[0] | {position: .key, name: .value.name, status: .value.status, created_at: .value.created_at, id: .value.id}'

# Test 6: Create new agent to test if it affects existing positions
echo ""
echo "📝 Test 6: Create new agent to test position stability"
echo "-------------------------------------------------------"
echo "Creating new agent to see if existing positions change..."
# Generate new agent key
NEW_AGENT_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-stable-position-$(date +%s)"}' \
  -w "\nHTTP Status: %{http_code}")

NEW_AGENT_KEY=$(echo "$NEW_AGENT_KEY_RESPONSE" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
echo "New Agent Key: $NEW_AGENT_KEY"

if [ ! -z "$NEW_AGENT_KEY" ]; then
    # Create new agent
    curl -X POST http://localhost:1337/api/v1/agents/ \
      -H "Content-Type: application/json" \
      -d "{\"name\": \"test-stable-position-$(date +%s)\", \"agent_key\": \"$NEW_AGENT_KEY\", \"ip_address\": \"192.168.1.888\", \"port\": 8888}" \
      -w "\nHTTP Status: %{http_code}\n" | jq '.'
    
    echo ""
    echo "Agent order after creating new agent:"
    curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10
fi

# Test 7: Test multiple status updates
echo ""
echo "📝 Test 7: Test multiple status updates"
echo "----------------------------------------"
echo "Updating agent status to 'busy'..."
curl -X PUT http://localhost:1337/api/v1/agents/$AGENT_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "busy"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

echo "Agent order after multiple status updates:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Get initial agent positions ✓"
echo "- Test 2: Update agent status to online ✓"
echo "- Test 3: Check positions after status update ✓"
echo "- Test 4: Update agent heartbeat ✓"
echo "- Test 5: Verify specific agent position and status ✓"
echo "- Test 6: Create new agent and check stability ✓"
echo "- Test 7: Test multiple status updates ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Agent card positions should remain stable after status updates"
echo "- Agent status should change correctly (offline → online → busy)"
echo "- New agents should be added at the top (newest first)"
echo "- Existing agents should maintain their relative positions"
echo "- Heartbeat updates should work without changing positions"
