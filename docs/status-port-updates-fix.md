# Status and Port Updates Fix

## Problem Description

Agent mengalami error saat mencoba mengupdate status dan port:

```
🔄 Updating agent status to online and port to 8081...
⚠️ Warning: Failed to update agent status to online: gagal update agent info: 

🔄 Updating agent status to offline and restoring port to 8080...
⚠️ Warning: Failed to update agent status to offline: gagal update agent info: 
```

**Result:**
- Status berhasil diupdate menjadi `offline` (mungkin oleh sistem lain)
- Port tidak berhasil diupdate dari `8081` ke `8080`
- Database tetap menunjukkan port `8081` meskipun agent sudah offline

## Root Cause Analysis

### **1. Wrong Endpoint Usage**

**Before (Problematic):**
```go
func (a *Agent) updateAgentInfo(agentID uuid.UUID, ip string, port int, capabilities string, status string) error {
    // Using wrong endpoint
    url := fmt.Sprintf("%s/api/v1/agents/%s", a.ServerURL, agentID.String())
    httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
    // ...
}
```

**Problem:**
- Endpoint `PUT /api/v1/agents/{id}` tidak tersedia di server
- Request gagal dengan error "gagal update agent info"
- Port tidak berhasil diupdate

### **2. Available Endpoints**

Server menyediakan endpoint yang berbeda:

1. **`POST /api/v1/agents/update-data`** - Untuk update data (ip, port, capabilities) tanpa status
2. **`PUT /api/v1/agents/{id}/status`** - Untuk update status saja
3. **`PUT /api/v1/agents/{id}`** - Tidak tersedia (yang digunakan sebelumnya)

## Solution Implemented

### **1. Fixed updateAgentInfo Function**

**After (Fixed):**
```go
func (a *Agent) updateAgentInfo(agentID uuid.UUID, ip string, port int, capabilities string, status string) error {
    // Use the correct endpoint for updating agent data
    req := struct {
        AgentKey     string `json:"agent_key"`
        IPAddress    string `json:"ip_address"`
        Port         int    `json:"port"`
        Capabilities string `json:"capabilities"`
    }{
        AgentKey:     a.AgentKey,
        IPAddress:    ip,
        Port:         port,
        Capabilities: capabilities,
    }

    jsonData, _ := json.Marshal(req)
    url := fmt.Sprintf("%s/api/v1/agents/update-data", a.ServerURL)

    httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := a.Client.Do(httpReq)
    if err != nil {
        return fmt.Errorf("failed to update agent data: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to update agent data: %s", string(body))
    }

    // If status needs to be updated, use the status endpoint
    if status != "" {
        statusReq := struct {
            Status string `json:"status"`
        }{
            Status: status,
        }

        statusData, _ := json.Marshal(statusReq)
        statusURL := fmt.Sprintf("%s/api/v1/agents/%s/status", a.ServerURL, agentID.String())

        statusHttpReq, _ := http.NewRequest(http.MethodPut, statusURL, bytes.NewBuffer(statusData))
        statusHttpReq.Header.Set("Content-Type", "application/json")

        statusResp, err := a.Client.Do(statusHttpReq)
        if err != nil {
            return fmt.Errorf("failed to update agent status: %w", err)
        }
        defer statusResp.Body.Close()

        if statusResp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(statusResp.Body)
            return fmt.Errorf("failed to update agent status: %s", string(body))
        }
    }

    return nil
}
```

### **2. Key Changes**

1. **Separated Data and Status Updates:**
   - Data updates (ip, port, capabilities) → `POST /api/v1/agents/update-data`
   - Status updates → `PUT /api/v1/agents/{id}/status`

2. **Correct Endpoint Usage:**
   - Changed from `PUT /api/v1/agents/{id}` to `POST /api/v1/agents/update-data`
   - Added separate call to `PUT /api/v1/agents/{id}/status` for status updates

3. **Better Error Handling:**
   - More descriptive error messages
   - Separate error handling for data vs status updates

4. **Proper Request Structure:**
   - Data update includes `agent_key` field
   - Status update only includes `status` field

## How It Works Now

### **1. Agent Startup Flow**

