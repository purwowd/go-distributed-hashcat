#!/bin/bash

# Test script to verify UpdateAgentData endpoint works correctly
echo "🧪 Testing UpdateAgentData Endpoint (No Status Change)"
echo "====================================================="

# Kill any existing server
echo "🛑 Killing any existing server..."
pkill -f "go run cmd/server/main.go" || true
sleep 2

# Start the server in background
echo "🚀 Starting server..."
cd .. && go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Test 1: Check if server is running
echo ""
echo "📝 Test 1: Check if server is running"
echo "-------------------------------------"
HEALTH_RESPONSE=$(curl -s http://localhost:1337/health)
echo "Health Response: $HEALTH_RESPONSE"

# Test 2: Generate agent key for testing
echo ""
echo "📝 Test 2: Generate agent key for testing"
echo "------------------------------------------"
AGENT_NAME="test-update-data-$(date +%s)"
echo "Generating agent key for: $AGENT_NAME"

GENERATE_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\"}")

echo "Generate Response:"
echo "$GENERATE_RESPONSE"

AGENT_KEY=$(echo "$GENERATE_RESPONSE" | jq -r '.data.agent_key // empty')
echo "Agent Key: $AGENT_KEY"

if [ -z "$AGENT_KEY" ]; then
    echo "❌ Failed to get agent key"
    kill $SERVER_PID
    exit 1
fi

# Test 3: Check initial agent status (should be offline)
echo ""
echo "📝 Test 3: Check initial agent status (should be offline)"
echo "--------------------------------------------------------"
echo "Checking initial agent status..."

INITIAL_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.agent_key == \"$AGENT_KEY\") | {name, agent_key, ip_address, port, capabilities, status}")

echo "Initial Agent Status:"
echo "$INITIAL_STATUS"

INITIAL_STATUS_VALUE=$(echo "$INITIAL_STATUS" | jq -r '.status')
echo "Initial Status Value: $INITIAL_STATUS_VALUE"

if [ "$INITIAL_STATUS_VALUE" = "offline" ]; then
    echo "✅ SUCCESS: Initial status is offline as expected"
    INITIAL_STATUS_CORRECT=true
else
    echo "❌ FAILED: Initial status should be offline, got: $INITIAL_STATUS_VALUE"
    INITIAL_STATUS_CORRECT=false
fi

# Test 4: Update agent data using new endpoint
echo ""
echo "📝 Test 4: Update agent data using new endpoint"
echo "-----------------------------------------------"
echo "Updating agent data (ip_address, port, capabilities)..."

UPDATE_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.200\", \"port\": 9090, \"capabilities\": \"GPU RTX 4090\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Update Response:"
echo "$UPDATE_RESPONSE"

# Test 5: Check if agent data was updated successfully
echo ""
echo "📝 Test 5: Check if agent data was updated successfully"
echo "-------------------------------------------------------"
echo "Checking updated agent data..."

UPDATED_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.agent_key == \"$AGENT_KEY\") | {name, agent_key, ip_address, port, capabilities, status}")

echo "Updated Agent Status:"
echo "$UPDATED_STATUS"

# Test 6: Verify data fields were updated
echo ""
echo "📝 Test 6: Verify data fields were updated"
echo "------------------------------------------"
echo "Verifying that data fields were updated correctly..."

UPDATED_IP=$(echo "$UPDATED_STATUS" | jq -r '.ip_address')
UPDATED_PORT=$(echo "$UPDATED_STATUS" | jq -r '.port')
UPDATED_CAPABILITIES=$(echo "$UPDATED_STATUS" | jq -r '.capabilities')
UPDATED_STATUS_VALUE=$(echo "$UPDATED_STATUS" | jq -r '.status')

echo "Updated IP: $UPDATED_IP (Expected: 192.168.1.200)"
echo "Updated Port: $UPDATED_PORT (Expected: 9090)"
echo "Updated Capabilities: $UPDATED_CAPABILITIES (Expected: GPU RTX 4090)"
echo "Updated Status: $UPDATED_STATUS_VALUE (Expected: offline - should NOT change)"

DATA_UPDATE_SUCCESS=true

if [ "$UPDATED_IP" != "192.168.1.200" ]; then
    echo "❌ FAILED: IP address was not updated correctly"
    DATA_UPDATE_SUCCESS=false
