# Documentation Index

Comprehensive documentation for the distributed hashcat system.

## 📚 Documentation Overview

### **🚀 Quick Start & Deployment**

| Document | Description | Audience |
|----------|-------------|----------|
| **[Quick Start WireGuard](quick-start-wireguard.md)** | 15-minute rapid deployment guide | DevOps, Beginners |
| **[WireGuard Deployment](wireguard-deployment.md)** | Complete production setup with VPN | System Administrators |
| **[General Deployment](deployment.md)** | Standard deployment options | DevOps Teams |

### **🏗️ System Architecture**

| Document | Description | Technical Level |
|----------|-------------|-----------------|
| **[Architecture Guide](architecture.md)** | System design and components | Intermediate |
| **[Database Migrations](migrations.md)** | Database schema management | Advanced |

### **🔌 API Documentation**

| Document | Description | Usage |
|----------|-------------|-------|
| **[API Reference](api.md)** | Complete REST API documentation | Developers |
| **[API Schema](api-schema.json)** | OpenAPI specification | Integration |

## 🎯 Quick Navigation

### **For System Administrators**
1. Start with [Quick Start WireGuard](quick-start-wireguard.md) for rapid setup
2. Move to [WireGuard Deployment](wireguard-deployment.md) for production
3. Reference [Architecture Guide](architecture.md) for system understanding

### **For Developers**
1. Review [Architecture Guide](architecture.md) for system overview
2. Study [API Reference](api.md) for integration
3. Use [API Schema](api-schema.json) for automated tooling

### **For DevOps Engineers**
1. Choose deployment: [WireGuard](wireguard-deployment.md) or [Standard](deployment.md)
2. Setup [Database Migrations](migrations.md) for schema management
3. Monitor using [Architecture Guide](architecture.md) patterns

## 🔧 Deployment Scenarios

### **Scenario 1: Secure Multi-Node Setup**
**Use Case**: Production environment with remote GPU workers
**Documents**:
- [Quick Start WireGuard](quick-start-wireguard.md) - Setup basics
- [WireGuard Deployment](wireguard-deployment.md) - Complete guide

### **Scenario 2: Local Development**
**Use Case**: Single machine or local network testing
**Documents**:
- [General Deployment](deployment.md) - Standard setup
- [API Reference](api.md) - Development integration

### **Scenario 3: Cloud Infrastructure**
**Use Case**: Scalable cloud deployment
**Documents**:
- [Architecture Guide](architecture.md) - Scaling patterns
- [WireGuard Deployment](wireguard-deployment.md) - Secure networking

## 📊 Document Complexity Matrix

```
    Simple ←→ Advanced
    ┌─────────────────────┐
    │ Quick Start WG      │ ← Start here
    ├─────────────────────┤
    │ API Reference       │
    │ Deployment Guide    │
    ├─────────────────────┤
    │ WireGuard Deploy    │
    │ Architecture        │
    ├─────────────────────┤
    │ Migrations          │ ← Advanced
    │ API Schema          │
    └─────────────────────┘
```

## 🚀 Getting Started Workflow

### **New to the System?**
```bash
1. Read: Quick Start WireGuard (15 min)
2. Deploy: Follow the quick guide
3. Explore: Access dashboard and API
4. Scale: Move to full WireGuard deployment
```

### **Integrating with Existing Systems?**
```bash
1. Study: Architecture Guide
2. Reference: API Documentation  
3. Implement: Using API Schema
4. Deploy: Choose appropriate deployment guide
```

### **Administering Production?**
```bash
1. Deploy: WireGuard Production Setup
2. Manage: Database Migrations
3. Monitor: Architecture patterns
4. Scale: Additional nodes via WireGuard
```

## 🔗 External Resources

- **GitHub Repository**: Main codebase and issues
- **Interactive API Docs**: `http://localhost:1337/docs` (when running)
- **Dashboard**: `http://localhost:3000` (development) or `http://15.15.15.1:3000` (WireGuard)

## 📈 Documentation Maintenance

### **Update Frequency**
- **API Reference**: Updated with each API change
- **Deployment Guides**: Updated with infrastructure changes
- **Architecture**: Updated with major system changes
- **Quick Start**: Kept simple and stable

### **Contributing to Documentation**
1. Follow existing document structure
2. Include practical examples
3. Test all commands and procedures
4. Update this index when adding new docs

---

**💡 Tip**: Bookmark this page as your documentation hub! All guides are designed to be self-contained but reference each other where helpful. 
