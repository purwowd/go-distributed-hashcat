-- Migration: 006_create_users_table.sql
-- Description: Create users table for authentication system
-- Author: System
-- Date: 2025-01-08

-- +migrate Up
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    last_login DATETIME
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Insert default admin user (password: admin123)
-- Password hash for 'admin123' using bcrypt with cost 10
INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@hashcat.local',
    '$2b$10$BO6WWqPlx/PsOj7GqnomHegRitXdBajhwRPeG0lQEiA7j6yj.g.M2',
    'admin',
    1,
    datetime('now'),
    datetime('now')
);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
