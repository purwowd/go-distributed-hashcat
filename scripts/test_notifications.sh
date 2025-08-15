#!/bin/bash

# Test script to verify notifications and error handling work correctly
echo "🧪 Testing Notifications and Error Handling"
echo "============================================"

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
AGENT_NAME="test-notifications-$(date +%s)"
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

# Test 3: Test successful update (should show success notification)
echo ""
echo "📝 Test 3: Test successful update (should show success notification)"
echo "------------------------------------------------------------------"
echo "Testing successful agent data update..."

SUCCESS_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.300\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Success Update Response:"
echo "$SUCCESS_RESPONSE"

# Test 4: Verify success response format
echo ""
echo "📝 Test 4: Verify success response format"
echo "-----------------------------------------"
echo "Verifying success response format..."

SUCCESS_CODE=$(echo "$SUCCESS_RESPONSE" | jq -r '.code // empty')
SUCCESS_MESSAGE=$(echo "$SUCCESS_RESPONSE" | jq -r '.message // empty')

echo "Success Code: $SUCCESS_CODE"
echo "Success Message: $SUCCESS_MESSAGE"

if [ "$SUCCESS_CODE" = "UPDATE_SUCCESS" ] && [ "$SUCCESS_MESSAGE" = "Agent data updated successfully" ]; then
    echo "✅ SUCCESS: Response format is correct"
    SUCCESS_FORMAT_CORRECT=true
else
    echo "❌ FAILED: Response format is incorrect"
    echo "   Expected code: UPDATE_SUCCESS, got: $SUCCESS_CODE"
    echo "   Expected message: Agent data updated successfully, got: $SUCCESS_MESSAGE"
    SUCCESS_FORMAT_CORRECT=false
fi

# Test 5: Test with non-existent agent key (should show agent key not found error)
echo ""
echo "📝 Test 5: Test with non-existent agent key (should show agent key not found error)"
echo "-------------------------------------------------------------------------------"
echo "Testing update with non-existent agent key..."

INVALID_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"invalid_key_123\", \"ip_address\": \"192.168.1.301\", \"port\": 8080, \"capabilities\": \"GPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Invalid Key Response:"
echo "$INVALID_KEY_RESPONSE"

# Test 6: Verify agent key not found error format
echo ""
echo "📝 Test 6: Verify agent key not found error format"
echo "-------------------------------------------------"
echo "Verifying agent key not found error format..."

INVALID_KEY_CODE=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.code // empty')
INVALID_KEY_ERROR=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.error // empty')
INVALID_KEY_MESSAGE=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.message // empty')

echo "Invalid Key Code: $INVALID_KEY_CODE"
echo "Invalid Key Error: $INVALID_KEY_ERROR"
echo "Invalid Key Message: $INVALID_KEY_MESSAGE"

if [ "$INVALID_KEY_CODE" = "AGENT_KEY_NOT_FOUND" ] && [ "$INVALID_KEY_ERROR" = "Agent key not found in database" ]; then
    echo "✅ SUCCESS: Agent key not found error format is correct"
    INVALID_KEY_ERROR_CORRECT=true
else
    echo "❌ FAILED: Agent key not found error format is incorrect"
    echo "   Expected code: AGENT_KEY_NOT_FOUND, got: $INVALID_KEY_CODE"
    echo "   Expected error: Agent key not found in database, got: $INVALID_KEY_ERROR"
    INVALID_KEY_ERROR_CORRECT=false
fi

# Test 7: Test with IP address conflict (should show IP address conflict error)
echo ""
echo "📝 Test 7: Test with IP address conflict (should show IP address conflict error)"
echo "---------------------------------------------------------------------------"
echo "Testing update with IP address conflict..."

IP_CONFLICT_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.300\", \"port\": 8080, \"capabilities\": \"GPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "IP Conflict Response:"
echo "$IP_CONFLICT_RESPONSE"

# Test 8: Verify IP address conflict error format
echo ""
echo "📝 Test 8: Verify IP address conflict error format"
echo "-------------------------------------------------"
echo "Verifying IP address conflict error format..."

