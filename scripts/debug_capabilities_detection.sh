#!/bin/bash

# Debug script for Capabilities Detection Issue
echo "🐛 Debugging Capabilities Detection Issue"
echo "========================================="

echo ""
echo "🔍 Problem: Agent detects 'GPU' but hashcat -I shows 'CPU'"
echo "Expected: capabilities = 'CPU'"
echo "Actual:   capabilities = 'GPU'"
echo ""

# Test 1: Check hashcat availability and output
echo "📝 Test 1: Check hashcat availability and output"
echo "------------------------------------------------"

if command -v hashcat &> /dev/null; then
    echo "✅ hashcat is available"
    echo ""
    echo "🔍 Running hashcat -I to see actual output:"
    echo "--------------------------------------------"
    hashcat -I
    echo "--------------------------------------------"
else
    echo "❌ hashcat is not available - this explains the fallback to basic detection"
fi

# Test 2: Check GPU detection commands
echo ""
echo "📝 Test 2: Check GPU detection commands"
echo "---------------------------------------"

echo "🔍 Checking nvidia-smi:"
if command -v nvidia-smi &> /dev/null; then
    echo "✅ nvidia-smi found"
    echo "🔍 Testing nvidia-smi output:"
    nvidia-smi --query-gpu=name --format=csv,noheader,nounits 2>/dev/null || echo "⚠️ nvidia-smi failed to run"
else
    echo "❌ nvidia-smi not found"
fi

echo ""
echo "🔍 Checking rocm-smi:"
if command -v rocm-smi &> /dev/null; then
    echo "✅ rocm-smi found"
    echo "🔍 Testing rocm-smi output:"
    rocm-smi --list-gpus 2>/dev/null || echo "⚠️ rocm-smi failed to run"
else
    echo "❌ rocm-smi not found"
fi

echo ""
echo "🔍 Checking intel_gpu_top:"
if command -v intel_gpu_top &> /dev/null; then
    echo "✅ intel_gpu_top found"
    echo "🔍 Testing intel_gpu_top output:"
    timeout 2 intel_gpu_top -J -s 1 2>/dev/null || echo "⚠️ intel_gpu_top failed to run"
else
    echo "❌ intel_gpu_top not found"
fi

# Test 3: Check system files that might trigger GPU detection
echo ""
echo "📝 Test 3: Check system files that might trigger GPU detection"
echo "--------------------------------------------------------------"

echo "🔍 Checking /proc/driver/nvidia:"
if [ -d "/proc/driver/nvidia" ]; then
    echo "✅ /proc/driver/nvidia exists - this triggers GPU detection"
    ls -la /proc/driver/nvidia/ | head -5
else
    echo "❌ /proc/driver/nvidia does not exist"
fi

echo ""
echo "🔍 Checking /sys/class/drm:"
if [ -d "/sys/class/drm" ]; then
    echo "✅ /sys/class/drm exists"
    echo "🔍 DRM devices found:"
    ls -la /sys/class/drm/ | grep "card"
    
    # Check if there are GPU cards (not just card0)
    GPU_CARDS=$(ls /sys/class/drm/ | grep "^card" | grep -v "^card0$" | wc -l)
    if [ "$GPU_CARDS" -gt 0 ]; then
        echo "⚠️ Found $GPU_CARDS GPU cards - this triggers GPU detection"
        ls /sys/class/drm/ | grep "^card" | grep -v "^card0$"
    else
        echo "✅ Only card0 found (integrated graphics) - should not trigger GPU detection"
    fi
else
    echo "❌ /sys/class/drm does not exist"
fi

# Test 4: Simulate the hasGPU() function logic
echo ""
echo "📝 Test 4: Simulate the hasGPU() function logic"
echo "-----------------------------------------------"

echo "🔍 Simulating hasGPU() function checks:"

# Check nvidia-smi
if command -v nvidia-smi &> /dev/null; then
    echo "1. nvidia-smi found ✓"
    if nvidia-smi --query-gpu=name --format=csv,noheader,nounits &>/dev/null; then
        echo "   → nvidia-smi runs successfully ✓"
        echo "   → hasGPU() would return true (GPU detected)"
        GPU_DETECTED=true
    else
        echo "   → nvidia-smi failed to run ✗"
        GPU_DETECTED=false
    fi
else
    echo "1. nvidia-smi not found ✗"
    GPU_DETECTED=false
fi

# Check rocm-smi
if command -v rocm-smi &> /dev/null; then
    echo "2. rocm-smi found ✓"
    if rocm-smi --list-gpus &>/dev/null; then
        echo "   → rocm-smi runs successfully ✓"
        echo "   → hasGPU() would return true (GPU detected)"
        GPU_DETECTED=true
    else
        echo "   → rocm-smi failed to run ✗"
    fi
else
    echo "2. rocm-smi not found ✗"
fi

