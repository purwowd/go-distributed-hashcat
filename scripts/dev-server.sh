#!/bin/bash

# Development server script
echo "🚀 Starting Hashcat Distributed Server in Development Mode..."

# Change to project root directory
cd "$(dirname "$0")/.."

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "📋 Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  No .env file found, using default values"
fi

# Set development environment (override .env if needed)
export GIN_MODE=debug
export SERVER_PORT=1337

# Create necessary directories
mkdir -p data
mkdir -p uploads

# Run the server
go run cmd/server/main.go 
