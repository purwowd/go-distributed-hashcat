# Stable Card Position Solution

## 🎯 **Masalah yang Diatasi**

Sebelumnya, ketika user melakukan update pada agent (misal: ganti IP address, update status, dll), posisi card agent bisa berubah dan berpindah tempat. Ini menyebabkan:

- ❌ User bingung karena card "hilang" dari posisi semula
- ❌ Layout berubah-ubah setiap kali ada update
- ❌ User experience yang buruk dan tidak konsisten

## ✅ **Solusi yang Diimplementasikan**

### **1. Backend Stable Sorting**

#### **Repository Level (SQL)**
```sql
-- Sebelumnya:
ORDER BY created_at DESC

-- Sekarang:
ORDER BY created_at DESC, id ASC
```

**Keuntungan**:
- `created_at DESC` memastikan agent baru muncul di atas (newest first)
- `id ASC` memberikan urutan yang stabil untuk agent dengan `created_at` yang sama

#### **Implementation di `agent_repository.go`**
```go
r.getAllStmt, err = r.db.DB().Prepare(`
    SELECT id, name, ip_address, port, status, capabilities, agent_key, last_seen, created_at, updated_at
    FROM agents ORDER BY created_at DESC, id ASC
`)
```

### **2. Frontend Stable Sorting**

#### **Alpine.js Implementation**
```typescript
// Subscribe to agent store changes
agentStore.subscribe(() => {
    const state = agentStore.getState()
    
    // ✅ Implement stable sorting to maintain card positions
    const agents = state.agents || []
    // Sort by created_at DESC, then by ID ASC for stable ordering
    const stableSortedAgents = agents.sort((a, b) => {
        // First sort by created_at DESC (newest first)
        const dateComparison = new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        if (dateComparison !== 0) {
            return dateComparison
        }
        // If dates are equal, sort by ID ASC for stable ordering
        return a.id.localeCompare(b.id)
    })
    
    // ✅ Force Alpine.js reactivity by creating new array reference
    this.reactiveAgents = [...stableSortedAgents]
})
```

### **3. Agent Status Management**

#### **New Endpoints**
- `PUT /api/v1/agents/:id/status` - Update agent status
- `PUT /api/v1/agents/:id/heartbeat` - Update agent heartbeat

#### **Status Values**
- `online` - Agent aktif dan siap menerima job
- `offline` - Agent tidak aktif (default)
- `busy` - Agent sedang menjalankan job
- `error` - Agent mengalami error

#### **Implementation di Handler**
```go
func (h *AgentHandler) UpdateAgentStatus(c *gin.Context) {
    // Validate status
    validStatuses := []string{"online", "offline", "busy", "error"}
    
    // Update agent status
    if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), id, req.Status); err != nil {
        // Handle error
    }
    
    // Update last seen
    if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), id); err != nil {
        // Log error but don't fail the request
    }
}
```

## 🔧 **Cara Kerja Stable Position**

### **Scenario 1: Create New Agent**
1. User buat agent baru
2. Agent baru mendapat `created_at` terbaru
3. Agent baru muncul di posisi atas (newest first)
4. Agent lama tetap di posisi yang sama

### **Scenario 2: Update Existing Agent**
1. User update agent (misal: ganti IP address, update status)
2. `created_at` tidak berubah (hanya `updated_at` yang berubah)
3. Agent tetap di posisi yang sama
4. Card tidak berpindah tempat

### **Scenario 3: Multiple Agents dengan Timestamp Sama**
1. Jika ada agent dengan `created_at` yang sama
2. Urutan ditentukan oleh `id ASC` (UUID)
3. Urutan tetap konsisten dan stabil

## 🎨 **User Experience Improvements**

### **Sebelumnya (Unstable)**:
- ❌ Agent card bisa berpindah posisi setelah update
- ❌ User bingung karena card "hilang" dari posisi semula
- ❌ Layout berubah-ubah setiap kali ada update

### **Sekarang (Stable)**:
- ✅ Agent card tetap di posisi yang sama setelah update
- ✅ Layout konsisten dan predictable
- ✅ User bisa dengan mudah menemukan agent yang dicari
- ✅ Agent baru tetap muncul di atas (newest first)

## 🧪 **Testing**

### **Test Scripts Created**:
1. `test_stable_card_position.sh` - Test backend stable sorting
2. `test_frontend_stable_sorting.sh` - Test frontend implementation

### **Test Scenarios**:
1. **Initial Position**: Posisi awal agent
2. **After Status Update**: Posisi setelah update status (harus sama)
3. **After Create**: Posisi setelah create agent baru
4. **Position Stability**: Memastikan posisi relatif tetap

## 📁 **Files Modified**

### **Backend**:
- `internal/infrastructure/repository/agent_repository.go` - Stable SQL sorting
- `internal/usecase/agent_usecase.go` - Added UpdateAgentLastSeen method
- `internal/delivery/http/handler/agent_handler.go` - New status endpoints
- `internal/delivery/http/router.go` - New routes

### **Frontend**:
- `frontend/src/main.ts` - Stable sorting implementation

### **Documentation**:
- `docs/stable-card-position-solution.md` - This file

## 🚀 **How to Use**

### **1. Update Agent Status**
```bash
curl -X PUT http://localhost:1337/api/v1/agents/{agent_id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "online"}'
```

### **2. Update Agent Heartbeat**
```bash
curl -X PUT http://localhost:1337/api/v1/agents/{agent_id}/heartbeat
```

### **3. Frontend Behavior**
- Agent cards maintain their positions after updates
- New agents appear at the top
- Status changes don't affect card positions
- Layout remains consistent

## 🎯 **Expected Results**

**Sekarang ketika Anda add new agent atau update existing agent, posisi card agent akan tetap stabil dan tidak berpindah tempat!**

- **Agent baru** → Muncul di atas (newest first)
- **Agent yang di-update** → Tetap di posisi yang sama
- **Layout konsisten** → User experience yang lebih baik
- **Status management** → Agent status bisa di-update tanpa mengubah posisi

## 🔍 **Troubleshooting**

### **Jika Card Masih Berpindah**:
1. Check browser console untuk error JavaScript
2. Verify backend sorting query (`ORDER BY created_at DESC, id ASC`)
3. Check frontend stable sorting implementation
4. Ensure cache invalidation works correctly

### **Jika Status Tidak Berubah**:
1. Check API endpoint response
2. Verify agent ID is correct
3. Check backend logs for errors
4. Ensure UpdateAgentStatus method exists in usecase

## 📚 **References**

- [Alpine.js Documentation](https://alpinejs.dev/)
- [Gin Framework](https://gin-gonic.com/)
- [SQLite ORDER BY](https://www.sqlite.org/lang_select.html#orderby)
- [Go UUID Package](https://pkg.go.dev/github.com/google/uuid)
