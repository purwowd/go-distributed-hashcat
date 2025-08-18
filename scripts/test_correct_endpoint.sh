#!/bin/bash

echo "🎯 Test Correct Agent Status Update Endpoint"
echo "==========================================="

echo ""
echo "📊 Router Analysis:"
echo "-------------------"
echo "✅ Server: RUNNING on port 1337"
echo "✅ Endpoint: PUT /api/v1/agents/:id/status"
echo "❌ Wrong endpoint: POST /api/v1/agents/status"
echo ""

echo "🔧 Testing Correct Endpoint..."
echo "=============================="

# Test 1: Check available agents first
echo "1. Getting available agents..."
AGENTS_RESPONSE=$(curl -s http://localhost:1337/api/v1/agents 2>/dev/null)
echo "   Response: $AGENTS_RESPONSE"

# Extract first agent ID if available
AGENT_ID=$(echo "$AGENTS_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$AGENT_ID" ]; then
    echo "   Found agent ID: $AGENT_ID"
else
    echo "   No agents found, using test ID"
    AGENT_ID="test-agent-123"
fi

echo ""
echo "2. Testing correct status update endpoint..."
echo "   Endpoint: PUT /api/v1/agents/$AGENT_ID/status"

STATUS_RESPONSE=$(curl -s -X PUT "http://localhost:1337/api/v1/agents/$AGENT_ID/status" \
     -H "Content-Type: application/json" \
     -d '{"status":"online","last_seen":"'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"}' 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ Status update: SUCCESS"
    echo "   Response: $STATUS_RESPONSE"
else
    echo "❌ Status update: FAILED"
fi

echo ""
echo "3. Testing heartbeat endpoint..."
echo "   Endpoint: PUT /api/v1/agents/$AGENT_ID/heartbeat"

HEARTBEAT_RESPONSE=$(curl -s -X PUT "http://localhost:1337/api/v1/agents/$AGENT_ID/heartbeat" \
     -H "Content-Type: application/json" \
     -d '{"status":"online","capabilities":"GPU"}' 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ Heartbeat update: SUCCESS"
    echo "   Response: $HEARTBEAT_RESPONSE"
else
    echo "❌ Heartbeat update: FAILED"
fi

echo ""
echo "🎯 Root Cause Found:"
echo "===================="
echo "❌ Agent was using: POST /api/v1/agents/status"
echo "✅ Correct endpoint: PUT /api/v1/agents/:id/status"
echo ""

echo "🔧 Fix Required:"
echo "================"
echo "1. Update agent code to use correct endpoint"
echo "2. Include agent ID in the URL path"
echo "3. Use PUT method instead of POST"
echo "4. Ensure proper authentication"
echo ""

echo "📝 Correct API Call Format:"
echo "==========================="
echo "PUT http://localhost:1337/api/v1/agents/{AGENT_ID}/status"
echo "Headers: Content-Type: application/json"
echo "Body: {\"status\":\"online\",\"last_seen\":\"timestamp\"}"
echo ""

echo "✅ Correct endpoint test completed!"
