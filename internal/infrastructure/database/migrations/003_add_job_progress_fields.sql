-- Migration: 003_add_job_progress_fields.sql
-- Description: Add fields for tracking job progress, speed, ETA, and word counts
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Add new columns to jobs table
ALTER TABLE jobs ADD COLUMN total_words INTEGER DEFAULT 0;
ALTER TABLE jobs ADD COLUMN processed_words INTEGER DEFAULT 0;

-- Update existing jobs to have default values
UPDATE jobs SET total_words = 0 WHERE total_words IS NULL;
UPDATE jobs SET processed_words = 0 WHERE processed_words IS NULL;

-- Create index for better performance on progress queries
CREATE INDEX IF NOT EXISTS idx_jobs_status_progress ON jobs(status, progress);
CREATE INDEX IF NOT EXISTS idx_jobs_speed ON jobs(speed);
CREATE INDEX IF NOT EXISTS idx_jobs_total_words ON jobs(total_words);

-- +migrate Down
DROP INDEX IF EXISTS idx_jobs_total_words;
DROP INDEX IF EXISTS idx_jobs_speed;
DROP INDEX IF EXISTS idx_jobs_status_progress;
