# Status and Port Updates for Agent

## Overview

Agent sekarang memiliki fitur otomatis untuk mengupdate status dan port berdasarkan kondisi running:

- **Startup**: Status → `online`, Port → `8081`
- **Running**: Maintains `online` status dan port `8081`
- **Shutdown (Ctrl+C)**: Status → `offline`, Port → `8080`

## Features

### ✅ **Automatic Status Updates**
- Status berubah otomatis dari `offline` ke `online` saat agent start
- Status berubah otomatis dari `online` ke `offline` saat agent stop
- Semua perubahan status langsung di-reflect ke database

### ✅ **Automatic Port Updates**
- Port berubah otomatis dari `8080` ke `8081` saat agent running
- Port restore otomatis dari `8081` ke `8080` saat agent shutdown
- Port management yang konsisten dan predictable

### ✅ **Real-time Database Updates**
- Database terupdate secara real-time dengan status dan port terbaru
- Monitoring agent status menjadi lebih akurat
- Port tracking untuk debugging dan management

## How It Works

### **1. Agent Startup Flow**

```
Agent Binary Starts
         ↓
Register with Server
         ↓
Update Status: offline → online
         ↓
Update Port: 8080 → 8081
         ↓
Start Background Services
         ↓
Agent Running (Status: online, Port: 8081)
```

### **2. Agent Shutdown Flow**

```
Ctrl+C Signal Received
         ↓
Graceful Shutdown Process
         ↓
Update Status: online → offline
         ↓
Restore Port: 8081 → 8080
         ↓
Cleanup Resources
         ↓
Agent Exited (Status: offline, Port: 8080)
```

### **3. Database State Changes**

| State | Status | Port | Capabilities | Description |
|-------|--------|------|--------------|-------------|
| **Initial** | `offline` | `8080` | (empty) | After agent key generation |
| **Running** | `online` | `8081` | `CPU` | When agent is active |
| **Shutdown** | `offline` | `8080` | `CPU` | After Ctrl+C |

## Implementation Details

### **Startup Status and Port Update**

```go
// ✅ Update status to online and port to 8081 when agent starts running
log.Printf("🔄 Updating agent status to online and port to 8081...")
if err := agent.updateAgentInfo(agent.ID, ip, 8081, capabilities, "online"); err != nil {
    log.Printf("⚠️ Warning: Failed to update agent status to online: %v", err)
} else {
    log.Printf("✅ Agent status updated to online with port 8081")
}
```

### **Shutdown Status and Port Update**

```go
// ✅ Update status to offline and restore original port 8080 before shutdown
log.Printf("🔄 Updating agent status to offline and restoring port to 8080...")
if err := agent.updateAgentInfo(agent.ID, ip, 8080, capabilities, "offline"); err != nil {
    log.Printf("⚠️ Warning: Failed to update agent status to offline: %v", err)
} else {
    log.Printf("✅ Agent status updated to offline with port 8080")
}
```

### **Port Management Logic**

```go
// Port changes during agent lifecycle:
// 1. Initial: 8080 (from database or default)
// 2. Startup: 8080 → 8081 (agent becomes active)
// 3. Running: 8081 (maintained while active)
// 4. Shutdown: 8081 → 8080 (restored to original)
```

## Usage Examples

### **1. Basic Agent Startup**

```bash
sudo ./bin/agent \
  --server http://172.15.2.76:1337 \
  --name GPU-Agent \
  --agent-key "3730b5d6"
```

**Expected Output:**
```
✅ IP address validation passed: 30.30.30.39 is a valid local IP
🔍 Auto-detected capabilities using hashcat -I: CPU
🔍 Detected device type: CPU
✅ CPU device detected: CPU
✅ Capabilities updated successfully
🔄 Updating agent status to online and port to 8081...
✅ Agent status updated to online with port 8081
✅ Agent registered successfully
```

**Database State:**
```
GPU-Agent, 30.30.30.39, 8081, online, CPU, 3730b5d6, ...
```

### **2. Agent Shutdown with Ctrl+C**

**When you press Ctrl+C:**
```
^C
Shutting down agent...
🔄 Updating agent status to offline and restoring port to 8080...
✅ Agent status updated to offline with port 8080
Agent exited
```

**Database State:**
```
GPU-Agent, 30.30.30.39, 8080, offline, CPU, 3730b5d6, ...
```

### **3. Agent with Specific Port**

```bash
sudo ./bin/agent \
  --server http://172.15.2.76:1337 \
  --name GPU-Agent \
  --port 8082 \
  --agent-key "3730b5d6"
```

**Port Behavior:**
- **Startup**: Port 8082 → 8081 (overridden to running port)
- **Running**: Port 8081 (maintained)
- **Shutdown**: Port 8081 → 8080 (restored to original)

## Benefits

### **1. Real-time Monitoring**
- ✅ **Status tracking**: Bisa monitor agent status secara real-time
- ✅ **Port tracking**: Bisa track port changes untuk debugging
- ✅ **Database consistency**: Database selalu up-to-date

### **2. Better Management**
- ✅ **Predictable behavior**: Port dan status berubah secara predictable
- ✅ **Easy debugging**: Port 8081 = running, Port 8080 = stopped
- ✅ **Resource tracking**: Bisa track resource usage berdasarkan status

### **3. Improved Reliability**
- ✅ **Automatic updates**: Tidak perlu manual update status/port
- ✅ **Consistent state**: Status dan port selalu konsisten
- ✅ **Graceful shutdown**: Clean shutdown dengan state restoration

## Testing

### **Run Test Script**

```bash
./scripts/test_status_port_updates.sh
```

### **Manual Testing**

1. **Start agent and check status:**
   ```bash
   sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key"
   # Check database: status should be 'online', port should be 8081
   ```

2. **Stop agent with Ctrl+C and check status:**
   ```bash
   # Press Ctrl+C in the terminal
   # Check database: status should be 'offline', port should be 8080
   ```

3. **Monitor database changes:**
   ```bash
   # Watch database changes in real-time
   # Status and port should update automatically
   ```

## Troubleshooting

### **Common Issues**

1. **Status not updating to online:**
   - Check if `updateAgentInfo` function is working
   - Verify database connection
   - Check agent logs for errors

2. **Port not changing to 8081:**
   - Check if port update is being called
   - Verify port parameter in update request
   - Check database for port changes

3. **Status not updating to offline:**
   - Check if shutdown signal is being received
   - Verify shutdown process is working
   - Check if update is called before exit

### **Debug Mode**

Enable debug logging to track status and port updates:

```bash
sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key" 2>&1 | grep -E "(🔄|✅|⚠️|❌)"
```

## Configuration

### **Port Configuration**

- **Default port**: 8080 (from database or default)
- **Running port**: 8081 (hardcoded for running state)
- **Restore port**: 8080 (restored on shutdown)

### **Status Configuration**

- **Initial status**: `offline` (from database)
- **Running status**: `online` (set on startup)
- **Shutdown status**: `offline` (set on shutdown)

## Conclusion

Fitur status dan port updates yang baru memberikan:

- **🎯 Real-time Monitoring**: Status dan port terupdate secara real-time
- **🔄 Automatic Management**: Tidak perlu manual update status/port
- **📊 Consistent State**: Database selalu reflect state yang akurat
- **🚀 Better Debugging**: Port tracking untuk troubleshooting
- **⚡ Improved Reliability**: Graceful shutdown dengan state restoration

Agent sekarang memiliki lifecycle management yang lebih baik dengan status dan port yang selalu up-to-date di database! 🎉
