# Multi-Agent Job Failure Fix

## Masalah yang Ditemukan

Berdasarkan analisis log dan gambar yang Anda tunjukkan, ditemukan bahwa:

- **Job single agent ("d")**: ✅ **BERHASIL** - menemukan password "Starbucks2025@!!"
- **Job multi-agent ("craking keke")**: ❌ **GAGAL SEMUA** - error "Hashcat execution failed: exit status 255"

## Root Cause Analysis

### Masalah Utama: Wordlist Content Handling

**Masalah**: Ketika job multi-agent dibuat, backend mengirim wordlist content sebagai string (bukan file), tetapi agent mencoba menggunakan string tersebut sebagai path file.

**Alur Masalah**:
1. Frontend mengirim wordlist content ke backend
2. Backend membuat job dengan wordlist content sebagai string
3. Agent menerima job dengan wordlist = "admin\ntehbotolsosro\nadmin.admin\n..."
4. Agent mencoba menggunakan string tersebut sebagai path file
5. Hashcat gagal karena file tidak ditemukan
6. Exit status 255 = "Invalid arguments"

### Perbedaan Single vs Multi-Agent

**Single Agent (Berhasil)**:
```
Job.Wordlist = "wordlist-test.txt" (filename)
Agent: Mencari file "wordlist-test.txt" → Ditemukan → Hashcat berhasil
```

**Multi-Agent (Gagal)**:
```
Job.Wordlist = "admin\ntehbotolsosro\nadmin.admin\n..." (content)
Agent: Mencari file "admin\ntehbotolsosro\nadmin.admin\n..." → Tidak ditemukan → Hashcat gagal
```

## Solusi yang Diimplementasi

### 1. **Enhanced Wordlist Content Detection**

**File**: `cmd/agent/main.go`

**Perubahan**: Menambahkan deteksi apakah wordlist adalah content atau path file

**Kode Baru**:
```go
// Check if wordlist contains newlines (indicating it's content, not a path)
if strings.Contains(job.Wordlist, "\n") {
    // This is wordlist content, create a temporary file
    tempDir := filepath.Join(a.UploadDir, "temp")
    if err := os.MkdirAll(tempDir, 0755); err != nil {
        return fmt.Errorf("failed to create temp directory: %w", err)
    }
    
    wordlistFile := filepath.Join(tempDir, fmt.Sprintf("wordlist-%s.txt", job.ID.String()))
    if err := os.WriteFile(wordlistFile, []byte(job.Wordlist), 0644); err != nil {
        return fmt.Errorf("failed to create wordlist file: %w", err)
    }
    
    localWordlist = wordlistFile
    log.Printf("📝 Created wordlist file from content: %s", localWordlist)
    log.Printf("📋 Wordlist content preview: %s", strings.Split(job.Wordlist, "\n")[0])
} else {
    // Fallback to wordlist filename resolution (existing logic)
    // ...
}
```

### 2. **Logging Enhancement**

**Perubahan**: Menambahkan logging yang lebih detail untuk debugging

**Log Baru**:
```
📝 Created wordlist file from content: /root/uploads/temp/wordlist-12345678-1234-1234-1234-123456789abc.txt
📋 Wordlist content preview: admin
```

## Hasil yang Diharapkan

Setelah implementasi fix ini:

1. **Multi-agent jobs akan berhasil**: Agent akan membuat file temporary dari wordlist content
2. **Wordlist distribution akan bekerja**: Setiap agent akan menerima wordlist yang berbeda
3. **Log yang informatif**: Agent akan menampilkan log yang jelas tentang wordlist handling

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

2. **Buat job multi-agent**:
   - Pilih multiple agent saat membuat job
   - Job akan otomatis terbagi ke setiap agent

3. **Monitor hasil**:
   - Setiap agent akan berhasil menjalankan job
   - Tidak ada lagi error "Hashcat execution failed: exit status 255"

## Log yang Diharapkan

Setelah fix, log akan menunjukkan:

```
📝 Created wordlist file from content: /root/uploads/temp/wordlist-12345678-1234-1234-1234-123456789abc.txt
📋 Wordlist content preview: admin
🔨 Running hashcat with args: [-m 2500 -a 0 /root/uploads/temp/Starbucks_20250526_140536.hccapx /root/uploads/temp/wordlist-12345678-1234-1234-1234-123456789abc.txt -w 4 --status --status-timer=2 --potfile-disable --outfile /root/uploads/temp/cracked-12345678-1234-1234-1234-123456789abc.txt --outfile-format 2]
✅ Job completion sent successfully to server
```

## Verifikasi Fix

Untuk memverifikasi bahwa fix sudah bekerja:

1. **Cek log agent**: Tidak ada lagi error "Hashcat execution failed: exit status 255"
2. **Cek job status**: Multi-agent jobs akan berhasil (completed/failed dengan alasan yang valid)
3. **Cek wordlist files**: Agent akan membuat file temporary untuk wordlist content

## Technical Details

### Hashcat Exit Status 255
- **Exit 0**: Password found
- **Exit 1**: Password not found (exhausted)
- **Exit 255**: Invalid arguments (file not found, invalid mode, etc.)

### Wordlist Content Detection
- **Contains newlines**: Content (create temporary file)
- **No newlines**: Path/UUID (use existing logic)

### Temporary File Management
- **Location**: `/root/uploads/temp/wordlist-{job-id}.txt`
- **Permissions**: 0644 (readable by hashcat)
- **Cleanup**: Automatic cleanup after job completion
