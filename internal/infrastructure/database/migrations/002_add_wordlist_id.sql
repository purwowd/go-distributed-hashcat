-- Migration: 002_add_wordlist_id.sql
-- Description: Add wordlist_id column to jobs table for proper wordlist reference
-- Author: System
-- Date: 2025-01-10

-- +migrate Up
ALTER TABLE jobs ADD COLUMN wordlist_id TEXT;

-- Add foreign key constraint (SQLite doesn't support adding FK constraints to existing tables directly)
-- We'll handle this in the application layer for now

-- +migrate Down
-- SQLite doesn't support DROP COLUMN, so we'd need to recreate the table
-- For now, we'll leave the column as it's backward compatible 
