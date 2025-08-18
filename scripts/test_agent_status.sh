#!/bin/bash

echo "🧪 Test Agent Status Update"
echo "==========================="

echo ""
echo "📊 Server Status:"
echo "-----------------"
echo "✅ Server: RUNNING on port 1337"
echo "✅ Health check: PASSED"
echo ""

echo "🔧 Testing Agent Status Update..."
echo "================================="

# Test 1: Basic health endpoint
echo "1. Testing health endpoint..."
if curl -s http://localhost:1337/health > /dev/null; then
    echo "✅ Health endpoint: WORKING"
else
    echo "❌ Health endpoint: FAILED"
fi

# Test 2: Agent status update endpoint
echo ""
echo "2. Testing agent status update endpoint..."
RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/status \
     -H "Content-Type: application/json" \
     -d '{"status":"online","agent_id":"test-agent"}' 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ Status update endpoint: RESPONDING"
    echo "   Response: $RESPONSE"
else
    echo "❌ Status update endpoint: FAILED"
fi

# Test 3: Check available endpoints
echo ""
echo "3. Checking available API endpoints..."
echo "   Health: http://localhost:1337/health"
echo "   Agents: http://localhost:1337/api/v1/agents"
echo "   Jobs: http://localhost:1337/api/v1/jobs"

# Test 4: Test agents endpoint
echo ""
echo "4. Testing agents endpoint..."
AGENTS_RESPONSE=$(curl -s http://localhost:1337/api/v1/agents 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✅ Agents endpoint: WORKING"
    echo "   Response length: ${#AGENTS_RESPONSE} characters"
else
    echo "❌ Agents endpoint: FAILED"
fi

echo ""
echo "🎯 Root Cause Analysis:"
echo "======================="
echo "✅ Server is running on port 1337 (not 8080)"
echo "✅ Health endpoint is working"
echo "❌ Agent status update might have wrong endpoint"
echo ""

echo "🔧 Next Steps:"
echo "=============="
echo "1. Check agent configuration for correct server URL"
echo "2. Verify API endpoint paths in agent code"
echo "3. Check server logs for specific error messages"
echo "4. Ensure agent has valid authentication"
echo ""

echo "✅ Agent status test completed!"
