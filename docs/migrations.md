# Database Migrations

Sistem migration untuk mengelola perubahan database schema secara terstruktur dan versioned.

## 📋 Overview

Migration system ini menyediakan:
- ✅ **Automatic versioning** dengan timestamp
- ✅ **Up/Down migrations** untuk rollback
- ✅ **Migration tracking** di database
- ✅ **CLI commands** untuk kemudahan penggunaan
- ✅ **Transaction safety** untuk data integrity

## 🚀 Quick Start

### 1. Generate Migration Baru

```bash
# Generate migration untuk model baru
./server migrate generate "add users table"

# Generate migration untuk perubahan schema
./server migrate generate "add index to jobs table"

# Generate migration untuk data migration
./server migrate generate "migrate old data format"
```

### 2. Edit Migration File

File migration akan dibuat di `internal/infrastructure/database/migrations/` dengan format:
```
20241201123045_add_users_table.sql
```

Edit file migration:

```sql
-- Migration: 20241201123045_add_users_table.sql
-- Description: add users table
-- Author: john
-- Date: 2024-12-01

-- +migrate Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
```

### 3. Run Migrations

```bash
# Apply semua pending migrations
./server migrate up

# Rollback migration terakhir
./server migrate down

# Check status migrations
./server migrate status
```

## 📝 Migration Commands

### Generate Migration

```bash
./server migrate generate [name]
```

**Examples:**
```bash
./server migrate generate "add users table"
./server migrate generate "add foreign key to jobs"
./server migrate generate "modify agents table"
./server migrate generate "add_indexes_for_performance"
```

**Tips:**
- Gunakan nama descriptive yang jelas
- Spaces akan diconvert ke underscores
- Nama akan di-lowercase otomatis

### Run Migrations

```bash
# Run all pending migrations
./server migrate up
```

**Output:**
```
🔄 Running migration 20241201123045: add_users_table
✅ Applied migration 20241201123045: add_users_table
🔄 Running migration 20241201123156: add_indexes_for_performance
✅ Applied migration 20241201123156: add_indexes_for_performance
🎉 Successfully applied 2 migrations
```

### Rollback Migration

```bash
# Rollback the last applied migration
./server migrate down
```

**Output:**
```
🔄 Rolling back migration 20241201123156: add_indexes_for_performance
✅ Rolled back migration 20241201123156: add_indexes_for_performance
```

### Check Status

```bash
./server migrate status
```

**Output:**
```
📊 Migration Status:
==================
Version 20241201123045: add_users_table [✅ Applied]
Version 20241201123156: add_indexes_for_performance [❌ Pending]
Version 20241201124530: modify_agents_table [❌ Pending]

📈 Summary: 3 total, 1 applied, 2 pending
```

## 🏗️ Migration Best Practices

### 1. Structure Migration yang Baik

```sql
-- +migrate Up
-- Always start with CREATE statements
CREATE TABLE new_table (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Then add indexes
CREATE INDEX idx_new_table_name ON new_table(name);

-- Finally add foreign keys (if any)
-- ALTER TABLE new_table ADD CONSTRAINT ...

-- +migrate Down
-- Reverse order: drop constraints first, then indexes, then tables
-- DROP CONSTRAINT ...
DROP INDEX IF EXISTS idx_new_table_name;
DROP TABLE IF EXISTS new_table;
```

### 2. Naming Conventions

| Type | Example | Description |
|------|---------|-------------|
| Create table | `add_users_table` | Table creation |
| Modify table | `modify_users_add_email` | Table modification |
| Add index | `add_index_users_email` | Index creation |
| Data migration | `migrate_old_user_format` | Data transformation |
| Foreign key | `add_fk_jobs_agents` | Foreign key relations |

### 3. Safe Migration Patterns

#### ✅ Safe Operations
```sql
-- Add new table
CREATE TABLE new_table (...);

-- Add new column (with default)
ALTER TABLE users ADD COLUMN phone TEXT DEFAULT '';

-- Add index
CREATE INDEX idx_users_phone ON users(phone);

-- Add NOT NULL constraint (after setting default)
-- Step 1: Add column with default
ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';
-- Step 2: Later migration to add NOT NULL
-- ALTER TABLE users ALTER COLUMN status SET NOT NULL;
```

