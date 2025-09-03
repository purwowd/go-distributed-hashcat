#!/bin/bash

# Test Script for Speed Reset Endpoint
# Tests if the /speed-reset endpoint is working correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="https://0229f6ee72a3.ngrok-free.app"
AGENT_NAME="test-speed-reset-agent"
AGENT_KEY="test-speed-reset-key-$(date +%s)"

echo -e "${BLUE}🧪 Testing Speed Reset Endpoint${NC}"
echo "====================================="
echo "Server URL: $SERVER_URL"
echo "Agent Name: $AGENT_NAME"
echo "Agent Key: $AGENT_KEY"
echo ""

# Function to log with timestamp
log() {
    echo -e "[$(date '+%H:%M:%S')] $1"
}

# Function to check if server is running
check_server() {
    log "${BLUE}🔍 Checking server connectivity...${NC}"
    if curl -s "$SERVER_URL/health" > /dev/null; then
        log "${GREEN}✅ Server is running${NC}"
    else
        log "${RED}❌ Server is not running or not accessible${NC}"
        exit 1
    fi
}

# Function to create agent
create_agent() {
    log "${BLUE}🔧 Creating test agent...${NC}"
    
    response=$(curl -s -X POST "$SERVER_URL/api/v1/agents/generate-key" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$AGENT_NAME\"}")
    
    if echo "$response" | grep -q "agent_key"; then
        log "${GREEN}✅ Agent created successfully${NC}"
        echo "$response" | jq -r '.data.agent_key' > /tmp/speed_reset_agent_key.txt
    else
        log "${RED}❌ Failed to create agent: $response${NC}"
        exit 1
    fi
}

# Function to start agent
start_agent() {
    log "${BLUE}🚀 Starting test agent...${NC}"
    
    AGENT_KEY=$(cat /tmp/speed_reset_agent_key.txt)
    
    response=$(curl -s -X POST "$SERVER_URL/api/v1/agents/startup" \
        -H "Content-Type: application/json" \
        -d "{
            \"agent_key\": \"$AGENT_KEY\",
            \"ip_address\": \"192.168.1.200\",
            \"port\": 8080,
            \"capabilities\": \"hashcat,benchmark\"
        }")
    
    if echo "$response" | grep -q "successfully registered"; then
        log "${GREEN}✅ Agent started successfully${NC}"
        # Extract agent ID
        echo "$response" | jq -r '.data.id' > /tmp/speed_reset_agent_id.txt
    else
        log "${RED}❌ Failed to start agent: $response${NC}"
        exit 1
    fi
}

# Function to get agent info
get_agent_info() {
    AGENT_ID=$(cat /tmp/speed_reset_agent_id.txt)
    
    response=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID")
    
    if echo "$response" | grep -q "id"; then
        log "${GREEN}✅ Agent info retrieved${NC}"
        SPEED=$(echo "$response" | jq -r '.data.speed')
        STATUS=$(echo "$response" | jq -r '.data.status')
        log "${YELLOW}📊 Current Speed: $SPEED H/s${NC}"
        log "${YELLOW}📊 Current Status: $STATUS${NC}"
    else
        log "${RED}❌ Failed to get agent info: $response${NC}"
        exit 1
    fi
}

# Function to test speed reset endpoint
test_speed_reset() {
    log "${BLUE}🔄 Testing speed reset endpoint...${NC}"
    
    AGENT_ID=$(cat /tmp/speed_reset_agent_id.txt)
    
    # First set a non-zero speed
    log "${BLUE}📝 Setting speed to 5000 H/s...${NC}"
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed" \
        -H "Content-Type: application/json" \
        -d "{\"speed\": 5000}")
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ Speed set to 5000 H/s${NC}"
    else
        log "${RED}❌ Failed to set speed: $response${NC}"
        return 1
    fi
    
    # Verify speed was set
    sleep 2
    CURRENT_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    log "${YELLOW}📊 Speed after setting: $CURRENT_SPEED H/s${NC}"
    
    if [ "$CURRENT_SPEED" != "5000" ]; then
        log "${RED}❌ Speed was not set correctly. Expected: 5000, Got: $CURRENT_SPEED${NC}"
        return 1
    fi
    
    # Now test speed reset endpoint
    log "${BLUE}🔄 Calling speed-reset endpoint...${NC}"
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed-reset")
    
    log "${BLUE}📡 Speed reset response: $response${NC}"
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ Speed reset endpoint call successful${NC}"
    else
        log "${RED}❌ Speed reset endpoint call failed: $response${NC}"
        return 1
    fi
    
    # Wait a moment and verify speed was reset
    sleep 3
    RESET_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    log "${YELLOW}📊 Speed after reset: $RESET_SPEED H/s${NC}"
    
    if [ "$RESET_SPEED" = "0" ]; then
        log "${GREEN}✅ Speed successfully reset to 0${NC}"
    else
        log "${RED}❌ Speed was not reset to 0. Current speed: $RESET_SPEED${NC}"
        return 1
    fi
}

# Function to test speed reset with different HTTP methods
test_http_methods() {
    log "${BLUE}🔧 Testing different HTTP methods for speed reset...${NC}"
    
    AGENT_ID=$(cat /tmp/speed_reset_agent_id.txt)
    
    # Test GET method (should fail)
    log "${BLUE}📝 Testing GET method (should fail)...${NC}"
    response=$(curl -s -X GET "$SERVER_URL/api/v1/agents/$AGENT_ID/speed-reset")
    log "${YELLOW}📡 GET response: $response${NC}"
    
    # Test POST method (should fail)
    log "${BLUE}📝 Testing POST method (should fail)...${NC}"
    response=$(curl -s -X POST "$SERVER_URL/api/v1/agents/$AGENT_ID/speed-reset")
    log "${YELLOW}📡 POST response: $response${NC}"
    
    # Test PUT method (should work)
    log "${BLUE}📝 Testing PUT method (should work)...${NC}"
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed-reset")
    log "${YELLOW}📡 PUT response: $response${NC}"
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ PUT method working correctly${NC}"
    else
        log "${RED}❌ PUT method not working: $response${NC}"
    fi
}

# Function to cleanup
cleanup() {
    log "${BLUE}🧹 Cleaning up test data...${NC}"
    
    # Remove temporary files
    rm -f /tmp/speed_reset_agent_key.txt /tmp/speed_reset_agent_id.txt
    
    log "${GREEN}✅ Cleanup completed${NC}"
}

# Main test execution
main() {
    echo -e "${BLUE}🚀 Starting Speed Reset Endpoint Tests${NC}"
    echo ""
    
    check_server
    create_agent
    start_agent
    get_agent_info
    
    echo ""
    log "${BLUE}⏳ Waiting 5 seconds for agent to stabilize...${NC}"
    sleep 5
    
    test_speed_reset
    test_http_methods
    
    echo ""
    echo -e "${GREEN}🎉 Speed Reset Endpoint Tests Completed!${NC}"
    echo "================================================"
    echo ""
    echo -e "${GREEN}✅ Test Results:${NC}"
    echo "   ✅ Speed reset endpoint accessible"
    echo "   ✅ Speed can be set to non-zero value"
    echo "   ✅ Speed reset endpoint resets speed to 0"
    echo "   ✅ PUT method works correctly"
    echo ""
    echo -e "${GREEN}🚀 The speed reset endpoint is working correctly!${NC}"
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
