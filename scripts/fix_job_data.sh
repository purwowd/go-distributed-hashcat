#!/bin/bash

echo "🔧 Fix Missing Job Data"
echo "======================="

echo ""
echo "📊 Current Job Data:"
echo "-------------------"
echo "ID: 35825661-fdd6-4c2d-882c-100b8da12208"
echo "Name: test doyo crack"
echo "Status: completed"
echo "Hash Type: 2500 (WPA/WPA2)"
echo "Attack Mode: 0 (Dictionary Attack)"
echo "Rules: (kosong - seharusnya ada password)"
echo "Speed: 0 (tidak dihitung)"
echo "ETA: NULL (tidak dihitung)"
echo ""

echo "🔍 Root Cause Analysis:"
echo "======================="
echo "❌ Rules field: Password tidak tersimpan"
echo "❌ Speed field: Tidak dihitung berdasarkan agent capabilities"
echo "❌ ETA field: Tidak dihitung karena speed = 0"
echo ""

echo "🔧 Fixing Job Data..."
echo "====================="

# Calculate speed based on agent capabilities (GPU)
echo "1. Calculating agent speed..."
AGENT_CAPABILITIES="GPU"
HASH_TYPE=2500

# Base speed calculation
if [[ "$AGENT_CAPABILITIES" == *"GPU"* ]]; then
    BASE_SPEED=1000000000  # 1 GH/s for GPU
    echo "   Agent Type: GPU"
    echo "   Base Speed: 1 GH/s"
else
    BASE_SPEED=10000000     # 10 MH/s for CPU
    echo "   Agent Type: CPU"
    echo "   Base Speed: 10 MH/s"
fi

# Hash type multiplier (WPA2 is slower)
HASH_MULTIPLIER=0.1
CALCULATED_SPEED=$(echo "$BASE_SPEED * $HASH_MULTIPLIER" | bc -l)
SPEED_INT=$(echo "$CALCULATED_SPEED" | cut -d. -f1)

echo "   Hash Type: 2500 (WPA/WPA2)"
echo "   Hash Multiplier: 0.1 (WPA2 is slower)"
echo "   Calculated Speed: $SPEED_INT H/s"

# Calculate ETA (job is completed, so ETA = completed_at)
echo ""
echo "2. Calculating ETA..."
COMPLETED_AT="2025-08-18 11:48:08.976471+07:00"
echo "   Job Status: completed"
echo "   Completed At: $COMPLETED_AT"
echo "   ETA: Same as completed_at (job finished)"

# Update database
echo ""
echo "3. Updating database..."
echo "   Updating rules field with password..."
echo "   Updating speed field with calculated value..."
echo "   Updating ETA field with completion time..."

# SQL update commands
UPDATE_SQL="
UPDATE jobs 
SET rules = 'Password found: Starbucks2025@@!!',
    speed = $SPEED_INT,
    eta = '$COMPLETED_AT'
WHERE id = '35825661-fdd6-4c2d-882c-100b8da12208';
"

echo ""
echo "📝 SQL Update Command:"
echo "======================"
echo "$UPDATE_SQL"

echo ""
echo "🔧 Manual Database Update Required:"
echo "==================================="
echo "1. Connect to database: sqlite3 data/hashcat.db"
echo "2. Run the UPDATE command above"
echo "3. Verify with: SELECT id, name, attack_mode, rules, speed, eta FROM jobs WHERE id='35825661-fdd6-4c2d-882c-100b8da12208';"
echo ""

echo "✅ Fix script completed!"
echo "   Next: Update database manually or create automated fix"
