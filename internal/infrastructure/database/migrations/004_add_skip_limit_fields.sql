-- Migration: 004_add_skip_limit_fields.sql
-- Description: Add skip and word_limit fields to jobs table for distributed cracking
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Add skip and word_limit columns for hashcat distributed cracking
ALTER TABLE jobs ADD COLUMN skip INTEGER;
ALTER TABLE jobs ADD COLUMN word_limit INTEGER;

-- Add index for efficient querying of distributed jobs
CREATE INDEX IF NOT EXISTS idx_jobs_skip_limit ON jobs(skip, word_limit);

-- +migrate Down
DROP INDEX IF EXISTS idx_jobs_skip_limit;
