# Dynamic Hash Type Feature

## Overview
Fitur ini membuat kolom Hash Type menjadi dinamis berdasarkan file yang dipilih, khususnya untuk file WiFi handshake (.hccapx) yang akan otomatis menggunakan hash type 2500 (WPA/WPA2).

## Perubahan yang Dibuat

### 1. HTML Changes (job-modal.html)
- Mengubah Hash Type dari statis menjadi dinamis menggunakan Alpine.js computed properties
- Dari hardcoded values menjadi reactive berdasarkan file yang dipilih

### 2. JavaScript Changes (main.ts)
- Menambahkan 3 computed properties baru untuk hash type:
  - `selectedHashTypeValue`: Nilai hash type (contoh: "2500")
  - `selectedHashTypeName`: Nama hash type (contoh: "WPA/WPA2")
  - `selectedHashTypeDescription`: Deskripsi detail hash type

## Mapping File Type ke Hash Type

### WiFi Handshake Files (Hash Type: 2500)
| File Extension | File Type | Hash Type | Description |
|----------------|-----------|-----------|-------------|
| `.hccapx` | `hccapx` | 2500 | Modern WiFi handshake format |
| `.hccap` | `hccap` | 2500 | Legacy WiFi handshake format |
| `.cap` | `cap` | 2500 | Raw 802.11 packet captures |
| `.pcap` | `pcap` | 2500 | Packet capture files |

### Generic Hash Files (Hash Type: 0)
| File Extension | File Type | Hash Type | Description |
|----------------|-----------|-----------|-------------|
| `.hash` | `hash` | 0 | Generic hash format |
| `.txt` | `txt` | 0 | Text-based hash files |
| Other | `hash` | 0 | Default for unknown types |

## Cara Kerja

### 1. File Selection
```typescript
// Ketika user memilih file
jobForm.hash_file_id = "selected-file-id"
```

### 2. Automatic Detection
```typescript
// Sistem otomatis mendeteksi file type
const selectedFile = this.hashFiles.find(f => f.id === this.jobForm.hash_file_id)
const fileType = selectedFile.type?.toLowerCase()
```

### 3. Hash Type Determination
```typescript
// Berdasarkan file type, hash type ditentukan
switch (fileType) {
    case 'hccapx':
    case 'hccap':
    case 'cap':
    case 'pcap':
        return '2500' // WPA/WPA2
    case 'hash':
    default:
        return '0' // Generic hash
}
```

### 4. UI Update
- **Hash Type Name**: Berubah dari "WPA/WPA2" menjadi sesuai file
- **Hash Type Value**: Berubah dari "2500" menjadi sesuai file
- **Description**: Berubah menjadi deskripsi yang spesifik

## Contoh Implementasi

### File .hccapx (WiFi Handshake)
```typescript
// Input
jobForm.hash_file_id = "wifi-handshake-123"

// Output
selectedHashTypeValue = "2500"
selectedHashTypeName = "WPA/WPA2"
selectedHashTypeDescription = "WiFi handshake file (.hccapx) - WPA/WPA2 cracking"
```

### File Generic Hash
```typescript
// Input
jobForm.hash_file_id = "generic-hash-456"

// Output
selectedHashTypeValue = "0"
selectedHashTypeName = "Generic Hash"
selectedHashTypeDescription = "Generic hash file - various hash types supported"
```

## Keuntungan

1. **Automatic Detection**: Hash type otomatis terdeteksi berdasarkan file
2. **User Experience**: User tidak perlu manual set hash type
3. **Accuracy**: Mengurangi kesalahan dalam pemilihan hash type
4. **Flexibility**: Mendukung berbagai jenis file WiFi
5. **Real-time Updates**: UI berubah secara instan ketika file berubah

## Backend Integration

### File Type Detection
Backend otomatis mendeteksi file type saat upload:
```go
func (u *hashFileUsecase) determineFileType(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".hccapx":
        return "hccapx"
    case ".hccap":
        return "hccap"
    case ".cap":
        return "cap"
    case ".pcap":
        return "pcap"
    default:
        return "hash"
    }
}
```

### Database Storage
File type disimpan di database untuk referensi:
```sql
CREATE TABLE hash_files (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    orig_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    type TEXT NOT NULL,  -- hccapx, hccap, cap, pcap, hash
    created_at DATETIME NOT NULL
);
```

## Testing

### Test Cases
1. **WiFi Handshake Files**
   - Upload file .hccapx → Hash Type harus 2500
   - Upload file .hccap → Hash Type harus 2500
   - Upload file .cap → Hash Type harus 2500

2. **Generic Hash Files**
   - Upload file .hash → Hash Type harus 0
   - Upload file .txt → Hash Type harus 0

3. **Edge Cases**
   - Tidak ada file dipilih → Default WPA/WPA2 (2500)
   - File tidak ditemukan → Default WPA/WPA2 (2500)

## Future Enhancements

### Support untuk Hash Types Lain
```typescript
// Bisa ditambahkan untuk hash types lain
case 'md5':
    return '0' // MD5 hashes
case 'sha1':
    return '100' // SHA1 hashes
case 'ntlm':
    return '1000' // NTLM hashes
```

### File Validation
```typescript
// Validasi file sebelum set hash type
if (fileType === 'hccapx' && !isValidHccapxFile(selectedFile)) {
    return '0' // Fallback ke generic hash
}
```

## Kesimpulan

Fitur ini membuat sistem lebih cerdas dengan:
- **Otomatis mendeteksi** hash type berdasarkan file yang dipilih
- **Khusus untuk WiFi** handshake files (.hccapx, .hccap, .cap, .pcap)
- **Fallback ke generic** hash untuk file yang tidak dikenali
- **Real-time updates** di UI tanpa perlu refresh
- **User-friendly** dengan deskripsi yang jelas
