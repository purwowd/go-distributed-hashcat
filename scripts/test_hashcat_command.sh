#!/bin/bash

echo "🧪 Testing Hashcat Command that Failed"
echo "======================================"

# The exact command that failed
HASH_FILE="/root/uploads/temp/Starbucks_20250526_140536.hccapx"
WORDLIST="/root/uploads/temp/wordlist-test.txt"
OUTFILE="/root/uploads/temp/cracked-test.txt"

echo "1. Testing with the exact command that failed:"
echo "hashcat -m 2500 -a 0 $HASH_FILE $WORDLIST -w 4 --status --status-timer=2 --potfile-disable --outfile $OUTFILE --outfile-format 2"

echo ""
echo "2. First, let's check if files exist and are readable..."
if [ -r "$HASH_FILE" ]; then
    echo "✅ Hash file is readable: $HASH_FILE"
    file "$HASH_FILE"
    ls -la "$HASH_FILE"
else
    echo "❌ Hash file is NOT readable: $HASH_FILE"
    exit 1
fi

if [ -r "$WORDLIST" ]; then
    echo "✅ Wordlist is readable: $WORDLIST"
    file "$WORDLIST"
    ls -la "$WORDLIST"
    echo "First few lines:"
    head -3 "$WORDLIST"
else
    echo "❌ Wordlist is NOT readable: $WORDLIST"
    exit 1
fi

echo ""
echo "3. Testing hashcat with verbose output..."
echo "Running: hashcat -m 2500 -a 0 $HASH_FILE $WORDLIST --dry-run -v"
hashcat -m 2500 -a 0 "$HASH_FILE" "$WORDLIST" --dry-run -v 2>&1

echo ""
echo "4. Testing hashcat with minimal arguments..."
echo "Running: hashcat -m 2500 -a 0 $HASH_FILE $WORDLIST --dry-run"
hashcat -m 2500 -a 0 "$HASH_FILE" "$WORDLIST" --dry-run 2>&1

echo ""
echo "5. Testing hashcat mode 2500 specifically..."
echo "Running: hashcat -m 2500 --help | head -10"
hashcat -m 2500 --help 2>/dev/null | head -10

echo ""
echo "6. Testing with a simple wordlist first..."
echo "Creating a simple test wordlist..."
echo "password123" > /tmp/simple-test.txt
echo "admin" >> /tmp/simple-test.txt
echo "test" >> /tmp/simple-test.txt

echo "Testing with simple wordlist:"
hashcat -m 2500 -a 0 "$HASH_FILE" /tmp/simple-test.txt --dry-run 2>&1

echo ""
echo "7. Checking hashcat error codes..."
echo "Hashcat exit status 255 usually means:"
echo "- Invalid arguments"
echo "- File not found or not readable"
echo "- Unsupported hash type"
echo "- Memory allocation failed"
echo "- GPU driver issues"

echo ""
echo "🔍 Test completed. Check the output above for specific errors."