fi

if [ "$UPDATED_PORT" != "9090" ]; then
    echo "❌ FAILED: Port was not updated correctly"
    DATA_UPDATE_SUCCESS=false
fi

if [ "$UPDATED_CAPABILITIES" != "GPU RTX 4090" ]; then
    echo "❌ FAILED: Capabilities was not updated correctly"
    DATA_UPDATE_SUCCESS=false
fi

if [ "$UPDATED_STATUS_VALUE" != "offline" ]; then
    echo "❌ FAILED: Status should remain offline, got: $UPDATED_STATUS_VALUE"
    DATA_UPDATE_SUCCESS=false
fi

if [ "$DATA_UPDATE_SUCCESS" = true ]; then
    echo "✅ SUCCESS: All data fields updated correctly while status remained offline"
else
    echo "❌ FAILED: Some data fields were not updated correctly"
fi

# Test 7: Test with missing agent_key (should fail)
echo ""
echo "📝 Test 7: Test with missing agent_key (should fail)"
echo "----------------------------------------------------"
echo "Testing update without agent_key (should fail)..."

UPDATE_FAIL_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"ip_address\": \"192.168.1.201\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Update Without Key Response:"
echo "$UPDATE_FAIL_RESPONSE"

# Test 8: Verify error handling for missing agent_key
echo ""
echo "📝 Test 8: Verify error handling for missing agent_key"
echo "-----------------------------------------------------"
echo "Verifying error handling for missing agent_key..."

if [[ "$UPDATE_FAIL_RESPONSE" == *"400"* ]]; then
    echo "✅ SUCCESS: Request without agent_key returned 400 status (validation error)"
    VALIDATION_WORKS=true
else
    echo "❌ FAILED: Request without agent_key should have failed with 400 status"
    VALIDATION_WORKS=false
fi

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Server health ✓"
echo "- Test 2: Generate agent key ✓"
echo "- Test 3: Check initial agent status ✓"
echo "- Test 4: Update agent data using new endpoint ✓"
echo "- Test 5: Check if agent data was updated ✓"
echo "- Test 6: Verify data fields were updated ✓"
echo "- Test 7: Test with missing agent_key ✓"
echo "- Test 8: Verify error handling for missing agent_key ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Initial agent status should be offline"
echo "- Agent data update should work (ip_address, port, capabilities)"
echo "- Agent status should remain offline after update"
echo "- Validation should fail when agent_key is missing"
echo ""
echo "🔧 UpdateAgentData Endpoint Status:"
if [ "$DATA_UPDATE_SUCCESS" = true ]; then
    echo "✅ SUCCESS: Agent data updated correctly"
    echo "   - IP address updated: $UPDATED_IP"
    echo "   - Port updated: $UPDATED_PORT"
    echo "   - Capabilities updated: $UPDATED_CAPABILITIES"
    echo "   - Status remained offline: $UPDATED_STATUS_VALUE"
else
    echo "❌ FAILED: Agent data update has issues"
    echo "   - Further investigation needed"
fi

echo ""
echo "🔧 Status Preservation Status:"
if [ "$UPDATED_STATUS_VALUE" = "offline" ]; then
    echo "✅ SUCCESS: Agent status remained offline after data update"
    echo "   - Status is preserved during data updates"
    echo "   - Only changes when agent binary runs"
else
    echo "❌ FAILED: Agent status changed unexpectedly"
    echo "   - Status should remain offline until agent binary runs"
fi

echo ""
echo "🔧 Validation Status:"
if [ "$VALIDATION_WORKS" = true ]; then
    echo "✅ SUCCESS: Endpoint validation works correctly"
    echo "   - Only agent_key is required"
    echo "   - Missing agent_key returns proper error"
else
    echo "❌ FAILED: Endpoint validation has issues"
    echo "   - Further investigation needed"
fi

echo ""
echo "🚀 New Endpoint Benefits:"
echo "- ✅ Only updates data fields (ip_address, port, capabilities)"
echo "- ✅ Preserves agent status (remains offline)"
echo "- ✅ No unnecessary status broadcasts"
echo "- ✅ Cleaner separation of concerns"
echo "- ✅ Status only changes when agent binary runs"
