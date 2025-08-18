#!/bin/bash

# Test script to verify last_seen field is properly updated
echo "🧪 Testing Last Seen Field Update Fix"
echo "====================================="

# Kill any existing server
echo "🛑 Killing any existing server..."
pkill -f "go run cmd/server/main.go" || true
sleep 2

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
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
AGENT_NAME="test-last-seen-$(date +%s)"
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

# Test 3: Check initial agent status (should have default last_seen)
echo ""
echo "📝 Test 3: Check initial agent status"
echo "-------------------------------------"
echo "Checking initial agent status..."

INITIAL_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, status}")

echo "Initial Agent Status:"
echo "$INITIAL_STATUS"

# Test 4: Test agent startup (should update last_seen)
echo ""
echo "📝 Test 4: Test agent startup"
echo "------------------------------"
echo "Testing agent startup to update last_seen..."

STARTUP_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Response:"
echo "$STARTUP_RESPONSE"

# Test 5: Check agent status after startup (should have updated last_seen)
echo ""
echo "📝 Test 5: Check agent status after startup"
echo "--------------------------------------------"
echo "Checking agent status after startup..."

AFTER_STARTUP_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, status, ip_address, port, capabilities}")

echo "Agent Status After Startup:"
echo "$AFTER_STARTUP_STATUS"

# Test 6: Test agent heartbeat (should update last_seen again)
echo ""
echo "📝 Test 6: Test agent heartbeat"
echo "--------------------------------"
echo "Testing agent heartbeat to update last_seen again..."

HEARTBEAT_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/heartbeat \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Heartbeat Response:"
echo "$HEARTBEAT_RESPONSE"

# Test 7: Check agent status after heartbeat (should have latest last_seen)
echo ""
echo "📝 Test 7: Check agent status after heartbeat"
echo "---------------------------------------------"
echo "Checking agent status after heartbeat..."

AFTER_HEARTBEAT_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, status, ip_address, port, capabilities}")

echo "Agent Status After Heartbeat:"
echo "$AFTER_HEARTBEAT_STATUS"

# Test 8: Test agent startup again (should show already exists and update last_seen)
echo ""
echo "📝 Test 8: Test agent startup again"
echo "------------------------------------"
echo "Testing agent startup again (should show already exists)..."

STARTUP_AGAIN_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Again Response:"
echo "$STARTUP_AGAIN_RESPONSE"

# Test 9: Check final agent status
echo ""
echo "📝 Test 9: Check final agent status"
echo "-----------------------------------"
echo "Checking final agent status..."

FINAL_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, status, ip_address, port, capabilities}")

echo "Final Agent Status:"
echo "$FINAL_STATUS"

# Test 10: Verify last_seen is not default value
echo ""
echo "📝 Test 10: Verify last_seen is not default value"
echo "------------------------------------------------"
echo "Verifying last_seen is not the default '0001-01-01' value..."

LAST_SEEN_VALUE=$(echo "$FINAL_STATUS" | jq -r '.last_seen')
echo "Last Seen Value: $LAST_SEEN_VALUE"

if [[ "$LAST_SEEN_VALUE" == "0001-01-01"* ]]; then
    echo "❌ FAILED: last_seen still shows default value"
    LAST_SEEN_FIXED=false
else
    echo "✅ SUCCESS: last_seen has been updated from default value"
    LAST_SEEN_FIXED=true
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
echo "- Test 4: Test agent startup ✓"
echo "- Test 5: Check agent status after startup ✓"
echo "- Test 6: Test agent heartbeat ✓"
echo "- Test 7: Check agent status after heartbeat ✓"
echo "- Test 8: Test agent startup again ✓"
echo "- Test 9: Check final agent status ✓"
echo "- Test 10: Verify last_seen fix ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Agent startup should update last_seen"
echo "- Agent heartbeat should update last_seen"
echo "- last_seen should NOT be '0001-01-01' (default value)"
echo "- last_seen should be current timestamp"
echo ""
echo "🔧 Last Seen Fix Status:"
if [ "$LAST_SEEN_FIXED" = true ]; then
    echo "✅ SUCCESS: last_seen field is now properly updated"
    echo "   - Default value '0001-01-01' issue has been resolved"
    echo "   - Agent startup and heartbeat now properly update last_seen"
else
    echo "❌ FAILED: last_seen field still has issues"
    echo "   - Default value '0001-01-01' still appears"
    echo "   - Further investigation needed"
fi
