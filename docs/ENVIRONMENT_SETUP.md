# Environment Configuration Guide

## Overview
This project uses environment variables for configuration instead of hardcoded values. This makes it easy to deploy to different environments without changing code.

## Backend Configuration

### File: `.env`
Create a `.env` file in the root directory with your configuration:

```bash
# Copy from template
cp .env.example .env

# Edit the values
nano .env
```

### Key Variables:
- `HASHCAT_SERVER_HOST`: Server IP address (default: localhost)
- `HASHCAT_SERVER_PORT`: Server port (default: 1337)
- `HASHCAT_DATABASE_PATH`: Database file path
- `HASHCAT_UPLOAD_DIRECTORY`: Upload directory path

### Examples:
```bash
# Local development
HASHCAT_SERVER_HOST=localhost
HASHCAT_SERVER_PORT=1337

# Production server
HASHCAT_SERVER_HOST=30.30.30.39
HASHCAT_SERVER_PORT=1337

# Custom configuration
HASHCAT_SERVER_HOST=192.168.1.100
HASHCAT_SERVER_PORT=8080
```

## Frontend Configuration

### File: `frontend/.env`
Create a `.env` file in the frontend directory:

```bash
# Copy from template
cp frontend/.env.example frontend/.env

# Edit the values
nano frontend/.env
```

### Key Variables:
- `VITE_API_BASE_URL`: Backend API URL

### Examples:
```bash
# Local development
VITE_API_BASE_URL=http://localhost:1337

# Production server
VITE_API_BASE_URL=http://192.118.30.2:1337

# Custom configuration
VITE_API_BASE_URL=http://192.168.1.100:8080
```

## How It Works

### Backend
- Uses `viper` library to load environment variables
- Falls back to YAML config if `.env` not found
- Environment variables override YAML values

### Frontend
- Uses Vite's environment variable system
- Variables prefixed with `VITE_` are available in browser
- Falls back to default values if not set

## Quick Setup

1. **Backend:**
   ```bash
   cp .env.example .env
   # Edit .env with your server IP
   ```

2. **Frontend:**
   ```bash
   cp frontend/.env.example frontend/.env
   # Edit frontend/.env with your server IP
   ```

3. **Run:**
   ```bash
   # Backend
   go run cmd/server/main.go
   
   # Frontend (in another terminal)
   cd frontend
   npm run dev
   ```

## Security Notes
- `.env` files are gitignored for security
- Never commit `.env` files to version control
- Use `.env.example` as templates for documentation
