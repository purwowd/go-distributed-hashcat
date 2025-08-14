#!/bin/bash

# Test Script untuk Agent Validation
# Pastikan server berjalan di localhost:1337

SERVER_URL="http://localhost:1337"
API_BASE="$SERVER_URL/api/v1"

echo "🧪 Testing Agent Validation Logic"
echo "=================================="
echo ""

# Colors untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function untuk print status
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "FAILED" ]; then
        echo -e "${RED}❌ $message${NC}"
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ️  $message${NC}"
    elif [ "$status" = "WARNING" ]; then
        echo -e "${YELLOW}⚠️  $message${NC}"
    fi
}

# Function untuk test API
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    
    echo "Testing: $test_name"
    echo "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    echo "Response Code: $http_code"
    echo "Response Body: $response_body"
    
    if [ "$http_code" = "$expected_status" ]; then
        print_status "SUCCESS" "Test passed: Expected $expected_status, got $http_code"
    else
        print_status "FAILED" "Test failed: Expected $expected_status, got $http_code"
    fi
    
    echo ""
    echo "---"
    echo ""
}

# Test 1: Generate Agent Key
echo "📋 Test 1: Generate Agent Key"
print_status "INFO" "Creating agent key for 'TestAgent'"
test_api "POST" "$API_BASE/agent-keys" '{"name": "TestAgent"}' "201" "Generate Agent Key"

# Extract agent key from response (assuming response format)
AGENT_KEY=$(echo "$response_body" | grep -o '"agent_key":"[^"]*"' | cut -d'"' -f4)
if [ -z "$AGENT_KEY" ]; then
    AGENT_KEY="test_key_123" # Fallback untuk testing
    print_status "WARNING" "Could not extract agent key, using fallback: $AGENT_KEY"
fi

print_status "INFO" "Generated Agent Key: $AGENT_KEY"
echo ""

# Test 2: Create Agent dengan Agent Key yang Valid
echo "📋 Test 2: Create Agent dengan Agent Key yang Valid"
print_status "INFO" "Creating agent with valid agent key and matching name"
test_api "POST" "$API_BASE/agents" "{
    \"name\": \"TestAgent\",
    \"ip_address\": \"192.168.1.100\",
    \"port\": 8081,
    \"capabilities\": \"NVIDIA RTX 4090\",
    \"agent_key\": \"$AGENT_KEY\"
}" "201" "Create Agent with Valid Key and Matching Name"

# Test 3: Create Agent dengan Agent Key yang Tidak Valid
echo "📋 Test 3: Create Agent dengan Agent Key yang Tidak Valid"
print_status "INFO" "Testing with invalid agent key"
test_api "POST" "$API_BASE/agents" "{
    \"name\": \"TestAgent2\",
    \"ip_address\": \"192.168.1.101\",
    \"port\": 8082,
    \"capabilities\": \"NVIDIA RTX 4080\",
    \"agent_key\": \"invalid_key_123\"
}" "400" "Create Agent with Invalid Key"

# Test 4: Create Agent dengan Nama yang Tidak Sesuai
echo "📋 Test 4: Create Agent dengan Nama yang Tidak Sesuai"
print_status "INFO" "Testing with mismatched agent name"
test_api "POST" "$API_BASE/agents" "{
    \"name\": \"WrongName\",
    \"ip_address\": \"192.168.1.102\",
    \"port\": 8083,
    \"capabilities\": \"NVIDIA RTX 4070\",
    \"agent_key\": \"$AGENT_KEY\"
}" "400" "Create Agent with Mismatched Name"

# Test 5: Create Agent dengan IP Address yang Sudah Digunakan
echo "📋 Test 5: Create Agent dengan IP Address yang Sudah Digunakan"
print_status "INFO" "Testing with duplicate IP address"
test_api "POST" "$API_BASE/agents" "{
    \"name\": \"TestAgent3\",
    \"ip_address\": \"192.168.1.100\",
    \"port\": 8084,
    \"capabilities\": \"NVIDIA RTX 4060\",
    \"agent_key\": \"$AGENT_KEY\"
}" "409" "Create Agent with Duplicate IP"

# Test 6: Create Agent tanpa Agent Key
echo "📋 Test 6: Create Agent tanpa Agent Key"
print_status "INFO" "Testing without agent key"
test_api "POST" "$API_BASE/agents" "{
    \"name\": \"TestAgent4\",
    \"ip_address\": \"192.168.1.103\",
    \"port\": 8085,
    \"capabilities\": \"NVIDIA RTX 4050\"
}" "400" "Create Agent without Agent Key"

echo "🎯 Testing Complete!"
echo ""
print_status "INFO" "Summary:"
echo "- Test 1: Generate Agent Key (Expected: 201)"
echo "- Test 2: Valid Agent Creation (Expected: 201)"
echo "- Test 3: Invalid Agent Key (Expected: 400)"
echo "- Test 4: Mismatched Name (Expected: 400)"
echo "- Test 5: Duplicate IP (Expected: 409)"
echo "- Test 6: Missing Agent Key (Expected: 400)"
echo ""
print_status "INFO" "Check the responses above to verify validation logic is working correctly."
