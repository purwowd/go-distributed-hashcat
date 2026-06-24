-- Migration: 007_add_agent_hardware_fields.sql
-- Description: Add resource_type (CPU/GPU) and processor model fields to agents
-- Date: 2026-06-24

-- +migrate Up
ALTER TABLE agents ADD COLUMN resource_type TEXT;
ALTER TABLE agents ADD COLUMN processor TEXT;

-- +migrate Down
ALTER TABLE agents DROP COLUMN processor;
ALTER TABLE agents DROP COLUMN resource_type;
