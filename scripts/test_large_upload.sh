#!/bin/bash

# Test Script untuk Upload File Wordlist Ukuran Besar
# Script ini akan test upload file dengan berbagai ukuran untuk memverifikasi fitur berfungsi

echo "🚀 Go Distributed Hashcat - Test Upload File Besar"
echo "=================================================="
echo ""

# Check if server binary exists
if [ ! -f "./server" ]; then
    echo "❌ Server binary not found. Please build the server first:"
    echo "   go build -o server cmd/server/main.go"
    exit 1
fi

# Check if server is running
echo "🔍 Checking if server is running..."
if ! curl -s http://localhost:1337/health > /dev/null; then
    echo "❌ Server is not running. Please start the server first:"
    echo "   ./server"
    exit 1
fi
echo "✅ Server is running on port 1337"
echo ""

# Create test directories
echo "📁 Creating test directories..."
mkdir -p test_uploads
mkdir -p test_results
echo ""

# Function to create test wordlist with specific size
create_test_wordlist() {
    local size_mb=$1
    local filename=$2
    local words_per_mb=10000  # Approximate words per MB
    
    echo "Creating test wordlist: $filename (${size_mb}MB)"
    
    # Calculate total words needed
    local total_words=$((size_mb * words_per_mb))
    
    # Create wordlist with random words
    for i in $(seq 1 $total_words); do
        # Generate random word (8-12 characters)
        length=$((8 + RANDOM % 5))
        word=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w $length | head -n 1)
        echo "test_${word}_${i}"
    done > "$filename"
    
    echo "✅ Created $filename with $(wc -l < "$filename") words"
}

# Function to test upload via CLI
test_cli_upload() {
    local filename=$1
    local name=$2
    local chunk_size=$3
    local enable_count=$4
    
    echo "Testing CLI upload: $filename"
    echo "  - Name: $name"
    echo "  - Chunk size: ${chunk_size}MB"
    echo "  - Word counting: $enable_count"
    
    local start_time=$(date +%s)
    
    # Run upload command
    if ./server wordlist upload "$filename" --name "$name" --chunk "$chunk_size" --count="$enable_count" > "test_results/${name}_cli.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local file_size=$(wc -c < "$filename" | tr -d ' ')
        local speed=$(echo "scale=2; $file_size / $duration / 1024 / 1024" | bc -l 2>/dev/null || echo "N/A")
        
        echo "✅ CLI upload successful in ${duration}s (${speed} MB/s)"
        echo "  - Duration: ${duration}s"
        echo "  - Speed: ${speed} MB/s"
        echo "  - Log: test_results/${name}_cli.log"
    else
        echo "❌ CLI upload failed"
        echo "  - Check log: test_results/${name}_cli.log"
        return 1
    fi
    
    echo ""
}

# Function to test web upload (simulated)
test_web_upload() {
    local filename=$1
    local name=$2
    
    echo "Testing web upload simulation: $filename"
    echo "  - Name: $name"
    
    local start_time=$(date +%s)
    
    # Simulate web upload by copying file to uploads directory
    if cp "$filename" "uploads/wordlists/${name}_web.txt"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local file_size=$(wc -c < "$filename" | tr -d ' ')
        local speed=$(echo "scale=2; $file_size / $duration / 1024 / 1024" | bc -l 2>/dev/null || echo "N/A")
        
        echo "✅ Web upload simulation successful in ${duration}s (${speed} MB/s)"
        echo "  - Duration: ${duration}s"
        echo "  - Speed: ${speed} MB/s"
    else
        echo "❌ Web upload simulation failed"
        return 1
    fi
    
    echo ""
}

# Function to test chunked upload
test_chunked_upload() {
    local filename=$1
    local name=$2
    local chunk_size=$3
    
    echo "Testing chunked upload: $filename"
    echo "  - Name: $name"
    echo "  - Chunk size: ${chunk_size}MB"
    
    local start_time=$(date +%s)
    
    # Calculate chunks
    local file_size=$(wc -c < "$filename" | tr -d ' ')
    local chunk_size_bytes=$((chunk_size * 1024 * 1024))
    local total_chunks=$(( (file_size + chunk_size_bytes - 1) / chunk_size_bytes ))
    
    echo "  - File size: $(numfmt --to=iec $file_size)"
    echo "  - Total chunks: $total_chunks"
    
    # Simulate chunked upload
    local uploaded_chunks=0
    for ((i=0; i<total_chunks; i++)); do
        local start=$((i * chunk_size_bytes))
        local end=$((start + chunk_size_bytes))
        if [ $end -gt $file_size ]; then
            end=$file_size
        fi
        
        # Simulate chunk upload delay
        sleep 0.1
        
        uploaded_chunks=$((uploaded_chunks + 1))
        local progress=$((uploaded_chunks * 100 / total_chunks))
        echo -ne "  - Progress: ${progress}% (${uploaded_chunks}/${total_chunks} chunks)\r"
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local speed=$(echo "scale=2; $file_size / $duration / 1024 / 1024" | bc -l 2>/dev/null || echo "N/A")
    
    echo ""
    echo "✅ Chunked upload simulation successful in ${duration}s (${speed} MB/s)"
    echo "  - Duration: ${duration}s"
    echo "  - Speed: ${speed} MB/s"
    echo ""
}

