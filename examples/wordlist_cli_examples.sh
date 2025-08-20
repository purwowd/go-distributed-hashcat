#!/bin/bash

# Wordlist CLI Examples for Go Distributed Hashcat
# This script demonstrates the new CLI commands for managing wordlist files

echo "🚀 Go Distributed Hashcat - Wordlist CLI Examples"
echo "=================================================="
echo ""

# Check if server binary exists
if [ ! -f "./server" ]; then
    echo "❌ Server binary not found. Please build the server first:"
    echo "   go build -o server cmd/server/main.go"
    exit 1
fi

echo "✅ Server binary found"
echo ""

# Create test directories
echo "📁 Creating test directories..."
mkdir -p data uploads
echo ""

# Example 1: Basic wordlist upload
echo "📝 Example 1: Basic wordlist upload"
echo "-----------------------------------"
echo "Creating test wordlist with 1000 words..."
for i in {1..1000}; do echo "password$i"; done > test_basic.txt
echo "Uploading wordlist..."
./server wordlist upload test_basic.txt --name "basic_test"
echo ""

# Example 2: Upload with custom name and chunk size
echo "📝 Example 2: Upload with custom name and chunk size"
echo "---------------------------------------------------"
echo "Creating larger test wordlist with 50000 words..."
for i in {1..50000}; do echo "word$i"; done > test_custom.txt
echo "Uploading with custom name and 25MB chunk size..."
./server wordlist upload test_custom.txt --name "custom_large" --chunk 25
echo ""

# Example 3: Fast upload without word counting
echo "📝 Example 3: Fast upload without word counting"
echo "-----------------------------------------------"
echo "Creating test wordlist with 10000 words..."
for i in {1..10000}; do echo "fast$i"; done > test_fast.txt
echo "Uploading without word counting for speed..."
./server wordlist upload test_fast.txt --name "fast_upload" --count=false
echo ""

# Example 4: List all wordlists
echo "📝 Example 4: List all wordlists"
echo "--------------------------------"
./server wordlist list
echo ""

# Example 5: Show help for upload command
echo "📝 Example 5: Upload command help"
echo "---------------------------------"
./server wordlist upload --help
echo ""

# Example 6: Show help for wordlist commands
echo "📝 Example 6: Wordlist commands help"
echo "------------------------------------"
./server wordlist --help
echo ""

# Cleanup
echo "🧹 Cleaning up test files..."
rm -f test_basic.txt test_custom.txt test_fast.txt

echo ""
echo "✅ CLI examples completed successfully!"
echo ""
echo "💡 Tips for large wordlists (1M+ words):"
echo "   - Use --chunk 50 for files larger than 1GB"
echo "   - Use --count=false for fastest uploads"
echo "   - Monitor progress with smaller chunk sizes"
echo "   - Ensure sufficient disk space in uploads directory"
echo ""
echo "🔧 Available commands:"
echo "   ./server wordlist upload <path> [options]"
echo "   ./server wordlist list"
echo "   ./server wordlist delete <id>"
echo "   ./server wordlist --help"
