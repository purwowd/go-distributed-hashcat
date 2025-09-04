# Agent Speed Feature

## Overview

Fitur Agent Speed memungkinkan sistem untuk mendeteksi dan menyimpan kecepatan hashcat benchmark dari setiap agent secara realtime. Kecepatan ini akan digunakan untuk optimasi distribusi job dan monitoring performa agent.

## Fitur Utama

### 1. Auto-Detection Speed
- Agent secara otomatis menjalankan `hashcat -b -m 2500` saat startup
- Parsing output untuk mengekstrak kecepatan dalam H/s (Hashes per second)
- Update database secara realtime

### 2. Real-time Updates
- Speed diupdate setiap kali agent startup
- Broadcast via WebSocket untuk real-time monitoring
- Data speed tersimpan di database sampai agent berhenti

### 3. Database Integration
- Field `speed` ditambahkan ke tabel `agents`
- Default value: 0 (untuk agent yang belum di-benchmark)
- Tipe data: INTEGER untuk menyimpan H/s

## Cara Kerja

### 1. Agent Startup
```bash
sudo ./bin/agent --server https://f62940c62e3c.ngrok-free.app --name "agent-A" --agent-key "4c3418d2" --ip "172.15.1.94"
```

### 2. Benchmark Process
1. Agent mendaftar ke server dan status menjadi "online"
2. Otomatis menjalankan `hashcat -b -m 2500`
3. Parse output untuk mengekstrak speed:
   ```
   Speed.#1.........:     1928 H/s (66.04ms) @ Accel:256 Loops:512 Thr:1 Vec:8
   ```
4. Update database dengan speed yang terdeteksi

### 3. Speed Persistence
- Speed tersimpan di database sampai agent berhenti
- Ketika agent berhenti, speed tetap tersimpan (tidak hilang)
- Speed akan diupdate ulang saat agent startup berikutnya

## API Endpoints

### Update Agent Speed
```http
PUT /api/v1/agents/{id}/speed
Content-Type: application/json

{
  "speed": 1928
}
```

### Response
```json
{
  "message": "Agent speed updated successfully",
  "data": {
    "id": "agent-uuid",
    "speed": 1928
  }
}
```

## Database Schema

### Tabel `agents`
```sql
ALTER TABLE agents ADD COLUMN speed INTEGER DEFAULT 0;
```

### Migration File
```sql
-- Migration: 005_add_speed_to_agents.sql
-- Description: Add speed field to agents table for storing hashcat benchmark speed
-- Author: System
-- Date: 2025-09-02

-- +migrate Up
ALTER TABLE agents ADD COLUMN speed INTEGER DEFAULT 0;

-- +migrate Down
ALTER TABLE agents DROP COLUMN speed;
```

## WebSocket Events

### Agent Speed Update
```json
{
  "type": "agent_speed",
  "data": {
    "agent_id": "agent-uuid",
    "speed": 1928
  },
  "timestamp": "2025-09-03T01:28:10Z"
}
```

## Code Structure

### 1. Domain Model
```go
type Agent struct {
    // ... existing fields ...
    Speed        int64     `json:"speed" db:"speed"` // Hash rate dalam H/s dari benchmark
    // ... existing fields ...
}
```

### 2. Repository Interface
```go
type AgentRepository interface {
    // ... existing methods ...
    UpdateSpeed(ctx context.Context, id uuid.UUID, speed int64) error
}
```

### 3. Usecase Interface
```go
type AgentUsecase interface {
    // ... existing methods ...
    UpdateAgentSpeed(ctx context.Context, id uuid.UUID, speed int64) error
}
```

### 4. Handler
```go
func (h *AgentHandler) UpdateAgentSpeed(c *gin.Context)
```

## Testing

### Manual Test
```bash
# Test script
./scripts/test_agent_speed.sh

# Manual test
curl -X PUT http://localhost:1337/api/v1/agents/{id}/speed \
  -H "Content-Type: application/json" \
  -d '{"speed": 1928}'
```

