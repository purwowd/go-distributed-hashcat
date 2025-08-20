# Go Distributed Hashcat

A distributed password cracking system using hashcat, built with Go.

## Features

- Distributed password cracking across multiple agents
- Support for various hash types and attack modes
- Real-time progress monitoring via WebSocket
- Agent health monitoring and automatic failover
- Web-based management interface
- RESTful API for automation
- Database migrations support

## CLI Commands

### Server Management

```bash
# Start the server
./server

# Database migrations
./server migrate up          # Run pending migrations
./server migrate down        # Rollback last migration
./server migrate status      # Show migration status
./server migrate generate    # Generate new migration
```

### Wordlist Management

The server includes optimized CLI commands for managing wordlist files, especially designed for large files with 1 million+ words:

```bash
# Upload a wordlist file with optimized processing
./server wordlist upload /path/to/wordlist.txt

# Upload with custom name
./server wordlist upload /path/to/rockyou.txt --name "rockyou_2021"

# Upload large files with custom chunk size (in MB)
./server wordlist upload /path/to/large_wordlist.txt --chunk 50

# Disable word counting for faster uploads
./server wordlist upload /path/to/wordlist.txt --count=false

# List all uploaded wordlists
./server wordlist list

# Delete a wordlist by ID
./server wordlist delete 123e4567-e89b-12d3-a456-426614174000
```

#### Wordlist Upload Features

- **Optimized Processing**: Uses buffered I/O and efficient file handling for large files
- **Progress Reporting**: Real-time progress updates during upload
- **Configurable Chunking**: Adjustable chunk size for memory-efficient processing
- **Word Counting**: Automatic word count calculation (can be disabled for speed)
- **File Validation**: Automatic file format detection and validation
- **Memory Efficient**: Processes files in chunks to handle very large wordlists

#### Performance Tips for Large Wordlists

- Use `--chunk 50` for files larger than 1GB
- Disable word counting with `--count=false` for fastest uploads
- Ensure sufficient disk space in the upload directory
- For files with 10M+ words, consider using larger chunk sizes

## Installation

```bash
# Install dependencies (Ubuntu)
sudo apt update && sudo apt install git nodejs npm golang-go hashcat sqlite3 -y

# Clone and build
git clone https://github.com/purwowd/go-distributed-hashcat.git
cd go-distributed-hashcat
make build

# Start backend and frontend
./bin/server &
cd frontend && npm install && npm run dev
```

**Access**: http://localhost:3000

### 🗝️ Agent Key Setup

1. Buka dashboard server di browser: http://localhost:3000
2. Masuk ke menu **Agent Key**.
3. Buat agent name baru (misal: `gpu-worker-01`), lalu copy agent key yang di-generate.
4. Simpan agent key untuk digunakan pada worker.

### 🖥️ GPU Worker (remote machine)

```bash
./bin/agent --server http://YOUR_SERVER:1337 --name gpu-worker-01 --ip "AGENT_IP" --agent-key "AGENT_KEY"
```
- Ganti `YOUR_SERVER` dengan IP server.
- Ganti `AGENT_IP` dengan IP worker.
- Ganti `AGENT_KEY` dengan agent key yang sudah di-copy dari dashboard.

## 🏗️ Architecture

```
Frontend (TypeScript) ←→ REST API (Go) ←→ Agent Network (GPU)
    localhost:3000         localhost:1337      Port 8080+
```

**Clean Architecture Layers**:
- **Frontend**: TypeScript + Alpine.js + Tailwind CSS
- **Domain**: Core business logic (`internal/domain/`)
- **Use Cases**: Application logic (`internal/usecase/`)
- **Infrastructure**: Database, external services
- **Delivery**: HTTP handlers, CLI

## 🌐 Production Deploy

```bash
# Server
./bin/server --host 0.0.0.0 --port 1337 &
cd frontend && npm run build && python3 -m http.server 3000 &

# GPU workers (remote machines)
./bin/agent --server http://YOUR_SERVER:1337 --name gpu-worker-01 --ip "AGENT_IP" --agent-key "AGENT_KEY"
```

## 🔌 API Overview

RESTful API v1 with complete OpenAPI documentation.

**Interactive Docs**: http://localhost:1337/docs

```bash
# Example usage
curl -X POST -H "Content-Type: application/json" \
  -d '{"name": "WiFi Crack", "hash_file_id": "uuid"}' \
  http://localhost:1337/api/v1/jobs/
```

## 📚 Documentation

| Document | Purpose | Time |
|----------|---------|------|
| [`docs/01-quick-start.md`](docs/01-quick-start.md) | 15-minute setup | 15 min |
| [`docs/02-deployment.md`](docs/02-deployment.md) | Production deployment | 30 min |
| [`docs/03-api-reference.md`](docs/03-api-reference.md) | Complete API docs | 20 min |
| [`docs/04-architecture.md`](docs/04-architecture.md) | System design | 15 min |
| [`docs/05-database-migrations.md`](docs/05-database-migrations.md) | Database schema | 20 min |
| [`docs/06-wireguard-deployment.md`](docs/06-wireguard-deployment.md) | Secure VPN setup | 45 min |
| [`docs/99-performance.md`](docs/99-performance.md) | Benchmarks | 10 min |

## 🧪 Testing

```bash
# Run all tests
./scripts/run_tests.sh --all

# Quick benchmark
./scripts/run_tests.sh --benchmark-quick
```

**Performance** (Apple M3):
- API: 72K ops/sec, 18.59µs latency
- Memory: <100MB usage
- Database: 1000+ ops/sec

## 🔧 Build Commands

```bash
make dev         # Development servers
make test        # Run tests  
make build       # Production build
make docker      # Docker build
```

## 📈 Performance

- **Throughput**: 10K+ requests/second
- **Latency**: <5ms average
- **Scalability**: 100+ agents, 1000+ jobs
- **GPU Support**: RTX 4090 (1.2M H/s), RTX 3080 (800K H/s)

---

**🎯 For beginners**: [`docs/01-quick-start.md`](docs/01-quick-start.md)  
**🔧 For production**: [`docs/02-deployment.md`](docs/02-deployment.md)  
**🔒 For security**: [`docs/06-wireguard-deployment.md`](docs/06-wireguard-deployment.md)