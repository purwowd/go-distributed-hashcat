# Agent Usage with X-Agent-Key Authentication

## Overview
Dengan implementasi X-Agent-Key, setiap agent sekarang memerlukan authentication key untuk berkomunikasi dengan server. Ini meningkatkan keamanan sistem distributed hashcat.

## Langkah-langkah Menjalankan Agent

### 1. Generate Agent Key dari Dashboard

1. Buka dashboard di browser: `http://localhost:1337`
2. Klik tab **"Agent Keys"** di navigasi
3. Klik tombol **"Generate Key"**
4. Isi form:
   - **Name**: Nama untuk agent (contoh: "Production Agent 1")
   - **Description**: Deskripsi opsional
5. Klik **"Generate"**
6. **PENTING**: Copy agent key yang ditampilkan (hanya ditampilkan sekali!)

### 2. Jalankan Agent dengan Agent Key

#### Menggunakan Command Line Flag:
```bash
./bin/agent --agent-key "your-64-character-agent-key-here" \
           --server "http://localhost:1337" \
           --name "my-agent" \
           --capabilities "GPU"
```

#### Menggunakan Environment Variable (Recommended):
```bash
export AGENT_KEY="your-64-character-agent-key-here"
export SERVER_URL="http://localhost:1337"
export AGENT_NAME="my-agent"

./bin/agent
```

#### Contoh Lengkap:
```bash
# Set environment variables
export AGENT_KEY="a10dab61fa0c583ebe1289ed5b704608158f104c5fe169150dcceb4e12114080"
export SERVER_URL="http://localhost:1337"
export AGENT_NAME="gpu-agent-1"

# Run agent
./bin/agent --capabilities "GPU" --upload-dir "/opt/hashcat/uploads"
```

### 3. Parameter Agent

| Parameter | Environment Variable | Default | Deskripsi |
|-----------|---------------------|---------|-----------|
| `--agent-key` | `AGENT_KEY` | - | **REQUIRED** Agent authentication key |
| `--server` | `SERVER_URL` | `http://localhost:1337` | Server URL |
| `--name` | `AGENT_NAME` | `agent-{hostname}` | Agent name |
| `--ip` | - | auto-detect | Agent IP address |
| `--port` | - | `8081` | Agent port |
| `--capabilities` | - | `GPU` | Agent capabilities |
| `--upload-dir` | - | `/root/uploads` | Local uploads directory |

### 4. Verifikasi Agent Registration

Setelah agent berjalan, Anda akan melihat:

```
2025/06/16 04:45:30 Agent gpu-agent-1 (12345678-1234-1234-1234-123456789abc) registered successfully
2025/06/16 04:45:30 Local upload directory: /opt/hashcat/uploads
2025/06/16 04:45:30 Found 5 local files
```

Di dashboard, agent akan muncul dengan status "online" di tab **Agents**.

## Security Best Practices

### 1. Keamanan Agent Key
- **Jangan share** agent key di public repositories
- **Simpan dengan aman** agent key (gunakan secret management)
- **Revoke key** jika tidak digunakan lagi
- **Generate key baru** secara berkala

### 2. Environment Variables
```bash
# Buat file .env untuk development
echo "AGENT_KEY=your-key-here" > .env
echo "SERVER_URL=http://localhost:1337" >> .env
echo "AGENT_NAME=dev-agent" >> .env

# Load environment variables
source .env
./bin/agent
```

### 3. Production Deployment
```bash
# Systemd service example
[Unit]
Description=Hashcat Distributed Agent
After=network.target

[Service]
Type=simple
User=hashcat
Environment=AGENT_KEY=your-production-key
Environment=SERVER_URL=https://your-server.com
Environment=AGENT_NAME=prod-agent-01
ExecStart=/opt/hashcat/bin/agent --capabilities "GPU" --upload-dir "/opt/hashcat/uploads"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

### Error: "Agent key is required"
```
❌ Agent key is required. Use --agent-key flag or set AGENT_KEY environment variable
```
**Solusi**: Pastikan agent key sudah di-set via flag atau environment variable.

### Error: "401 Unauthorized"
```
Failed to register with server: 401 Unauthorized
```
**Solusi**: 
- Pastikan agent key valid dan belum expired
- Pastikan agent key belum di-revoke
- Generate agent key baru jika perlu

### Error: "Agent not found"
```
Failed to register with server: Agent not found
```
**Solusi**: Agent key mungkin sudah di-revoke. Generate agent key baru.

## Management Commands

### List Active Agent Keys
```bash
curl -s http://localhost:1337/api/v1/agent-keys/ | jq .
```

### Revoke Agent Key
```bash
curl -X DELETE http://localhost:1337/api/v1/agent-keys/{agent-key}/revoke
```

## Migration dari Agent Lama

Jika Anda memiliki agent yang berjalan tanpa agent key:

1. **Stop agent lama**:
   ```bash
   pkill -f "./bin/agent"
   ```

2. **Generate agent key** dari dashboard

3. **Update deployment script** dengan agent key

4. **Restart agent** dengan agent key baru

## Docker Deployment

```dockerfile
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache hashcat

# Copy agent binary
COPY bin/agent /usr/local/bin/agent

# Set environment variables
ENV AGENT_KEY=""
ENV SERVER_URL="http://localhost:1337"
ENV AGENT_NAME="docker-agent"

# Run agent
CMD ["agent", "--capabilities", "GPU", "--upload-dir", "/data/uploads"]
```

```bash
# Run with Docker
docker run -d \
  -e AGENT_KEY="your-agent-key" \
  -e SERVER_URL="http://your-server:1337" \
  -e AGENT_NAME="docker-agent-1" \
  -v /opt/uploads:/data/uploads \
  your-hashcat-agent:latest
``` 
