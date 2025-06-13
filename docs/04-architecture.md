# 🏗️ System Architecture

Modern distributed hashcat system with clean architecture and horizontal scalability.

## 📋 Overview

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Frontend** | TypeScript + Alpine.js | Web dashboard |
| **Backend** | Go + Gin + Clean Architecture | REST API server |
| **Database** | SQLite + WAL mode | Persistent storage |
| **Agents** | Go + Hashcat integration | GPU workers |
| **Communication** | HTTP REST + JSON | Inter-service protocol |

## 🏛️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    DISTRIBUTED HASHCAT SYSTEM                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────┐ │
│  │   Frontend      │    │   Backend       │    │  Database   │ │
│  │  (TypeScript)   │◄──►│     (Go)        │◄──►│  (SQLite)   │ │
│  │  localhost:3000 │    │ localhost:1337  │    │   +WAL      │ │
│  └─────────────────┘    └─────────────────┘    └─────────────┘ │
│                                   │                            │
│                     ┌─────────────┴─────────────┐              │
│                     │                           │              │
│                ┌────▼─────┐                ┌─────▼─────┐       │
│                │  Agent   │                │  Agent    │       │
│                │  (GPU)   │                │  (GPU)    │       │
│                └──────────┘                └───────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

## 🔧 Component Details

### **Frontend Layer (TypeScript)**
```typescript
├── Alpine.js         // Reactive framework
├── Tailwind CSS     // Utility-first styling  
├── Vite            // Build tool & dev server
└── TypeScript      // Type safety
```

**Features**: Real-time monitoring, file upload, responsive design

### **Backend Layer (Go)**
```go
internal/
├── domain/          // Business entities
├── usecase/         // Application logic
├── delivery/        // HTTP handlers (Gin)
└── infrastructure/  // Database & external services
```

**Architecture**: Clean Architecture with dependency injection

### **Database Layer (SQLite)**
```sql
-- Core entities with relationships
agents ──┐
         ├── jobs ──┐
         │          ├── hash_files
         │          └── wordlists
         └── agent_files
```

**Optimizations**: WAL mode, indexed foreign keys, connection pooling

### **Agent Layer**
```bash
├── Agent Discovery   # Auto-registration
├── File Management   # Local file sync
├── Job Execution     # Hashcat process control
└── Progress Reporting # Real-time status updates
```

## 🔄 Data Flow

### **Job Lifecycle**
```
Frontend → Backend → Database → Agents → Hashcat Execution → Progress Updates
```

1. **Create**: Frontend creates job via API
2. **Validate**: Backend validates and stores job
3. **Queue**: Job enters queue for execution
4. **Assign**: Backend assigns job to available agent
5. **Execute**: Agent executes hashcat process
6. **Monitor**: Real-time progress updates

### **File Management Flow**
```
Upload → Validation → Storage → Distribution → Cleanup
```

## 🛡️ Security Architecture

### **Network Security**
```
Public Internet → VPN Gateway → Private Network → Control Server → GPU Workers
```

### **Security Layers**
| Layer | Protection | Implementation |
|-------|------------|----------------|
| **Network** | VPN encryption | WireGuard tunneling |
| **Transport** | TLS/SSL | HTTPS certificates |
| **Application** | Input validation | Sanitization & limits |
| **Database** | Transaction safety | SQLite WAL mode |

## 🚀 Deployment Patterns

### **Single Node (Development)**
```
┌─────────────────────────────────┐
│         Single Machine          │
│  Frontend  Backend  Database    │
│    :3000    :1337   SQLite      │
│         Local Agent             │
└─────────────────────────────────┘
```

### **Multi-Node (Production)**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Control Node   │    │  Worker Node 1  │    │  Worker Node 2  │
│  Frontend       │    │     Agent       │    │     Agent       │
│  Backend        │    │    (2x RTX)     │    │    (4x A100)    │
│  Database       │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────── VPN Network (WireGuard) ─────────┘
```

## 📊 Performance Characteristics

### **System Performance**
- **API Response**: <5ms average
- **Database Throughput**: 1000+ ops/sec
- **Memory Usage**: <100MB backend
- **Concurrent Support**: 100+ agents, 1000+ jobs

### **Optimization Features**
- **Database**: SQLite WAL mode for concurrent access
- **Caching**: In-memory caching
- **Connection Pooling**: Efficient database connections
- **Async Processing**: Non-blocking I/O operations

## 🔧 Technology Stack

### **Backend Technologies**
- **Language**: Go 1.24+
- **Framework**: Gin HTTP router
- **Database**: SQLite 3.35+ with WAL mode
- **Architecture**: Clean Architecture pattern

### **Frontend Technologies**
- **Language**: TypeScript 5.0+
- **Framework**: Alpine.js for reactivity
- **Styling**: Tailwind CSS
- **Build Tool**: Vite

### **Infrastructure**
- **Container**: Docker support
- **VPN**: WireGuard for secure communication
- **Process Management**: Systemd services
- **Monitoring**: Built-in health checks

**Next Steps**: [`05-database-migrations.md`](05-database-migrations.md) for database details
