-- Migration: 004_add_skip_limit_fields.sql
-- Description: Add index for skip and word_limit fields in jobs table for distributed cracking
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Note: skip and word_limit columns already exist in initial schema
-- Only add index for efficient querying of distributed jobs
CREATE INDEX IF NOT EXISTS idx_jobs_skip_limit ON jobs(skip, word_limit);

-- +migrate Down
DROP INDEX IF EXISTS idx_jobs_skip_limit;
