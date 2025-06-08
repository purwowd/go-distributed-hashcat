# 🔌 Distributed Hashcat API Documentation

Complete REST API reference for the distributed hashcat system.

## 📋 Overview

- **Base URL**: `http://localhost:1337`
- **API Version**: `v1`
- **Content Type**: `application/json` (except file uploads)
- **Authentication**: None (add authentication for production)

## 📊 Health Check

### `GET /health`

Check server health and status.

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1749341114
}
```

**Status Codes:**
- `200` - Server is healthy

---

## 👥 Agents API

Manage distributed cracking agents.

### `GET /api/v1/agents/`

List all registered agents.

**Response:**
```json
{
  "data": [
    {
      "id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
      "name": "GPU-Server-01",
      "ip_address": "192.168.1.100",
      "port": 8080,
      "status": "online",
      "capabilities": "RTX 4090, OpenCL",
      "last_seen": "2025-06-08T07:06:20.571239+07:00",
      "created_at": "2025-06-08T07:06:20.571239+07:00",
      "updated_at": "2025-06-08T07:06:20.571239+07:00"
    }
  ]
}
```

### `POST /api/v1/agents/`

Register a new agent.

**Request Body:**
```json
{
  "name": "GPU-Server-01",
  "ip_address": "192.168.1.100",
  "port": 8080,
  "capabilities": "RTX 4090, OpenCL"
}
```

**Response:**
```json
{
  "data": {
    "id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
    "name": "GPU-Server-01",
    "ip_address": "192.168.1.100",
    "port": 8080,
    "status": "online",
    "capabilities": "RTX 4090, OpenCL",
    "last_seen": "2025-06-08T07:06:20.571239+07:00",
    "created_at": "2025-06-08T07:06:20.571239+07:00",
    "updated_at": "2025-06-08T07:06:20.571239+07:00"
  }
}
```

**Status Codes:**
- `201` - Agent created successfully
- `400` - Invalid request body
- `500` - Internal server error

### `GET /api/v1/agents/{id}`

Get agent by ID.

**Parameters:**
- `id` (path) - Agent UUID

**Response:**
```json
{
  "data": {
    "id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
    "name": "GPU-Server-01",
    // ... other agent fields
  }
}
```

**Status Codes:**
- `200` - Success
- `400` - Invalid agent ID
- `404` - Agent not found

### `POST /api/v1/agents/{id}/heartbeat`

Update agent heartbeat.

**Parameters:**
- `id` (path) - Agent UUID

**Response:**
```json
{
  "message": "Heartbeat updated"
}
```

**Status Codes:**
- `200` - Heartbeat updated
- `400` - Invalid agent ID
- `500` - Internal server error

### `POST /api/v1/agents/{id}/files`

Register agent's local files with server.

**Parameters:**
- `id` (path) - Agent UUID

**Request Body:**
```json
{
  "agent_id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
  "files": {
    "rockyou.txt": {
      "name": "rockyou.txt",
      "path": "/root/uploads/wordlists/rockyou.txt",
      "size": 139921507,
      "type": "wordlist",
      "hash": "a1b2c3d4e5f6",
      "mod_time": "2025-06-08T00:00:00Z"
    }
  }
}
```

**Response:**
```json
{
  "message": "Agent files registered successfully",
  "agent_id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
  "file_count": 1
}
```

### `DELETE /api/v1/agents/{id}`

Delete an agent.

**Parameters:**
- `id` (path) - Agent UUID

**Response:**
```json
{
  "message": "Agent deleted successfully"
}
```

**Status Codes:**
- `200` - Agent deleted
- `400` - Invalid agent ID
- `500` - Internal server error

### `GET /api/v1/agents/{id}/jobs`

Get all jobs assigned to a specific agent.

**Parameters:**
- `id` (path) - Agent UUID

**Response:**
```json
{
  "data": [
    {
      "id": "job-uuid",
      "name": "WiFi Crack Job",
      "status": "pending",
      "hash_type": 2500,
      "attack_mode": 0,
      "agent_id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28",
      "progress": 0,
      "created_at": "2025-06-08T07:00:00Z"
    }
  ]
}
```

**Status Codes:**
- `200` - Success
- `400` - Invalid agent ID
- `500` - Internal server error

### `GET /api/v1/agents/{id}/jobs/next`

Get the next available job for an agent to execute (used by agent polling).

**Parameters:**
- `id` (path) - Agent UUID

**Response (when job available):**
```json
{
  "data": {
    "id": "job-uuid",
    "name": "WiFi Crack Job",
    "status": "pending",
    "hash_type": 2500,
    "attack_mode": 0,
    "hash_file": "/uploads/hashfiles/capture.hccapx",
    "wordlist": "rockyou.txt",
    "agent_id": "3aad5b48-bb85-4e2a-9814-bd9313ba8f28"
  }
}
```

**Response (when no jobs available):**
```json
{
  "data": null,
  "message": "No available jobs"
}
```

**Status Codes:**
- `200` - Success (whether job found or not)
- `400` - Invalid agent ID
- `500` - Internal server error

---

## ⚙️ Jobs API

Manage cracking jobs.

### `GET /api/v1/jobs/`

List all jobs.

**Response:**
```json
{
  "data": [
    {
      "id": "job-uuid",
      "name": "WiFi Crack Job",
      "status": "pending",
      "hash_type": 2500,
      "attack_mode": 0,
      "hash_file": "capture.hccapx",
      "hash_file_id": "hashfile-uuid",
      "wordlist": "rockyou.txt",
      "rules": "",
      "agent_id": null,
      "progress": 0,
      "speed": 0,
      "eta": null,
      "result": "",
      "created_at": "2025-06-08T07:00:00Z",
      "updated_at": "2025-06-08T07:00:00Z",
      "started_at": null,
      "completed_at": null
    }
  ]
}
```

### `POST /api/v1/jobs/`

Create a new job.

**Request Body:**
```json
{
  "name": "WiFi Crack Job",
  "hash_type": 2500,
  "attack_mode": 0,
  "hash_file_id": "hashfile-uuid",
  "wordlist_id": "wordlist-uuid",
  "rules": ""
}
```

**Response:**
```json
{
  "data": {
    "id": "job-uuid",
    "name": "WiFi Crack Job",
    "status": "pending",
    // ... other job fields
  }
}
```

**Status Codes:**
- `201` - Job created
- `400` - Invalid request body
- `500` - Internal server error

### `GET /api/v1/jobs/{id}`

Get job by ID.

**Parameters:**
- `id` (path) - Job UUID

**Response:**
```json
{
  "data": {
    "id": "job-uuid",
    "name": "WiFi Crack Job",
    // ... job details
  }
}
```

### `POST /api/v1/jobs/{id}/start`

Start a job.

**Parameters:**
- `id` (path) - Job UUID

**Response:**
```json
{
  "message": "Job started successfully"
}
```

**Status Codes:**
- `200` - Job started
- `400` - Invalid job ID or job cannot be started
- `500` - Internal server error

### `POST /api/v1/jobs/{id}/pause`

Pause a running job.

**Parameters:**
- `id` (path) - Job UUID

**Response:**
```json
{
  "message": "Job paused successfully"
}
```

### `POST /api/v1/jobs/{id}/resume`

Resume a paused job.

**Parameters:**
- `id` (path) - Job UUID

**Response:**
```json
{
  "message": "Job resumed successfully"
}
```

### `POST /api/v1/jobs/{id}/complete`

Mark job as completed.

**Parameters:**
- `id` (path) - Job UUID

**Request Body:**
```json
{
  "result": "password123:hash"
}
```

**Response:**
```json
{
  "message": "Job completed successfully"
}
```

### `POST /api/v1/jobs/{id}/fail`

Mark job as failed.

**Parameters:**
- `id` (path) - Job UUID

**Request Body:**
```json
{
  "error": "Reason for failure"
}
```

### `PUT /api/v1/jobs/{id}/progress`

Update job progress.

**Parameters:**
- `id` (path) - Job UUID

**Request Body:**
```json
{
  "progress": 25.5,
  "speed": 1234567,
  "eta": "2025-06-08T08:00:00Z"
}
```

**Response:**
```json
{
  "message": "Progress updated successfully"
}
```

### `DELETE /api/v1/jobs/{id}`

Delete a job.

**Parameters:**
- `id` (path) - Job UUID

**Response:**
```json
{
  "message": "Job deleted successfully"
}
```

---

## 📁 Hash Files API

Manage hash files for cracking.

### `GET /api/v1/hashfiles/`

List all hash files.

**Response:**
```json
{
  "data": [
    {
      "id": "hashfile-uuid",
      "name": "capture-01.hccapx",
      "orig_name": "wifi-handshake.hccapx",
      "path": "uploads/hashfiles/capture-01.hccapx",
      "size": 2048,
      "type": "hccapx",
      "created_at": "2025-06-08T07:00:00Z"
    }
  ]
}
```

### `POST /api/v1/hashfiles/upload`

Upload a hash file.

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file` - Hash file to upload

