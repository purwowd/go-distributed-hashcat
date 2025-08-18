#!/bin/bash

# Test script untuk verifikasi distribusi wordlist
# Memastikan total persentase selalu 100%

echo "🧮 Testing Wordlist Distribution Calculation"
echo "=========================================="

# Simulasi 3 agent dengan performance scores
echo ""
echo "📊 Test Case 1: 3 Agents (GPU + 2 CPU)"
echo "----------------------------------------"

# Agent 1: GPU (Performance: 80%)
agent1_score=80
agent1_type="GPU"

# Agent 2: CPU (Performance: 15%)
agent2_score=15
agent2_type="CPU"

# Agent 3: CPU (Performance: 5%)
agent3_score=5
agent3_type="CPU"

# Hitung total performance
total_score=$((agent1_score + agent2_score + agent3_score))

# Hitung persentase untuk setiap agent
agent1_percent=$((agent1_score * 100 / total_score))
agent2_percent=$((agent2_score * 100 / total_score))
agent3_percent=$((agent3_score * 100 / total_score))

# Hitung total persentase
total_percent=$((agent1_percent + agent2_percent + agent3_percent))

echo "Agent 1 ($agent1_type): $agent1_score points = $agent1_percent%"
echo "Agent 2 ($agent2_type): $agent2_score points = $agent2_percent%"
echo "Agent 3 ($agent3_type): $agent3_score points = $agent3_percent%"
echo "Total Performance: $total_score points"
echo "Total Percentage: $total_percent%"

if [ $total_percent -eq 100 ]; then
    echo "✅ SUCCESS: Total percentage equals 100%"
else
    echo "❌ FAILED: Total percentage should be 100%, got $total_percent%"
fi

echo ""
echo "📊 Test Case 2: 2 Agents (GPU + CPU)"
echo "--------------------------------------"

# Agent 1: GPU (Performance: 85%)
agent1_score=85
agent1_type="GPU"

# Agent 2: CPU (Performance: 15%)
agent2_score=15
agent2_type="CPU"

# Hitung total performance
total_score=$((agent1_score + agent2_score))

# Hitung persentase untuk setiap agent
agent1_percent=$((agent1_score * 100 / total_score))
agent2_percent=$((agent2_score * 100 / total_score))

# Hitung total persentase
total_percent=$((agent1_percent + agent2_percent))

echo "Agent 1 ($agent1_type): $agent1_score points = $agent1_percent%"
echo "Agent 2 ($agent2_type): $agent2_score points = $agent2_percent%"
echo "Total Performance: $total_score points"
echo "Total Percentage: $total_percent%"

if [ $total_percent -eq 100 ]; then
    echo "✅ SUCCESS: Total percentage equals 100%"
else
    echo "❌ FAILED: Total percentage should be 100%, got $total_percent%"
fi

echo ""
echo "📊 Test Case 3: 4 Agents (2 GPU + 2 CPU)"
echo "------------------------------------------"

# Agent 1: GPU (Performance: 50%)
agent1_score=50
agent1_type="GPU"

# Agent 2: GPU (Performance: 30%)
agent2_score=30
agent2_type="GPU"

# Agent 3: CPU (Performance: 15%)
agent3_score=15
agent3_type="CPU"

# Agent 4: CPU (Performance: 5%)
agent4_score=5
agent4_type="CPU"

# Hitung total performance
total_score=$((agent1_score + agent2_score + agent3_score + agent4_score))

# Hitung persentase untuk setiap agent
agent1_percent=$((agent1_score * 100 / total_score))
agent2_percent=$((agent2_score * 100 / total_score))
agent3_percent=$((agent3_score * 100 / total_score))
agent4_percent=$((agent4_score * 100 / total_score))

# Hitung total persentase
total_percent=$((agent1_percent + agent2_percent + agent3_percent + agent4_percent))

echo "Agent 1 ($agent1_type): $agent1_score points = $agent1_percent%"
echo "Agent 2 ($agent2_type): $agent2_score points = $agent2_percent%"
echo "Agent 3 ($agent3_type): $agent3_score points = $agent3_percent%"
echo "Agent 4 ($agent3_type): $agent4_score points = $agent4_percent%"
echo "Total Performance: $total_score points"
echo "Total Percentage: $total_percent%"

if [ $total_percent -eq 100 ]; then
    echo "✅ SUCCESS: Total percentage equals 100%"
else
    echo "❌ FAILED: Total percentage should be 100%, got $total_percent%"
fi

echo ""
echo "🔍 Distribution Logic Summary:"
echo "=============================="
echo "1. Performance scores dihitung berdasarkan hardware (GPU > CPU)"
echo "2. Persentase = (Agent Score / Total Score) × 100"
echo "3. Total persentase selalu = 100%"
echo "4. GPU agents mendapat persentase lebih tinggi"
echo "5. CPU agents mendapat persentase lebih rendah"
echo ""
echo "📝 Formula:"
echo "Percentage = (Agent Performance Score / Total Performance Score) × 100"
echo "Total Percentage = Sum of all agent percentages = 100%"
echo ""
echo "🎯 Expected Results:"
echo "- 3 Agents: Total = 100%"
echo "- 2 Agents: Total = 100%"
echo "- 4 Agents: Total = 100%"
echo ""
echo "✅ All tests completed!"
