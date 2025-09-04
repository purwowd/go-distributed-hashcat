-- Migration: 005_add_speed_to_agents.sql
-- Description: Add speed field to agents table for storing hashcat benchmark speed
-- Author: System
-- Date: 2025-09-02

-- +migrate Up
ALTER TABLE agents ADD COLUMN speed INTEGER DEFAULT 0;

-- +migrate Down
ALTER TABLE agents DROP COLUMN speed;