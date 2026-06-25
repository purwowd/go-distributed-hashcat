-- +migrate Up
ALTER TABLE wordlists ADD COLUMN source TEXT NOT NULL DEFAULT 'uploaded';
CREATE INDEX IF NOT EXISTS idx_wordlists_orig_name ON wordlists(orig_name);
CREATE INDEX IF NOT EXISTS idx_wordlists_source ON wordlists(source);

-- +migrate Down
-- SQLite cannot drop columns easily; leave source column in place on rollback.
