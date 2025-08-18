#!/bin/bash

# Test script for Hashcat -I Parsing
echo "🧪 Testing Hashcat -I Parsing Logic"
echo "===================================="

# Test 1: Check if hashcat is available
echo ""
echo "📝 Test 1: Check hashcat availability"
echo "-------------------------------------"
if command -v hashcat &> /dev/null; then
    echo "✅ hashcat is available"
    HASHCAT_VERSION=$(hashcat --version | head -n1)
    echo "   Version: $HASHCAT_VERSION"
else
    echo "❌ hashcat is not available"
    echo "   Note: This test requires hashcat to be installed"
    exit 1
fi

# Test 2: Test hashcat -I command and parse output
echo ""
echo "📝 Test 2: Test hashcat -I parsing"
echo "----------------------------------"
echo "Running: hashcat -I"
echo ""

# Run hashcat -I and capture output
HASHCAT_OUTPUT=$(hashcat -I 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ hashcat -I executed successfully"
    echo ""
    echo "Raw output (first 30 lines):"
    echo "$HASHCAT_OUTPUT" | head -30
    
    echo ""
    echo "🔍 Parsing test results:"
    echo "========================="
    
    # Count total lines
    TOTAL_LINES=$(echo "$HASHCAT_OUTPUT" | wc -l)
    echo "Total lines: $TOTAL_LINES"
    
    # Find device section headers
    DEVICE_SECTIONS=$(echo "$HASHCAT_OUTPUT" | grep -c "Backend Device ID #")
    echo "Device sections found: $DEVICE_SECTIONS"
    
    # Find Type lines
    TYPE_LINES=$(echo "$HASHCAT_OUTPUT" | grep -c "Type...........:")
    echo "Type lines found: $TYPE_LINES"
    
    # Extract device types
    echo ""
    echo "Device types found:"
    echo "==================="
    DEVICE_TYPES=$(echo "$HASHCAT_OUTPUT" | grep "Type...........:" | sed 's/.*Type...........: //')
    
    if [ -n "$DEVICE_TYPES" ]; then
        echo "$DEVICE_TYPES"
        
        # Check for GPU
        GPU_FOUND=$(echo "$DEVICE_TYPES" | grep -i "GPU" | wc -l)
        echo ""
        echo "GPU devices: $GPU_FOUND"
        
        # Check for CPU
        CPU_FOUND=$(echo "$DEVICE_TYPES" | grep -i "CPU" | wc -l)
        echo "CPU devices: $CPU_FOUND"
        
        # Determine capabilities
        echo ""
        echo "🎯 Capabilities determination:"
        echo "=============================="
        
        if [ $GPU_FOUND -gt 0 ]; then
            echo "✅ GPU detected - Capabilities should be: GPU"
        elif [ $CPU_FOUND -gt 0 ]; then
            echo "✅ CPU detected - Capabilities should be: CPU"
        else
            echo "⚠️ Unknown device types - Should fallback to basic detection"
        fi
        
    else
        echo "❌ No device types found"
    fi
    
else
    echo "❌ hashcat -I failed to execute"
    echo "Error output:"
    hashcat -I
fi

echo ""
echo "🎯 Test Summary:"
echo "================"
echo "✅ hashcat availability: $(command -v hashcat &> /dev/null && echo "Available" || echo "Not available")"
echo "✅ hashcat -I execution: $(hashcat -I &>/dev/null && echo "Success" || echo "Failed")"
echo "✅ Device sections: $DEVICE_SECTIONS"
echo "✅ Type lines: $TYPE_LINES"
echo "✅ Device types: $(echo "$DEVICE_TYPES" | wc -l)"
echo ""
echo "🔍 Expected behavior:"
echo "- Agent should detect device types from hashcat -I output"
echo "- If GPU found: return 'GPU'"
echo "- If CPU found: return 'CPU'"
echo "- If parsing fails: fallback to basic detection"