**Response:**
```json
{
  "data": {
    "id": "hashfile-uuid",
    "name": "capture-01.hccapx",
    "orig_name": "wifi-handshake.hccapx",
    "path": "uploads/hashfiles/capture-01.hccapx",
    "size": 2048,
    "type": "hccapx",
    "created_at": "2025-06-08T07:00:00Z"
  }
}
```

**Status Codes:**
- `201` - File uploaded successfully
- `400` - No file uploaded or invalid file
- `500` - Upload failed

### `GET /api/v1/hashfiles/{id}`

Get hash file details.

**Parameters:**
- `id` (path) - Hash file UUID

**Response:**
```json
{
  "data": {
    "id": "hashfile-uuid",
    "name": "capture-01.hccapx",
    // ... file details
  }
}
```

### `GET /api/v1/hashfiles/{id}/download`

Download a hash file.

**Parameters:**
- `id` (path) - Hash file UUID

**Response:** Binary file content

**Headers:**
- `Content-Disposition: attachment; filename=original-name.hccapx`
- `Content-Type: application/octet-stream`

**Status Codes:**
- `200` - File download successful
- `404` - File not found
- `500` - Download failed

### `DELETE /api/v1/hashfiles/{id}`

Delete a hash file.

**Parameters:**
- `id` (path) - Hash file UUID

