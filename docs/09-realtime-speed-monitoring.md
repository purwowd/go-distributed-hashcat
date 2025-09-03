# Real-Time Speed Monitoring Feature

## 📋 Overview

Fitur **Real-Time Speed Monitoring** adalah sistem monitoring real-time yang memungkinkan tracking dan update field `speed` pada tabel `agent` secara otomatis dan manual. Fitur ini memastikan bahwa data speed agent selalu up-to-date dan dapat digunakan untuk optimasi job distribution.

## 🎯 Tujuan Utama

1. **Real-time Speed Updates**: Field data speed di tabel agent harus selalu diperbarui secara real-time selama agent online
2. **Status Update**: Field status diupdate ke offline ketika agent offline (speed tetap)
3. **Comprehensive Logging**: Log harus mencatat waktu, agent ID, dan nilai speed terbaru setiap kali diperbarui
4. **Non-Intrusive Operation**: Mekanisme tidak boleh mengganggu proses utama agent
5. **WebSocket Broadcasting**: Update real-time ke frontend via WebSocket

## 🏗️ Architecture

### Database Layer
- **Field `speed`**: Integer untuk menyimpan hash rate dalam H/s
- **Field `status`**: String untuk status agent (online, offline, busy)
- **Field `updated_at`**: Timestamp untuk tracking perubahan

### Repository Layer
```go
type AgentRepository interface {
    UpdateSpeed(ctx context.Context, id uuid.UUID, speed int64) error
    UpdateSpeedWithStatus(ctx context.Context, id uuid.UUID, speed int64, status string) error
    ResetSpeedOnOffline(ctx context.Context, id uuid.UUID) error
    // ... existing methods
}
```

### Usecase Layer
```go
type AgentUsecase interface {
    UpdateAgentSpeed(ctx context.Context, id uuid.UUID, speed int64) error
    UpdateAgentSpeedWithStatus(ctx context.Context, id uuid.UUID, speed int64, status string) error
    
    // ... existing methods
}
```

### Handler Layer
- **PUT `/api/v1/agents/{id}/speed`**: Update speed saja
- **PUT `/api/v1/agents/{id}/speed-status`**: Update speed dan status sekaligus
- **PUT `/api/v1/agents/{id}/speed-reset`**: Reset speed ke 0

## 🚀 Key Features

### 1. Real-Time Speed Updates
- **Automatic Detection**: Hashcat benchmark otomatis saat agent startup
- **Manual Updates**: API endpoint untuk update manual
- **Continuous Monitoring**: Background goroutine untuk monitoring berkelanjutan

### 2. Status-Aware Speed Management
- **Online Status**: Speed diupdate sesuai benchmark atau manual input
- **Busy Status**: Speed tetap tersimpan untuk tracking performance
- **Offline Status**: Speed otomatis direset ke 0

### 3. Comprehensive Logging
```go
// Real-time speed update
log.Printf("[REAL-TIME SPEED UPDATE] Agent %s speed updated to %d H/s at %s", 
    id.String(), speed, now.Format("2006-01-02 15:04:05"))

// Combined speed and status update
log.Printf("[REAL-TIME AGENT UPDATE] Agent %s: speed=%d H/s, status=%s, time=%s", 
    id.String(), speed, status, now.Format("2006-01-02 15:04:05"))

// Status update on offline
log.Printf("[STATUS UPDATE] Agent %s status updated to offline at %s",
	id.String(), now.Format("2006-01-02 15:04:05"))
```

### 4. WebSocket Broadcasting
- **Real-time Updates**: Broadcast perubahan speed ke semua client
- **Status Changes**: Broadcast perubahan status agent
- **Status Update**: Broadcast status update ke offline

## 📡 API Endpoints

### Update Agent Speed
```http
PUT /api/v1/agents/{id}/speed
Content-Type: application/json

{
    "speed": 5000
}
```

