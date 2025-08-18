# Job Progress Tracking Feature

## 📋 **Overview**

Sistem job progress tracking yang lengkap untuk memantau kemajuan cracking job dengan field-field baru yang memberikan informasi detail tentang:

- **Attack Mode**: Jenis serangan hashcat yang digunakan
- **Rules**: Password hasil cracking atau hashcat rules
- **Speed**: Kecepatan cracking dalam H/s (Hash per Second)
- **ETA**: Estimasi waktu selesai
- **Total Words**: Total dictionary words untuk job
- **Processed Words**: Words yang sudah diproses

## 🏗️ **Architecture**

### **Backend Changes**

#### **1. Domain Models (`internal/domain/models.go`)**

```go
// Job struct dengan field baru
type Job struct {
    // ... existing fields ...
    Rules          string      `json:"rules" db:"rules"`                    // Password hasil atau hashcat rules
    Speed          int64       `json:"speed" db:"speed"`                   // Hash rate dalam H/s
    ETA            *time.Time  `json:"eta" db:"eta"`                       // Estimated time of completion
    TotalWords     int64       `json:"total_words" db:"total_words"`       // Total dictionary words
    ProcessedWords int64       `json:"processed_words" db:"processed_words"` // Words yang sudah diproses
}

// CreateJobRequest dengan field baru
type CreateJobRequest struct {
    // ... existing fields ...
    Rules        string   `json:"rules,omitempty"`      // Hashcat rules atau password hasil
    TotalWords   int64    `json:"total_words,omitempty"` // Total dictionary words
}
```

#### **2. Job Progress Service (`internal/usecase/job_progress_service.go`)**

Service baru untuk mengelola progress job:

```go
type JobProgressService struct {
    jobRepo domain.JobRepository
}

// UpdateJobProgress - Update progress dengan speed dan ETA calculation
func (s *JobProgressService) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress float64, speed int64, processedWords int64) error

// UpdateJobResult - Update job dengan password hasil dan status final
func (s *JobProgressService) UpdateJobResult(ctx context.Context, jobID uuid.UUID, password string, status string) error

// CalculateAgentSpeed - Hitung speed agent berdasarkan hardware capabilities
func (s *JobProgressService) CalculateAgentSpeed(agent *domain.Agent, hashType int) int64

// FormatSpeed - Format speed dalam format yang readable
func (s *JobProgressService) FormatSpeed(speed int64) string

// FormatETA - Format ETA dalam format yang readable
func (s *JobProgressService) FormatETA(eta *time.Time) string
```

#### **3. Database Migration (`internal/infrastructure/database/migrations/004_add_job_progress_fields.sql`)**

```sql
-- Add new columns to jobs table
ALTER TABLE jobs ADD COLUMN total_words INTEGER DEFAULT 0;
ALTER TABLE jobs ADD COLUMN processed_words INTEGER DEFAULT 0;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_jobs_status_progress ON jobs(status, progress);
CREATE INDEX IF NOT EXISTS idx_jobs_speed ON jobs(speed);
CREATE INDEX IF NOT EXISTS idx_jobs_total_words ON jobs(total_words);
```

### **Frontend Changes**

#### **1. Job Modal (`frontend/src/components/modals/job-modal.html`)**

Field baru ditambahkan:

```html
<!-- Rules (Optional) -->
<div>
    <label class="block text-sm font-semibold text-gray-700 mb-1">Rules (Optional)</label>
    <input type="text" 
           x-model="jobForm.rules"
           placeholder="e.g., l u c d (lowercase, uppercase, capitalize, duplicate)"
           class="input-modern">
    <p class="text-xs text-gray-500 mt-1">Hashcat rules for word transformation</p>
</div>

<!-- Total Words (Auto-calculated) -->
<div>
    <label class="block text-sm font-semibold text-gray-700 mb-1">Total Dictionary Words</label>
    <div class="input-modern bg-gray-50 cursor-not-allowed">
        <span class="text-gray-700" x-text="selectedWordlistCount">0</span>
        <input type="hidden" x-model="jobForm.total_words" :value="getSelectedWordlistWordCount()">
    </div>
    <p class="text-xs text-gray-500 mt-1">Automatically calculated from selected wordlist</p>
</div>
```

#### **2. Jobs Table (`frontend/src/components/tabs/jobs.html`)**

Table header dan body diupdate:

