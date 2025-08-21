# Agent Startup Solution

## 🎯 **Masalah yang Diatasi**

Sebelumnya, ketika menjalankan agent dengan command:
```bash
sudo ./bin/agent --server http://30.30.30.102:1337 --name test-agent-003 --ip "30.30.30.39" --agent-key "d8675fb7"
```

Agent mengalami beberapa masalah:
1. **404 Error pada Heartbeat**: Endpoint `/api/v1/agents/{id}/heartbeat` tidak ditemukan
2. **Tidak ada validasi agent key**: Agent bisa start tanpa validasi
3. **Tidak ada logic untuk update data**: Data agent tidak ter-update dengan benar
4. **Status management tidak konsisten**: Status tidak berubah sesuai kondisi

## ✅ **Solusi yang Diimplementasikan**

### **1. Endpoint Agent Startup (`POST /api/v1/agents/startup`)**

Endpoint ini akan melakukan validasi dan update data agent saat startup:

#### **Request Body:**
```json
{
  "name": "test-agent-003",
  "agent_key": "d8675fb7",
  "ip_address": "30.30.30.39",
  "port": 8080,
  "capabilities": "CPU"
}
```

#### **Logic Flow:**
1. **Validasi Agent Key**: Cek apakah agent key ada di database
2. **Validasi Agent Name**: Cek apakah nama sesuai dengan agent key
3. **Check Existing Data**: Cek apakah agent sudah punya IP, port, capabilities
4. **Update Logic**:
   - Jika sudah ada data → Update status ke "online" saja
   - Jika belum ada data → Update semua data + status "online"

#### **Response Codes:**
- `AGENT_ALREADY_EXISTS`: Agent sudah ada dengan data lengkap
- `AGENT_UPDATED`: Agent data berhasil di-update
- `AGENT_KEY_NOT_FOUND`: Agent key tidak ditemukan
- `AGENT_NAME_MISMATCH`: Nama tidak sesuai dengan agent key

### **2. Endpoint Agent Heartbeat (`POST /api/v1/agents/heartbeat`)**

Endpoint ini untuk update heartbeat agent menggunakan agent key:

#### **Request Body:**
```json
{
  "agent_key": "d8675fb7"
}
```

#### **Response:**
```json
{
  "message": "Agent heartbeat updated successfully",
  "data": {
    "id": "uuid",
    "name": "test-agent-003",
    "status": "online",
    "updated_at": "2025-08-14T21:37:18+07:00"
  }
}
```

### **3. Implementation Details**

#### **Handler Implementation:**
```go
// AgentStartup handles agent startup and validation
func (h *AgentHandler) AgentStartup(c *gin.Context) {
    // 1. Validate request body
    // 2. Check agent key exists
    // 3. Validate agent name matches
    // 4. Check if agent has existing data
    // 5. Update accordingly (status only or full data)
}
```

#### **Usecase Methods Added:**
```go
type AgentUsecase interface {
    // ... existing methods ...
    GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error)
}
```

#### **Router Configuration:**
```go
agents := v1.Group("/agents")
{
    agents.POST("/startup", agentHandler.AgentStartup)     // New
    agents.POST("/heartbeat", agentHandler.AgentHeartbeat) // New
    // ... existing routes ...
}
```

## 🔧 **Cara Kerja Agent Startup**

### **Scenario 1: Agent Baru (Belum Ada Data)**
1. Agent kirim request ke `/api/v1/agents/startup`
2. Validasi agent key dan nama
3. Update data agent (IP, port, capabilities)
4. Set status ke "online"
5. Response: `AGENT_UPDATED`

### **Scenario 2: Agent Sudah Ada (Data Lengkap)**
1. Agent kirim request ke `/api/v1/agents/startup`
2. Validasi agent key dan nama
3. Deteksi agent sudah punya data lengkap
4. Update status ke "online" saja
5. Response: `AGENT_ALREADY_EXISTS`

### **Scenario 3: Agent Key Tidak Valid**
1. Agent kirim request dengan agent key salah
2. Validasi gagal
3. Response: `AGENT_KEY_NOT_FOUND`
4. Agent berhenti (tidak bisa start)

### **Scenario 4: Agent Name Mismatch**
1. Agent kirim request dengan nama salah
2. Validasi nama gagal
3. Response: `AGENT_NAME_MISMATCH`
4. Agent berhenti (tidak bisa start)

