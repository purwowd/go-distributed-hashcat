-- +migrate Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME NULL
);

-- Create indexes for better performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Insert default admin user (password: admin123)
-- Password hash generated with: golang.org/x/crypto/bcrypt, cost 10
INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at) 
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@hashcat-dashboard.local',
    '$2a$10$8K1p/a0dURXAMkcspf.8Nu2YCb7d9A0VGMeAUvJqh7YdRwBd/Gxhe',
    'admin',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users; 
