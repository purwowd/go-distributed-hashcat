-- Migration: Create agent_keys table
-- This table stores agent authentication keys separately from agents table

CREATE TABLE IF NOT EXISTS agent_keys (
    id TEXT PRIMARY KEY,
    agent_key TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'active', -- active, expired, revoked
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    last_used_at DATETIME,
    agent_id TEXT, -- Reference to agents.id when key is used
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
);

-- Index for fast key lookups
CREATE INDEX IF NOT EXISTS idx_agent_keys_key ON agent_keys(agent_key);
CREATE INDEX IF NOT EXISTS idx_agent_keys_status ON agent_keys(status);
CREATE INDEX IF NOT EXISTS idx_agent_keys_agent_id ON agent_keys(agent_id); 
