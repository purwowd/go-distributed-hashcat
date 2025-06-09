#!/bin/bash

# Simple Deployment Script for Go Distributed Hashcat
# This script builds and deploys both backend and frontend

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Deployment Process${NC}"
echo "=================================="

# Build backend
echo -e "${YELLOW}📦 Building Backend...${NC}"
go build -o bin/server ./cmd/server
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Backend built successfully${NC}"
else
    echo -e "${RED}❌ Backend build failed${NC}"
    exit 1
fi

# Build agent (optional)
echo -e "${YELLOW}📦 Building Agent...${NC}"
go build -o bin/agent ./cmd/agent
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Agent built successfully${NC}"
else
    echo -e "${RED}❌ Agent build failed${NC}"
    exit 1
fi

# Build frontend
echo -e "${YELLOW}📦 Building Frontend...${NC}"
cd frontend
npm ci --silent
npm run build:prod
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Frontend built successfully${NC}"
    cd ..
else
    echo -e "${RED}❌ Frontend build failed${NC}"
    exit 1
fi

# Create deployment package
echo -e "${YELLOW}📋 Creating deployment package...${NC}"
mkdir -p deploy
cp bin/server deploy/
cp bin/agent deploy/
cp -r frontend/dist deploy/frontend
cp -r configs deploy/
cp .env-example deploy/

echo -e "${GREEN}✅ Deployment package created in ./deploy/${NC}"
echo ""
echo -e "${BLUE}Deployment Information:${NC}"
echo "📁 Backend binary: deploy/server"
echo "📁 Agent binary: deploy/agent" 
echo "📁 Frontend files: deploy/frontend/"
echo "📁 Config files: deploy/configs/"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Copy deploy/ directory to your server"
echo "2. Set environment variables (copy from .env-example)"
echo "3. Run: ./server"
echo "4. Configure web server to serve frontend/ directory" 