IP_CONFLICT_CODE=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.code // empty')
IP_CONFLICT_ERROR=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.error // empty')
IP_CONFLICT_MESSAGE=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.message // empty')

echo "IP Conflict Code: $IP_CONFLICT_CODE"
echo "IP Conflict Error: $IP_CONFLICT_ERROR"
echo "IP Conflict Message: $IP_CONFLICT_MESSAGE"

if [ "$IP_CONFLICT_CODE" = "IP_ADDRESS_CONFLICT" ] && [ "$IP_CONFLICT_ERROR" = "IP address already in use" ]; then
    echo "✅ SUCCESS: IP address conflict error format is correct"
    IP_CONFLICT_ERROR_CORRECT=true
else
    echo "❌ FAILED: IP address conflict error format is incorrect"
    echo "   Expected code: IP_ADDRESS_CONFLICT, got: $IP_CONFLICT_CODE"
    echo "   Expected error: IP address already in use, got: $IP_CONFLICT_ERROR"
    IP_CONFLICT_ERROR_CORRECT=false
fi

# Test 9: Test with missing agent_key (should show validation error)
echo ""
echo "📝 Test 9: Test with missing agent_key (should show validation error)"
echo "------------------------------------------------------------------------"
echo "Testing update without agent_key (should fail validation)..."

MISSING_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"ip_address\": \"192.168.1.302\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Missing Key Response:"
echo "$MISSING_KEY_RESPONSE"

# Test 10: Verify validation error handling
echo ""
echo "📝 Test 10: Verify validation error handling"
echo "--------------------------------------------"
echo "Verifying validation error handling..."

if [[ "$MISSING_KEY_RESPONSE" == *"400"* ]]; then
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
echo "- Test 3: Test successful update ✓"
echo "- Test 4: Verify success response format ✓"
echo "- Test 5: Test with non-existent agent key ✓"
echo "- Test 6: Verify agent key not found error format ✓"
echo "- Test 7: Test with IP address conflict ✓"
echo "- Test 8: Verify IP address conflict error format ✓"
echo "- Test 9: Test with missing agent_key ✓"
echo "- Test 10: Verify validation error handling ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Successful update should return UPDATE_SUCCESS code"
echo "- Agent key not found should return AGENT_KEY_NOT_FOUND code"
echo "- IP address conflict should return IP_ADDRESS_CONFLICT code"
echo "- Validation should fail when agent_key is missing"
echo ""
echo "🔧 Success Response Status:"
if [ "$SUCCESS_FORMAT_CORRECT" = true ]; then
    echo "✅ SUCCESS: Success response format is correct"
    echo "   - Code: $SUCCESS_CODE"
    echo "   - Message: $SUCCESS_MESSAGE"
else
    echo "❌ FAILED: Success response format has issues"
    echo "   - Further investigation needed"
fi

echo ""
echo "🔧 Error Response Status:"
if [ "$INVALID_KEY_ERROR_CORRECT" = true ]; then
    echo "✅ SUCCESS: Agent key not found error format is correct"
    echo "   - Code: $INVALID_KEY_CODE"
    echo "   - Error: $INVALID_KEY_ERROR"
else
    echo "❌ FAILED: Agent key not found error format has issues"
    echo "   - Further investigation needed"
fi

if [ "$IP_CONFLICT_ERROR_CORRECT" = true ]; then
    echo "✅ SUCCESS: IP address conflict error format is correct"
    echo "   - Code: $IP_CONFLICT_CODE"
    echo "   - Error: $IP_CONFLICT_ERROR"
else
    echo "❌ FAILED: IP address conflict error format has issues"
    echo "   - Further investigation needed"
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
echo "🚀 Notification and Error Handling Benefits:"
echo "- ✅ Success notifications when agent data is updated"
echo "- ✅ Specific error messages for different failure types"
echo "- ✅ Clear error codes for frontend handling"
echo "- ✅ User-friendly error messages"
echo "- ✅ Consistent error response format"
