# Wordlist Distribution Fix

## Masalah yang Ditemukan

Berdasarkan analisis log dan kode, ditemukan bahwa **frontend sudah menunjukkan distribusi wordlist yang benar** (17%, 67%, 16% untuk 6 kata), tetapi **backend belum mengimplementasikan pembagian wordlist yang sebenarnya**. Semua agent masih menerima wordlist yang sama.

## Solusi yang Diimplementasi

### 1. **Implementasi Wordlist Distribution di Backend**

**File**: `internal/usecase/job_usecase.go`

**Perubahan**:
- Menambahkan logika untuk menghitung performance agent (sama seperti frontend)
- Membagi wordlist berdasarkan performance ratio
- Membuat separate wordlist untuk setiap agent

**Kode Baru**:
```go
// Calculate agent performance scores (similar to frontend)
type AgentPerformance struct {
    AgentID      uuid.UUID
    Name         string
    Capabilities string
    Speed        int
    Weight       float64
}

var agentPerformances []AgentPerformance
totalSpeed := 0

for _, agentID := range agentIDs {
    agent, err := u.agentRepo.GetByID(ctx, agentID)
    if err != nil {
        continue
    }
    
    speed := 1 // Default untuk CPU
    if strings.Contains(strings.ToLower(agent.Capabilities), "gpu") {
        speed = 5 // GPU lebih cepat
    } else if strings.Contains(strings.ToLower(agent.Capabilities), "rtx") {
        speed = 8 // RTX lebih cepat lagi
    } else if strings.Contains(strings.ToLower(agent.Capabilities), "gtx") {
        speed = 6 // GTX lebih cepat
    }
    
    totalSpeed += speed
    
    agentPerformances = append(agentPerformances, AgentPerformance{
        AgentID:      agent.ID,
        Name:         agent.Name,
        Capabilities: agent.Capabilities,
        Speed:        speed,
        Weight:       0, // Will be calculated below
    })
}

// Calculate weights
for i := range agentPerformances {
    agentPerformances[i].Weight = float64(agentPerformances[i].Speed) / float64(totalSpeed)
}
```

### 2. **Wordlist Content Distribution**

**Perubahan**:
- Parse wordlist content dari request
- Bagi wordlist berdasarkan performance ratio
- Buat separate wordlist untuk setiap agent

**Kode Baru**:
```go
// Get agent's wordlist segment
var agentWordlist string
if len(wordlistContent) > 0 {
    endIndex := currentIndex + wordCount
    if endIndex > len(wordlistContent) {
        endIndex = len(wordlistContent)
    }
    
    agentWords := wordlistContent[currentIndex:endIndex]
    agentWordlist = strings.Join(agentWords, "\n")
    currentIndex = endIndex
} else {
    // Fallback to original wordlist if we can't parse it
    agentWordlist = req.Wordlist
}

subJob := &domain.Job{
    // ... other fields
    Wordlist:       agentWordlist, // Use distributed wordlist
    TotalWords:     int64(wordCount),
    // ... other fields
}
```

### 3. **Frontend Enhancement untuk Wordlist Content**

**File**: `frontend/src/main.ts`

**Perubahan**:
- Mengambil wordlist content dari server
- Mengirim content ke backend untuk distribusi

**Kode Baru**:
```typescript
// Get wordlist content for distribution
let wordlistContent = ''
if (selectedWordlist && selectedWordlist.content) {
    wordlistContent = selectedWordlist.content
} else if (selectedWordlist && selectedWordlist.path) {
    // Try to fetch wordlist content from server
    try {
        const response = await fetch(`/api/v1/wordlists/${selectedWordlist.id}/content`)
        if (response.ok) {
            wordlistContent = await response.text()
        }
    } catch (error) {
        console.warn('Failed to fetch wordlist content:', error)
    }
}

const jobPayload = {
    // ... other fields
    wordlist: wordlistContent || wordlistName,  // Send content if available, otherwise name
    // ... other fields
}
```

