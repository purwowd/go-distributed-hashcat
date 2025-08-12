# 🔓 Distributed Hashcat

**Modern distributed password cracking system** with Go backend, TypeScript frontend, and clean architecture.

[![Go](https://img.shields.io/badge/Go-1.24-blue)](https://golang.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0-blue)](https://www.typescriptlang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 🚀 Key Features

- **🔥 High Performance**: Multi-GPU distributed cracking
- **🌐 Web Dashboard**: Real-time monitoring with TypeScript UI  
- **🏗️ Clean Architecture**: Go backend with domain-driven design
- **🔒 Secure**: WireGuard VPN support
- **📊 Real-time Updates**: Live progress tracking

**Stack**: Go 1.24 + Gin + SQLite + TypeScript + Alpine.js + Tailwind CSS

## ⚡ Quick Start

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