#!/bin/bash

# Script untuk debug agent status update
# Mengatasi error "Gagal update status agent"

echo "🔍 Debug Agent Status Update Issue"
echo "=================================="

echo ""
echo "📊 Error Analysis:"
echo "------------------"
echo "✅ Hashcat execution: SUCCESS"
echo "✅ Job completion: SUCCESS" 
echo "❌ Agent status update: FAILED"
echo ""

echo "🔧 Step 1: Check Network Connectivity"
echo "===================================="

# Check if agent can reach server
echo "Testing connection to server..."
if ping -c 1 127.0.0.1 > /dev/null 2>&1; then
    echo "✅ Local network: OK"
else
    echo "❌ Local network: FAILED"
fi

# Check if port is accessible (assuming default port 8080)
echo "Testing server port accessibility..."
if nc -z 127.0.0.1 8080 2>/dev/null; then
    echo "✅ Server port 8080: ACCESSIBLE"
else
    echo "❌ Server port 8080: NOT ACCESSIBLE"
    echo "   - Server might be down"
    echo "   - Port might be different"
    echo "   - Firewall blocking connection"
fi

echo ""
echo "🔧 Step 2: Check Agent Configuration"
echo "===================================="

# Check agent config files
echo "Checking agent configuration..."
if [ -f "/root/.agent_config" ]; then
    echo "✅ Agent config file: EXISTS"
    echo "Content:"
    cat /root/.agent_config | head -10
else
    echo "❌ Agent config file: NOT FOUND"
fi

# Check environment variables
echo ""
echo "Environment variables:"
env | grep -i agent | head -5
env | grep -i server | head -5

echo ""
echo "🔧 Step 3: Test API Endpoints"
echo "=============================="

# Test basic HTTP endpoints
echo "Testing API endpoints..."

# Test server health
echo "Testing server health endpoint..."
if curl -s http://127.0.0.1:8080/health > /dev/null 2>&1; then
    echo "✅ Health endpoint: RESPONDING"
else
    echo "❌ Health endpoint: NOT RESPONDING"
fi

# Test agent status update endpoint
echo "Testing agent status update endpoint..."
if curl -s -X POST http://127.0.0.1:8080/api/v1/agents/status \
     -H "Content-Type: application/json" \
     -d '{"status":"online"}' > /dev/null 2>&1; then
    echo "✅ Status update endpoint: RESPONDING"
else
    echo "❌ Status update endpoint: NOT RESPONDING"
fi

echo ""
echo "🔧 Step 4: Check Agent Logs"
echo "============================"

# Check recent agent logs
echo "Recent agent logs (last 20 lines):"
if [ -f "/var/log/agent.log" ]; then
    tail -20 /var/log/agent.log
elif [ -f "/root/agent.log" ]; then
    tail -20 /root/agent.log
else
    echo "❌ Agent log file not found"
    echo "Checking system logs..."
    journalctl -u agent --since "1 hour ago" | tail -10
fi

echo ""
echo "🔧 Step 5: Check Server Status"
echo "=============================="

# Check if server process is running
echo "Checking server process..."
if pgrep -f "go-distributed-hashcat" > /dev/null; then
    echo "✅ Server process: RUNNING"
    ps aux | grep "go-distributed-hashcat" | grep -v grep
else
    echo "❌ Server process: NOT RUNNING"
fi

# Check server logs
echo ""
echo "Server logs (if accessible):"
if [ -f "/var/log/server.log" ]; then
    tail -10 /var/log/server.log
else
    echo "❌ Server log file not found"
fi

echo ""
echo "🔧 Step 6: Manual Status Update Test"
echo "===================================="

# Test manual status update
echo "Testing manual status update..."
AGENT_ID=$(cat /root/.agent_config | grep "agent_id" | cut -d'=' -f2 2>/dev/null)
if [ -n "$AGENT_ID" ]; then
    echo "Agent ID: $AGENT_ID"
    echo "Attempting manual status update..."
    
    curl -v -X PUT "http://127.0.0.1:8080/api/v1/agents/$AGENT_ID/status" \
         -H "Content-Type: application/json" \
         -d '{"status":"online","last_seen":"'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"}'
else
    echo "❌ Agent ID not found in config"
fi

echo ""
echo "🔧 Step 7: Quick Fixes to Try"
echo "=============================="

echo "1. Restart agent service:"
echo "   sudo systemctl restart agent"
echo ""

echo "2. Check server is running:"
echo "   cd /path/to/go-distributed-hashcat"
echo "   ./server"
echo ""

echo "3. Verify agent configuration:"
echo "   - Check server URL/port"
echo "   - Verify agent key"
echo "   - Check network connectivity"
echo ""

echo "4. Test with curl:"
echo "   curl -X POST http://127.0.0.1:8080/api/v1/agents/status \\"
echo "        -H 'Content-Type: application/json' \\"
echo "        -d '{\"status\":\"online\"}'"
echo ""

echo "🎯 Summary:"
echo "==========="
echo "✅ Job execution: SUCCESS"
echo "❌ Status update: FAILED"
echo "🔧 Check: Network, Server, API endpoints, Configuration"
echo ""

echo "✅ Debug completed. Check the output above for issues."