### 4. **Endpoint untuk Wordlist Content**

**File**: `internal/delivery/http/handler/wordlist_handler.go`

**Perubahan**:
- Menambahkan endpoint `/api/v1/wordlists/:id/content`
- Mengembalikan konten wordlist sebagai plain text

**Kode Baru**:
```go
func (h *WordlistHandler) GetWordlistContent(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
        return
    }

    wordlist, err := h.wordlistUsecase.GetWordlist(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    // Read file content
    content, err := os.ReadFile(wordlist.Path)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read wordlist content"})
        return
    }

    // Return content as plain text
    c.Header("Content-Type", "text/plain")
    c.Data(http.StatusOK, "text/plain", content)
}
```

## Hasil yang Diharapkan

Setelah implementasi fix ini:

1. **Wordlist akan terbagi dengan benar**: Setiap agent akan menerima bagian wordlist yang sesuai dengan performance ratio
2. **Distribusi sesuai frontend**: Backend akan menggunakan logika yang sama dengan frontend untuk menghitung distribusi
3. **Log yang informatif**: Backend akan menampilkan log distribusi yang detail

## Cara Testing

1. **Restart server dan agent**:
   ```bash
   # Restart server
   sudo ./bin/server
   
   # Restart agent di masing-masing machine
   sudo ./bin/agent --server http://172.15.3.241:1337 --name tes-agent-cpu-01 --ip "172.15.1.94" --agent-key "c59a995d"
   sudo ./bin/agent --server http://172.15.3.241:1337 --name tes-agent-cpu-02 --ip "172.15.1.196" --agent-key "99230bc5"
   sudo ./bin/agent --server http://172.15.3.241:1337 --name tes-agent-gpu-01 --ip "30.30.30.39" --agent-key "56465c86"
   ```

2. **Buat job dengan multiple agent**:
   - Pilih multiple agent saat membuat job
   - Wordlist akan otomatis terbagi ke setiap agent

3. **Monitor hasil**:
   - Setiap agent akan menerima wordlist yang berbeda
   - Distribusi akan sesuai dengan yang ditampilkan di frontend

## Log yang Diharapkan

Setelah fix, log akan menunjukkan:

```
✅ Created job "doyo (Part 1 - tes-agent-cpu-02)" for agent tes-agent-cpu-02 with 1 words (16.7%)
✅ Created job "doyo (Part 2 - tes-agent-gpu-01)" for agent tes-agent-gpu-01 with 4 words (66.7%)
✅ Created job "doyo (Part 3 - tes-agent-cpu-01)" for agent tes-agent-cpu-01 with 1 words (16.7%)

// Agent akan menerima wordlist yang berbeda
Found assigned job: doyo (Part 1 - tes-agent-cpu-02)
Found assigned job: doyo (Part 2 - tes-agent-gpu-01)  
Found assigned job: doyo (Part 3 - tes-agent-cpu-01)

// Wordlist content yang berbeda untuk setiap agent
// tes-agent-cpu-02: ["admin"]
// tes-agent-gpu-01: ["tehbotolsosro", "admin.admin", "bambang1234", "Starbucks2025@@!!"]
// tes-agent-cpu-01: ["makanyuk"]
```

## Verifikasi Distribusi

Untuk memverifikasi bahwa wordlist sudah terbagi dengan benar, Anda dapat:

1. **Cek wordlist di setiap agent**:
   ```bash
   sudo cat /root/uploads/temp/wordlist-test.txt
   ```

2. **Bandingkan dengan distribusi frontend**:
   - Frontend: 17%, 67%, 16%
   - Backend: 1 kata, 4 kata, 1 kata (dari total 6 kata)

3. **Monitor job execution**:
   - Setiap agent akan menjalankan wordlist yang berbeda
   - Tidak ada duplikasi kata antar agent