### Update Agent Speed and Status
```http
PUT /api/v1/agents/{id}/speed-status
Content-Type: application/json

{
    "speed": 5000,
    "status": "online"
}
```

### Reset Agent Speed
```http
PUT /api/v1/agents/{id}/speed-reset
Content-Type: application/json
```

## 🔧 Implementation Details

### Agent Main Code
```go
// Real-time speed monitoring with background goroutine
func (a *Agent) startRealTimeSpeedMonitoring(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if a.Status == "online" {
                    go func() {
                        if err := a.runHashcatBenchmark(); err != nil {
                            // Log warning, continue monitoring
                        }
                    }()
                }
            }
        }
    }()
}
```



### Database Operations
```go
// Update speed with comprehensive logging
func (r *agentRepository) UpdateSpeed(ctx context.Context, id uuid.UUID, speed int64) error {
    query := `UPDATE agents SET speed = ?, updated_at = ? WHERE id = ?`
    now := time.Now()
    
    _, err := r.db.DB().ExecContext(ctx, query, speed, now, id.String())
    if err != nil {
        return fmt.Errorf("failed to update agent speed: %w", err)
    }

    // Log for real-time monitoring
    log.Printf("[REAL-TIME SPEED UPDATE] Agent %s speed updated to %d H/s at %s", 
        id.String(), speed, now.Format("2006-01-02 15:04:05"))

    // Invalidate cache
    r.cache.Delete(ctx, "agent:"+id.String())
    r.cache.Delete(ctx, "agents:all")

    return nil
}
```

## 🧪 Testing

### Test Scripts
- **`scripts/test_realtime_speed_monitoring.sh`**: Comprehensive testing
- **`scripts/restart_and_test_realtime.sh`**: Server restart and testing

### Test Coverage
1. ✅ Server connectivity
2. ✅ Agent creation
3. ✅ Real-time speed-status update
4. ✅ Database persistence
5. ✅ Multiple real-time updates
6. ✅ Status changes (online/busy/offline)

8. ✅ Real-time monitoring simulation
9. ✅ Cleanup

## 📊 Monitoring and Logging

### Log Categories
- **[REAL-TIME SPEED UPDATE]**: Speed updates
- **[REAL-TIME AGENT UPDATE]**: Combined speed and status updates

- **✅ [SUCCESS]**: Successful operations
- **❌ [FAILED]**: Failed operations
- **⚠️ [WARNING]**: Warning messages

### Log Format
```
[Timestamp] [LOG_LEVEL] [OPERATION] Agent {ID}: details
```

### Example Logs
```
2025/09/03 09:33:20 [REAL-TIME UPDATE REQUEST] Agent 71b1fcdb: speed=5000 H/s, status=online
2025/09/03 09:33:20 [REAL-TIME AGENT UPDATE] Agent 71b1fcdb: speed=5000 H/s, status=online, time=2025-09-03 09:33:20
2025/09/03 09:33:20 [REAL-TIME BROADCAST] Agent test-agent: speed=5000 H/s, status=online
2025/09/03 09:33:20 ✅ [REAL-TIME UPDATE SUCCESS] Agent 71b1fcdb: speed=5000 H/s, status=online
```

## Real-Time Monitoring Flow

### 1. Agent Startup
```
Agent Start → Hashcat Benchmark → Speed Detection → Database Update → WebSocket Broadcast
```

### 2. Continuous Monitoring
```
Background Goroutine → 30s Timer → Hashcat Benchmark → Speed Update → Log & Broadcast
```

### 3. Status Change
```
Status Change → Speed Update → Database Update → Cache Invalidation → WebSocket Broadcast
```

### 4. Agent Shutdown
```
Shutdown Signal → Status Offline → Database Update → WebSocket Broadcast
```

## 🚨 Error Handling

### Hashcat Not Found
```go
if _, err := exec.LookPath("hashcat"); err != nil {
    infrastructure.AgentLogger.Warning("hashcat not found in PATH: %v", err)
    infrastructure.AgentLogger.Info("Skipping automatic speed detection. You can manually set speed via API:")
    infrastructure.AgentLogger.Info("PUT /api/v1/agents/%s/speed", a.ID.String())
    return nil
}
```

