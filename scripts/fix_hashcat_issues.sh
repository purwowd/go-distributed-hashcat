#!/bin/bash

echo "🔧 Fixing Common Hashcat Issues"
echo "==============================="

echo "1. Checking and updating hashcat..."
if command -v hashcat &> /dev/null; then
    echo "✅ Hashcat is installed"
    CURRENT_VERSION=$(hashcat --version | head -1)
    echo "Current version: $CURRENT_VERSION"
    
    # Check if we need to update
    echo "Checking for updates..."
    if command -v apt-get &> /dev/null; then
        echo "Ubuntu/Debian system detected"
        sudo apt-get update
        sudo apt-get install -y hashcat
    elif command -v yum &> /dev/null; then
        echo "CentOS/RHEL system detected"
        sudo yum update -y hashcat
    else
        echo "Manual installation required. Download from: https://hashcat.net/hashcat/"
    fi
else
    echo "❌ Hashcat not found. Installing..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y hashcat
    elif command -v yum &> /dev/null; then
        sudo yum install -y hashcat
    else
        echo "Please install hashcat manually"
        exit 1
    fi
fi

echo ""
echo "2. Checking GPU drivers and OpenCL..."
if command -v nvidia-smi &> /dev/null; then
    echo "✅ NVIDIA GPU detected"
    nvidia-smi --query-gpu=name,driver_version --format=csv,noheader,nounits
else
    echo "ℹ️  No NVIDIA GPU detected"
fi

if command -v clinfo &> /dev/null; then
    echo "✅ OpenCL is available"
    clinfo | grep "Platform Name" | head -3
else
    echo "⚠️  OpenCL not found. Installing..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get install -y ocl-icd-opencl-dev
    elif command -v yum &> /dev/null; then
        sudo yum install -y ocl-icd
    fi
fi

echo ""
echo "3. Checking system resources..."
echo "Memory:"
free -h
echo ""
echo "Disk space:"
df -h /root/uploads/temp

echo ""
echo "4. Testing hashcat with different modes..."
echo "Testing CPU mode:"
hashcat -m 2500 -a 0 --help | grep -E "(2500|WPA)" | head -3

echo ""
echo "5. Creating test hashcat command..."
echo "Testing minimal command:"
TEST_HASH="/root/uploads/temp/Starbucks_20250526_140536.hccapx"
TEST_WORDLIST="/root/uploads/temp/wordlist-test.txt"

if [ -f "$TEST_HASH" ] && [ -f "$TEST_WORDLIST" ]; then
    echo "✅ Test files found"
    echo "Running: hashcat -m 2500 -a 0 $TEST_HASH $TEST_WORDLIST --dry-run"
    hashcat -m 2500 -a 0 "$TEST_HASH" "$TEST_WORDLIST" --dry-run 2>&1
    
    if [ $? -eq 0 ]; then
        echo "✅ Hashcat command works with --dry-run"
    else
        echo "❌ Hashcat command still fails"
        echo "Error code: $?"
    fi
else
    echo "❌ Test files not found"
    echo "Hash file: $TEST_HASH"
    echo "Wordlist: $TEST_WORDLIST"
fi

echo ""
echo "6. Checking file permissions..."
if [ -f "$TEST_HASH" ]; then
    echo "Hash file permissions:"
    ls -la "$TEST_HASH"
fi

if [ -f "$TEST_WORDLIST" ]; then
    echo "Wordlist permissions:"
    ls -la "$TEST_WORDLIST"
fi

echo ""
echo "7. Testing with different hashcat options..."
echo "Testing without -w flag:"
hashcat -m 2500 -a 0 "$TEST_HASH" "$TEST_WORDLIST" --dry-run 2>&1

echo ""
echo "Testing with different work profile:"
hashcat -m 2500 -a 0 "$TEST_HASH" "$TEST_WORDLIST" -w 1 --dry-run 2>&1

echo ""
echo "🔧 Fix script completed. Check the output above for resolved issues."
