#!/bin/bash

echo "🔧 Update Job Rules Field"
echo "========================="

echo ""
echo "📊 Current Job Data:"
echo "-------------------"
sqlite3 data/hashcat.db "SELECT id, name, attack_mode, rules, speed, eta FROM jobs WHERE id='35825661-fdd6-4c2d-882c-100b8da12208';"

echo ""
echo "🔧 Updating rules field..."
echo "=========================="

# Create temporary SQL file
cat > /tmp/update_job.sql << 'EOF'
UPDATE jobs 
SET rules = 'Password found: Starbucks2025@@!!'
WHERE id = '35825661-fdd6-4c2d-882c-100b8da12208';
EOF

echo "✅ SQL file created: /tmp/update_job.sql"
echo "   Content:"
cat /tmp/update_job.sql

echo ""
echo "🔧 Executing update..."
echo "====================="

# Execute the SQL file
if sqlite3 data/hashcat.db < /tmp/update_job.sql; then
    echo "✅ Update successful!"
else
    echo "❌ Update failed!"
    exit 1
fi

echo ""
echo "📊 Updated Job Data:"
echo "-------------------"
sqlite3 data/hashcat.db "SELECT id, name, attack_mode, rules, speed, eta FROM jobs WHERE id='35825661-fdd6-4c2d-882c-100b8da12208';"

echo ""
echo "🧹 Cleaning up..."
rm -f /tmp/update_job.sql
echo "✅ Temporary files cleaned up"

echo ""
echo "✅ Job rules update completed!"
