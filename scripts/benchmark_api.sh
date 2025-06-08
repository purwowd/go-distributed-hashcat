#!/bin/bash

echo "🚀 API Performance Benchmark Test"
echo "=================================="

# Test server connectivity first
echo "Testing server connectivity..."
if curl -s http://localhost:1337/health > /dev/null; then
    echo "✅ Server is running on port 1337"
else
    echo "❌ Server not accessible on port 1337"
    exit 1
fi

# Test frontend connectivity
echo "Testing frontend connectivity..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend is running on port 3000"
else
    echo "⚠️  Frontend not accessible on port 3000 (may not be started)"
fi

echo ""
echo "🔧 Testing API Response Times:"
echo "------------------------------"

# Test Agents API
echo -n "GET /api/v1/agents/: "
curl -w "%{time_total}s" -o /dev/null -s http://localhost:1337/api/v1/agents/
echo ""

# Test Jobs API
echo -n "GET /api/v1/jobs/: "
curl -w "%{time_total}s" -o /dev/null -s http://localhost:1337/api/v1/jobs/
echo ""

# Test Hash Files API
echo -n "GET /api/v1/hashfiles/: "
curl -w "%{time_total}s" -o /dev/null -s http://localhost:1337/api/v1/hashfiles/
echo ""

# Test Wordlists API
echo -n "GET /api/v1/wordlists/: "
curl -w "%{time_total}s" -o /dev/null -s http://localhost:1337/api/v1/wordlists/
echo ""

# Test Health API
echo -n "GET /health: "
curl -w "%{time_total}s" -o /dev/null -s http://localhost:1337/health
echo ""

echo ""
echo "📊 Multiple Request Test (5 requests each):"
echo "--------------------------------------------"

# Multiple requests test
for endpoint in "agents" "jobs" "hashfiles" "wordlists"; do
    echo "Testing /api/v1/$endpoint/ (5 requests):"
    for i in {1..5}; do
        TIME=$(curl -w "%{time_total}" -o /dev/null -s http://localhost:1337/api/v1/$endpoint/)
        echo "  Request $i: ${TIME}s"
    done
    echo ""
done

echo ""
echo "🌐 Frontend Integration Test:"
echo "-----------------------------"
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend dashboard accessible"
    echo "   - Dashboard: http://localhost:3000"
    echo "   - API Proxy: Frontend → Backend (port 3000 → 1337)"
else
    echo "ℹ️  To start frontend:"
    echo "   cd frontend && npm install && npm run dev"
fi

echo ""
echo "✅ Benchmark completed!" 