# Check intel_gpu_top
if command -v intel_gpu_top &> /dev/null; then
    echo "3. intel_gpu_top found ✓"
    if timeout 2 intel_gpu_top -J -s 1 &>/dev/null; then
        echo "   → intel_gpu_top runs successfully ✓"
        echo "   → hasGPU() would return true (GPU detected)"
        GPU_DETECTED=true
    else
        echo "   → intel_gpu_top failed to run ✗"
    fi
else
    echo "3. intel_gpu_top not found ✗"
fi

# Check /proc/driver/nvidia
if [ -d "/proc/driver/nvidia" ]; then
    echo "4. /proc/driver/nvidia exists ✓"
    echo "   → hasGPU() would return true (GPU detected)"
    GPU_DETECTED=true
else
    echo "4. /proc/driver/nvidia does not exist ✗"
fi

# Check /sys/class/drm for GPU cards
if [ -d "/sys/class/drm" ]; then
    GPU_CARDS=$(ls /sys/class/drm/ | grep "^card" | grep -v "^card0$" | wc -l)
    if [ "$GPU_CARDS" -gt 0 ]; then
        echo "5. GPU cards found in /sys/class/drm ✓"
        echo "   → Found $GPU_CARDS GPU cards"
        echo "   → hasGPU() would return true (GPU detected)"
        GPU_DETECTED=true
    else
        echo "5. Only integrated graphics in /sys/class/drm ✗"
    fi
else
    echo "5. /sys/class/drm does not exist ✗"
fi

# Test 5: Expected vs Actual behavior
echo ""
echo "📝 Test 5: Expected vs Actual behavior"
echo "--------------------------------------"

echo "🔍 Expected behavior:"
echo "   hashcat -I shows: Type...........: CPU"
echo "   Agent should detect: CPU"
echo "   Database should update to: capabilities = 'CPU'"
echo ""

echo "🔍 Actual behavior:"
echo "   hashcat -I shows: Type...........: CPU"
echo "   Agent detects: GPU (incorrect!)"
echo "   Database updates to: capabilities = 'GPU' (incorrect!)"
echo ""

echo "🔍 Root cause analysis:"
if [ "$GPU_DETECTED" = true ]; then
    echo "   ❌ hasGPU() returns true due to one of these:"
    echo "      - nvidia-smi command available and working"
    echo "      - rocm-smi command available and working"
    echo "      - intel_gpu_top command available and working"
    echo "      - /proc/driver/nvidia exists"
    echo "      - GPU cards found in /sys/class/drm"
    echo ""
    echo "   💡 Solution: Fix hasGPU() to be more accurate or prioritize hashcat -I output"
else
    echo "   ✅ hasGPU() returns false (correct)"
    echo "   💡 Problem must be elsewhere in the code"
fi

# Test 6: Recommended fixes
echo ""
echo "📝 Test 6: Recommended fixes"
echo "----------------------------"

echo "🔧 Fix 1: Prioritize hashcat -I output over fallback"
echo "   - If hashcat -I succeeds, use its output"
echo "   - Only fallback to hasGPU() if hashcat -I fails"
echo ""

echo "🔧 Fix 2: Make hasGPU() more accurate"
echo "   - Check if GPU commands actually work, not just exist"
echo "   - Verify GPU output is meaningful"
echo "   - Don't rely on file existence alone"
echo ""

echo "🔧 Fix 3: Add more logging to detectCapabilitiesWithHashcat"
echo "   - Log each step of parsing"
echo "   - Show why fallback is triggered"
echo "   - Display actual hashcat -I output being parsed"
echo ""

# Test 7: Immediate debugging steps
echo ""
echo "📝 Test 7: Immediate debugging steps"
echo "------------------------------------"

echo "🚀 To debug this issue immediately:"
echo ""
echo "1. 🔍 Check if hashcat -I is actually being executed:"
echo "   sudo ./bin/agent ... 2>&1 | grep -E '(🔍|⚠️|✅|❌)'"
echo ""
echo "2. 🔍 Look for hashcat -I output in logs:"
echo "   sudo ./bin/agent ... 2>&1 | grep -E 'Hashcat -I output'"
echo ""
echo "3. 🔍 Check if fallback is triggered:"
echo "   sudo ./bin/agent ... 2>&1 | grep -E 'falling back to basic detection'"
echo ""
echo "4. 🔍 Verify hasGPU() result:"
echo "   sudo ./bin/agent ... 2>&1 | grep -E 'Starting GPU detection'"
echo ""

echo ""
echo "🎯 Debug Summary:"
echo "================="
echo ""
echo "✅ hashcat -I output shows: Type...........: CPU"
echo "❌ Agent detects: GPU (incorrect)"
echo "🔍 Root cause: hasGPU() fallback returning true"
echo "💡 Solution: Fix hasGPU() logic or prioritize hashcat -I output"
echo ""
echo "🚀 Next step: Run the agent with detailed logging to see exactly what's happening!"
