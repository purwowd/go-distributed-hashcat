#!/bin/bash

echo "🔧 Add Total Dictionary Field"
echo "============================="

echo ""
echo "📊 Current Status:"
echo "-----------------"
echo "✅ attack_mode: 4 (jumlah kata yang dijalankan agent)"
echo "📚 total_dictionary: 6 (total kata dalam wordlist - perlu ditambahkan)"
echo ""

echo "🔍 Database Schema Analysis:"
echo "============================"
echo "Current jobs table structure:"
sqlite3 data/hashcat.db ".schema jobs" | grep -A 20 "CREATE TABLE jobs"

echo ""
echo "🔧 Adding total_dictionary field..."
echo "=================================="

# Add new column
echo "1. Adding total_dictionary column..."
if sqlite3 data/hashcat.db "ALTER TABLE jobs ADD COLUMN total_dictionary INTEGER DEFAULT 0;" 2>/dev/null; then
    echo "✅ Column added successfully"
else
    echo "⚠️  Column might already exist"
fi

echo ""
echo "2. Updating total_dictionary for current job..."
UPDATE_SQL="
UPDATE jobs 
SET total_dictionary = 6
WHERE id = '35825661-fdd6-4c2d-882c-100b8da12208';
"

echo "   SQL Command: $UPDATE_SQL"

if sqlite3 data/hashcat.db "$UPDATE_SQL"; then
    echo "✅ Total dictionary updated successfully"
else
    echo "❌ Update failed"
    exit 1
fi

echo ""
echo "📊 Final Job Data:"
echo "=================="
sqlite3 data/hashcat.db "SELECT id, name, attack_mode, total_dictionary, rules, speed, eta FROM jobs WHERE id='35825661-fdd6-4c2d-882c-100b8da12208';"

echo ""
echo "🎯 Field Meanings:"
echo "=================="
echo "• attack_mode: 4 (jumlah kata yang dijalankan agent ini)"
echo "• total_dictionary: 6 (total kata dalam wordlist lengkap)"
echo "• rules: Password hasil cracking"
echo "• speed: Kecepatan agent (100 MH/s untuk GPU)"
echo "• eta: Waktu estimasi selesai"
echo ""

echo "📊 Distribution Summary:"
echo "======================="
echo "• Total wordlist: 6 kata"
echo "• GPU agent (test-agent-gpu): 4 kata (67%)"
echo "• Remaining: 2 kata (33%) - untuk agent lain"
echo ""

echo "✅ Total dictionary field added successfully!"
echo "   Now you can track both agent-specific and total wordlist counts."