```
Agent Binary Starts
         ↓
Register with Server
         ↓
Update Data: POST /api/v1/agents/update-data
         ↓
Update Status: PUT /api/v1/agents/{id}/status
         ↓
Agent Running (Status: online, Port: 8081)
```

### **2. Agent Shutdown Flow**

```
Ctrl+C Signal Received
         ↓
Graceful Shutdown Process
         ↓
Update Data: POST /api/v1/agents/update-data (port: 8080)
         ↓
Update Status: PUT /api/v1/agents/{id}/status (offline)
         ↓
Agent Exited (Status: offline, Port: 8080)
```

### **3. API Calls Made**

**Startup:**
1. `POST /api/v1/agents/update-data` - Update port to 8081
2. `PUT /api/v1/agents/{id}/status` - Update status to online

**Shutdown:**
1. `POST /api/v1/agents/update-data` - Restore port to 8080
2. `PUT /api/v1/agents/{id}/status` - Update status to offline

## Expected Results

### **Before Fix:**
```
🔄 Updating agent status to online and port to 8081...
⚠️ Warning: Failed to update agent status to online: gagal update agent info: 

🔄 Updating agent status to offline and restoring port to 8080...
⚠️ Warning: Failed to update agent status to offline: gagal update agent info: 
```

**Database State:**
```
GPU-Agent, 30.30.30.39, 8081, offline, GPU, 3730b5d6, ...
```
- Status: offline (updated by system)
- Port: 8081 (failed to update)

### **After Fix:**
```
🔄 Updating agent status to online and port to 8081...
✅ Agent status updated to online with port 8081

🔄 Updating agent status to offline and restoring port to 8080...
✅ Agent status updated to offline with port 8080
```

**Database State:**
```
GPU-Agent, 30.30.30.39, 8080, offline, GPU, 3730b5d6, ...
```
- Status: offline (successfully updated)
- Port: 8080 (successfully restored)

## Testing

### **Run Test Script**

```bash
./scripts/test_fixed_status_port_updates.sh
```

### **Manual Testing**

1. **Start agent and check for errors:**
   ```bash
   sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key"
   # Should see: "✅ Agent status updated to online with port 8081"
   ```

2. **Stop agent with Ctrl+C and check for errors:**
   ```bash
   # Press Ctrl+C
   # Should see: "✅ Agent status updated to offline with port 8080"
   ```

3. **Check database state:**
   ```bash
   # Port should change from 8081 to 8080
   # Status should change from online to offline
   ```

## Benefits of the Fix

### **1. Successful Updates**
- ✅ **Status updates**: online ↔ offline berhasil
- ✅ **Port updates**: 8081 ↔ 8080 berhasil
- ✅ **No more errors**: Tidak ada lagi "Failed to update agent info"

### **2. Correct Endpoint Usage**
- ✅ **Data updates**: Menggunakan endpoint yang benar
- ✅ **Status updates**: Menggunakan endpoint yang benar
- ✅ **Proper separation**: Data dan status diupdate secara terpisah

### **3. Better Error Handling**
- ✅ **Descriptive errors**: Error message yang lebih jelas
- ✅ **Separate handling**: Error handling untuk data vs status
- ✅ **Proper logging**: Log yang informatif untuk debugging

## Troubleshooting

### **If Updates Still Fail:**

1. **Check server endpoints:**
   ```bash
   # Verify endpoints are available
   curl -X POST http://localhost:1337/api/v1/agents/update-data
   curl -X PUT http://localhost:1337/api/v1/agents/{id}/status
   ```

2. **Check agent logs:**
   ```bash
   # Look for specific error messages
   sudo ./bin/agent ... 2>&1 | grep -E "(🔄|✅|⚠️|❌)"
   ```

3. **Verify database connection:**
   ```bash
   # Check if server can connect to database
   # Check if agent can reach server
   ```

## Conclusion

Fix yang diterapkan menyelesaikan masalah:

- **🎯 Correct Endpoints**: Menggunakan endpoint yang tersedia dan benar
- **🔄 Successful Updates**: Status dan port berhasil diupdate
- **📊 Consistent State**: Database selalu reflect state yang akurat
- **🚀 Better Reliability**: Tidak ada lagi error "Failed to update agent info"

Agent sekarang akan berhasil mengupdate status dan port tanpa error, memberikan monitoring yang akurat dan reliable! 🎉
