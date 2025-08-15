#!/bin/bash

# Debug script for Capabilities Detection Issue
echo "🐛 Debugging Capabilities Detection Issue"
echo "========================================="

echo "📝 Problem Analysis:"
echo "Agent detects: GPU ❌"
echo "Should detect: CPU ✅"
echo "Hashcat -I shows: Type...........: CPU"
echo ""

# Test 1: Check hashcat -I output parsing
echo "📝 Test 1: Hashcat -I Output Parsing"
echo "-------------------------------------"
echo "Running hashcat -I to verify output:"
hashcat -I | grep -A 20 "Backend Device ID #"

echo ""
echo "🔍 Parsing test:"
echo "Looking for 'Type...........:' lines:"
hashcat -I | grep "Type...........:"

echo ""
echo "Extracting device type:"
DEVICE_TYPE=$(hashcat -I | grep "Type...........:" | sed 's/.*Type...........: //')
echo "Device type extracted: '$DEVICE_TYPE'"

if [ "$DEVICE_TYPE" = "CPU" ]; then
    echo "✅ Device type correctly extracted: CPU"
else
    echo "❌ Device type extraction failed: '$DEVICE_TYPE'"
fi

# Test 2: Simulate the Go parsing logic
echo ""
echo "📝 Test 2: Simulate Go Parsing Logic"
echo "-------------------------------------"
echo "Simulating the exact parsing logic from Go code:"

HASHCAT_OUTPUT=$(hashcat -I)
LINES_COUNT=$(echo "$HASHCAT_OUTPUT" | wc -l)
echo "Total lines: $LINES_COUNT"

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
done <<< "$HASHCAT_OUTPUT"

echo ""
echo "📝 Test 3: Parsing Results"
echo "==========================="
echo "Total device types found: ${#DEVICE_TYPES_FOUND[@]}"
echo "Device types: ${DEVICE_TYPES_FOUND[*]}"

if [ ${#DEVICE_TYPES_FOUND[@]} -gt 0 ]; then
    echo ""
    echo "🎯 Expected Go function result:"
    
    # Simulate the Go logic
    GPU_DETECTED=false
    CPU_DETECTED=false
    
    for deviceType in "${DEVICE_TYPES_FOUND[@]}"; do
        echo "Checking device type: '$deviceType'"
        if echo "$deviceType" | grep -qi "GPU"; then
            GPU_DETECTED=true
            echo "✅ GPU device detected: $deviceType"
            echo "   -> Function should return: GPU"
            break
        fi
    done
    
    if [ "$GPU_DETECTED" = false ]; then
        for deviceType in "${DEVICE_TYPES_FOUND[@]}"; do
            echo "Checking device type for CPU: '$deviceType'"
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

# Test 4: Check if there are any GPU-related commands or files
echo ""
echo "📝 Test 4: System GPU Detection Check"
echo "====================================="
echo "Checking for GPU-related commands and files:"

# Check nvidia-smi
if command -v nvidia-smi &> /dev/null; then
    echo "🔍 nvidia-smi command found"
    echo "   Testing if it works:"
    if nvidia-smi --query-gpu=name --format=csv,noheader,nounits 2>/dev/null; then
        echo "   ✅ nvidia-smi works - GPU detected"
    else
        echo "   ⚠️ nvidia-smi found but doesn't work"
    fi
else
    echo "🔍 nvidia-smi command not found"
fi

# Check rocm-smi
if command -v rocm-smi &> /dev/null; then
    echo "🔍 rocm-smi command found"
    echo "   Testing if it works:"
    if rocm-smi --list-gpus 2>/dev/null; then
        echo "   ✅ rocm-smi works - GPU detected"
    else
        echo "   ⚠️ rocm-smi found but doesn't work"
    fi
else
    echo "🔍 rocm-smi command not found"
fi

# Check intel_gpu_top
if command -v intel_gpu_top &> /dev/null; then
    echo "🔍 intel_gpu_top command found"
    echo "   Testing if it works:"
    if intel_gpu_top -J -s 1 2>/dev/null; then
        echo "   ✅ intel_gpu_top works - GPU detected"
    else
        echo "   ⚠️ intel_gpu_top found but doesn't work"
    fi
else
    echo "🔍 intel_gpu_top command not found"
fi

# Check for GPU drivers in /proc
if [ -d "/proc/driver/nvidia" ]; then
    echo "🔍 Found NVIDIA driver in /proc/driver/nvidia"
else
    echo "🔍 No NVIDIA driver in /proc/driver/nvidia"
fi

# Check for GPU devices in /sys/class/drm
if [ -d "/sys/class/drm" ]; then
    echo "🔍 Checking /sys/class/drm for GPU devices:"
    for file in /sys/class/drm/card*; do
        if [ -e "$file" ]; then
            echo "   Found: $(basename "$file")"
        fi
    done
else
    echo "🔍 /sys/class/drm directory not found"
fi

echo ""
echo "🎯 Debug Summary:"
echo "=================="
echo "✅ Hashcat -I output: Type...........: CPU"
echo "✅ Parsing logic: Should extract 'CPU'"
echo "✅ Expected result: Function should return 'CPU'"
echo "❌ Actual result: Function returns 'GPU'"
echo ""
echo "🔍 Possible issues:"
echo "1. Hashcat -I parsing is failing"
echo "2. Function is falling back to basic detection"
echo "3. hasGPU() function is returning true incorrectly"
echo "4. There's a bug in the parsing loop"
echo ""
echo "💡 Next steps:"
echo "1. Check agent logs for detailed parsing information"
echo "2. Verify if hashcat -I command is being executed"
echo "3. Check if parsing is finding the 'Type...........:' line"
echo "4. Verify fallback detection logic"
