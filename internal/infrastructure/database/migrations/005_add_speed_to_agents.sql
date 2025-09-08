-- Migration: 005_add_speed_to_agents.sql
-- Description: Add speed field to agents table for storing hashcat benchmark speed
-- Author: System
-- Date: 2025-09-02

-- +migrate Up
-- Note: speed column already exists in current schema, this migration is a no-op
-- ALTER TABLE agents ADD COLUMN speed INTEGER DEFAULT 0;

-- +migrate Down
-- Note: speed column should not be dropped as it's part of the base schema
-- ALTER TABLE agents DROP COLUMN speed;