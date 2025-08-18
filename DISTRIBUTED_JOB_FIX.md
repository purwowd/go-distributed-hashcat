# Distributed Job System Fix

## Masalah yang Ditemukan

Berdasarkan analisis log dan kode, ditemukan dua masalah utama dalam sistem distributed job:

### 1. Job Tidak Terbagi ke Multiple Agent
**Masalah**: Job yang dibuat dengan multiple agent assignment hanya diassign ke agent pertama saja.

**Penyebab**: Di `internal/usecase/job_usecase.go` line 103-104, ada TODO comment yang menunjukkan bahwa distributed job creation belum diimplementasi dengan benar:

```go
// For now, assign to first agent (legacy compatibility)
// TODO: Implement proper distributed job creation
job.AgentID = &agentIDs[0]
```

### 2. Job Dijalankan Berulang Kali
**Masalah**: Agent terus mengambil job yang sama setelah selesai, menyebabkan job dijalankan berulang kali.

**Penyebab**: Agent menggunakan endpoint `/api/v1/jobs?status=pending` yang mengembalikan semua job pending, bukan hanya job yang diassign ke agent tersebut.

## Solusi yang Diimplementasi

### 1. Implementasi Distributed Job Creation yang Benar

**File**: `internal/usecase/job_usecase.go`

**Perubahan**:
- Mengganti logika assignment yang hanya mengassign ke agent pertama
- Membuat separate job untuk setiap agent ketika multiple agent dipilih
- Membuat master job untuk monitoring overall progress
- Membuat sub-jobs untuk setiap agent dengan nama yang berbeda

**Kode Baru**:
```go
// Create separate job for each agent (distributed job creation)
if len(agentIDs) > 1 {
    // Create master job record
    masterJob := &domain.Job{
        ID:             uuid.New(),
        Name:           fmt.Sprintf("%s (Master)", req.Name),
        Status:         "distributed",
        // ... other fields
    }

    // Create sub-jobs for each agent
    var subJobs []*domain.Job
    for i, agentID := range agentIDs {
        subJob := &domain.Job{
            ID:             uuid.New(),
            Name:           fmt.Sprintf("%s (Part %d - %s)", req.Name, i+1, agentName),
            Status:         "pending",
            AgentID:        &agentID,
            // ... other fields
        }
        // Save sub-job
        if err := u.jobRepo.Create(ctx, subJob); err != nil {
            return nil, fmt.Errorf("failed to create sub-job %d: %w", i, err)
        }
        subJobs = append(subJobs, subJob)
    }

    // Return the first sub-job as the primary result
    return subJobs[0], nil
}
```

### 2. Perbaikan Job Assignment untuk Agent

**File**: `cmd/agent/main.go`

**Perubahan**:
- Mengubah cara agent mengambil job dari endpoint umum ke endpoint spesifik
- Menggunakan `/api/v1/jobs/agent/{agentID}` untuk mendapatkan job yang diassign ke agent tersebut

**Kode Baru**:
```go
func (a *Agent) checkForNewJob() error {
    // Use the specific endpoint for getting available job for this agent
    url := fmt.Sprintf("%s/api/v1/jobs/agent/%s", a.ServerURL, a.ID.String())
    resp, err := a.Client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var response struct {
        Data *domain.Job `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return err
    }

    // Check if we got a job
    if response.Data != nil {
        log.Printf("Found assigned job: %s", response.Data.Name)
        a.CurrentJob = response.Data
        go a.executeJob(response.Data)
    }

    return nil
}
```

### 3. Penambahan Endpoint untuk Agent Job

**File**: `internal/delivery/http/router.go`

**Perubahan**:
- Menambahkan endpoint `/api/v1/jobs/agent/:id` untuk mendapatkan job yang tersedia untuk agent tertentu

**Kode Baru**:
```go
jobs.GET("/agent/:id", jobHandler.GetAvailableJobForAgent)
```

## Hasil yang Diharapkan

Setelah implementasi fix ini:

1. **Job akan terbagi ke multiple agent**: Ketika job dibuat dengan multiple agent, sistem akan membuat separate job untuk setiap agent dengan nama yang berbeda (Part 1, Part 2, dst).

2. **Job tidak akan dijalankan berulang kali**: Agent hanya akan mengambil job yang diassign kepadanya, dan job yang sudah selesai tidak akan diambil lagi.

3. **Monitoring yang lebih baik**: Master job akan tersedia untuk monitoring overall progress dari semua sub-jobs.

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
   - Job akan otomatis terbagi ke setiap agent

3. **Monitor hasil**:
   - Setiap agent akan menjalankan job terpisah
   - Job tidak akan dijalankan berulang kali
   - Master job akan tersedia untuk monitoring

## Log yang Diharapkan

Setelah fix, log akan menunjukkan:

```
✅ Created job "doyo (Part 1 - tes-agent-cpu-01)" for agent tes-agent-cpu-01
✅ Created job "doyo (Part 2 - tes-agent-cpu-02)" for agent tes-agent-cpu-02  
✅ Created job "doyo (Part 3 - tes-agent-gpu-01)" for agent tes-agent-gpu-01

// Agent akan mengambil job yang diassign kepadanya
Found assigned job: doyo (Part 1 - tes-agent-cpu-01)
Found assigned job: doyo (Part 2 - tes-agent-cpu-02)
Found assigned job: doyo (Part 3 - tes-agent-gpu-01)

// Job tidak akan dijalankan berulang kali
Job completed: doyo (Part 1 - tes-agent-cpu-01)
Job completed: doyo (Part 2 - tes-agent-cpu-02)
Job completed: doyo (Part 3 - tes-agent-gpu-01)
```
