-- Migration: 009_create_agent_local_files.sql
-- Description: Store per-agent local file inventory reported by agents (no file sync)
-- Date: 2026-06-24

-- +migrate Up
CREATE TABLE IF NOT EXISTS agent_local_files (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL DEFAULT 0,
    type TEXT NOT NULL DEFAULT 'wordlist',
    hash TEXT,
    mod_time DATETIME,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    UNIQUE(agent_id, name, type)
);

CREATE INDEX IF NOT EXISTS idx_agent_local_files_agent_id ON agent_local_files(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_local_files_name ON agent_local_files(name);
CREATE INDEX IF NOT EXISTS idx_agent_local_files_type ON agent_local_files(type);

-- +migrate Down
DROP INDEX IF EXISTS idx_agent_local_files_type;
DROP INDEX IF EXISTS idx_agent_local_files_name;
DROP INDEX IF EXISTS idx_agent_local_files_agent_id;
DROP TABLE IF EXISTS agent_local_files;
