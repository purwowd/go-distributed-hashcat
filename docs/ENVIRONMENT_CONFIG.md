# Environment Configuration Guide

This guide explains how to configure the distributed hashcat system using environment variables instead of hardcoded URLs.

## Overview

The system now supports environment-based configuration for:
- Backend server configuration
- Frontend API base URL
- Agent server connection
- WebSocket CORS origins
- Development and production environments

## Quick Setup

### 1. Backend Configuration

Create `.env` file in the root directory:

```bash
cp .env.example .env
```

Edit `.env` with your server IP:

```bash
# Backend Environment Configuration for Distributed Hashcat

# Server Configuration
HASHCAT_SERVER_PORT=1337
HASHCAT_SERVER_HOST=0.0.0.0

# Database Configuration
HASHCAT_DATABASE_TYPE=sqlite
HASHCAT_DATABASE_PATH=./data/hashcat.db

# Upload Configuration
HASHCAT_UPLOAD_DIRECTORY=./uploads

# CORS Configuration
HASHCAT_FRONTEND_URL=http://30.30.30.39:3000

# API Base URL (for reference)
API_BASE_URL=http://30.30.30.39:1337

# Development Configuration
GIN_MODE=debug

# Production Override (uncomment for production)
# GIN_MODE=release
# HASHCAT_FRONTEND_URL=https://your-frontend-domain.com
# API_BASE_URL=https://your-api-domain.com
```

### 2. Frontend Configuration

Create `.env` file in the frontend directory:

```bash
cp frontend/.env.example frontend/.env
```

Edit `frontend/.env` with your server IP:

```bash
# Environment Configuration for Distributed Hashcat Frontend

# API Configuration
VITE_API_BASE_URL=http://30.30.30.39:1337

# Development Configuration
VITE_DEV_PORT=3000
VITE_DEV_HOST=true

# Build Configuration
VITE_APP_VERSION=1.0.0
VITE_BUILD_TIME=

# Feature Flags
VITE_ENABLE_HOT_RELOAD=true
VITE_ENABLE_LAZY_LOADING=false
VITE_ENABLE_COMPONENT_CACHING=true
VITE_ENABLE_PERFORMANCE_MONITORING=true

# Optimization Settings
VITE_ENABLE_BUNDLE_SPLITTING=false
VITE_ENABLE_TREESHAKING=false
VITE_ENABLE_MINIFICATION=false
VITE_ENABLE_COMPRESSION=false

# Production Override (uncomment for production)
# VITE_API_BASE_URL=https://api.your-domain.com
# VITE_ENABLE_LAZY_LOADING=true
# VITE_ENABLE_BUNDLE_SPLITTING=true
# VITE_ENABLE_TREESHAKING=true
# VITE_ENABLE_MINIFICATION=true
# VITE_ENABLE_COMPRESSION=true
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
| `HASHCAT_FRONTEND_URL` | Frontend URL for CORS | http://localhost:3000 | http://30.30.30.39:3000 |
| `GIN_MODE` | Gin framework mode | debug | debug/release |

### Frontend Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `VITE_API_BASE_URL` | Backend API URL | http://localhost:1337 | http://30.30.30.39:1337 |
| `VITE_DEV_PORT` | Development server port | 3000 | 3000 |
| `VITE_DEV_HOST` | Allow external connections | true | true |
| `VITE_ENABLE_HOT_RELOAD` | Enable hot reload | true | true |
| `VITE_ENABLE_LAZY_LOADING` | Enable lazy loading | false | false |
| `VITE_ENABLE_COMPONENT_CACHING` | Enable component caching | true | true |
| `VITE_ENABLE_PERFORMANCE_MONITORING` | Enable performance monitoring | true | true |

### Agent Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `HASHCAT_SERVER_URL` | Server URL for agent connection | http://localhost:1337 | http://30.30.30.39:1337 |

## Running the System

### 1. Start Backend Server

**Option A: Using the development script (Recommended)**
```bash
bash scripts/dev-server.sh
```

**Option B: Manual start**
```bash
# Load environment from .env file
export $(cat .env | grep -v "^#" | xargs)
export GIN_MODE=debug
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

4. **Frontend shows "Loading components..." indefinitely**
   - Check `HASHCAT_FRONTEND_URL` in backend `.env` matches frontend URL
   - Verify CORS headers: `curl -H "Origin: http://30.30.30.39:3000" http://30.30.30.39:1337/api/v1/agents/`
   - Restart both backend and frontend servers

5. **Script dev-server.sh doesn't load environment variables**
   - Ensure `.env` file exists in project root
   - Check file permissions: `ls -la .env`
   - Use manual start method as alternative

### Debug Commands

```bash
# Check environment variables
env | grep HASHCAT
env | grep VITE

# Verify configuration files exist
ls -la .env frontend/.env

# Test backend connectivity
curl http://30.30.30.39:1337/health

# Test CORS configuration
curl -H "Origin: http://30.30.30.39:3000" http://30.30.30.39:1337/api/v1/agents/

# Check if servers are running
netstat -tlnp | grep -E "(1337|3000)"
lsof -i :1337
lsof -i :3000

# Test API response
curl -s http://30.30.30.39:1337/api/v1/agents/ | python3 -c "import json, sys; data=json.load(sys.stdin); print('Agents:', len(data.get('data', [])))"
```

## Migration from Hardcoded URLs

If you're upgrading from a version with hardcoded URLs:

1. Copy the appropriate `.env.example` files to `.env`
2. Update the IP addresses in the `.env` files
3. Restart all services
4. Verify connectivity

The system will automatically use environment variables when available, falling back to sensible defaults when not configured.

## Recent Changes

### v1.1.0 - Environment Variables Implementation

- ✅ **Removed hardcoded URLs** from `router.go` and `websocket_handler.go`
- ✅ **Added CORS configuration** via `HASHCAT_FRONTEND_URL` environment variable
- ✅ **Updated dev-server.sh** to automatically load `.env` file
- ✅ **Enhanced frontend configuration** with feature flags and optimization settings
- ✅ **Improved error handling** and debugging capabilities

### Key Benefits

- 🔧 **Flexible configuration** - Easy to change URLs without code changes
- 🚀 **Environment-specific settings** - Different configs for dev/staging/production
- 🔒 **Security** - `.env` files are gitignored
- 🐛 **Better debugging** - Comprehensive troubleshooting guide
- 📚 **Complete documentation** - Step-by-step setup instructions
