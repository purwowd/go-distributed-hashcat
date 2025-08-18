-- Migration: Add job progress tracking fields
-- Date: 2025-08-17
-- Description: Add fields for tracking job progress, speed, ETA, and word counts

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

-- Add comments to explain the new fields
COMMENT ON COLUMN jobs.total_words IS 'Total dictionary words for this job';
COMMENT ON COLUMN jobs.processed_words IS 'Number of words already processed';
COMMENT ON COLUMN jobs.speed IS 'Hash rate in H/s (hashes per second)';
COMMENT ON COLUMN jobs.eta IS 'Estimated time of completion';
COMMENT ON COLUMN jobs.rules IS 'Hashcat rules or cracked password result';
