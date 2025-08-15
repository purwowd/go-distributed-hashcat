#!/bin/bash

# Debug script for Agent Capabilities Issue
echo "🐛 Debugging Agent Capabilities Detection Issue"
echo "==============================================="

echo "📝 Problem Analysis:"
echo "Agent detects: GPU ❌"
echo "Should detect: CPU ✅"
echo "Hashcat -I shows: Type...........: CPU"
echo ""

# Test 1: Check hashcat availability and execution
echo "📝 Test 1: Hashcat Availability and Execution"
echo "----------------------------------------------"
if command -v hashcat &> /dev/null; then
    echo "✅ hashcat is available"
    echo "   Path: $(which hashcat)"
    echo "   Version: $(hashcat --version | head -n1)"
    
    echo ""
    echo "🔍 Testing hashcat -I execution:"
    if hashcat -I &>/dev/null; then
        echo "✅ hashcat -I executes successfully"
    else
        echo "❌ hashcat -I fails to execute"
        echo "   Error output:"
        hashcat -I
    fi
else
    echo "❌ hashcat is not available"
    echo "   This explains why agent falls back to basic detection"
fi

# Test 2: Check hashcat -I output parsing
echo ""
echo "📝 Test 2: Hashcat -I Output Parsing"
echo "-------------------------------------"
if command -v hashcat &> /dev/null; then
    echo "Running hashcat -I and parsing output:"
    echo ""
    
    # Capture hashcat -I output
    HASHCAT_OUTPUT=$(hashcat -I 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ hashcat -I executed successfully"
        echo ""
        echo "Raw output (first 30 lines):"
        echo "$HASHCAT_OUTPUT" | head -30
        
        echo ""
        echo "🔍 Parsing test:"
        echo "================"
        
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
            echo "🎯 Expected capabilities:"
            echo "========================="
            
            if [ $GPU_FOUND -gt 0 ]; then
                echo "✅ GPU detected - Capabilities should be: GPU"
            elif [ $CPU_FOUND -gt 0 ]; then
                echo "✅ CPU detected - Capabilities should be: CPU"
            else
                echo "⚠️ Unknown device types - Should fallback to basic detection"
            fi
            
        else
            echo "❌ No device types found"
            echo "   This explains why agent falls back to basic detection"
        fi
        
    else
        echo "❌ hashcat -I failed to execute"
        echo "   This explains why agent falls back to basic detection"
    fi
else
    echo "Skipping hashcat -I test (hashcat not available)"
fi

# Test 3: Check system GPU detection (what basic detection would find)
echo ""
echo "📝 Test 3: System GPU Detection (Basic Detection)"
echo "=================================================="
echo "This simulates what the agent's basic detection would find:"
echo ""

# Check nvidia-smi
if command -v nvidia-smi &> /dev/null; then
    echo "🔍 nvidia-smi command found"
    echo "   Testing if it works:"
    if nvidia-smi --query-gpu=name --format=csv,noheader,nounits 2>/dev/null; then
        echo "   ✅ nvidia-smi works - GPU detected"
        echo "   ❌ This explains why agent returns GPU!"
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
        echo "   ❌ This explains why agent returns GPU!"
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
        echo "   ❌ This explains why agent returns GPU!"
    else
        echo "   ⚠️ intel_gpu_top found but doesn't work"
    fi
else
    echo "🔍 intel_gpu_top command not found"
fi

# Check for GPU drivers in /proc
if [ -d "/proc/driver/nvidia" ]; then
    echo "🔍 Found NVIDIA driver in /proc/driver/nvidia"
    echo "   ❌ This explains why agent returns GPU!"
else
    echo "🔍 No NVIDIA driver in /proc/driver/nvidia"
fi

# Check for GPU devices in /sys/class/drm
if [ -d "/sys/class/drm" ]; then
    echo "🔍 Checking /sys/class/drm for GPU devices:"
    GPU_CARDS=0
    for file in /sys/class/drm/card*; do
        if [ -e "$file" ]; then
            echo "   Found: $(basename "$file")"
            GPU_CARDS=$((GPU_CARDS + 1))
        fi
    done
    
    if [ $GPU_CARDS -gt 1 ]; then
        echo "   ❌ Multiple GPU cards found - This explains why agent returns GPU!"
    elif [ $GPU_CARDS -eq 1 ]; then
        echo "   ℹ️ Single GPU card found (card0 is usually integrated)"
    else
        echo "   ℹ️ No additional GPU cards found"
    fi
else
    echo "🔍 /sys/class/drm directory not found"
fi

# Test 4: Check if there are any GPU-related environment variables
echo ""
echo "📝 Test 4: GPU Environment Variables"
echo "===================================="
echo "Checking for GPU-related environment variables:"

if [ -n "$CUDA_VISIBLE_DEVICES" ]; then
    echo "🔍 CUDA_VISIBLE_DEVICES: $CUDA_VISIBLE_DEVICES"
fi

if [ -n "$GPU_DEVICE_ORDINAL" ]; then
    echo "🔍 GPU_DEVICE_ORDINAL: $GPU_DEVICE_ORDINAL"
fi

if [ -n "$MIG_GPU_UUID" ]; then
    echo "🔍 MIG_GPU_UUID: $MIG_GPU_UUID"
fi

# Test 5: Check for any GPU-related processes
echo ""
echo "📝 Test 5: GPU-Related Processes"
echo "================================="
echo "Checking for GPU-related processes:"

GPU_PROCESSES=$(ps aux | grep -E "(nvidia|rocm|intel_gpu|cuda)" | grep -v grep | wc -l)
if [ $GPU_PROCESSES -gt 0 ]; then
    echo "🔍 Found $GPU_PROCESSES GPU-related processes:"
    ps aux | grep -E "(nvidia|rocm|intel_gpu|cuda)" | grep -v grep
else
    echo "🔍 No GPU-related processes found"
fi

echo ""
echo "🎯 Debug Summary:"
echo "=================="
echo "✅ Hashcat -I output: Type...........: CPU"
echo "✅ Parsing logic: Should extract 'CPU'"
echo "✅ Expected result: Function should return 'CPU'"
echo "❌ Actual result: Function returns 'GPU'"
echo ""
echo "🔍 Root Cause Analysis:"
echo "========================"

if command -v hashcat &> /dev/null; then
    if hashcat -I &>/dev/null; then
        echo "✅ hashcat is available and working"
        echo "❌ Agent should use hashcat -I detection"
        echo "❌ If agent still returns GPU, parsing is failing"
    else
        echo "❌ hashcat -I is failing"
        echo "✅ Agent falls back to basic detection"
        echo "❌ Basic detection is returning GPU incorrectly"
    fi
else
    echo "❌ hashcat is not available"
    echo "✅ Agent falls back to basic detection"
    echo "❌ Basic detection is returning GPU incorrectly"
fi

echo ""
echo "💡 Next Steps:"
echo "==============="
echo "1. Check agent logs for detailed parsing information"
echo "2. Verify if hashcat -I command is being executed"
echo "3. Check if parsing is finding the 'Type...........:' line"
echo "4. Verify fallback detection logic"
echo "5. Check if any GPU detection method is returning true incorrectly"
echo ""
echo "🔧 Quick Fix:"
echo "============="
echo "If hashcat is available but agent still returns GPU:"
echo "1. Check agent logs for parsing errors"
echo "2. Verify the exact parsing logic in Go code"
echo "3. Check if there's a bug in the parsing loop"
echo ""
echo "If hashcat is not available or failing:"
echo "1. Fix the hasGPU() function to not return false positives"
echo "2. Ensure GPU detection methods are accurate"
echo "3. Add more robust GPU detection logic"
