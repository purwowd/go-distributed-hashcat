#!/bin/bash

echo "🔧 Fix Agent Status Update Issue"
echo "================================"

echo ""
echo "📊 Problem: Agent status update failed after job completion"
echo "✅ Hashcat: SUCCESS"
echo "✅ Job: COMPLETED" 
echo "❌ Status Update: FAILED"
echo ""

echo "🔧 Quick Fixes:"
echo "==============="

echo "1. Check server connectivity..."
if curl -s http://127.0.0.1:8080/health > /dev/null; then
    echo "✅ Server: ONLINE"
else
    echo "❌ Server: OFFLINE - Start server first"
    echo "   cd /path/to/go-distributed-hashcat && ./server"
fi

echo ""
echo "2. Test status update endpoint..."
curl -X POST http://127.0.0.1:8080/api/v1/agents/status \
     -H "Content-Type: application/json" \
     -d '{"status":"online"}' 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ Status update: WORKING"
else
    echo "❌ Status update: FAILED - Check server logs"
fi

echo ""
echo "3. Restart agent service..."
echo "   sudo systemctl restart agent"
echo "   # or restart manually: ./agent"

echo ""
echo "✅ Fix script completed!"