### Hashcat Benchmark Test
```bash
# Test benchmark output parsing
hashcat -b -m 2500

# Expected output format
Speed.#1.........:     1928 H/s (66.04ms) @ Accel:256 Loops:512 Thr:1 Vec:8
```

## Error Handling

### 1. Hashcat Not Available
- Warning log: "hashcat not found, falling back to basic detection"
- Agent tetap berjalan tanpa speed information
- Speed tetap 0 di database

### 2. Benchmark Failed
- Warning log: "Hashcat benchmark completed with error"
- Agent tetap berjalan tanpa speed information
- Speed tetap 0 di database

### 3. API Update Failed
- Error log: "Failed to update agent speed"
- Agent tetap berjalan tanpa speed information
- Speed tetap 0 di database

## Monitoring & Debugging

### Log Messages
```
[INFO] Starting hashcat benchmark to detect agent speed...
[INFO] Detected hashcat speed: 1928 H/s
[SUCCESS] Agent speed updated successfully: 1928 H/s
[WARNING] Hashcat benchmark completed with error: exit status 1
[ERROR] Failed to update agent speed: API call failed
```

### Database Query
```sql
-- Check agent speed
SELECT id, name, speed, status FROM agents WHERE speed > 0;

-- Check speed distribution
SELECT 
    CASE 
        WHEN speed = 0 THEN 'No Speed'
        WHEN speed < 1000 THEN 'Low (<1K H/s)'
        WHEN speed < 10000 THEN 'Medium (1K-10K H/s)'
        WHEN speed < 100000 THEN 'High (10K-100K H/s)'
        ELSE 'Very High (>100K H/s)'
    END as speed_category,
    COUNT(*) as agent_count
FROM agents 
GROUP BY speed_category;
```

## Performance Considerations

### 1. Benchmark Time
- `hashcat -b -m 2500` biasanya memakan waktu 5-30 detik
- Hanya dijalankan saat agent startup
- Tidak mempengaruhi performa job cracking

### 2. Database Updates
- Update speed hanya saat benchmark selesai
- Tidak ada polling atau update berulang
- Minimal impact pada database performance

### 3. WebSocket Broadcasting
- Broadcast hanya saat speed berubah
- Tidak ada continuous streaming
- Minimal network overhead

## Future Enhancements

### 1. Multiple Hash Types
- Benchmark untuk berbagai hash types (MD5, SHA1, etc.)
- Speed matrix per hash type
- Optimasi job assignment berdasarkan hash type

### 2. Dynamic Speed Updates
- Periodic re-benchmark untuk monitoring degradation
- Speed trending dan alerting
- Performance degradation detection

### 3. Speed-based Job Assignment
- Job assignment berdasarkan speed agent
- Load balancing berdasarkan kapasitas
- Predictive job completion time

## Troubleshooting

### Common Issues

#### 1. Speed Always 0
```bash
# Check hashcat availability
which hashcat

# Check benchmark output
hashcat -b -m 2500

# Check agent logs
tail -f /var/log/agent.log
```

#### 2. Benchmark Hangs
```bash
# Kill hanging hashcat process
pkill -f "hashcat -b -m 2500"

# Check system resources
htop
free -h
```

#### 3. Database Update Fails
```bash
# Check database connection
sqlite3 data/hashcat.db ".tables"

# Check agent table schema
sqlite3 data/hashcat.db ".schema agents"

# Check API endpoint
curl -v http://localhost:1337/api/v1/agents/{id}/speed
```

## Conclusion

Fitur Agent Speed memberikan visibility yang lebih baik terhadap kapasitas setiap agent, memungkinkan optimasi distribusi job dan monitoring performa sistem secara realtime. Implementasi ini dirancang untuk minimal overhead dengan maksimal benefit untuk sistem distributed hashcat.
