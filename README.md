# 🔓 Distributed Hashcat

**Modern distributed password cracking system** with Go backend, TypeScript frontend, and clean architecture.

[![Go](https://img.shields.io/badge/Go-1.24.7-blue)](https://golang.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2-blue)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-5.0.10-646CFF)](https://vitejs.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-20.19.5-green)](https://nodejs.org/)
[![npm](https://img.shields.io/badge/npm-10.8.2-red)](https://npmjs.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 🚀 Key Features

- **🔥 High Performance**: Multi-GPU distributed cracking
- **🌐 Web Dashboard**: Real-time monitoring with TypeScript UI  
- **🏗️ Clean Architecture**: Go backend with domain-driven design
- **🔒 Secure**: WireGuard VPN support
- **📊 Real-time Updates**: Live progress tracking
- **✅ Production Ready**: Fully tested and optimized builds

**Stack**: Go 1.24.7 + Gin + SQLite + TypeScript 5.9.2 + Vite 5.0.10 + Alpine.js + Tailwind CSS + Node.js 20.19.5

## ⚡ Quick Start

### 🏗️ Build from Source (Local Development)

```bash
# Clone repository
git clone https://github.com/purwowd/go-distributed-hashcat.git
cd go-distributed-hashcat

# Build server and agent binaries
go build -o server cmd/server/main.go
go build -o agent cmd/agent/main.go

# Or build to bin/ directory
go build -o bin/server cmd/server/main.go
go build -o bin/agent cmd/agent/main.go

# Create symlinks (optional)
ln -sf bin/server server
ln -sf bin/agent agent
```

### 🚀 Run Locally

#### **Development Mode (Recommended)**
```bash
# Start server in development mode with hot reload
bash scripts/dev-server.sh

# In another terminal, start agent
./agent --server http://localhost:1337 --agent-key YOUR_AGENT_KEY
# or  
./bin/agent --server http://localhost:1337 --agent-key YOUR_AGENT_KEY
```

#### **Production Mode**
```bash
# Start server
./server
# or
./bin/server

# In another terminal, start agent
./agent --server http://localhost:1337 --agent-key YOUR_AGENT_KEY
# or  
./bin/agent --server http://localhost:1337 --agent-key YOUR_AGENT_KEY
```

### 📦 Production Build (Ubuntu)

```bash
# Install dependencies (Ubuntu)
sudo apt update && sudo apt install git nodejs npm hashcat sqlite3 -y

# Install Go 1.24.7 (required)
wget https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/purwowd/go-distributed-hashcat.git
cd go-distributed-hashcat
make build

# Start backend and frontend
./bin/server &
cd frontend && npm install && npm run dev
```

**Access**: http://localhost:3000

## 🎨 Development Workflow

### **Complete Development Setup**
```bash
# Terminal 1: Start backend
bash scripts/dev-server.sh

# Terminal 2: Start frontend
cd frontend
npm run dev
```

**Access**: 
- Backend API: http://localhost:1337
- Frontend Dashboard: http://localhost:3000
- API Documentation: http://localhost:1337/docs

### **Backend Development**
```bash
# Start backend in development mode (with hot reload)
bash scripts/dev-server.sh
```

### **Frontend Development**

The frontend is built with **Vite 5.0.10** and **TypeScript 5.9.2** for fast development and optimized builds:

```bash
cd frontend

# Install dependencies
npm install

# Development server (with hot reload)
npm run dev

# Type checking
npm run type-check

# Build for production
npm run build

# Build for production with optimizations
npm run build:prod

# Preview production build
npm run preview

# Code formatting
npm run format

# Linting and auto-fix
npm run lint:fix

# Clean build artifacts
npm run clean
```

**Frontend Features**:
- ⚡ **Vite 5.0.10**: Lightning-fast build tool
- 🔷 **TypeScript 5.9.2**: Type-safe development with latest features
- 🎨 **Tailwind CSS 3.4.0**: Utility-first styling
- 🏔️ **Alpine.js 3.13.3**: Lightweight reactive framework
- 📦 **ES Modules**: Modern JavaScript modules
- 🔧 **ESLint + Prettier**: Code quality and formatting
- 🎯 **Production Ready**: Optimized builds with Terser

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
- **Frontend**: Vite 5.0.10 + TypeScript 5.9.2 + Alpine.js 3.13.3 + Tailwind CSS 3.4.0
- **Domain**: Core business logic (`internal/domain/`)
- **Use Cases**: Application logic (`internal/usecase/`)
- **Infrastructure**: Database, external services
- **Delivery**: HTTP handlers, CLI

**Technology Stack**:
- **Backend**: Go 1.24.7 + Gin + SQLite + CGO
- **Frontend**: Vite 5.0.10 + TypeScript 5.9.2 + Alpine.js 3.13.3 + Tailwind CSS 3.4.0
- **Build**: Make + Terser 5.43.1 + ESLint 8.56.0 + Prettier 3.1.1
- **Development**: Hot reload, Type checking, Linting, Auto-formatting

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
| [`docs/07-agent-speed-feature.md`](docs/07-agent-speed-feature.md) | Agent speed detection | 15 min |
| [`docs/09-realtime-speed-monitoring.md`](docs/09-realtime-speed-monitoring.md) | Real-time monitoring | 20 min |
| [`docs/11-speed-based-distribution.md`](docs/11-speed-based-distribution.md) | Smart job distribution | 15 min |
| [`docs/ENVIRONMENT_SETUP.md`](docs/ENVIRONMENT_SETUP.md) | Environment configuration | 10 min |
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
# Development
bash scripts/dev-server.sh    # Start backend in development mode
make dev                      # Development servers
make test                     # Run tests  

# Production
make build                    # Production build
make docker                   # Docker build
```

## 📈 Performance

- **Throughput**: 10K+ requests/second
- **Latency**: <5ms average
- **Scalability**: 100+ agents, 1000+ jobs
- **GPU Support**: RTX 4090 (1.2M H/s), RTX 3080 (800K H/s)

## 🔧 Troubleshooting

### Go Version Issues
```bash
# Check Go version (must be 1.24+)
go version

# If you get "cannot compile Go 1.24 code" error:
# Install Go 1.24.7 manually
wget https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Frontend TypeScript Errors
```bash
cd frontend

# Install missing dependencies
npm install

# Check for TypeScript errors
npm run type-check

# Fix linting issues
npm run lint:fix
```

**Recent Fixes** (v1.0.0):
- ✅ **setTimeout Type Issues**: Fixed TypeScript errors with `setTimeout` return types
- ✅ **Build Optimization**: Improved build process with proper type checking
- ✅ **Code Quality**: Enhanced ESLint and Prettier configuration

### Build Issues
```bash
# Clean and rebuild
make clean
make build

# Check if binaries were created
ls -la bin/
```

---

**🎯 For beginners**: [`docs/01-quick-start.md`](docs/01-quick-start.md)  
**🔧 For production**: [`docs/02-deployment.md`](docs/02-deployment.md)  
**🔒 For security**: [`docs/06-wireguard-deployment.md`](docs/06-wireguard-deployment.md)