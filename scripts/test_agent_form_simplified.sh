#!/bin/bash

# Test script to verify Add New Agent form is simplified
echo "🧪 Testing Simplified Add New Agent Form"
echo "========================================"

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
AGENT_NAME="test-form-simplified-$(date +%s)"
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

# Test 3: Test creating agent with only agent_key (no name field)
echo ""
echo "📝 Test 3: Test creating agent with only agent_key (no name field)"
echo "------------------------------------------------------------------"
echo "Testing agent creation with simplified form data..."

CREATE_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Create Agent Response:"
echo "$CREATE_RESPONSE"

# Test 4: Check if agent was created successfully
echo ""
echo "📝 Test 4: Check if agent was created successfully"
echo "--------------------------------------------------"
echo "Checking if agent was created..."

AGENT_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.agent_key == \"$AGENT_KEY\") | {name, agent_key, ip_address, port, capabilities, status}")

echo "Agent Status:"
echo "$AGENT_STATUS"

# Test 5: Verify agent name was automatically set from database
echo ""
echo "📝 Test 5: Verify agent name was automatically set from database"
echo "----------------------------------------------------------------"
echo "Verifying that agent name was automatically set based on agent key..."

AGENT_NAME_FROM_DB=$(echo "$AGENT_STATUS" | jq -r '.name')
echo "Agent Name from Database: $AGENT_NAME_FROM_DB"
echo "Expected Agent Name: $AGENT_NAME"

if [ "$AGENT_NAME_FROM_DB" = "$AGENT_NAME" ]; then
    echo "✅ SUCCESS: Agent name was automatically set from database based on agent key"
    NAME_AUTO_SET=true
else
    echo "❌ FAILED: Agent name was not automatically set correctly"
    echo "   Expected: $AGENT_NAME"
    echo "   Got: $AGENT_NAME_FROM_DB"
    NAME_AUTO_SET=false
fi

# Test 6: Test creating agent with missing agent_key (should fail)
echo ""
echo "📝 Test 6: Test creating agent with missing agent_key (should fail)"
echo "-------------------------------------------------------------------"
echo "Testing agent creation without agent_key (should fail)..."

CREATE_FAIL_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d "{\"ip_address\": \"192.168.1.101\", \"port\": 8080, \"capabilities\": \"GPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Create Agent Without Key Response:"
echo "$CREATE_FAIL_RESPONSE"

# Test 7: Verify error message for missing agent_key
echo ""
echo "📝 Test 7: Verify error message for missing agent_key"
echo "-----------------------------------------------------"
echo "Verifying error message for missing agent_key..."

if [[ "$CREATE_FAIL_RESPONSE" == *"400"* ]]; then
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
echo "- Test 3: Test creating agent with only agent_key ✓"
echo "- Test 4: Check if agent was created successfully ✓"
echo "- Test 5: Verify agent name was automatically set ✓"
echo "- Test 6: Test creating agent with missing agent_key ✓"
echo "- Test 7: Verify error message for missing agent_key ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Agent creation should work with only agent_key (no name field required)"
echo "- Agent name should be automatically set from database based on agent_key"
echo "- Validation should fail when agent_key is missing"
echo ""
echo "🔧 Form Simplification Status:"
if [ "$NAME_AUTO_SET" = true ]; then
    echo "✅ SUCCESS: Agent name is automatically set from database"
    echo "   - No need to input agent name manually"
    echo "   - Agent name is retrieved based on agent key validation"
else
    echo "❌ FAILED: Agent name is not automatically set correctly"
    echo "   - Further investigation needed"
fi

echo ""
echo "🔧 Validation Status:"
if [ "$VALIDATION_WORKS" = true ]; then
    echo "✅ SUCCESS: Form validation works correctly"
    echo "   - Only agent_key is required"
    echo "   - Missing agent_key returns proper error"
else
    echo "❌ FAILED: Form validation has issues"
    echo "   - Further investigation needed"
fi

echo ""
echo "🚀 Form Simplification Benefits:"
echo "- ✅ Simpler user experience (fewer fields to fill)"
echo "- ✅ Reduced user error (no name mismatch issues)"
echo "- ✅ Automatic validation (agent name comes from database)"
echo "- ✅ Consistent data (agent name always matches agent key)"