#### ⚠️ Potentially Risky Operations
```sql
-- Drop column (can cause data loss)
ALTER TABLE users DROP COLUMN old_column;

-- Rename column (might break existing code)
ALTER TABLE users RENAME COLUMN old_name TO new_name;

-- Change column type (might cause data loss)
ALTER TABLE users ALTER COLUMN age INTEGER;
```

### 4. Complex Migration Example

Model baru dengan relasi:

```sql
-- +migrate Up
-- Create notifications table
CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'info',
    read_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);

-- Add foreign key constraint (if users table exists)
-- Note: Enable foreign keys in SQLite first
PRAGMA foreign_keys = ON;

-- Insert default notification types
INSERT INTO notification_types (type, description) VALUES 
('info', 'Information notification'),
('warning', 'Warning notification'),
('error', 'Error notification'),
('success', 'Success notification');

-- +migrate Down
-- Drop foreign key constraint first (SQLite handles this on table drop)
-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_read_at;
DROP INDEX IF EXISTS idx_notifications_type;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_user_id;

-- Drop table
DROP TABLE IF EXISTS notifications;

-- Clean up notification types
DELETE FROM notification_types WHERE type IN ('info', 'warning', 'error', 'success');
```

## 🔧 Advanced Usage

### Custom Migrations Directory

```bash
./server migrate --migrations-dir ./custom/migrations up
```

### Development Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feature/user-management
   ```

2. **Generate migration**
   ```bash
   ./server migrate generate "add users table"
   ```

3. **Edit migration file**
   - Add UP SQL untuk membuat table
   - Add DOWN SQL untuk rollback

4. **Test migration**
   ```bash
   # Apply migration
   ./server migrate up
   
   # Test rollback
   ./server migrate down
   
   # Apply again
   ./server migrate up
   ```

5. **Commit migration file**
   ```bash
   git add internal/infrastructure/database/migrations/
   git commit -m "Add users table migration"
   ```

### Production Deployment

1. **Backup database** sebelum migration
   ```bash
   cp production.db production.db.backup
   ```

2. **Run migrations**
   ```bash
   ./server migrate status  # Check pending migrations
   ./server migrate up      # Apply migrations
   ```

3. **Verify deployment**
   ```bash
   ./server migrate status  # Confirm all applied
   ```

## 🚨 Troubleshooting

### Migration Failed

```bash
# Check status
./server migrate status

# Review last migration
cat internal/infrastructure/database/migrations/[last_migration].sql

# Fix SQL syntax and re-run
./server migrate up
```

### Rollback Issues

```bash
# If rollback fails, check DOWN section
# Make sure DOWN SQL is correct reverse of UP SQL

# Manually fix database if needed
sqlite3 data/hashcat.db
```

### Database Locked

```bash
# Check if server is running
ps aux | grep server

# Stop server before running migrations
kill [server_pid]

# Run migration
./server migrate up

# Restart server
./server
```

## 📊 Migration Tracking

Sistem tracking menggunakan table `schema_migrations`:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checksum TEXT NOT NULL
);
```

**Fields:**
- `version`: Timestamp-based version number
- `name`: Migration name
- `applied_at`: When migration was applied
- `checksum`: SQL content verification

## 🎯 Example Scenarios

### Adding New Model

```bash
# 1. Generate migration
./server migrate generate "add products table"

# 2. Edit migration file
# 3. Apply migration
./server migrate up

# 4. Create Go model, repository, usecase
# 5. Update API handlers
```

### Modifying Existing Model

```bash
# 1. Generate migration
./server migrate generate "add category to products"

# 2. Edit migration file with ALTER TABLE
# 3. Apply migration
./server migrate up

# 4. Update Go structs and code
```

### Performance Optimization

```bash
# 1. Generate migration
./server migrate generate "add indexes for query performance"

# 2. Add CREATE INDEX statements
# 3. Apply migration
./server migrate up

# 4. Test query performance
```

Dengan sistem migration ini, development dan deployment database schema jadi lebih terorganisir dan aman! 🚀 
