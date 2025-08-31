# Distributed Cracking with Skip/Limit Implementation

## Overview
The system now supports true distributed hashcat cracking using `--skip` and `--limit` parameters instead of creating separate wordlist files. This allows multiple agents to work on the same wordlist simultaneously with optimal performance.

## How It Works

### 1. Range-Based Distribution
- Wordlist is divided into ranges based on agent performance
- Each agent gets a specific range: `--skip N --limit M`
- No physical wordlist segmentation required

### 2. Agent Performance Calculation
- GPU agents get higher priority and larger ranges
- Performance score determines wordlist allocation
- Automatic load balancing based on capabilities

### 3. Hashcat Command Generation
```bash
# Agent 1 (high performance)
hashcat -m 0 -a 0 hash.txt wordlist.txt --skip 0 --limit 5000

# Agent 2 (medium performance)  
hashcat -m 0 -a 0 hash.txt wordlist.txt --skip 5000 --limit 3000

# Agent 3 (low performance)
hashcat -m 0 -a 0 hash.txt wordlist.txt --skip 8000 --limit 2000
```

## Database Schema
New fields added to `jobs` table:
- `skip`: Starting position in wordlist (--skip parameter)
- `word_limit`: Number of words to process (--limit parameter)

## API Usage

### Create Distributed Job
```bash
curl -X POST http://localhost:1337/api/jobs/distributed \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Distributed MD5 Crack",
    "hash_type": 0,
    "attack_mode": 0, 
    "hash_file_id": "uuid-here",
    "wordlist_id": "uuid-here",
    "auto_distribute": true
  }'
```

### Benefits
- **Efficiency**: No file I/O overhead for wordlist segmentation
- **Performance**: True parallel processing on same wordlist
- **Scalability**: Easy to add/remove agents dynamically
- **Memory**: Lower memory usage per agent
- **Speed**: Faster job distribution and startup

## Testing
1. Start server: `go run cmd/server/main.go`
2. Register multiple agents with different capabilities
3. Create distributed job via API
4. Verify agents receive jobs with correct skip/limit values
5. Monitor hashcat execution with proper parameters

## Migration
Database migration `005_add_skip_limit_fields.sql` adds required columns:
```sql
ALTER TABLE jobs ADD COLUMN skip INTEGER;
ALTER TABLE jobs ADD COLUMN word_limit INTEGER;
```
