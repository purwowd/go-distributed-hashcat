#!/bin/bash

# Script to clean up old jobs that might have incorrect skip/limit parameters
# This script will delete all pending and running jobs to force recreation with correct logic

echo "🧹 Cleaning up old jobs with potentially incorrect skip/limit parameters..."

# Check if database exists
if [ ! -f "data/hashcat.db" ]; then
    echo "❌ Database not found at data/hashcat.db"
    exit 1
fi

# Backup database first
echo "📦 Creating database backup..."
cp data/hashcat.db data/hashcat.db.backup.$(date +%Y%m%d_%H%M%S)

# Delete all pending and running jobs
echo "🗑️  Deleting all pending and running jobs..."
sqlite3 data/hashcat.db "DELETE FROM jobs WHERE status IN ('pending', 'running');"

# Show remaining jobs
echo "📊 Remaining jobs:"
sqlite3 data/hashcat.db "SELECT id, name, status, agent_id FROM jobs;"

echo "✅ Cleanup completed!"
echo "💡 You can now create new jobs with the corrected distribution logic."
echo "🔄 Restart the server and agents to ensure clean state."
