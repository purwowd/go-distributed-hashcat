# Agent Port Default Feature

## Overview
Fitur ini memungkinkan agent untuk didaftarkan tanpa menyertakan port, dan sistem akan secara otomatis mengisi port default 8080.

## Perubahan yang Dibuat

### 1. Domain Model (`internal/domain/models.go`)
- Mengubah field `Port` di `CreateAgentRequest` dari `binding:"required"` menjadi `json:"port,omitempty"`
- Port sekarang bersifat optional dalam request JSON

### 2. Handler (`internal/delivery/http/handler/agent_handler.go`)
- Menambahkan logika default di method `RegisterAgent`
- Jika `req.Port == 0` (nilai default Go untuk int), maka port akan diset ke 8080

### 3. Repository (`internal/infrastructure/repository/agent_repository.go`)
- Memperbarui query `GetByNameAndIP` untuk juga mempertimbangkan port
- Query sekarang mencari berdasarkan `name`, `ip_address`, dan `port`

## Cara Kerja

### Request tanpa port:
```json
{
  "name": "my-agent",
  "ip_address": "192.168.1.100",
  "capabilities": "gpu,cpu"
}
```

### Request dengan port:
```json
{
  "name": "my-agent",
  "ip_address": "192.168.1.100",
  "port": 9090,
  "capabilities": "gpu,cpu"
}
```

## Testing

Test case baru telah ditambahkan untuk memverifikasi:
- Agent dapat dibuat tanpa port (default 8080)
- Agent dapat dibuat dengan port yang dispesifikasikan
- Validasi tetap bekerja untuk field required lainnya

## Kompatibilitas

- Semua API endpoint yang ada tetap kompatibel
- Database schema tidak berubah
- Test yang ada tetap berjalan dengan baik 