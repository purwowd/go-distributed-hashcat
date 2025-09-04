#!/bin/bash

# Script untuk restart server dan test fitur real-time speed monitoring

echo "Restarting Server and Testing Real-Time Speed Monitoring"
echo "============================================================"

# Check if server is running
echo "🔍 Checking current server status..."
if curl -s "http://localhost:1337/health" > /dev/null 2>&1; then
    echo "✅ Server is currently running"
    
    # Get server PID
    SERVER_PID=$(lsof -ti:1337)
    if [ ! -z "$SERVER_PID" ]; then
        echo "🛑 Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID
        sleep 2
        
        # Check if server stopped
        if curl -s "http://localhost:1337/health" > /dev/null 2>&1; then
            echo "⚠️  Server still running, force killing..."
            kill -9 $SERVER_PID
            sleep 1
        fi
    fi
else
    echo "ℹ️  Server is not running"
fi

# Build server
echo "🔨 Building server..."
make build-server

# Start server in background
echo "🚀 Starting server in background..."
./bin/server > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
echo "⏳ Waiting for server to start..."
for i in {1..10}; do
    if curl -s "http://localhost:1337/health" > /dev/null 2>&1; then
        echo "✅ Server started successfully (PID: $SERVER_PID)"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo "❌ Server failed to start after 10 attempts"
        echo "Server logs:"
        tail -20 server.log
        exit 1
    fi
    
    echo "  Attempt $i/10..."
    sleep 2
done

# Test new endpoints
echo ""
echo "🧪 Testing new real-time endpoints..."

# Create test agent
echo "📝 Creating test agent..."
CREATE_RESPONSE=$(curl -s -X POST "http://localhost:1337/api/v1/agents/generate-key" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-realtime-restart"}')

AGENT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "✅ Test agent created with ID: $AGENT_ID"

# Test speed-status endpoint
echo "Testing speed-status endpoint..."
SPEED_STATUS_RESPONSE=$(curl -s -X PUT "http://localhost:1337/api/v1/agents/$AGENT_ID/speed-status" \
    -H "Content-Type: application/json" \
    -d '{"speed": 5000, "status": "online"}')

if [ $? -eq 0 ]; then
    echo "✅ speed-status endpoint working: $SPEED_STATUS_RESPONSE"
else
    echo "❌ speed-status endpoint failed: $SPEED_STATUS_RESPONSE"
fi

# Test speed-reset endpoint
echo "Testing speed-reset endpoint..."
SPEED_RESET_RESPONSE=$(curl -s -X PUT "http://localhost:1337/api/v1/agents/$AGENT_ID/speed-reset" \
    -H "Content-Type: application/json")

if [ $? -eq 0 ]; then
    echo "✅ speed-reset endpoint working: $SPEED_RESET_RESPONSE"
else
    echo "❌ speed-reset endpoint failed: $SPEED_RESET_RESPONSE"
fi

# Clean up
echo "🧹 Cleaning up test agent..."
curl -s -X DELETE "http://localhost:1337/api/v1/agents/$AGENT_ID" > /dev/null

echo ""
echo "🎯 Testing completed!"
echo "Server is running with PID: $SERVER_PID"
echo "Server logs: server.log"
echo ""
echo "To stop server: kill $SERVER_PID"
echo "To view logs: tail -f server.log"
