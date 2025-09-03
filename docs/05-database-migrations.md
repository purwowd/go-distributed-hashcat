# Database Schema & Migrations

Database management, schema migrations, and operations for the distributed hashcat system.

## 📋 Table of Contents

- [Schema Overview](#-schema-overview)
- [Core Tables](#-core-tables)
- [Relationships](#-relationships)
- [Migration System](#-migration-system)
- [Common Operations](#-common-operations)
- [Performance Optimization](#-performance-optimization)
- [Backup & Recovery](#-backup--recovery)

## 📊 Schema Overview

SQLite database with clean relational design:

```sql
-- Core entities hierarchy
agents → jobs → hash_files, wordlists
       ↓
   agent_files (file synchronization)
```

**Features**: UUID primary keys, foreign key constraints, optimized indexes, WAL mode

## 🗄️ Core Tables

### **Agents Table**
```sql
CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    ip_address TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 8080,
    status TEXT NOT NULL DEFAULT 'offline',
    capabilities TEXT,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_last_seen ON agents(last_seen);
```

### **Jobs Table**
```sql
CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    hash_file_id TEXT,
    wordlist_id TEXT,
    attack_mode INTEGER NOT NULL DEFAULT 0,
    hash_type INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    agent_id TEXT,
    progress REAL DEFAULT 0.0,
    speed INTEGER DEFAULT 0,
    result TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hash_file_id) REFERENCES hash_files(id),
    FOREIGN KEY (wordlist_id) REFERENCES wordlists(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_agent_id ON jobs(agent_id);
```

### **Hash Files & Wordlists Tables**
```sql
CREATE TABLE hash_files (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    orig_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wordlists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    orig_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    word_count INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### **Agent Files Table**
```sql
CREATE TABLE agent_files (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    type TEXT NOT NULL,
    hash TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_files_agent_id ON agent_files(agent_id);
CREATE UNIQUE INDEX idx_agent_files_unique ON agent_files(agent_id, name);
```

## 🔗 Relationships

### **Entity Relationship**
```
agents (1) ──→ (N) jobs
agents (1) ──→ (N) agent_files
hash_files (1) ──→ (N) jobs
wordlists (1) ──→ (N) jobs
```

### **Foreign Key Constraints**
| Table | Column | References | Action |
|-------|--------|------------|--------|
| jobs | hash_file_id | hash_files(id) | SET NULL |
| jobs | wordlist_id | wordlists(id) | SET NULL |
| jobs | agent_id | agents(id) | SET NULL |
| agent_files | agent_id | agents(id) | CASCADE |

## Migration System

### **Migration Structure**
```
internal/infrastructure/database/
├── sqlite.go           # Database connection & setup
├── migrations/         # Migration files
│   ├── 001_initial.sql
│   ├── 002_add_wordlist_id.sql
│   └── 003_agent_files.sql
└── schema.sql         # Current schema
```

### **Adding New Migrations**
```sql
-- Example: 004_add_job_priority.sql
ALTER TABLE jobs ADD COLUMN priority INTEGER DEFAULT 1;
CREATE INDEX idx_jobs_priority ON jobs(priority);

-- Update internal/infrastructure/database/sqlite.go
-- Add "004_add_job_priority.sql" to migrations slice
```

## 💼 Common Operations

### **CRUD Operations**

```sql
-- Create agent
INSERT INTO agents (id, name, ip_address, port, capabilities) 
VALUES (?, ?, ?, ?, ?);

-- Update agent status
UPDATE agents SET status = ?, last_seen = CURRENT_TIMESTAMP WHERE id = ?;

-- Create job
INSERT INTO jobs (id, name, hash_file_id, attack_mode, hash_type) 
VALUES (?, ?, ?, ?, ?);

-- Update job progress
UPDATE jobs SET progress = ?, speed = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- List active jobs
SELECT * FROM jobs WHERE status IN ('pending', 'running') ORDER BY created_at;

-- Get agent with jobs
SELECT a.*, COUNT(j.id) as job_count 
FROM agents a 
LEFT JOIN jobs j ON a.id = j.agent_id 
WHERE a.status = 'online' 
GROUP BY a.id;
```

## 📊 Performance Optimization

### **Database Configuration**
```sql
-- SQLite optimizations (already configured)
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=10000;
PRAGMA temp_store=memory;
PRAGMA mmap_size=268435456;
```

### **Indexing Strategy**
```sql
-- Performance indexes
CREATE INDEX idx_jobs_status_created ON jobs(status, created_at);
CREATE INDEX idx_agents_status_last_seen ON agents(status, last_seen);
CREATE INDEX idx_agent_files_agent_type ON agent_files(agent_id, type);
```

## Backup & Recovery

### **Database Backup**
```bash
# SQLite backup
sqlite3 data/hashcat.db ".backup backup_$(date +%Y%m%d_%H%M%S).db"

# WAL checkpoint (for consistency)
sqlite3 data/hashcat.db "PRAGMA wal_checkpoint(FULL);"

# Automated backup script
#!/bin/bash
DB_PATH="data/hashcat.db"
BACKUP_DIR="backups"
DATE=$(date +%Y%m%d_%H%M%S)
sqlite3 $DB_PATH ".backup $BACKUP_DIR/hashcat_$DATE.db"
```

### **Recovery**
```bash
# Restore from backup
cp backups/hashcat_20241225_120000.db data/hashcat.db

# Check database integrity
sqlite3 data/hashcat.db "PRAGMA integrity_check;"
```

## 🔍 Monitoring & Maintenance

### **Database Health**
```sql
-- Check database size
SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();

-- Check table stats
SELECT name, seq FROM sqlite_sequence;

-- Check index usage
EXPLAIN QUERY PLAN SELECT * FROM jobs WHERE status = 'running';
```

### **Maintenance Tasks**
```sql
-- Vacuum database (periodic cleanup)
VACUUM;

-- Analyze statistics (for query optimization)
ANALYZE;

-- Reindex all tables
REINDEX;
```

**Next Steps**: [`06-wireguard-deployment.md`](06-wireguard-deployment.md) for secure VPN setup
