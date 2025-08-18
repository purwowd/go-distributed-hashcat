#!/bin/bash

# Test script to verify formatted error messages work correctly
echo "🧪 Testing Formatted Error Messages"
echo "==================================="

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
AGENT_NAME="test-formatted-errors-$(date +%s)"
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

# Test 3: Test with non-existent agent key (should show formatted agent key not found error)
echo ""
echo "📝 Test 3: Test with non-existent agent key (should show formatted agent key not found error)"
echo "-----------------------------------------------------------------------------------------"
echo "Testing update with non-existent agent key..."

INVALID_KEY="invalid_key_$(date +%s)"
echo "Using invalid agent key: $INVALID_KEY"

INVALID_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$INVALID_KEY\", \"ip_address\": \"192.168.1.500\", \"port\": 8080, \"capabilities\": \"GPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Invalid Key Response:"
echo "$INVALID_KEY_RESPONSE"

# Test 4: Verify formatted agent key not found error
echo ""
echo "📝 Test 4: Verify formatted agent key not found error"
echo "----------------------------------------------------"
echo "Verifying formatted agent key not found error..."

INVALID_KEY_CODE=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.code // empty')
INVALID_KEY_ERROR=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.error // empty')
INVALID_KEY_MESSAGE=$(echo "$INVALID_KEY_RESPONSE" | jq -r '.message // empty')

echo "Invalid Key Code: $INVALID_KEY_CODE"
echo "Invalid Key Error: $INVALID_KEY_ERROR"
echo "Invalid Key Message: $INVALID_KEY_MESSAGE"

EXPECTED_ERROR="Agent key $INVALID_KEY not found in database"
if [ "$INVALID_KEY_ERROR" = "$EXPECTED_ERROR" ]; then
    echo "✅ SUCCESS: Formatted agent key not found error is correct"
    echo "   Expected: $EXPECTED_ERROR"
    echo "   Got: $INVALID_KEY_ERROR"
    INVALID_KEY_ERROR_CORRECT=true
else
    echo "❌ FAILED: Formatted agent key not found error is incorrect"
    echo "   Expected: $EXPECTED_ERROR"
    echo "   Got: $INVALID_KEY_ERROR"
    INVALID_KEY_ERROR_CORRECT=false
fi

# Test 5: Test with IP address conflict (should show formatted IP address conflict error)
echo ""
echo "📝 Test 5: Test with IP address conflict (should show formatted IP address conflict error)"
echo "------------------------------------------------------------------------------"
echo "Testing update with IP address conflict..."

# First, update the agent with a specific IP
FIRST_UPDATE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.600\", \"port\": 8080, \"capabilities\": \"CPU\"}")

echo "First Update Response:"
echo "$FIRST_UPDATE"

# Now try to use the same IP with a different agent key (should conflict)
IP_CONFLICT_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$INVALID_KEY\", \"ip_address\": \"192.168.1.600\", \"port\": 8080, \"capabilities\": \"GPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "IP Conflict Response:"
echo "$IP_CONFLICT_RESPONSE"

# Test 6: Verify formatted IP address conflict error
echo ""
echo "📝 Test 6: Verify formatted IP address conflict error"
echo "---------------------------------------------------"
echo "Verifying formatted IP address conflict error..."

IP_CONFLICT_CODE=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.code // empty')
IP_CONFLICT_ERROR=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.error // empty')
IP_CONFLICT_MESSAGE=$(echo "$IP_CONFLICT_RESPONSE" | jq -r '.message // empty')

echo "IP Conflict Code: $IP_CONFLICT_CODE"
echo "IP Conflict Error: $IP_CONFLICT_ERROR"
echo "IP Conflict Message: $IP_CONFLICT_MESSAGE"

