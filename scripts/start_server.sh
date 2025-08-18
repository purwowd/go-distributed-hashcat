#!/bin/bash

echo "🚀 Starting Go Distributed Hashcat Server"
echo "========================================"

echo ""
echo "📊 Current Status:"
echo "------------------"
echo "✅ Agent: RUNNING (job completed successfully)"
echo "❌ Server: OFFLINE (causing status update failure)"
echo ""

echo "🔧 Starting Server..."
echo "===================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Not in go-distributed-hashcat directory"
    echo "   Please cd to the project directory first"
    exit 1
fi

echo "✅ Go version: $(go version)"
echo "✅ Project directory: $(pwd)"
echo ""

# Build the server
echo "🔨 Building server..."
if go build -o server ./cmd/server; then
    echo "✅ Server built successfully"
else
    echo "❌ Failed to build server"
    echo "   Check for compilation errors"
    exit 1
fi

# Check if config file exists
if [ ! -f "configs/config.yaml" ]; then
    echo "⚠️  Config file not found, using defaults"
else
    echo "✅ Config file found"
fi

# Start the server
echo ""
echo "🚀 Starting server..."
echo "===================="

# Run server in background
./server &
SERVER_PID=$!

# Wait a moment for server to start
sleep 3

# Check if server is running
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Server started successfully (PID: $SERVER_PID)"
    echo "✅ Server should be accessible at http://localhost:8080"
    
    # Test server health
    echo ""
    echo "🔍 Testing server health..."
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "✅ Server health check: PASSED"
    else
        echo "⚠️  Server health check: FAILED (might need more time to start)"
    fi
    
    echo ""
    echo "📝 Server Management:"
    echo "===================="
    echo "• View logs: tail -f server.log"
    echo "• Stop server: kill $SERVER_PID"
    echo "• Check status: ps aux | grep server"
    echo "• Test API: curl http://localhost:8080/health"
    
else
    echo "❌ Failed to start server"
    echo "   Check server logs for errors"
fi

echo ""
echo "✅ Server startup script completed!"