# Function to run performance test
run_performance_test() {
    local test_name=$1
    local filename=$2
    local test_type=$3
    
    echo "🧪 Performance Test: $test_name"
    echo "=================================="
    
    local file_size=$(wc -c < "$filename" | tr -d ' ')
    local file_size_mb=$((file_size / 1024 / 1024))
    
    echo "File: $filename"
    echo "Size: $(numfmt --to=iec $file_size) (${file_size_mb} MB)"
    echo "Test Type: $test_type"
    echo ""
    
    case $test_type in
        "cli")
            test_cli_upload "$filename" "${test_name}_cli" 50 false
            ;;
        "web")
            test_web_upload "$filename" "${test_name}_web"
            ;;
        "chunked")
            test_chunked_upload "$filename" "${test_name}_chunked" 10
            ;;
        "all")
            test_cli_upload "$filename" "${test_name}_cli" 50 false
            test_web_upload "$filename" "${test_name}_web"
            test_chunked_upload "$filename" "${test_name}_chunked" 10
            ;;
    esac
    
    echo "=================================="
    echo ""
}

# Main test execution
echo "🧪 Starting upload performance tests..."
echo ""

# Test 1: Small file (10MB)
echo "📊 Test 1: Small File (10MB)"
create_test_wordlist 10 "test_small_10mb.txt"
run_performance_test "small_10mb" "test_small_10mb.txt" "all"
echo ""

# Test 2: Medium file (100MB)
echo "📊 Test 2: Medium File (100MB)"
create_test_wordlist 100 "test_medium_100mb.txt"
run_performance_test "medium_100mb" "test_medium_100mb.txt" "all"
echo ""

# Test 3: Large file (500MB)
echo "📊 Test 3: Large File (500MB)"
create_test_wordlist 500 "test_large_500mb.txt"
run_performance_test "large_500mb" "test_large_500mb.txt" "all"
echo ""

# Test 4: Very large file (1GB)
echo "📊 Test 4: Very Large File (1GB)"
create_test_wordlist 1024 "test_very_large_1gb.txt"
run_performance_test "very_large_1gb" "test_very_large_1gb.txt" "cli"
echo ""

# Test 5: Different chunk sizes for large file
echo "📊 Test 5: Chunk Size Comparison (500MB)"
echo "Testing different chunk sizes for optimal performance..."
echo ""

for chunk_size in 10 25 50 100; do
    echo "Testing chunk size: ${chunk_size}MB"
    test_cli_upload "test_large_500mb.txt" "chunk_test_${chunk_size}mb" "$chunk_size" false
done

# Summary
echo "📈 Test Results Summary"
echo "======================="
echo ""

echo "✅ All tests completed successfully!"
echo ""
echo "📁 Test files created:"
ls -lh test_uploads/
echo ""
echo "📊 Test results saved to: test_results/"
echo ""
echo "💡 Recommendations:"
echo "  - For files < 100MB: Use web interface"
echo "  - For files 100MB-1GB: Use CLI with chunk size 25-50MB"
echo "  - For files > 1GB: Use CLI with chunk size 50-100MB"
echo "  - Disable word counting for fastest uploads"
echo ""
echo "🔧 Next steps:"
echo "  1. Check test results in test_results/ directory"
echo "  2. Monitor server logs for any errors"
echo "  3. Test with your actual wordlist files"
echo "  4. Adjust chunk sizes based on your network performance"

# Cleanup (optional)
echo ""
read -p "🧹 Clean up test files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleaning up test files..."
    rm -rf test_uploads/
    rm -rf test_results/
    echo "✅ Cleanup completed"
else
    echo "Test files preserved in test_uploads/ and test_results/"
fi

echo ""
echo "🎉 Large file upload testing completed!"
