#!/bin/bash

# Wordlist CLI Performance Benchmark
# Tests the CLI with different file sizes and settings

echo "🚀 Go Distributed Hashcat - Wordlist CLI Performance Benchmark"
echo "==============================================================="
echo ""

# Check if server binary exists
if [ ! -f "./server" ]; then
    echo "❌ Server binary not found. Please build the server first:"
    echo "   go build -o server cmd/server/main.go"
    exit 1
fi

# Create test directories
mkdir -p data uploads

# Function to create test wordlist
create_wordlist() {
    local size=$1
    local filename=$2
    echo "Creating wordlist with $size words..."
    for i in $(seq 1 $size); do
        echo "word$i"
    done > "$filename"
}

# Function to benchmark upload
benchmark_upload() {
    local filename=$1
    local name=$2
    local chunk=$3
    local count=$4
    
    echo "Benchmarking: $filename (chunk: ${chunk}MB, count: $count)"
    
    # Get file size
    local filesize=$(wc -c < "$filename" | tr -d ' ')
    local wordcount=$(wc -l < "$filename" | tr -d ' ')
    
    echo "  File size: $(numfmt --to=iec $filesize)"
    echo "  Word count: $(numfmt --grouping $wordcount)"
    
    # Time the upload
    local start_time=$(date +%s.%N)
    ./server wordlist upload "$filename" --name "$name" --chunk "$chunk" --count="$count" > /dev/null 2>&1
    local end_time=$(date +%s.%N)
    
    # Calculate duration
    local duration=$(echo "$end_time - $start_time" | bc -l)
    local speed=$(echo "scale=2; $filesize / $duration / 1024 / 1024" | bc -l)
    
    echo "  Duration: ${duration}s"
    echo "  Speed: ${speed} MB/s"
    echo ""
    
    # Store results
    echo "$name,$filesize,$wordcount,$chunk,$count,$duration,$speed" >> benchmark_results.csv
}

# Initialize results file
echo "Name,Size_Bytes,Word_Count,Chunk_MB,Count_Enabled,Duration_Sec,Speed_MBps" > benchmark_results.csv

echo "📊 Running benchmarks..."
echo ""

# Test 1: Small wordlist (1K words)
echo "🔍 Test 1: Small wordlist (1K words)"
create_wordlist 1000 "test_1k.txt"
benchmark_upload "test_1k.txt" "test_1k" 10 true
benchmark_upload "test_1k.txt" "test_1k_fast" 10 false

# Test 2: Medium wordlist (10K words)
echo "🔍 Test 2: Medium wordlist (10K words)"
create_wordlist 10000 "test_10k.txt"
benchmark_upload "test_10k.txt" "test_10k" 10 true
benchmark_upload "test_10k.txt" "test_10k_fast" 10 false

# Test 3: Large wordlist (100K words)
echo "🔍 Test 3: Large wordlist (100K words)"
create_wordlist 100000 "test_100k.txt"
benchmark_upload "test_100k.txt" "test_100k" 10 true
benchmark_upload "test_100k.txt" "test_100k_fast" 10 false

# Test 4: Different chunk sizes
echo "🔍 Test 4: Different chunk sizes (100K words)"
benchmark_upload "test_100k.txt" "test_100k_chunk25" 25 true
benchmark_upload "test_100k.txt" "test_100k_chunk50" 50 true

# Test 5: Very large wordlist (500K words)
echo "🔍 Test 5: Very large wordlist (500K words)"
create_wordlist 500000 "test_500k.txt"
benchmark_upload "test_500k.txt" "test_500k" 50 true
benchmark_upload "test_500k.txt" "test_500k_fast" 50 false

echo "📈 Benchmark Results Summary"
echo "============================"
echo ""

# Display results in a table
echo "Name                    | Size     | Words   | Chunk | Count | Time  | Speed"
echo "----------------------|----------|---------|-------|-------|-------|--------"
while IFS=',' read -r name size words chunk count duration speed; do
    if [ "$name" != "Name" ]; then
        printf "%-22s | %-8s | %-7s | %-5s | %-5s | %-5.2f | %-6.2f\n" \
               "$name" \
               "$(numfmt --to=iec $size)" \
               "$(numfmt --grouping $words)" \
               "${chunk}MB" \
               "$count" \
               "$duration" \
               "$speed"
    fi
done < benchmark_results.csv

echo ""
echo "💡 Performance Insights:"
echo ""

# Analyze results
echo "Fastest uploads:"
sort -t',' -k7 -nr benchmark_results.csv | head -3 | while IFS=',' read -r name size words chunk count duration speed; do
    if [ "$name" != "Name" ]; then
        echo "  - $name: ${speed} MB/s"
    fi
done

echo ""
echo "Slowest uploads:"
sort -t',' -k7 -n benchmark_results.csv | head -3 | while IFS=',' read -r name size words chunk count duration speed; do
    if [ "$name" != "Name" ]; then
        echo "  - $name: ${speed} MB/s"
    fi
done

echo ""
echo "📊 Recommendations:"
echo "  - For files < 100MB: Use default settings (chunk: 10MB, count: true)"
echo "  - For files 100MB-1GB: Use chunk: 25MB, count: true"
echo "  - For files > 1GB: Use chunk: 50MB, count: false for speed"
echo "  - Disable word counting for fastest uploads on large files"

# Cleanup
echo ""
echo "🧹 Cleaning up test files..."
rm -f test_*.txt
rm -f benchmark_results.csv

echo ""
echo "✅ Benchmark completed successfully!"
echo "Results saved to benchmark_results.csv"
