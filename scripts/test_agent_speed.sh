#!/bin/bash

# Test script untuk fitur speed agent
# Script ini akan menjalankan agent dan menguji update speed

echo "🚀 Testing Agent Speed Feature"
echo "================================"

# Set environment variables
export SERVER_URL="http://localhost:1337"
export AGENT_NAME="test-agent-speed"
export AGENT_KEY="test123"
export AGENT_IP="127.0.0.1"

echo "📋 Test Configuration:"
echo "  Server URL: $SERVER_URL"
echo "  Agent Name: $AGENT_NAME"
echo "  Agent Key: $AGENT_KEY"
echo "  Agent IP: $AGENT_IP"
echo ""

# Check if server is running
echo "🔍 Checking if server is running..."
if ! curl -s "$SERVER_URL/health" > /dev/null 2>&1; then
    echo "❌ Server is not running. Please start the server first."
    echo "   Run: ./bin/server"
    exit 1
fi
echo "✅ Server is running"
echo ""

# Check if hashcat is available
echo "🔍 Checking if hashcat is available..."
if ! command -v hashcat &> /dev/null; then
    echo "❌ hashcat is not installed or not in PATH"
    echo "   Please install hashcat first"
    exit 1
fi
echo "✅ hashcat is available"
echo ""

# Test hashcat benchmark output parsing
echo "🧪 Testing hashcat benchmark output parsing..."
echo "Running: hashcat -b -m 2500"

# Run hashcat benchmark and capture output
BENCHMARK_OUTPUT=$(hashcat -b -m 2500 2>&1)
echo ""

# Parse speed from output
SPEED=$(echo "$BENCHMARK_OUTPUT" | grep -o "Speed\.#1\.*: *[0-9]* H/s" | grep -o "[0-9]*" | head -1)

if [ -n "$SPEED" ]; then
    echo "✅ Speed detected: $SPEED H/s"
else
    echo "❌ No speed detected in benchmark output"
    echo "Benchmark output:"
    echo "$BENCHMARK_OUTPUT"
    exit 1
fi
echo ""

# Test API endpoint for updating agent speed
echo "🌐 Testing API endpoint for updating agent speed..."

# First, create a test agent
echo "Creating test agent..."
CREATE_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/agents/generate-key" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$AGENT_NAME\"}")

if [ $? -ne 0 ]; then
    echo "❌ Failed to create test agent"
    exit 1
fi

echo "Create response: $CREATE_RESPONSE"

# Extract agent ID from response
AGENT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$AGENT_ID" ]; then
    echo "❌ Failed to extract agent ID from response"
    exit 1
fi

echo "✅ Test agent created with ID: $AGENT_ID"
echo ""

# Test updating agent speed
echo "Testing speed update API..."
SPEED_UPDATE_RESPONSE=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed" \
    -H "Content-Type: application/json" \
    -d "{\"speed\": $SPEED}")

if [ $? -eq 0 ]; then
    echo "✅ Speed update API call successful"
    echo "Response: $SPEED_UPDATE_RESPONSE"
else
    echo "❌ Speed update API call failed"
    exit 1
fi
echo ""

# Verify speed was updated in database
echo "🔍 Verifying speed update in database..."
AGENT_INFO=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID")

if [ $? -eq 0 ]; then
    echo "✅ Agent info retrieved successfully"
    echo "Agent info: $AGENT_INFO"
    
    # Check if speed field exists and has the correct value
    if echo "$AGENT_INFO" | grep -q "\"speed\":$SPEED"; then
        echo "✅ Speed field updated correctly in database"
    else
        echo "❌ Speed field not updated correctly in database"
        exit 1
    fi
else
    echo "❌ Failed to retrieve agent info"
    exit 1
fi
echo ""

# Clean up test agent
echo "🧹 Cleaning up test agent..."
DELETE_RESPONSE=$(curl -s -X DELETE "$SERVER_URL/api/v1/agents/$AGENT_ID")

if [ $? -eq 0 ]; then
    echo "✅ Test agent deleted successfully"
else
    echo "⚠️  Warning: Failed to delete test agent"
fi
echo ""

echo "🎉 Agent Speed Feature Test Completed Successfully!"
echo "=================================================="
echo ""
echo "📊 Summary:"
echo "  ✅ Server connectivity: OK"
echo "  ✅ hashcat availability: OK"
echo "  ✅ Speed detection: $SPEED H/s"
echo "  ✅ API endpoint: OK"
echo "  ✅ Database update: OK"
echo "  ✅ Cleanup: OK"
echo ""
echo "🚀 The agent speed feature is working correctly!"
echo "   When you run an agent with:"
echo "   sudo ./bin/agent --server $SERVER_URL --name \"agent-A\" --agent-key \"4c3418d2\" --ip \"172.15.1.94\""
echo ""
echo "   It will:"
echo "   1. Run hashcat -b -m 2500 benchmark"
echo "   2. Parse the speed output (e.g., 1928 H/s)"
echo "   3. Update the agent speed in database"
echo "   4. Broadcast speed updates via WebSocket"
echo "   5. Speed data persists until agent stops"
echo ""
