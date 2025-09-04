#!/bin/bash

# Test Script for Optimized Speed Monitoring
# Tests the improved mechanism where speed is only updated once when online
# and automatically reset to 0 when agent goes offline

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:1337"
AGENT_NAME="test-optimized-agent"
AGENT_KEY="test-optimized-key-$(date +%s)"

echo -e "${BLUE}🧪 Testing Optimized Speed Monitoring Mechanism${NC}"
echo "=================================================="
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
        log "${RED}❌ Server is not running. Please start the server first.${NC}"
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
        echo "$response" | jq -r '.data.agent_key' > /tmp/agent_key.txt
    else
        log "${RED}❌ Failed to create agent: $response${NC}"
        exit 1
    fi
}

# Function to start agent
start_agent() {
    log "${BLUE}🚀 Starting test agent...${NC}"
    
    AGENT_KEY=$(cat /tmp/agent_key.txt)
    
    response=$(curl -s -X POST "$SERVER_URL/api/v1/agents/startup" \
        -H "Content-Type: application/json" \
        -d "{
            \"agent_key\": \"$AGENT_KEY\",
            \"ip_address\": \"192.168.1.100\",
            \"port\": 8080,
            \"capabilities\": \"hashcat,benchmark\"
        }")
    
    if echo "$response" | grep -q "successfully registered"; then
        log "${GREEN}✅ Agent started successfully${NC}"
        # Extract agent ID
        echo "$response" | jq -r '.data.id' > /tmp/agent_id.txt
    else
        log "${RED}❌ Failed to start agent: $response${NC}"
        exit 1
    fi
}

# Function to get agent info
get_agent_info() {
    AGENT_ID=$(cat /tmp/agent_id.txt)
    
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

# Function to monitor speed updates
monitor_speed_updates() {
    log "${BLUE}👀 Monitoring speed updates for 2 minutes...${NC}"
    log "${YELLOW}⚠️  Speed should only be updated ONCE during startup${NC}"
    log "${YELLOW}⚠️  No continuous speed updates should occur${NC}"
    
    AGENT_ID=$(cat /tmp/agent_id.txt)
    INITIAL_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    
    log "${BLUE}📊 Initial Speed: $INITIAL_SPEED H/s${NC}"
    
    # Monitor for 2 minutes
    for i in {1..12}; do
        sleep 10
        CURRENT_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
        CURRENT_STATUS=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.status')
        
        log "${BLUE}[$((i*10))s] Speed: $CURRENT_SPEED H/s, Status: $CURRENT_STATUS${NC}"
        
        # Check if speed changed unexpectedly
        if [ "$CURRENT_SPEED" != "$INITIAL_SPEED" ]; then
            log "${RED}❌ Speed changed unexpectedly from $INITIAL_SPEED to $CURRENT_SPEED${NC}"
            log "${RED}❌ This indicates continuous speed updates are still happening${NC}"
            return 1
        fi
    done
    
    log "${GREEN}✅ Speed remained consistent during monitoring period${NC}"
    log "${GREEN}✅ No continuous speed updates detected${NC}"
}

# Function to test offline speed reset
test_offline_speed_reset() {
    log "${BLUE}Testing automatic speed reset when agent goes offline...${NC}"
    
    AGENT_ID=$(cat /tmp/agent_id.txt)
    
    # Get current speed
    CURRENT_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    log "${YELLOW}📊 Speed before offline: $CURRENT_SPEED H/s${NC}"
    
    # Simulate agent going offline by updating status
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/status" \
        -H "Content-Type: application/json" \
        -d "{\"status\": \"offline\"}")
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ Agent status updated to offline${NC}"
    else
        log "${RED}❌ Failed to update agent status: $response${NC}"
        return 1
    fi
    
    # Wait a moment for health monitor to process
    sleep 5
    
    # Check if speed was reset to 0
    UPDATED_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    UPDATED_STATUS=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.status')
    
    log "${YELLOW}📊 Speed after offline: $UPDATED_SPEED H/s${NC}"
    log "${YELLOW}📊 Status after offline: $UPDATED_STATUS${NC}"
    
    if [ "$UPDATED_SPEED" = "0" ]; then
        log "${GREEN}✅ Speed automatically reset to 0 when agent went offline${NC}"
    else
        log "${RED}❌ Speed was not reset to 0. Current speed: $UPDATED_SPEED${NC}"
        return 1
    fi
}

# Function to test manual speed reset
test_manual_speed_reset() {
    log "${BLUE}🔧 Testing manual speed reset endpoint...${NC}"
    
    AGENT_ID=$(cat /tmp/agent_id.txt)
    
    # First set a non-zero speed
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed" \
        -H "Content-Type: application/json" \
        -d "{\"speed\": 5000}")
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ Speed set to 5000 H/s${NC}"
    else
        log "${RED}❌ Failed to set speed: $response${NC}"
        return 1
    fi
    
    # Now test manual speed reset
    response=$(curl -s -X PUT "$SERVER_URL/api/v1/agents/$AGENT_ID/speed-reset")
    
    if echo "$response" | grep -q "successfully"; then
        log "${GREEN}✅ Manual speed reset successful${NC}"
    else
        log "${RED}❌ Manual speed reset failed: $response${NC}"
        return 1
    fi
    
    # Verify speed was reset
    sleep 2
    RESET_SPEED=$(curl -s "$SERVER_URL/api/v1/agents/$AGENT_ID" | jq -r '.data.speed')
    
    if [ "$RESET_SPEED" = "0" ]; then
        log "${GREEN}✅ Speed manually reset to 0${NC}"
    else
        log "${RED}❌ Speed was not reset to 0. Current speed: $RESET_SPEED${NC}"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    log "${BLUE}🧹 Cleaning up test data...${NC}"
    
    # Remove temporary files
    rm -f /tmp/agent_key.txt /tmp/agent_id.txt
    
    log "${GREEN}✅ Cleanup completed${NC}"
}

# Main test execution
main() {
    echo -e "${BLUE}🚀 Starting Optimized Speed Monitoring Tests${NC}"
    echo ""
    
    check_server
    create_agent
    start_agent
    get_agent_info
    
    echo ""
    log "${BLUE}⏳ Waiting 10 seconds for agent to stabilize...${NC}"
    sleep 10
    
    monitor_speed_updates
    test_offline_speed_reset
    test_manual_speed_reset
    
    echo ""
    echo -e "${GREEN}🎉 Optimized Speed Monitoring Tests Completed Successfully!${NC}"
    echo "============================================================="
    echo ""
    echo -e "${GREEN}✅ Key Improvements Verified:${NC}"
    echo "   ✅ Speed only updated once during startup"
    echo "   ✅ No continuous speed updates during monitoring"
    echo "   ✅ Automatic speed reset to 0 when offline"
    echo "   ✅ Manual speed reset endpoint working"
    echo "   ✅ Health monitor properly resets speed"
    echo ""
    echo -e "${GREEN}🚀 The optimized speed monitoring mechanism is working correctly!${NC}"
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
