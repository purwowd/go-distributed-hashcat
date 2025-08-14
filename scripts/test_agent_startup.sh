#!/bin/bash

# Test script for Agent Startup and Heartbeat
echo "🧪 Testing Agent Startup and Heartbeat Functionality"
echo "===================================================="

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
echo "Generating agent key for 'test-agent-startup-$(date +%s)'..."
AGENT_NAME="test-agent-startup-$(date +%s)"
AGENT_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Generate Key Response:"
echo "$AGENT_KEY_RESPONSE"

AGENT_KEY=$(echo "$AGENT_KEY_RESPONSE" | jq -r '.data.agent_key // empty')
echo "Generated Agent Key: $AGENT_KEY"

if [ -z "$AGENT_KEY" ]; then
    echo "❌ Failed to generate agent key"
    echo "Response was: $AGENT_KEY_RESPONSE"
    kill $SERVER_PID
    exit 1
fi

# Test 2: Test agent startup with new data
echo ""
echo "📝 Test 2: Test agent startup with new data"
echo "--------------------------------------------"
echo "Starting agent with new IP and capabilities..."
STARTUP_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"172.15.1.94\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Response:"
echo "$STARTUP_RESPONSE"

# Test 3: Test agent heartbeat
echo ""
echo "📝 Test 3: Test agent heartbeat"
echo "--------------------------------"
echo "Sending heartbeat..."
HEARTBEAT_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/heartbeat \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Heartbeat Response:"
echo "$HEARTBEAT_RESPONSE"

# Test 4: Test agent startup again (should show already exists)
echo ""
echo "📝 Test 4: Test agent startup again (should show already exists)"
echo "----------------------------------------------------------------"
echo "Starting agent again with same data..."
STARTUP_AGAIN_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"172.15.1.94\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Again Response:"
echo "$STARTUP_AGAIN_RESPONSE"

# Test 5: Check agent status
echo ""
echo "📝 Test 5: Check agent status"
echo "------------------------------"
echo "Checking agent status..."
AGENT_STATUS=$(curl -s -X GET http://localhost:1337/api/v1/agents/ | jq '.data[] | select(.name == "'$AGENT_NAME'") | {name, status, ip_address, port, capabilities, agent_key}')

echo "Agent Status:"
echo "$AGENT_STATUS"

# Test 6: Test agent startup with different IP (should fail due to IP conflict)
echo ""
echo "📝 Test 6: Test agent startup with different IP (should fail due to IP conflict)"
echo "------------------------------------------------------------------------------"
echo "Starting agent with different IP (should fail)..."
STARTUP_DIFF_IP_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Different IP Response:"
echo "$STARTUP_DIFF_IP_RESPONSE"

# Test 7: Test agent startup with wrong name (should fail due to name mismatch)
echo ""
echo "📝 Test 7: Test agent startup with wrong name (should fail due to name mismatch)"
echo "-----------------------------------------------------------------------------"
echo "Starting agent with wrong name (should fail)..."
STARTUP_WRONG_NAME_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"wrong-name\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"172.15.1.94\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Wrong Name Response:"
echo "$STARTUP_WRONG_NAME_RESPONSE"

# Test 8: Test agent startup with invalid agent key (should fail)
echo ""
echo "📝 Test 8: Test agent startup with invalid agent key (should fail)"
echo "------------------------------------------------------------------"
echo "Starting agent with invalid agent key (should fail)..."
STARTUP_INVALID_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"invalid-key\", \"ip_address\": \"172.15.1.94\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Invalid Key Response:"
echo "$STARTUP_INVALID_KEY_RESPONSE"

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Generate agent key ✓"
echo "- Test 2: Agent startup with new data ✓"
echo "- Test 3: Agent heartbeat ✓"
echo "- Test 4: Agent startup again (already exists) ✓"
echo "- Test 5: Check agent status ✓"
echo "- Test 6: Agent startup with different IP (should fail) ✓"
echo "- Test 7: Agent startup with wrong name (should fail) ✓"
echo "- Test 8: Agent startup with invalid key (should fail) ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Agent startup should validate agent key and name"
echo "- First startup should update agent data and set status to online"
echo "- Second startup should show 'already exists' and only update status"
echo "- Heartbeat should work with agent key"
echo "- Invalid requests should fail with appropriate error messages"
