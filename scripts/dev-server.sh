#!/bin/bash

# Development server script
echo "🚀 Starting Hashcat Distributed Server in Development Mode..."

# Set development environment
export GIN_MODE=debug
export SERVER_PORT=1337

# Create necessary directories
mkdir -p data
mkdir -p uploads

# Run the server
cd "$(dirname "$0")/.."
go run cmd/server/main.go 
