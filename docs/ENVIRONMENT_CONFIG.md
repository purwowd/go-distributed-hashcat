# Environment Configuration Guide

This guide explains how to configure the distributed hashcat system using environment variables instead of hardcoded URLs.

## Overview

The system now supports environment-based configuration for:
- Backend server configuration
- Frontend API base URL
- Agent server connection
- WebSocket CORS origins

## Quick Setup

### 1. Backend Configuration

Create `.env` file in the root directory:

```bash
cp .env.example .env
```

Edit `.env` with your server IP:

```bash
# Backend Environment Configuration
HASHCAT_SERVER_PORT=1337
HASHCAT_SERVER_HOST=30.30.30.39

# Database Configuration
HASHCAT_DATABASE_TYPE=sqlite
HASHCAT_DATABASE_PATH=./data/hashcat.db

# Upload Configuration
HASHCAT_UPLOAD_DIRECTORY=./uploads

# API Base URL (for reference)
API_BASE_URL=http://30.30.30.39:1337
```

### 2. Frontend Configuration

Create `.env` file in the frontend directory:

```bash
cp frontend/.env.example frontend/.env
```

Edit `frontend/.env` with your server IP:

```bash
# Frontend Environment Variables
VITE_API_BASE_URL=http://30.30.30.39:1337
```

### 3. Agent Configuration

For agents, you can either:

**Option A: Use environment variable**
```bash
export HASHCAT_SERVER_URL=http://30.30.30.39:1337
./bin/agent --agent-key YOUR_AGENT_KEY
```

**Option B: Use command line parameter**
```bash
./bin/agent --server http://30.30.30.39:1337 --agent-key YOUR_AGENT_KEY
```

## Environment Variables Reference

### Backend Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `HASHCAT_SERVER_PORT` | Server port | 1337 | 1337 |
| `HASHCAT_SERVER_HOST` | Server host/IP | 0.0.0.0 | 30.30.30.39 |
| `HASHCAT_DATABASE_TYPE` | Database type | sqlite | sqlite |
| `HASHCAT_DATABASE_PATH` | Database file path | ./data/hashcat.db | ./data/hashcat.db |
| `HASHCAT_UPLOAD_DIRECTORY` | Upload directory | ./uploads | ./uploads |

### Frontend Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `VITE_API_BASE_URL` | Backend API URL | http://localhost:1337 | http://30.30.30.39:1337 |

### Agent Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `HASHCAT_SERVER_URL` | Server URL for agent connection | http://localhost:1337 | http://30.30.30.39:1337 |

## Running the System

### 1. Start Backend Server

```bash
# Load environment from .env file
go run cmd/server/main.go
```

### 2. Start Frontend

```bash
cd frontend
# Load environment from .env file
npm run dev
```

### 3. Start Agent

```bash
# Using environment variable
export HASHCAT_SERVER_URL=http://30.30.30.39:1337
./bin/agent --agent-key YOUR_AGENT_KEY

# Or using command line parameter
./bin/agent --server http://30.30.30.39:1337 --agent-key YOUR_AGENT_KEY
```

## Configuration Files

- `.env` - Backend environment variables (gitignored)
- `.env.example` - Backend environment template
- `frontend/.env` - Frontend environment variables (gitignored)
- `frontend/.env.example` - Frontend environment template
- `.env.agent.example` - Agent environment template

## Security Notes

- `.env` files are gitignored and should never be committed
- Use `.env.example` files as templates
- Change default values for production deployments
- Use strong passwords and secure configurations in production

## Troubleshooting

### Common Issues

1. **Frontend can't connect to backend**
   - Check `VITE_API_BASE_URL` in `frontend/.env`
   - Ensure backend is running on the specified host/port

2. **Agent can't connect to server**
   - Check `HASHCAT_SERVER_URL` environment variable
   - Verify server is accessible from agent machine

3. **WebSocket connection fails**
   - Check `HASHCAT_SERVER_HOST` in backend `.env`
   - Ensure CORS origins are properly configured

### Debug Commands

```bash
# Check environment variables
env | grep HASHCAT
env | grep VITE

# Verify configuration files exist
ls -la .env frontend/.env

# Test backend connectivity
curl http://30.30.30.39:1337/health
```

## Migration from Hardcoded URLs

If you're upgrading from a version with hardcoded URLs:

1. Copy the appropriate `.env.example` files to `.env`
2. Update the IP addresses in the `.env` files
3. Restart all services
4. Verify connectivity

The system will automatically use environment variables when available, falling back to sensible defaults when not configured.
