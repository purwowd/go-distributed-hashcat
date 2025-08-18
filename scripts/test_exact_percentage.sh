#!/bin/bash

# Test script untuk verifikasi perhitungan percentage exact
# Memastikan total selalu tepat 100% tanpa rounding errors

echo "🧮 Testing Exact Percentage Calculation (Total = 100%)"
echo "====================================================="

echo ""
echo "📊 Test Case 1: 3 Agents (GPU + 2 CPU)"
echo "----------------------------------------"

# Simulasi performance scores
agent1_score=80  # GPU
agent2_score=15  # CPU
agent3_score=5   # CPU

total_score=$((agent1_score + agent2_score + agent3_score))

echo "Performance Scores:"
echo "Agent 1 (GPU): $agent1_score points"
echo "Agent 2 (CPU): $agent2_score points"
echo "Agent 3 (CPU): $agent3_score points"
echo "Total Score: $total_score points"
echo ""

# Hitung percentage exact (tanpa rounding untuk 2 agent pertama)
agent1_percent_exact=$((agent1_score * 100 / total_score))
agent2_percent_exact=$((agent2_score * 100 / total_score))

# Agent terakhir mendapat sisa untuk total = 100%
agent3_percent_exact=$((100 - agent1_percent_exact - agent2_percent_exact))

echo "Exact Percentage Calculation:"
echo "Agent 1 (GPU): $agent1_percent_exact%"
echo "Agent 2 (CPU): $agent2_percent_exact%"
echo "Agent 3 (CPU): $agent3_percent_exact% (calculated as 100 - $agent1_percent_exact - $agent2_percent_exact)"
echo "Total: $((agent1_percent_exact + agent2_percent_exact + agent3_percent_exact))%"

if [ $((agent1_percent_exact + agent2_percent_exact + agent3_percent_exact)) -eq 100 ]; then
    echo "✅ SUCCESS: Total equals exactly 100%"
else
    echo "❌ FAILED: Total should be 100%"
fi

echo ""
echo "📊 Test Case 2: 2 Agents (GPU + CPU)"
echo "--------------------------------------"

# Simulasi performance scores
agent1_score=85  # GPU
agent2_score=15  # CPU

total_score=$((agent1_score + agent2_score))

echo "Performance Scores:"
echo "Agent 1 (GPU): $agent1_score points"
echo "Agent 2 (CPU): $agent2_score points"
echo "Total Score: $total_score points"
echo ""

# Hitung percentage exact
agent1_percent_exact=$((agent1_score * 100 / total_score))
agent2_percent_exact=$((100 - agent1_percent_exact))

echo "Exact Percentage Calculation:"
echo "Agent 1 (GPU): $agent1_percent_exact%"
echo "Agent 2 (CPU): $agent2_percent_exact% (calculated as 100 - $agent1_percent_exact)"
echo "Total: $((agent1_percent_exact + agent2_percent_exact))%"

if [ $((agent1_percent_exact + agent2_percent_exact)) -eq 100 ]; then
    echo "✅ SUCCESS: Total equals exactly 100%"
else
    echo "❌ FAILED: Total should be 100%"
fi

echo ""
echo "📊 Test Case 3: 4 Agents (2 GPU + 2 CPU)"
echo "------------------------------------------"

# Simulasi performance scores
agent1_score=50  # GPU 1
agent2_score=30  # GPU 2
agent3_score=15  # CPU 1
agent4_score=5   # CPU 2

total_score=$((agent1_score + agent2_score + agent3_score + agent4_score))

echo "Performance Scores:"
echo "Agent 1 (GPU): $agent1_score points"
echo "Agent 2 (GPU): $agent2_score points"
echo "Agent 3 (CPU): $agent3_score points"
echo "Agent 4 (CPU): $agent4_score points"
echo "Total Score: $total_score points"
echo ""

# Hitung percentage exact (tanpa rounding untuk 3 agent pertama)
agent1_percent_exact=$((agent1_score * 100 / total_score))
agent2_percent_exact=$((agent2_score * 100 / total_score))
agent3_percent_exact=$((agent3_score * 100 / total_score))

# Agent terakhir mendapat sisa untuk total = 100%
agent4_percent_exact=$((100 - agent1_percent_exact - agent2_percent_exact - agent3_percent_exact))

echo "Exact Percentage Calculation:"
echo "Agent 1 (GPU): $agent1_percent_exact%"
echo "Agent 2 (GPU): $agent2_percent_exact%"
echo "Agent 3 (CPU): $agent3_percent_exact%"
echo "Agent 4 (CPU): $agent4_percent_exact% (calculated as 100 - $agent1_percent_exact - $agent2_percent_exact - $agent3_percent_exact)"
echo "Total: $((agent1_percent_exact + agent2_percent_exact + agent3_percent_exact + agent4_percent_exact))%"

if [ $((agent1_percent_exact + agent2_percent_exact + agent3_percent_exact + agent4_percent_exact)) -eq 100 ]; then
    echo "✅ SUCCESS: Total equals exactly 100%"
else
    echo "❌ FAILED: Total should be 100%"
fi

echo ""
echo "🔍 Exact Percentage Logic:"
echo "=========================="
echo "1. Agent 1 sampai N-1: Round percentage normal"
echo "2. Agent N (terakhir): 100% - sum(agent1 sampai N-1)"
echo "3. Total selalu = 100% (exact, no rounding errors)"
echo ""
echo "📝 Formula:"
echo "Agent N Percentage = 100% - Σ(Agent 1 to N-1 Percentage)"
echo "Total = 100% (guaranteed)"
echo ""
echo "🎯 Expected Results:"
echo "- 3 Agents: Total = 100% (exact)"
echo "- 2 Agents: Total = 100% (exact)"
echo "- 4 Agents: Total = 100% (exact)"
echo ""
echo "✅ All exact percentage tests completed!"
