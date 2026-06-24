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
mkdir -p uploads/wordlists uploads/hash-files uploads/temp

# uploads/ must be writable by this user (wordlist/hash uploads fail with 500 otherwise).
if [ -d uploads/wordlists ] && [ ! -w uploads/wordlists ]; then
	echo "⚠️  uploads/ is not writable by $(whoami) (often root-owned from a past sudo run)."
	echo "    Fix uploads, then restart:"
	echo "    sudo chown -R $(whoami):$(whoami) \"$(pwd)/uploads\""
	exit 1
fi

# Run the server (prefer compiled binary when present)
if [ -x ./bin/server ]; then
	exec ./bin/server
else
	exec go run cmd/server/main.go
fi
