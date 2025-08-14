#!/bin/bash

# Test script for Generate Agent Key endpoint
echo "🧪 Testing Generate Agent Key Endpoint"
echo "======================================"

# Start the server in background
echo "🚀 Starting server..."
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test 1: Generate new agent key
echo ""
echo "📝 Test 1: Generate new agent key"
echo "--------------------------------"
curl -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-001"}' \
  -w "\nHTTP Status: %{http_code}\n"

# Test 2: Try to generate agent key with same name (should fail)
echo ""
echo "📝 Test 2: Try to generate agent key with same name (should fail)"
echo "----------------------------------------------------------------"
curl -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-001"}' \
  -w "\nHTTP Status: %{http_code}\n"

# Test 3: Generate another agent key
echo ""
echo "📝 Test 3: Generate another agent key"
echo "------------------------------------"
curl -X POST http://localhost:1337/api/v1/agents/generate-key \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-002"}' \
  -w "\nHTTP Status: %{http_code}\n"

# Test 4: List all agents to see the generated keys
echo ""
echo "📝 Test 4: List all agents"
echo "--------------------------"
curl -X GET http://localhost:1337/api/v1/agents/ \
  -w "\nHTTP Status: %{http_code}\n"

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ Test completed!"
