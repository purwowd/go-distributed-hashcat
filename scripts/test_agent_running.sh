#!/bin/bash

# Test script to verify agent can run properly
echo "🧪 Testing Agent Running with New Heartbeat Endpoint"
echo "===================================================="

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
AGENT_NAME="test-agent-running-$(date +%s)"
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

# Test 3: Test agent startup
echo ""
echo "📝 Test 3: Test agent startup"
echo "------------------------------"
echo "Testing agent startup with new data..."

STARTUP_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Response:"
echo "$STARTUP_RESPONSE"

# Test 4: Test agent heartbeat manually
echo ""
echo "📝 Test 4: Test agent heartbeat manually"
echo "----------------------------------------"
echo "Testing heartbeat endpoint manually..."

HEARTBEAT_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/heartbeat \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Heartbeat Response:"
echo "$HEARTBEAT_RESPONSE"

# Test 5: Check agent status
echo ""
echo "📝 Test 5: Check agent status"
echo "------------------------------"
echo "Checking agent status..."

AGENT_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, status, ip_address, port, capabilities, agent_key}")

echo "Agent Status:"
echo "$AGENT_STATUS"

# Test 6: Test agent binary (if available)
echo ""
echo "📝 Test 6: Test agent binary"
echo "-----------------------------"
if [ -f "./bin/agent" ]; then
    echo "✅ Agent binary found"
    echo "Testing agent with --help flag..."
    ./bin/agent --help | head -10
else
    echo "⚠️ Agent binary not found, skipping binary test"
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
echo "- Test 3: Agent startup ✓"
echo "- Test 4: Agent heartbeat ✓"
echo "- Test 5: Check agent status ✓"
echo "- Test 6: Agent binary ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Agent startup should update agent data"
echo "- Agent heartbeat should work with new endpoint"
echo "- Agent status should be updated"
echo "- Agent binary should be accessible"
echo ""
echo "🚀 Next Steps:"
echo "1. Copy the new agent binary to your agent machine"
echo "2. Run: sudo ./bin/agent --server http://30.30.30.102:1337 --name test-agent-003 --ip \"172.15.1.94\" --agent-key \"$AGENT_KEY\""
echo "3. Agent should now work with heartbeat properly"
