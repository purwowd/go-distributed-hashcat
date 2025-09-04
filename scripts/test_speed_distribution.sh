#!/bin/bash

# Test script untuk memverifikasi distribusi words berdasarkan actual speed data
# Data dari gambar: Agent-B (2,404 H/s), Agent-A (1,907 H/s), Agent-C (1,294 H/s)
# Total words: 758,561

echo "🧪 Testing Speed-Based Word Distribution"
echo "========================================"

# Data dari gambar
AGENT_B_SPEED=2404
AGENT_A_SPEED=1907
AGENT_C_SPEED=1294
TOTAL_WORDS=758561

echo "📊 Input Data:"
echo "  Agent-B Speed: $AGENT_B_SPEED H/s"
echo "  Agent-A Speed: $AGENT_A_SPEED H/s"
echo "  Agent-C Speed: $AGENT_C_SPEED H/s"
echo "  Total Words: $TOTAL_WORDS"
echo ""

# Hitung total speed
TOTAL_SPEED=$((AGENT_B_SPEED + AGENT_A_SPEED + AGENT_C_SPEED))
echo "📈 Total Speed: $TOTAL_SPEED H/s"
echo ""

# Hitung distribusi yang benar
AGENT_B_WORDS=$(( (AGENT_B_SPEED * TOTAL_WORDS) / TOTAL_SPEED ))
AGENT_A_WORDS=$(( (AGENT_A_SPEED * TOTAL_WORDS) / TOTAL_SPEED ))
AGENT_C_WORDS=$((TOTAL_WORDS - AGENT_B_WORDS - AGENT_A_WORDS))

# Hitung percentage
AGENT_B_PERCENT=$(( (AGENT_B_WORDS * 100) / TOTAL_WORDS ))
AGENT_A_PERCENT=$(( (AGENT_A_WORDS * 100) / TOTAL_WORDS ))
AGENT_C_PERCENT=$(( (AGENT_C_WORDS * 100) / TOTAL_WORDS ))

echo "✅ Correct Distribution Based on Speed:"
echo "======================================="
echo "Agent-B (Highest Speed):"
echo "  Speed: $AGENT_B_SPEED H/s"
echo "  Words: $AGENT_B_WORDS"
echo "  Percentage: $AGENT_B_PERCENT%"
echo ""

echo "Agent-A (Medium Speed):"
echo "  Speed: $AGENT_A_SPEED H/s"
echo "  Words: $AGENT_A_WORDS"
echo "  Percentage: $AGENT_A_PERCENT%"
echo ""

echo "Agent-C (Lowest Speed):"
echo "  Speed: $AGENT_C_SPEED H/s"
echo "  Words: $AGENT_C_WORDS"
echo "  Percentage: $AGENT_C_PERCENT%"
echo ""

# Verifikasi total
TOTAL_DISTRIBUTED=$((AGENT_B_WORDS + AGENT_A_WORDS + AGENT_C_WORDS))
echo "🔍 Verification:"
echo "  Total Distributed: $TOTAL_DISTRIBUTED"
echo "  Original Total: $TOTAL_WORDS"
echo "  Difference: $((TOTAL_DISTRIBUTED - TOTAL_WORDS))"
echo ""

# Bandingkan dengan distribusi lama (yang salah)
OLD_WORDS_PER_AGENT=$((TOTAL_WORDS / 3))
OLD_PERCENT=33

echo "❌ Old Distribution (Incorrect - Equal Split):"
echo "=============================================="
echo "All Agents:"
echo "  Words: $OLD_WORDS_PER_AGENT each"
echo "  Percentage: $OLD_PERCENT% each"
echo ""

echo "📊 Comparison:"
echo "=============="
echo "Agent-B: $AGENT_B_WORDS words (NEW) vs $OLD_WORDS_PER_AGENT words (OLD) = +$((AGENT_B_WORDS - OLD_WORDS_PER_AGENT)) words"
echo "Agent-A: $AGENT_A_WORDS words (NEW) vs $OLD_WORDS_PER_AGENT words (OLD) = +$((AGENT_A_WORDS - OLD_WORDS_PER_AGENT)) words"
echo "Agent-C: $AGENT_C_WORDS words (NEW) vs $OLD_WORDS_PER_AGENT words (OLD) = $((AGENT_C_WORDS - OLD_WORDS_PER_AGENT)) words"
echo ""

echo "🎯 Efficiency Improvement:"
echo "========================="
echo "Agent-B (fastest) gets $((AGENT_B_WORDS - OLD_WORDS_PER_AGENT)) more words"
echo "Agent-C (slowest) gets $((OLD_WORDS_PER_AGENT - AGENT_C_WORDS)) fewer words"
echo "This optimizes overall cracking performance!"
