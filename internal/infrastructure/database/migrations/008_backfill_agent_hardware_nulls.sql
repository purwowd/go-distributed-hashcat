-- Migration: 008_backfill_agent_hardware_nulls.sql
-- Description: Backfill NULL resource_type/processor values for legacy agent rows
-- Date: 2026-06-24

-- +migrate Up
UPDATE agents SET resource_type = '' WHERE resource_type IS NULL;
UPDATE agents SET processor = '' WHERE processor IS NULL;

-- +migrate Down
-- no-op