**Response:**
```json
{
  "message": "Hash file deleted successfully"
}
```

---

## 📋 Wordlists API

Manage wordlist files for dictionary attacks.

### `GET /api/v1/wordlists/`

List all wordlists.

**Response:**
```json
{
  "data": [
    {
      "id": "wordlist-uuid",
      "name": "rockyou.txt",
      "orig_name": "rockyou.txt",
      "path": "uploads/wordlists/rockyou.txt",
      "size": 139921507,
      "word_count": 14344392,
      "created_at": "2025-06-08T07:00:00Z"
    }
  ]
}
```

### `POST /api/v1/wordlists/upload`

Upload a wordlist file.

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file` - Wordlist file to upload

**Response:**
```json
{
  "data": {
    "id": "wordlist-uuid",
    "name": "rockyou.txt",
    "orig_name": "rockyou.txt",
    "path": "uploads/wordlists/rockyou.txt",
    "size": 139921507,
    "word_count": 14344392,
    "created_at": "2025-06-08T07:00:00Z"
  }
}
```

**Status Codes:**
- `201` - Wordlist uploaded successfully
- `400` - No file uploaded or invalid file
- `500` - Upload failed

### `GET /api/v1/wordlists/{id}`

Get wordlist details.

**Parameters:**
- `id` (path) - Wordlist UUID

**Response:**
```json
{
  "data": {
    "id": "wordlist-uuid",
    "name": "rockyou.txt",
    // ... wordlist details
  }
}
```

### `GET /api/v1/wordlists/{id}/download`

Download a wordlist file.

**Parameters:**
- `id` (path) - Wordlist UUID

**Response:** Text file content

**Headers:**
- `Content-Disposition: attachment; filename=rockyou.txt`
- `Content-Type: text/plain`

### `DELETE /api/v1/wordlists/{id}`

Delete a wordlist.

**Parameters:**
- `id` (path) - Wordlist UUID

**Response:**
```json
{
  "message": "Wordlist deleted successfully"
}
```

---

## 🔄 Job Assignment API

### `POST /api/v1/jobs/assign`

Assign pending jobs to available agents.

**Response:**
```json
{
  "message": "Jobs assigned successfully",
  "assigned_count": 3
}
```

---

## 📊 Status Codes

| Code | Description |
|------|-------------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request |
| `404` | Not Found |
| `500` | Internal Server Error |

---

## 🔒 Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

**Common Errors:**
- Invalid UUID format
- Missing required fields
- File upload failures
- Resource not found
- Database connection issues

---

## 📝 Usage Examples

### Complete Workflow Example

```bash
# 1. Register an agent
curl -X POST http://localhost:1337/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GPU-Server-01",
    "ip_address": "192.168.1.100",
    "port": 8080,
    "capabilities": "RTX 4090"
  }'

# 2. Upload hash file
curl -X POST http://localhost:1337/api/v1/hashfiles/upload \
  -F "file=@capture.hccapx"

# 3. Upload wordlist
curl -X POST http://localhost:1337/api/v1/wordlists/upload \
  -F "file=@rockyou.txt"

# 4. Create job
curl -X POST http://localhost:1337/api/v1/jobs/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WiFi Crack Job",
    "hash_type": 2500,
    "attack_mode": 0,
    "hash_file_id": "hashfile-uuid",
    "wordlist_id": "wordlist-uuid"
  }'

# 5. Start job
curl -X POST http://localhost:1337/api/v1/jobs/job-uuid/start

# 6. Check job status
curl http://localhost:1337/api/v1/jobs/job-uuid
```

### Performance Testing

Use the provided benchmark script:
```bash
bash scripts/benchmark_api.sh
```

---

## 🌐 Frontend Integration

The API is designed to work seamlessly with the modern TypeScript frontend:

- **Development**: Frontend proxy at `localhost:3000` → API at `localhost:1337`
- **Production**: Serve frontend static files from web server with API proxy
- **Real-time**: Frontend polls API every 10 seconds for updates

---

## 🔧 Rate Limiting & Performance

- **No rate limiting** currently implemented (add for production)
- **Response times**: < 5ms for most endpoints
- **Concurrent requests**: Supports high concurrency with Go's goroutines
- **File uploads**: Streaming uploads for large files
- **Database**: Optimized SQLite with WAL mode and indexes

---

## 🚀 Next Steps

1. **Authentication**: Add API keys or JWT tokens
2. **Rate Limiting**: Implement per-client rate limits
3. **WebSockets**: Real-time job progress updates
4. **Pagination**: For large datasets
5. **API Versioning**: Support v2 endpoints
6. **OpenAPI/Swagger**: Interactive API documentation 