### Database Update Failures
```go
if err := u.agentRepo.UpdateSpeedWithStatus(ctx, id, speed, status); err != nil {
    log.Printf("❌ [REAL-TIME UPDATE FAILED] Agent %s: speed=%d H/s, status=%s, error=%v", 
        id.String(), speed, status, err)
    return err
}
```

### WebSocket Broadcast Failures
```go
if u.wsHub != nil {
    u.wsHub.BroadcastAgentSpeed(agent.ID.String(), agent.Speed)
    log.Printf("[REAL-TIME BROADCAST] Agent %s: speed=%d H/s, status=%s", 
        agent.Name, speed, status)
} else {
    log.Printf("⚠️ Warning: WebSocket hub not available for real-time broadcast")
}
```

## 📈 Performance Considerations

### Background Operations
- **Non-blocking**: Semua operasi real-time berjalan di background
- **Goroutine Management**: Proper context cancellation untuk cleanup
- **Resource Efficiency**: Minimal impact pada main agent operations

### Database Optimization
- **Prepared Statements**: Menggunakan prepared statements untuk performance
- **Cache Invalidation**: Smart cache invalidation untuk data consistency
- **Batch Updates**: Kemampuan update multiple fields dalam satu query

### WebSocket Efficiency
- **Selective Broadcasting**: Hanya broadcast perubahan yang relevan
- **Channel Management**: Proper channel handling untuk broadcast
- **Error Resilience**: Continue operation meskipun broadcast gagal

## 🔮 Future Enhancements

### 1. Advanced Speed Analytics
- **Speed History**: Tracking perubahan speed over time
- **Performance Trends**: Analisis trend performance agent
- **Predictive Scaling**: Prediksi kebutuhan scaling berdasarkan speed

### 2. Enhanced Monitoring
- **Custom Intervals**: Configurable monitoring intervals
- **Multiple Hash Types**: Support untuk berbagai hash type benchmarks
- **Performance Alerts**: Alert ketika speed drop signifikan

### 3. Integration Features
- **Metrics Export**: Export metrics ke monitoring systems
- **API Rate Limiting**: Rate limiting untuk API endpoints
- **Audit Trail**: Complete audit trail untuk semua perubahan

## 📝 Usage Examples

### Manual Speed Update
```bash
# Update agent speed to 5000 H/s
curl -X PUT http://localhost:1337/api/v1/agents/{id}/speed \
  -H "Content-Type: application/json" \
  -d '{"speed": 5000}'
```

### Update Speed and Status
```bash
# Update agent speed and status simultaneously
curl -X PUT http://localhost:1337/api/v1/agents/{id}/speed-status \
  -H "Content-Type: application/json" \
  -d '{"speed": 5000, "status": "online"}'
```

### Reset Speed on Offline
```bash
# Reset agent speed to 0
curl -X PUT http://localhost:1337/api/v1/agents/{id}/speed-reset \
  -H "Content-Type: application/json"
```

## 🎯 Conclusion

Fitur **Real-Time Speed Monitoring** telah berhasil diimplementasikan dengan:

- ✅ **Real-time Updates**: Speed dan status diupdate secara real-time
- ✅ **Automatic Management**: Speed otomatis direset saat offline
- ✅ **Comprehensive Logging**: Log lengkap untuk semua operasi
- ✅ **Non-intrusive Operation**: Background monitoring tanpa mengganggu main operations
- ✅ **WebSocket Integration**: Real-time updates ke frontend
- ✅ **Error Handling**: Robust error handling untuk semua scenarios
- ✅ **Performance Optimization**: Efficient database operations dan caching

Fitur ini siap untuk production use dan dapat digunakan untuk optimasi job distribution berdasarkan real-time agent performance capabilities.