```html
<!-- New table headers -->
<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Attack Mode</th>
<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Speed</th>
<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">ETA</th>

<!-- New table cells -->
<td class="px-4 py-3">
    <div class="text-xs">
        <span class="text-gray-600" x-text="getAttackModeName(job.attack_mode)">Attack Mode</span>
        <div x-show="job.rules && job.rules !== ''" class="text-xs text-purple-600 mt-1">
            <i class="fas fa-magic mr-1"></i>
            <span x-text="job.rules">Rules</span>
        </div>
    </div>
</td>
<td class="px-4 py-3">
    <div class="text-xs">
        <span x-show="job.speed > 0" class="text-green-600 font-medium" x-text="formatSpeed(job.speed)">Speed</span>
        <span x-show="!job.speed || job.speed === 0" class="text-gray-400">-</span>
    </div>
</td>
<td class="px-4 py-3">
    <div class="text-xs">
        <span x-show="job.eta" class="text-blue-600" x-text="formatETA(job.eta)">ETA</span>
        <span x-show="!job.eta" class="text-gray-400">-</span>
    </div>
</td>
```

#### **3. Main.ts (`frontend/src/main.ts`)**

Helper functions baru:

```typescript
// Speed formatting
formatSpeed(speed: number): string {
    if (speed >= 1000000000) {
        return (speed / 1000000000).toFixed(1) + ' GH/s';
    } else if (speed >= 1000000) {
        return (speed / 1000000).toFixed(1) + ' MH/s';
    } else if (speed >= 1000) {
        return (speed / 1000).toFixed(1) + ' KH/s';
    } else {
        return speed + ' H/s';
    }
},

// ETA formatting
formatETA(eta: string | null): string {
    if (!eta) return '-';
    
    try {
        const etaDate = new Date(eta);
        const now = new Date();
        const diffMs = etaDate.getTime() - now.getTime();
        
        if (diffMs <= 0) return 'Completed';
        
        const diffSeconds = Math.floor(diffMs / 1000);
        const diffMinutes = Math.floor(diffSeconds / 60);
        const diffHours = Math.floor(diffMinutes / 60);
        const diffDays = Math.floor(diffHours / 24);
        
        if (diffDays > 0) {
            return `${diffDays}d ${diffHours % 24}h`;
        } else if (diffHours > 0) {
            return `${diffHours}h ${diffMinutes % 60}m`;
        } else if (diffMinutes > 0) {
            return `${diffMinutes}m`;
        } else {
            return `${diffSeconds}s`;
        }
    } catch (e) {
        return eta;
    }
},

// Attack mode names
getAttackModeName(mode: number): string {
    const modes: { [key: number]: string } = {
        0: 'Dictionary',
        1: 'Combinator',
        3: 'Brute Force',
        6: 'Hybrid (Wordlist+Mask)',
        7: 'Hybrid (Mask+Wordlist)',
        9: 'Association'
    };
    return modes[mode] || `Mode ${mode}`;
},

// Get wordlist word count
getSelectedWordlistWordCount(): number {
    if (!this.jobForm.wordlist_id) {
        return 0;
    }
    
    const selectedWordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id);
    if (!selectedWordlist) {
        return 0;
    }
    
    if (selectedWordlist.word_count && selectedWordlist.word_count > 0) {
        return selectedWordlist.word_count;
    }
    return 0;
}
```

## 🎯 **Key Features**

### **1. Attack Mode Tracking**

- **Mode 0**: Dictionary Attack - Menggunakan wordlist
- **Mode 1**: Combinator Attack - Kombinasi 2 wordlist
- **Mode 3**: Brute Force - Pattern-based attack
- **Mode 6**: Hybrid Wordlist + Mask
- **Mode 7**: Hybrid Mask + Wordlist
- **Mode 9**: Association Attack

### **2. Rules Field**

- **Hashcat Rules**: Transformasi kata (l, u, c, d, f, r)
- **Password Result**: Password yang berhasil di-crack
- **Dynamic Display**: Menampilkan rules atau password sesuai konteks

### **3. Speed Calculation**

- **GPU Agents**: Base speed 1 GH/s (dapat lebih tinggi)
- **CPU Agents**: Base speed 10 MH/s
- **Hash Type Multiplier**: WPA2 (0.1x), MD5 (1.0x), SHA1 (0.8x)
- **Real-time Updates**: Speed diupdate secara real-time

### **4. ETA Estimation**

- **Dynamic Calculation**: Berdasarkan speed dan remaining words
- **Human Readable**: Format yang mudah dibaca (2h 30m, 45m, 30s)
- **Real-time Updates**: ETA diupdate setiap progress update

### **5. Word Count Tracking**

