#!/bin/bash

echo "🔍 Hashcat Debug Script for Agent"
echo "=================================="

# Check if hashcat is installed
echo "1. Checking hashcat installation..."
if command -v hashcat &> /dev/null; then
    echo "✅ Hashcat is installed"
    hashcat --version | head -1
else
    echo "❌ Hashcat is NOT installed"
    echo "Please install hashcat first:"
    echo "  Ubuntu/Debian: sudo apt-get install hashcat"
    echo "  CentOS/RHEL: sudo yum install hashcat"
    echo "  Or download from: https://hashcat.net/hashcat/"
    exit 1
fi

echo ""
echo "2. Checking hashcat capabilities..."
hashcat -I 2>/dev/null | grep -E "(Device|Type|Name|Speed)" | head -10

echo ""
echo "3. Testing basic hashcat functionality..."
echo "Running: hashcat --help | head -5"
hashcat --help 2>/dev/null | head -5

echo ""
echo "4. Checking if we can access the files..."
if [ -f "/root/uploads/temp/Starbucks_20250526_140536.hccapx" ]; then
    echo "✅ Hash file exists: /root/uploads/temp/Starbucks_20250526_140536.hccapx"
    ls -la "/root/uploads/temp/Starbucks_20250526_140536.hccapx"
else
    echo "❌ Hash file NOT found: /root/uploads/temp/Starbucks_20250526_140536.hccapx"
fi

if [ -f "/root/uploads/temp/wordlist-test.txt" ]; then
    echo "✅ Wordlist exists: /root/uploads/temp/wordlist-test.txt"
    ls -la "/root/uploads/temp/wordlist-test.txt"
    echo "Wordlist content (first 5 lines):"
    head -5 "/root/uploads/temp/wordlist-test.txt"
else
    echo "❌ Wordlist NOT found: /root/uploads/temp/wordlist-test.txt"
fi

echo ""
echo "5. Testing hashcat with the actual files..."
echo "Running: hashcat -m 2500 -a 0 --help"
hashcat -m 2500 -a 0 --help 2>/dev/null | head -3

echo ""
echo "6. Checking system resources..."
echo "Available memory:"
free -h
echo ""
echo "Available disk space:"
df -h /root/uploads/temp

echo ""
echo "7. Testing hashcat dry run (without actual cracking)..."
echo "This will test if hashcat can parse the files correctly:"
hashcat -m 2500 -a 0 "/root/uploads/temp/Starbucks_20250526_140536.hccapx" "/root/uploads/temp/wordlist-test.txt" --dry-run 2>&1 | head -10

echo ""
echo "🔍 Debug completed. Check the output above for issues."
