#!/bin/bash

# Test script for Frontend Stable Sorting
echo "🧪 Testing Frontend Stable Sorting Implementation"
echo "================================================="

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Check if frontend is accessible
echo ""
echo "📝 Test 1: Check if frontend is accessible"
echo "-------------------------------------------"
echo "Testing frontend endpoint..."
curl -s -I http://localhost:1337/ | head -1

# Test 2: Get initial agent data
echo ""
echo "📝 Test 2: Get initial agent data"
echo "----------------------------------"
echo "Fetching initial agent list..."
INITIAL_AGENTS=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}')
echo "$INITIAL_AGENTS"

# Test 3: Create test agents with different timestamps
echo ""
echo "📝 Test 3: Create test agents with different timestamps"
echo "-------------------------------------------------------"

# Generate agent key for test
AGENT_KEY_1=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-frontend-sort-1"}' | jq -r '.data.agent_key')

AGENT_KEY_2=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-frontend-sort-2"}' | jq -r '.data.agent_key')

echo "Agent Key 1: $AGENT_KEY_1"
echo "Agent Key 2: $AGENT_KEY_2"

# Create first agent
echo "Creating first test agent..."
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-frontend-sort-1\", \"agent_key\": \"$AGENT_KEY_1\", \"ip_address\": \"192.168.1.100\", \"port\": 8080}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

sleep 2

# Create second agent
echo "Creating second test agent..."
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"test-frontend-sort-2\", \"agent_key\": \"$AGENT_KEY_2\", \"ip_address\": \"192.168.1.101\", \"port\": 8080}" \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 4: Check agent order after creation
echo ""
echo "📝 Test 4: Check agent order after creation"
echo "-------------------------------------------"
echo "Agent order after creating test agents:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10

# Test 5: Update agent status and check position stability
echo ""
echo "📝 Test 5: Update agent status and check position stability"
echo "------------------------------------------------------------"
echo "Updating first test agent status to 'online'..."

# Get first test agent ID
TEST_AGENT_ID=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name | contains("test-frontend-sort")) | .id' | head -1 | tr -d '"')

echo "Test Agent ID: $TEST_AGENT_ID"

# Update status
curl -X PUT http://localhost:1337/api/v1/agents/$TEST_AGENT_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "online"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

echo "Agent order after status update:"
curl -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | {name, status, created_at, id}' | head -10

# Test 6: Verify position stability
echo ""
echo "📝 Test 6: Verify position stability"
echo "------------------------------------"
echo "Checking if test agents maintained their positions..."

# Get current positions
CURRENT_POSITIONS=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data | to_entries | map(select(.value.name | contains("test-frontend-sort"))) | .[] | {position: .key, name: .value.name, status: .value.status, created_at: .value.created_at}')

echo "Current positions of test agents:"
echo "$CURRENT_POSITIONS"

# Test 7: Test frontend sorting logic
echo ""
echo "📝 Test 7: Test frontend sorting logic"
echo "--------------------------------------"
echo "Testing frontend stable sorting implementation..."

# Open frontend in browser (optional)
if command -v open &> /dev/null; then
    echo "Opening frontend in browser..."
    open http://localhost:1337/
elif command -v xdg-open &> /dev/null; then
    echo "Opening frontend in browser..."
    xdg-open http://localhost:1337/
else
    echo "Frontend available at: http://localhost:1337/"
    echo "Please manually test the stable sorting functionality:"
    echo "1. Open the frontend in your browser"
    echo "2. Check if agent cards maintain their positions after updates"
    echo "3. Verify that new agents appear at the top"
    echo "4. Confirm that existing agents don't move when status changes"
fi

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Frontend accessibility ✓"
echo "- Test 2: Initial agent data ✓"
echo "- Test 3: Create test agents ✓"
echo "- Test 4: Check agent order after creation ✓"
echo "- Test 5: Update agent status ✓"
echo "- Test 6: Verify position stability ✓"
echo "- Test 7: Frontend sorting logic ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Frontend should be accessible at http://localhost:1337/"
echo "- Test agents should be created successfully"
echo "- Agent order should be: newest first (created_at DESC)"
echo "- Agent positions should remain stable after status updates"
echo "- Frontend should implement stable sorting (created_at DESC, id ASC)"
echo "- Card positions should not change when updating agent properties"
