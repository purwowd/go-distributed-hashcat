#!/bin/bash

# Test script for Parsing Logic with Sample Data
echo "🧪 Testing Parsing Logic with Sample Hashcat -I Output"
echo "======================================================="

# Sample hashcat -I output (based on user's actual output)
SAMPLE_OUTPUT="hashcat (v6.1.1) starting...

OpenCL Info:
============

OpenCL Platform ID #1
  Vendor..: The pocl project
  Name....: Portable Computing Language
  Version.: OpenCL 2.0 pocl 1.8  Linux, None+Asserts, RELOC, LLVM 11.1.0, SLEEF, DISTRO, POCL_DEBUG

  Backend Device ID #1
    Type...........: CPU
    Vendor.ID......: 128
    Vendor.........: GenuineIntel
    Name...........: pthread-11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
    Version........: OpenCL 1.2 pocl HSTR: pthread-x86_64-pc-linux-gnu-goldmont
    Processor(s)...: 4
    Clock..........: 2803
    Memory.Total...: 2936 MB (limited to 1024 MB allocatable in one block)
    Memory.Free....: 2872 MB
    OpenCL.Version.: OpenCL C 1.2 pocl
    Driver.Version.: 1.8"

echo "📝 Test 1: Sample hashcat -I output"
echo "------------------------------------"
echo "$SAMPLE_OUTPUT"

echo ""
echo "📝 Test 2: Parsing test results"
echo "================================="

# Count total lines
TOTAL_LINES=$(echo "$SAMPLE_OUTPUT" | wc -l)
echo "Total lines: $TOTAL_LINES"

# Find device section headers
DEVICE_SECTIONS=$(echo "$SAMPLE_OUTPUT" | grep -c "Backend Device ID #")
echo "Device sections found: $DEVICE_SECTIONS"

# Find Type lines
TYPE_LINES=$(echo "$SAMPLE_OUTPUT" | grep -c "Type...........:")
echo "Type lines found: $TYPE_LINES"

# Extract device types
echo ""
echo "Device types found:"
echo "==================="
DEVICE_TYPES=$(echo "$SAMPLE_OUTPUT" | grep "Type...........:" | sed 's/.*Type...........: //')

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

echo ""
echo "📝 Test 3: Line-by-line parsing simulation"
echo "==========================================="

# Simulate the Go parsing logic
echo "Simulating Go parsing logic:"
echo ""

DEVICE_TYPES_FOUND=()
LINE_NUMBER=0

while IFS= read -r line; do
    LINE_NUMBER=$((LINE_NUMBER + 1))
    line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    # Look for device section headers
    if echo "$line" | grep -q "Backend Device ID #"; then
        echo "Line $LINE_NUMBER: Found device section header: '$line'"
        continue
    fi
    
    # Look for Type line
    if echo "$line" | grep -q "^Type...........:"; then
        echo "Line $LINE_NUMBER: Found Type line: '$line'"
        parts=$(echo "$line" | sed 's/^Type...........: //')
        if [ -n "$parts" ]; then
            DEVICE_TYPES_FOUND+=("$parts")
            echo "  -> Extracted device type: '$parts'"
        fi
    fi
done <<< "$SAMPLE_OUTPUT"

echo ""
echo "📝 Test 4: Final parsing results"
echo "================================="
echo "Total device types found: ${#DEVICE_TYPES_FOUND[@]}"
echo "Device types: ${DEVICE_TYPES_FOUND[*]}"

if [ ${#DEVICE_TYPES_FOUND[@]} -gt 0 ]; then
    echo ""
    echo "🎯 Expected Go function result:"
    
    # Simulate the Go logic
    GPU_DETECTED=false
    CPU_DETECTED=false
    
    for deviceType in "${DEVICE_TYPES_FOUND[@]}"; do
        if echo "$deviceType" | grep -qi "GPU"; then
            GPU_DETECTED=true
            echo "✅ GPU device detected: $deviceType"
            echo "   -> Function should return: GPU"
            break
        fi
    done
    
    if [ "$GPU_DETECTED" = false ]; then
        for deviceType in "${DEVICE_TYPES_FOUND[@]}"; do
            if echo "$deviceType" | grep -qi "CPU"; then
                CPU_DETECTED=true
                echo "✅ CPU device detected: $deviceType"
                echo "   -> Function should return: CPU"
                break
            fi
        done
    fi
    
    if [ "$GPU_DETECTED" = false ] && [ "$CPU_DETECTED" = false ]; then
        echo "⚠️ Could not determine capabilities from device types"
        echo "   -> Function should fallback to basic detection"
    fi
else
    echo "❌ No device types found"
    echo "   -> Function should fallback to basic detection"
fi

echo ""
echo "🎯 Test Summary:"
echo "================"
echo "✅ Sample data: Based on actual hashcat -I output"
echo "✅ Total lines: $TOTAL_LINES"
echo "✅ Device sections: $DEVICE_SECTIONS"
echo "✅ Type lines: $TYPE_LINES"
echo "✅ Device types: ${#DEVICE_TYPES_FOUND[@]}"
echo ""
echo "🔍 Expected behavior:"
echo "- Agent should detect 'CPU' from hashcat -I output"
echo "- Function should return 'CPU' as capabilities"
echo "- Database should be updated with 'CPU'"
echo ""
echo "❓ If agent still returns 'GPU', check:"
echo "1. Is the Go parsing logic working correctly?"
echo "2. Is there a fallback to basic detection?"
echo "3. Are there any errors in the parsing loop?"
