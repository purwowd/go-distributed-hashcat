-- Migration: 001_initial_schema.sql
-- Description: Create initial tables for agents, jobs, hash_files, and wordlists
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    port INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'offline',
    capabilities TEXT,
    last_seen DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Create hash_files table
CREATE TABLE IF NOT EXISTS hash_files (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    orig_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    type TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

-- Create wordlists table
CREATE TABLE IF NOT EXISTS wordlists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    orig_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL,
    word_count INTEGER,
    created_at DATETIME NOT NULL
);

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    hash_type INTEGER NOT NULL,
    attack_mode INTEGER NOT NULL,
    hash_file TEXT NOT NULL,
    hash_file_id TEXT,
    wordlist TEXT NOT NULL,
    wordlist_id TEXT,
    rules TEXT,
    agent_id TEXT,
    progress REAL DEFAULT 0,
    speed INTEGER DEFAULT 0,
    eta DATETIME,
    result TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    started_at DATETIME,
    completed_at DATETIME,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL,
    FOREIGN KEY (hash_file_id) REFERENCES hash_files(id) ON DELETE SET NULL,
    FOREIGN KEY (wordlist_id) REFERENCES wordlists(id) ON DELETE SET NULL
);

-- +migrate Down
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS wordlists;
DROP TABLE IF EXISTS hash_files;
DROP TABLE IF EXISTS agents; 
