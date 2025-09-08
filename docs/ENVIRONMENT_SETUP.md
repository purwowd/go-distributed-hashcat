# Environment Configuration Guide

## Overview
This project follows standard environment configuration practices with proper separation of concerns between development, staging, and production environments.

## File Structure
```
├── .env                    # Backend environment (gitignored)
├── .env.example           # Backend template
├── frontend/
│   ├── .env              # Frontend environment (gitignored)
│   └── .env.example      # Frontend template
└── ENVIRONMENT_SETUP.md  # This documentation
```

## Backend Configuration

### Environment File: `.env`
```bash
# Server Configuration
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

### Setup Commands:
```bash
# Copy template
cp .env.example .env

# Edit configuration
nano .env
```

## Frontend Configuration

### Environment File: `frontend/.env`
```bash
# Frontend Environment Variables
VITE_API_BASE_URL=http://30.30.30.39:1337
```

### Setup Commands:
```bash
# Copy template
cp frontend/.env.example frontend/.env

# Edit configuration
nano frontend/.env
```

## Environment Variables Naming Convention

### Backend (Go/Viper)
- Prefix: `HASHCAT_`
- Format: `HASHCAT_SECTION_VARIABLE`
- Examples:
  - `HASHCAT_SERVER_PORT`
  - `HASHCAT_DATABASE_PATH`
  - `HASHCAT_UPLOAD_DIRECTORY`

### Frontend (Vite)
- Prefix: `VITE_`
- Format: `VITE_VARIABLE_NAME`
- Examples:
  - `VITE_API_BASE_URL`
  - `VITE_APP_TITLE`
  - `VITE_DEBUG_MODE`

## Environment-Specific Configurations

### Development
```bash
# Backend
HASHCAT_SERVER_HOST=localhost
HASHCAT_SERVER_PORT=1337

# Frontend
VITE_API_BASE_URL=http://localhost:1337
```

### Production
```bash
# Backend
HASHCAT_SERVER_HOST=30.30.30.39
HASHCAT_SERVER_PORT=1337

# Frontend
VITE_API_BASE_URL=http://30.30.30.39:1337
```

### Staging
```bash
# Backend
HASHCAT_SERVER_HOST=staging-server.com
HASHCAT_SERVER_PORT=1337

# Frontend
VITE_API_BASE_URL=http://staging-server.com:1337
```

## Security Best Practices

### ✅ Implemented
- `.env` files are gitignored
- Template files (`.env.example`) are committed
- Sensitive data is not hardcoded
- Environment variables have proper prefixes
- Documentation exists for setup

### 🔒 Security Notes
- Never commit `.env` files to version control
- Use strong passwords in production
- Rotate secrets regularly
- Use different configurations per environment
- Validate environment variables on startup

## Quick Setup Guide

### 1. Initial Setup
```bash
# Backend
cp .env.example .env
nano .env  # Edit with your values

# Frontend
cp frontend/.env.example frontend/.env
nano frontend/.env  # Edit with your values
```

### 2. Run Application
```bash
# Backend
go run cmd/server/main.go

# Frontend (separate terminal)
cd frontend
npm run dev
```

### 3. Verify Configuration
```bash
# Check backend config
curl http://30.30.30.39:1337/health

# Check frontend config
# Open browser to http://localhost:3000
```

## Troubleshooting

### Common Issues
1. **Port already in use**: Change `HASHCAT_SERVER_PORT`
2. **CORS errors**: Ensure `VITE_API_BASE_URL` matches backend
3. **File not found**: Check if `.env` files exist
4. **Permission denied**: Check file permissions on `.env` files

### Debug Commands
```bash
# Check environment variables
env | grep HASHCAT
env | grep VITE

# Verify file existence
ls -la .env frontend/.env

# Check gitignore
grep -n "\.env" .gitignore
```

## Migration Guide

### From Hardcoded to Environment Variables
1. Identify hardcoded values in code
2. Create environment variables with proper naming
3. Update code to read from environment
4. Create template files
5. Update documentation
6. Test in all environments

## Compliance Checklist

- [x] Environment files are gitignored
- [x] Template files exist and are committed
- [x] Proper naming conventions are used
- [x] Documentation is complete
- [x] Security best practices are followed
- [x] Multiple environment support
- [x] Easy setup process
- [x] Troubleshooting guide available
