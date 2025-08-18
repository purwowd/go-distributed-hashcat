# Job Naming Fix

## Masalah yang Ditemukan

Berdasarkan feedback dari user, ditemukan bahwa:

1. **Job "Master" tidak diperlukan**: Job "tes cracking (Master)" yang muncul di dashboard tidak diperlukan dan membingungkan
2. **Format nama job terlalu panjang**: Format "tes cracking (Part 1 - tes-agent-cpu-01)" terlalu panjang dan tidak user-friendly

## Solusi yang Diimplementasi

### 1. **Menghilangkan Master Job**

**File**: `internal/usecase/job_usecase.go`

**Perubahan**: Menghapus pembuatan master job yang tidak diperlukan

**Kode yang Dihapus**:
```go
// Create master job record
masterJob := &domain.Job{
    ID:             uuid.New(),
    Name:           fmt.Sprintf("%s (Master)", req.Name),
    Status:         "distributed",
    // ... other fields
}

// Save master job
if err := u.jobRepo.Create(ctx, masterJob); err != nil {
    return nil, fmt.Errorf("failed to create master job: %w", err)
}
```

**Hasil**: Tidak ada lagi job "Master" yang muncul di dashboard

### 2. **Menyederhanakan Format Nama Job**

**Perubahan**: Mengubah format nama job dari "Part X - Agent" menjadi hanya "Agent"

**Kode Lama**:
```go
Name: fmt.Sprintf("%s (Part %d - %s)", req.Name, i+1, agentPerf.Name)
```

**Kode Baru**:
```go
Name: fmt.Sprintf("%s (%s)", req.Name, agentPerf.Name)
```

## Hasil yang Diharapkan

Setelah implementasi fix ini:

### **Sebelum Fix**:
```
tes cracking (Master) - Status: distributed
tes cracking (Part 1 - tes-agent-cpu-02) - Status: failed
tes cracking (Part 2 - tes-agent-gpu-01) - Status: completed
tes cracking (Part 3 - tes-agent-cpu-01) - Status: failed
```

### **Setelah Fix**:
```
tes cracking (tes-agent-cpu-02) - Status: failed
tes cracking (tes-agent-gpu-01) - Status: completed
tes cracking (tes-agent-cpu-01) - Status: failed
```

## Keuntungan Perubahan

1. **Dashboard lebih bersih**: Tidak ada job "Master" yang membingungkan
2. **Nama job lebih sederhana**: Format "tes cracking (tes-agent-cpu-01)" lebih mudah dibaca
3. **User experience lebih baik**: User dapat langsung melihat job untuk agent mana
4. **Kurangi kebingungan**: Tidak ada lagi job "Master" yang statusnya "distributed" dan "Pending..."

## Cara Testing

1. **Restart server**:
   ```bash
   sudo ./bin/server
   ```

2. **Buat job multi-agent**:
   - Pilih multiple agent saat membuat job
   - Job akan dibuat dengan format nama yang baru

3. **Monitor hasil**:
   - Tidak ada lagi job "Master"
   - Job akan memiliki format nama yang sederhana

## Log yang Diharapkan

Setelah fix, log akan menunjukkan:

```
✅ Created job "tes cracking (tes-agent-cpu-02)" for agent tes-agent-cpu-02 with 1 words (16.7%)
✅ Created job "tes cracking (tes-agent-gpu-01)" for agent tes-agent-gpu-01 with 4 words (66.7%)
✅ Created job "tes cracking (tes-agent-cpu-01)" for agent tes-agent-cpu-01 with 1 words (16.7%)
```

## Dashboard yang Diharapkan

Setelah fix, dashboard akan menampilkan:

| Job Details | Status | Progress | Resources | Result | Actions |
|-------------|--------|----------|-----------|--------|---------|
| tes cracking (tes-agent-cpu-02) | failed | 0% | admin, tes-agent-cpu-02 | Hashcat execution failed: exit status 255 | ... |
| tes cracking (tes-agent-gpu-01) | completed | 100% | tehbotolsosro admin.admin bambang1234 Starbucks2025@@!!, tes-agent-gpu-01 | Found: Starbucks2025@@!! | ... |
| tes cracking (tes-agent-cpu-01) | failed | 0% | makanyuk, tes-agent-cpu-01 | Hashcat execution failed: exit status 255 | ... |

## Technical Details

### Master Job Removal
- **Sebelum**: Master job dibuat untuk monitoring overall progress
- **Sesudah**: Tidak ada master job, hanya sub-jobs langsung
- **Alasan**: Master job tidak memberikan nilai tambah dan membingungkan user

### Job Naming Simplification
- **Format Lama**: `{job_name} (Part {number} - {agent_name})`
- **Format Baru**: `{job_name} ({agent_name})`
- **Contoh**: "tes cracking (tes-agent-cpu-01)"

### Backward Compatibility
- **Existing Jobs**: Tidak terpengaruh, tetap menggunakan format lama
- **New Jobs**: Menggunakan format baru
- **Database**: Tidak ada perubahan struktur database
