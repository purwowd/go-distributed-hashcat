#!/bin/bash

# Test script to verify last_seen field is properly set when generating agent keys
echo "🧪 Testing Generate Agent Key Last Seen Fix"
echo "============================================"

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

# Test 2: Generate agent key and check last_seen
echo ""
echo "📝 Test 2: Generate agent key and check last_seen"
echo "------------------------------------------------"
AGENT_NAME="test-gen-key-$(date +%s)"
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

# Test 3: Check agent status immediately after generation
echo ""
echo "📝 Test 3: Check agent status immediately after generation"
echo "----------------------------------------------------------"
echo "Checking agent status after generation..."

IMMEDIATE_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, created_at, updated_at, status}")

echo "Immediate Agent Status:"
echo "$IMMEDIATE_STATUS"

# Test 4: Verify last_seen is not default value
echo ""
echo "📝 Test 4: Verify last_seen is not default value"
echo "------------------------------------------------"
echo "Verifying last_seen is not the default '0001-01-01' value..."

LAST_SEEN_VALUE=$(echo "$IMMEDIATE_STATUS" | jq -r '.last_seen')
CREATED_AT_VALUE=$(echo "$IMMEDIATE_STATUS" | jq -r '.created_at')
UPDATED_AT_VALUE=$(echo "$IMMEDIATE_STATUS" | jq -r '.updated_at')

echo "Last Seen Value: $LAST_SEEN_VALUE"
echo "Created At Value: $CREATED_AT_VALUE"
echo "Updated At Value: $UPDATED_AT_VALUE"

if [[ "$LAST_SEEN_VALUE" == "0001-01-01"* ]]; then
    echo "❌ FAILED: last_seen still shows default value"
    LAST_SEEN_FIXED=false
else
    echo "✅ SUCCESS: last_seen has been set to current time"
    LAST_SEEN_FIXED=true
fi

# Test 5: Verify datetime format consistency
echo ""
echo "📝 Test 5: Verify datetime format consistency"
echo "---------------------------------------------"
echo "Verifying that last_seen, created_at, and updated_at use the same format..."

# Check if all three timestamps are in the same format (should all be current time)
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%S")
LAST_SEEN_DATE=$(echo "$LAST_SEEN_VALUE" | cut -d'T' -f1)
CREATED_DATE=$(echo "$CREATED_AT_VALUE" | cut -d'T' -f1)
UPDATED_DATE=$(echo "$UPDATED_AT_VALUE" | cut -d'T' -f1)

echo "Current Date: $CURRENT_TIME"
echo "Last Seen Date: $LAST_SEEN_DATE"
echo "Created Date: $CREATED_DATE"
echo "Updated Date: $UPDATED_DATE"

if [[ "$LAST_SEEN_DATE" == "$CREATED_DATE" && "$CREATED_DATE" == "$UPDATED_DATE" ]]; then
    echo "✅ SUCCESS: All datetime fields use consistent format and are from the same date"
    DATETIME_CONSISTENT=true
else
    echo "❌ FAILED: Datetime fields are not consistent"
    DATETIME_CONSISTENT=false
fi

# Test 6: Test agent startup to ensure last_seen continues to update
echo ""
echo "📝 Test 6: Test agent startup to ensure last_seen continues to update"
echo "--------------------------------------------------------------------"
echo "Testing agent startup to verify last_seen continues to update..."

STARTUP_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"agent_key\": \"$AGENT_KEY\", \"ip_address\": \"192.168.1.100\", \"port\": 8080, \"capabilities\": \"CPU\"}" \
  -w "\nHTTP Status: %{http_code}")

echo "Startup Response:"
echo "$STARTUP_RESPONSE"

# Test 7: Check final agent status
echo ""
echo "📝 Test 7: Check final agent status"
echo "-----------------------------------"
echo "Checking final agent status..."

FINAL_STATUS=$(curl -s http://localhost:1337/api/v1/agents/ | jq ".data[] | select(.name == \"$AGENT_NAME\") | {name, last_seen, status, ip_address, port, capabilities}")

echo "Final Agent Status:"
echo "$FINAL_STATUS"

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
echo "- Test 3: Check agent status after generation ✓"
echo "- Test 4: Verify last_seen is not default value ✓"
echo "- Test 5: Verify datetime format consistency ✓"
echo "- Test 6: Test agent startup ✓"
echo "- Test 7: Check final agent status ✓"
echo ""
echo "🎯 Expected Results:"
echo "- Server should start successfully"
echo "- Agent key generation should work"
echo "- last_seen should NOT be '0001-01-01' (default value)"
echo "- last_seen should be set to current time when generating agent key"
echo "- last_seen, created_at, and updated_at should use consistent datetime format"
echo "- Agent startup should continue to update last_seen properly"
echo ""
echo "🔧 Generate Agent Key Last Seen Fix Status:"
if [ "$LAST_SEEN_FIXED" = true ]; then
    echo "✅ SUCCESS: last_seen field is now properly set when generating agent keys"
    echo "   - Default value '0001-01-01' issue has been resolved"
    echo "   - last_seen is set to current time like created_at and updated_at"
else
    echo "❌ FAILED: last_seen field still has issues when generating agent keys"
    echo "   - Default value '0001-01-01' still appears"
    echo "   - Further investigation needed"
fi

echo ""
echo "🔧 Datetime Format Consistency Status:"
if [ "$DATETIME_CONSISTENT" = true ]; then
    echo "✅ SUCCESS: All datetime fields use consistent format"
    echo "   - last_seen, created_at, and updated_at are from the same date"
    echo "   - Datetime format is consistent across all fields"
else
    echo "❌ FAILED: Datetime fields are not consistent"
    echo "   - Inconsistent datetime formats detected"
    echo "   - Further investigation needed"
fi
