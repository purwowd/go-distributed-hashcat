#!/bin/bash

# Test script for Error Message Format
echo "🧪 Testing Error Message Format"
echo "================================"

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Try to create agent with wrong name for existing agent key
echo ""
echo "📝 Test 1: Try to create agent with wrong name for existing agent key"
echo "-------------------------------------------------------------------"
echo "Expected: Agent name 'wrong-name' does not match the name associated with agent key 'd8675fb7' (expected: 'test-agent-003')"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d '{"name": "wrong-name", "agent_key": "d8675fb7"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 2: Try to create agent with non-existent agent key
echo ""
echo "📝 Test 2: Try to create agent with non-existent agent key"
echo "-----------------------------------------------------------"
echo "Expected: Agent key 'invalid-key' not found in database. Please generate a valid agent key first"
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-004", "agent_key": "invalid-key"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Test 3: Try to generate agent key with duplicate name
echo ""
echo "📝 Test 3: Try to generate agent key with duplicate name"
echo "--------------------------------------------------------"
echo "Expected: An agent with this name already exists."
echo "Actual:"
curl -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-003"}' \
  -w "\nHTTP Status: %{http_code}\n" | jq '.'

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 Summary:"
echo "- Test 1: Should show clean error message without HTTP status prefix"
echo "- Test 2: Should show clean error message without HTTP status prefix"
echo "- Test 3: Should show clean error message without HTTP status prefix"