## 🎨 **User Experience Improvements**

### **Sebelumnya (Problematic)**:
- ❌ Agent bisa start tanpa validasi
- ❌ 404 error pada heartbeat
- ❌ Data agent tidak ter-update
- ❌ Status management tidak konsisten

### **Sekarang (Improved)**:
- ✅ Agent key harus valid untuk bisa start
- ✅ Agent name harus sesuai dengan agent key
- ✅ Data agent ter-update otomatis saat startup
- ✅ Status management konsisten (online/offline)
- ✅ Heartbeat berfungsi dengan agent key

## 🧪 **Testing**

### **Test Script Created:**
- `scripts/test_agent_startup.sh` - Test lengkap agent startup

### **Test Scenarios:**
1. **Generate Agent Key**: Buat agent key untuk testing
2. **Agent Startup New Data**: Test startup dengan data baru
3. **Agent Heartbeat**: Test heartbeat functionality
4. **Agent Startup Again**: Test startup kedua (should show already exists)
5. **Check Agent Status**: Verifikasi status agent
6. **Invalid IP Test**: Test dengan IP yang berbeda
7. **Wrong Name Test**: Test dengan nama yang salah
8. **Invalid Key Test**: Test dengan agent key yang salah

## 📁 **Files Modified**

### **Backend**:
- `internal/delivery/http/handler/agent_handler.go` - New endpoints
- `internal/usecase/agent_usecase.go` - Added GetByAgentKey method
- `internal/delivery/http/router.go` - New routes

### **Testing**:
- `scripts/test_agent_startup.sh` - Test script

### **Documentation**:
- `docs/agent-startup-solution.md` - This file

## 🚀 **How to Use**

### **1. Agent Startup**
```bash
curl -X POST http://localhost:1337/api/v1/agents/startup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-agent-003",
    "agent_key": "d8675fb7",
    "ip_address": "30.30.30.39",
    "port": 8080,
    "capabilities": "CPU"
  }'
```

### **2. Agent Heartbeat**
```bash
curl -X POST http://localhost:1337/api/v1/agents/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_key": "d8675fb7"
  }'
```

### **3. Agent Command**
```bash
sudo ./bin/agent \
  --server http://30.30.30.102:1337 \
  --name test-agent-003 \
  --ip "30.30.30.39" \
  --agent-key "d8675fb7"
```

## 🎯 **Expected Results**

**Sekarang ketika Anda menjalankan agent, sistem akan:**

1. **Validasi Agent Key**: Pastikan agent key ada di database
2. **Validasi Agent Name**: Pastikan nama sesuai dengan agent key
3. **Update Data Agent**: Update IP, port, capabilities jika belum ada
4. **Set Status Online**: Set status agent ke "online"
5. **Handle Heartbeat**: Update last seen saat heartbeat
6. **Graceful Shutdown**: Set status ke "offline" saat agent berhenti

### **Success Scenarios:**
- ✅ Agent start dengan data baru → Data ter-update + status online
- ✅ Agent start dengan data existing → Status online saja
- ✅ Heartbeat berfungsi → Last seen ter-update
- ✅ Agent stop → Status offline

### **Failure Scenarios:**
- ❌ Agent key tidak valid → Agent berhenti
- ❌ Agent name tidak sesuai → Agent berhenti
- ❌ Server tidak accessible → Agent berhenti

## 🔍 **Troubleshooting**

### **Jika Agent Tidak Bisa Start:**
1. Check agent key valid di database
2. Check agent name sesuai dengan agent key
3. Check server endpoint accessible
4. Check network connectivity

### **Jika Heartbeat 404:**
1. Use endpoint `/api/v1/agents/heartbeat` (not `/api/v1/agents/{id}/heartbeat`)
2. Send agent_key in request body
3. Check server logs for errors

### **Jika Status Tidak Update:**
1. Check agent startup response
2. Verify agent data in database
3. Check server logs for errors

## 📚 **References**

- [Gin Framework](https://gin-gonic.com/)
- [Go HTTP Testing](https://golang.org/pkg/net/http/httptest/)
- [Testify Mock](https://github.com/stretchr/testify#mock-package)
- [REST API Design](https://restfulapi.net/)
