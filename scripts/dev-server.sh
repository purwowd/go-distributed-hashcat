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

# data/ and uploads/ must be writable (past sudo runs often leave them root-owned).
for dir in data uploads; do
	if [ -d "$dir" ] && [ ! -w "$dir" ]; then
		echo "⚠️  $dir/ is not writable by $(whoami) (often root-owned from a past sudo run)."
		echo "    Fix ownership, then restart:"
		echo "    sudo chown -R $(whoami):$(whoami) \"$(pwd)/$dir\""
		exit 1
	fi
done

# Run the server (prefer compiled binary when present)
if [ -x ./bin/server ]; then
	exec ./bin/server
else
	exec go run cmd/server/main.go
fi
