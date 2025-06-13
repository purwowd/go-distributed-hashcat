# 🔌 API Reference

Complete REST API documentation for the distributed hashcat system.

## 📋 Overview

| Setting | Value |
|---------|--------|
| **Base URL** | `http://localhost:1337` |
| **Content Type** | `application/json` |
| **Interactive Docs** | `http://localhost:1337/docs` |

**Health Check**: `GET /health` → `{"status": "ok", "timestamp": 1749341114}`

## 👥 Agents API

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/agents/` | GET | List all agents |
| `/api/v1/agents/` | POST | Register new agent |
| `/api/v1/agents/{id}` | GET | Get agent by ID |
| `/api/v1/agents/{id}/heartbeat` | POST | Update heartbeat |

### Agent Object
```json
{
  "id": "uuid",
  "name": "GPU-Server-01", 
  "ip_address": "192.168.1.100",
  "port": 8080,
  "status": "online",
  "capabilities": "RTX 4090, OpenCL"
}
```

### Examples
```bash
# List agents
curl http://localhost:1337/api/v1/agents/

# Register agent
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d '{"name":"GPU-01","ip_address":"192.168.1.100","port":8080}'
```

## 💼 Jobs API

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/jobs/` | GET | List all jobs |
| `/api/v1/jobs/` | POST | Create new job |
| `/api/v1/jobs/{id}` | GET | Get job details |
| `/api/v1/jobs/{id}/start` | POST | Start job |
| `/api/v1/jobs/{id}/stop` | POST | Stop job |

### Job Object
```json
{
  "id": "uuid",
  "name": "WiFi Password Crack",
  "hash_file_id": "hash-uuid",
  "wordlist_id": "wordlist-uuid", 
  "attack_mode": 0,
  "hash_type": 2500,
  "status": "pending",
  "progress": 0.0,
  "agent_id": "agent-uuid"
}
```

### Status Values
- `pending` - Job created, waiting to start
- `running` - Job in progress
- `paused` - Job temporarily stopped
- `completed` - Job finished successfully
- `failed` - Job failed with error

### Examples
```bash
# Create job
curl -X POST http://localhost:1337/api/v1/jobs/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WiFi Crack",
    "hash_file_id": "hash-uuid",
    "attack_mode": 0,
    "hash_type": 2500
  }'

# Start job
curl -X POST http://localhost:1337/api/v1/jobs/{id}/start

# Check progress
curl http://localhost:1337/api/v1/jobs/{id}
```

## 📁 Hash Files API

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/hash-files/` | GET | List hash files |
| `/api/v1/hash-files/` | POST | Upload hash file |
| `/api/v1/hash-files/{id}` | GET | Get file details |
| `/api/v1/hash-files/{id}/download` | GET | Download file |

### Examples
```bash
# Upload hash file
curl -X POST http://localhost:1337/api/v1/hash-files/ \
  -F "file=@hashes.txt"

# List files
curl http://localhost:1337/api/v1/hash-files/
```

## 📚 Wordlists API

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/wordlists/` | GET | List wordlists |
| `/api/v1/wordlists/` | POST | Upload wordlist |
| `/api/v1/wordlists/{id}` | GET | Get wordlist details |
| `/api/v1/wordlists/{id}/download` | GET | Download wordlist |

### Examples
```bash
# Upload wordlist
curl -X POST http://localhost:1337/api/v1/wordlists/ \
  -F "file=@rockyou.txt"

# Use in job
curl -X POST http://localhost:1337/api/v1/jobs/ \
  -d '{"wordlist_id":"wordlist-uuid",...}'
```

## ⚠️ Error Handling

### Error Response Format
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2023-12-25T10:30:00Z"
}
```

### Common HTTP Status Codes
| Code | Meaning | Example |
|------|---------|---------|
| 200 | Success | Request completed |
| 201 | Created | Resource created |
| 400 | Bad Request | Invalid JSON/parameters |
| 404 | Not Found | Resource doesn't exist |
| 500 | Server Error | Internal error |

## 📊 Rate Limiting

- **Default**: 100 requests per minute per IP
- **File Uploads**: 10 uploads per minute
- **Headers**: `X-RateLimit-Remaining`, `X-RateLimit-Reset`

**Next Steps**: [`04-architecture.md`](04-architecture.md) for system design details