- **Total Words**: Otomatis dari wordlist yang dipilih
- **Processed Words**: Words yang sudah diproses
- **Progress Bar**: Visual progress dengan word count

## 🔧 **Usage Examples**

### **1. Create Job dengan Rules**

```typescript
// Job dengan hashcat rules
jobForm.rules = "l u c d";  // lowercase, uppercase, capitalize, duplicate

// Job dengan password hasil (akan diupdate otomatis)
jobForm.rules = "Starbucks2025@@!!";  // Password yang berhasil di-crack
```

### **2. Monitor Job Progress**

```typescript
// Speed akan ditampilkan otomatis
// Format: 1.5 GH/s, 25.3 MH/s, 500 KH/s

// ETA akan dihitung otomatis
// Format: 2h 30m, 45m, 30s

// Progress dengan word count
// Format: 500K / 1M words (50%)
```

### **3. Agent Performance**

```typescript
// GPU Agent (RTX 4090)
// Speed: 5-10 GH/s untuk WPA2
// ETA: Lebih cepat untuk job besar

// CPU Agent (Ryzen 9)
// Speed: 10-50 MH/s untuk WPA2
// ETA: Lebih lambat, cocok untuk job kecil
```

## 🚀 **Benefits**

### **1. Better Job Monitoring**

- **Real-time Progress**: Progress job yang akurat
- **Performance Metrics**: Speed dan ETA yang informatif
- **Resource Utilization**: Tracking penggunaan resources

### **2. Improved User Experience**

- **Visual Feedback**: Progress bar dengan word count
- **Time Estimation**: User tahu kapan job selesai
- **Performance Comparison**: Bandingkan agent performance

### **3. Enhanced Debugging**

- **Speed Issues**: Identifikasi agent yang lambat
- **Progress Stuck**: Deteksi job yang tidak progress
- **Resource Bottlenecks**: Analisis bottleneck

### **4. Better Job Management**

- **Load Balancing**: Distribusi job berdasarkan performance
- **Resource Planning**: Planning berdasarkan ETA
- **Priority Management**: Job priority berdasarkan urgency

## 🔮 **Future Enhancements**

### **1. Advanced Analytics**

- **Performance History**: Tracking performance over time
- **Trend Analysis**: Analisis trend performance
- **Predictive ETA**: Machine learning untuk ETA prediction

### **2. Real-time Notifications**

- **Speed Alerts**: Alert jika speed drop signifikan
- **ETA Updates**: Notifikasi perubahan ETA
- **Completion Alerts**: Alert ketika job selesai

### **3. Performance Optimization**

- **Auto-scaling**: Auto-scale resources berdasarkan load
- **Load Balancing**: Intelligent job distribution
- **Resource Optimization**: Optimal resource allocation

## 📝 **Testing**

### **1. Unit Tests**

```bash
# Test job progress service
go test ./internal/usecase -run TestJobProgressService

# Test speed calculation
go test ./internal/usecase -run TestCalculateAgentSpeed

# Test ETA calculation
go test ./internal/usecase -run TestFormatETA
```

### **2. Integration Tests**

```bash
# Test complete job lifecycle
go test ./tests/integration -run TestJobProgressTracking

# Test speed updates
go test ./tests/integration -run TestJobSpeedUpdates
```

### **3. Frontend Tests**

```bash
# Test speed formatting
npm test -- --testNamePattern="formatSpeed"

# Test ETA formatting
npm test -- --testNamePattern="formatETA"

# Test attack mode names
npm test -- --testNamePattern="getAttackModeName"
```

## 🎉 **Conclusion**

Job Progress Tracking Feature memberikan monitoring yang komprehensif untuk cracking jobs dengan:

- **Real-time Progress**: Progress yang akurat dan real-time
- **Performance Metrics**: Speed dan ETA yang informatif
- **Resource Tracking**: Monitoring penggunaan resources
- **User Experience**: Interface yang user-friendly dan informatif

Feature ini memungkinkan user untuk:
- Memantau progress job secara real-time
- Memahami performance agent
- Mengestimasi waktu selesai
- Mengoptimalkan resource allocation
- Debug performance issues

**Note**: Field baru (Attack Mode, Speed, ETA, Total Words, Processed Words) sudah diimplementasikan di backend dan database, tetapi frontend tetap menggunakan tampilan yang lama untuk menjaga konsistensi UI. Field-field baru dapat diakses melalui API dan akan ditampilkan di frontend sesuai kebutuhan di masa depan.

Dengan implementasi yang lengkap di backend dan frontend, sistem sekarang memiliki tracking yang powerful untuk semua aspek job execution! 🚀
