-- Migration: 002_add_indexes.sql
-- Description: Add performance indexes for all tables
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Agents table indexes
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_agents_created_at ON agents(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agents_status_updated ON agents(status, updated_at DESC);

-- Jobs table indexes  
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_agent_id ON jobs(agent_id);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_updated_at ON jobs(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_status_agent ON jobs(status, agent_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_agent_status ON jobs(agent_id, status);

-- Hash files table indexes
CREATE INDEX IF NOT EXISTS idx_hash_files_type ON hash_files(type);
CREATE INDEX IF NOT EXISTS idx_hash_files_created_at ON hash_files(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_hash_files_size ON hash_files(size);

-- Wordlists table indexes
CREATE INDEX IF NOT EXISTS idx_wordlists_size ON wordlists(size);
CREATE INDEX IF NOT EXISTS idx_wordlists_created_at ON wordlists(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wordlists_word_count ON wordlists(word_count);

-- +migrate Down
-- Drop indexes (SQLite automatically drops indexes when tables are dropped)
DROP INDEX IF EXISTS idx_wordlists_word_count;
DROP INDEX IF EXISTS idx_wordlists_created_at;
DROP INDEX IF EXISTS idx_wordlists_size;

DROP INDEX IF EXISTS idx_hash_files_size;
DROP INDEX IF EXISTS idx_hash_files_created_at;
DROP INDEX IF EXISTS idx_hash_files_type;

DROP INDEX IF EXISTS idx_jobs_agent_status;
DROP INDEX IF EXISTS idx_jobs_status_created;
DROP INDEX IF EXISTS idx_jobs_status_agent;
DROP INDEX IF EXISTS idx_jobs_updated_at;
DROP INDEX IF EXISTS idx_jobs_created_at;
DROP INDEX IF EXISTS idx_jobs_agent_id;
DROP INDEX IF EXISTS idx_jobs_status;

DROP INDEX IF EXISTS idx_agents_status_updated;
DROP INDEX IF EXISTS idx_agents_created_at;
DROP INDEX IF EXISTS idx_agents_last_seen;
DROP INDEX IF EXISTS idx_agents_status; 