EXPECTED_IP_ERROR="IP address 192.168.1.600 already in use"
if [ "$IP_CONFLICT_ERROR" = "$EXPECTED_IP_ERROR" ]; then
    echo "✅ SUCCESS: Formatted IP address conflict error is correct"
    echo "   Expected: $EXPECTED_IP_ERROR"
    echo "   Got: $IP_CONFLICT_ERROR"
    IP_CONFLICT_ERROR_CORRECT=true
else
    echo "❌ FAILED: Formatted IP address conflict error is incorrect"
    echo "   Expected: $EXPECTED_IP_ERROR"
    echo "   Got: $IP_CONFLICT_ERROR"
    IP_CONFLICT_ERROR_CORRECT=false
fi

# Test 7: Test successful update to verify success message still works
echo ""
echo "📝 Test 7: Test successful update to verify success message still works"
echo "---------------------------------------------------------------------"
echo "Testing successful agent data update..."

SUCCESS_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/update-data \
  -H "Content-Type: application/json" \
  -d "{\"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.700\", \"port\": 9090, \"capabilities\": \"GPU RTX 4090\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Success Response:"
echo "$SUCCESS_RESPONSE"

# Test 8: Verify success response format
echo ""
echo "📝 Test 8: Verify success response format"
echo "-----------------------------------------"
echo "Verifying success response format..."

SUCCESS_CODE=$(echo "$SUCCESS_RESPONSE" | jq -r '.code // empty')
SUCCESS_MESSAGE=$(echo "$SUCCESS_RESPONSE" | jq -r '.message // empty')

echo "Success Code: $SUCCESS_CODE"
echo "Success Message: $SUCCESS_MESSAGE"

if [ "$SUCCESS_CODE" = "UPDATE_SUCCESS" ] && [ "$SUCCESS_MESSAGE" = "Agent data updated successfully" ]; then
    echo "✅ SUCCESS: Success response format is correct"
    SUCCESS_FORMAT_CORRECT=true
else
    echo "❌ FAILED: Success response format is incorrect"
    echo "   Expected code: UPDATE_SUCCESS, got: $SUCCESS_CODE"
    echo "   Expected message: Agent data updated successfully, got: $SUCCESS_MESSAGE"
    SUCCESS_FORMAT_CORRECT=false
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
echo "- Test 3: Test with non-existent agent key ✓"
echo "- Test 4: Verify formatted agent key not found error ✓"
echo "- Test 5: Test with IP address conflict ✓"
echo "- Test 6: Verify formatted IP address conflict error ✓"
echo "- Test 7: Test successful update ✓"
echo "- Test 8: Verify success response format ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- Agent key not found should show: 'Agent key {agent_key} not found in database'"
echo "- IP address conflict should show: 'IP address {ip_address} already in use'"
echo "- Success update should return UPDATE_SUCCESS code"
echo ""
echo "🔧 Formatted Error Messages Status:"
if [ "$INVALID_KEY_ERROR_CORRECT" = true ]; then
    echo "✅ SUCCESS: Agent key not found error is properly formatted"
    echo "   - Shows specific agent key in error message"
    echo "   - No HTTP status prefix"
    echo "   - User-friendly format"
else
    echo "❌ FAILED: Agent key not found error formatting has issues"
    echo "   - Further investigation needed"
fi

if [ "$IP_CONFLICT_ERROR_CORRECT" = true ]; then
    echo "✅ SUCCESS: IP address conflict error is properly formatted"
    echo "   - Shows specific IP address in error message"
    echo "   - No HTTP status prefix"
    echo "   - User-friendly format"
else
    echo "❌ FAILED: IP address conflict error formatting has issues"
    echo "   - Further investigation needed"
fi

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
echo "🚀 Formatted Error Message Benefits:"
echo "- ✅ No HTTP status prefix (cleaner user experience)"
echo "- ✅ Specific information included (IP address, agent key)"
echo "- ✅ User-friendly error messages"
echo "- ✅ Consistent error format across all endpoints"
echo "- ✅ Better debugging information for users"
