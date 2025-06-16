#!/bin/bash

# Test script untuk menjalankan agent dengan agent key
# Usage: ./test_agent.sh

set -e

echo "🚀 Testing Agent with X-Agent-Key Authentication"
echo "================================================"

# Check if server is running
echo "📡 Checking if server is running..."
if ! curl -s http://localhost:1337/health > /dev/null; then
    echo "❌ Server is not running. Please start server first:"
    echo "   ./bin/server"
    exit 1
fi
echo "✅ Server is running"

# Generate agent key via API
echo "🔑 Generating agent key..."
AGENT_KEY_RESPONSE=$(curl -s -X POST http://localhost:1337/api/v1/agent-keys/generate \
    -H "Content-Type: application/json" \
    -d '{"name":"test-agent","description":"Test agent for development"}')

if [ $? -ne 0 ]; then
    echo "❌ Failed to generate agent key"
    exit 1
fi

AGENT_KEY=$(echo $AGENT_KEY_RESPONSE | jq -r '.agent_key')

if [ "$AGENT_KEY" = "null" ] || [ -z "$AGENT_KEY" ]; then
    echo "❌ Failed to extract agent key from response:"
    echo "$AGENT_KEY_RESPONSE"
    exit 1
fi

echo "✅ Agent key generated: ${AGENT_KEY:0:16}..."

# Set environment variables
export AGENT_KEY="$AGENT_KEY"
export SERVER_URL="http://localhost:1337"
export AGENT_NAME="test-agent-$(date +%s)"

echo "🔧 Environment variables set:"
echo "   AGENT_KEY: ${AGENT_KEY:0:16}..."
echo "   SERVER_URL: $SERVER_URL"
echo "   AGENT_NAME: $AGENT_NAME"

# Create test upload directory
TEST_UPLOAD_DIR="/tmp/hashcat-test-uploads"
mkdir -p "$TEST_UPLOAD_DIR"
echo "📁 Created test upload directory: $TEST_UPLOAD_DIR"

echo ""
echo "🎯 Starting agent..."
echo "   Press Ctrl+C to stop the agent"
echo ""

# Run agent
./bin/agent \
    --capabilities "GPU,CPU" \
    --upload-dir "$TEST_UPLOAD_DIR" \
    --ip "127.0.0.1" \
    --port 8081

echo ""
echo "🛑 Agent stopped"

# Cleanup
echo "🧹 Cleaning up test directory..."
rm -rf "$TEST_UPLOAD_DIR"
echo "✅ Cleanup complete" 
