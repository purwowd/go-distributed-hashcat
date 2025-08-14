# Agent Validation Logic

## Overview

Sistem sekarang memiliki validasi yang ketat untuk pembuatan agent baru. Validasi ini memastikan bahwa:

1. **Agent Key harus ada di database** sebelum agent bisa dibuat
2. **Agent Name harus sesuai** dengan nama yang terdaftar di agent key
3. **IP Address harus unik** (jika diisi)

## Flow Validasi

### 1. Validasi Agent Key
- ✅ **SUCCESS**: Agent key ditemukan di database
- ❌ **FAILED**: Agent key tidak ditemukan di database
  - Error: `agent key 'XXXX' not found in database. Please generate a valid agent key first`
  - HTTP Status: `400 Bad Request`
  - Error Code: `AGENT_KEY_NOT_FOUND`

### 2. Validasi Agent Name
- ✅ **SUCCESS**: Nama agent sesuai dengan nama di agent key
- ❌ **FAILED**: Nama agent tidak sesuai dengan nama di agent key
  - Error: `agent name 'AgentName' does not match the name associated with agent key 'XXXX' (expected: 'ExpectedName')`
  - HTTP Status: `400 Bad Request`
  - Error Code: `AGENT_NAME_MISMATCH`

### 3. Validasi IP Address (Optional)
- ✅ **SUCCESS**: IP address unik atau kosong
- ❌ **FAILED**: IP address sudah digunakan agent lain
  - Error: `IP address 192.168.1.100 is already used by agent AgentName`
  - HTTP Status: `409 Conflict`
  - Error Code: `IP_ADDRESS_CONFLICT`

## Contoh Response Error

### Agent Key Tidak Ditemukan
```json
{
  "error": "agent key 'abcd1234' not found in database. Please generate a valid agent key first",
  "code": "AGENT_KEY_NOT_FOUND",
  "message": "The provided agent key does not exist in the database. Please generate a valid agent key first."
}
```

### Nama Agent Tidak Sesuai
```json
{
  "error": "agent name 'MyAgent' does not match the name associated with agent key 'abcd1234' (expected: 'ExpectedAgent')",
  "code": "AGENT_NAME_MISMATCH",
  "message": "The agent name does not match the name associated with the provided agent key."
}
```

### IP Address Konflik
```json
{
  "error": "IP address 192.168.1.100 is already used by agent ExistingAgent",
  "code": "IP_ADDRESS_CONFLICT",
  "message": "The IP address is already in use by another agent."
}
```

## Cara Kerja

### 1. Generate Agent Key
Pertama, buat agent key melalui endpoint `/api/v1/agent-keys`:
```bash
curl -X POST http://localhost:1337/api/v1/agent-keys \
  -H "Content-Type: application/json" \
  -d '{"name": "MyAgent"}'
```

### 2. Create Agent dengan Validasi
Kemudian buat agent dengan agent key yang sudah ada:
```bash
curl -X POST http://localhost:1337/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MyAgent",
    "ip_address": "192.168.1.100",
    "port": 8081,
    "capabilities": "NVIDIA RTX 4090",
    "agent_key": "abcd1234"
  }'
```

## Implementasi Teknis

### Repository Layer
- Method baru: `GetByAgentKey(ctx, agentKey string) (*Agent, error)`
- Menggunakan prepared statement untuk performa optimal
- Cache support untuk query yang sering digunakan

### Usecase Layer
- Validasi berurutan: Agent Key → Agent Name → IP Address
- Error handling yang spesifik untuk setiap jenis validasi
- Support untuk update agent existing

### Handler Layer
- Response error yang informatif dengan error code
- HTTP status code yang sesuai untuk setiap jenis error
- Message yang user-friendly

## Keuntungan

1. **Security**: Mencegah pembuatan agent dengan key yang tidak valid
2. **Data Integrity**: Memastikan konsistensi antara agent key dan agent name
3. **User Experience**: Error message yang jelas dan actionable
4. **Performance**: Menggunakan cache dan prepared statements
5. **Maintainability**: Kode yang terstruktur dan mudah di-maintain

## Testing

Untuk test validasi ini, gunakan skenario berikut:

1. **Test Agent Key Invalid**: Gunakan agent key yang tidak ada di database
2. **Test Name Mismatch**: Gunakan nama yang berbeda dengan yang ada di agent key
3. **Test IP Conflict**: Gunakan IP address yang sudah digunakan agent lain
4. **Test Success Case**: Gunakan data yang valid

## Troubleshooting

### Error: "agent key not found"
- Pastikan agent key sudah dibuat melalui endpoint agent-keys
- Periksa apakah agent key yang diinput sudah benar

### Error: "agent name does not match"
- Pastikan nama agent sama persis dengan nama di agent key
- Periksa case sensitivity (huruf besar/kecil)

### Error: "IP address conflict"
- Gunakan IP address yang berbeda
- Atau kosongkan field IP address untuk auto-detection
