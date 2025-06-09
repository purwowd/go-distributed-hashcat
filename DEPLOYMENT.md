# 🚀 Deployment Guide - Go Distributed Hashcat

## 📋 Overview

This guide covers deploying the improved Go Distributed Hashcat system with all the latest enhancements including better error handling, job control, and production optimizations.

## 🎯 New Features Added

### ✅ **Backend Improvements**
- ✅ Added missing `StopJob` endpoint (`POST /api/v1/jobs/:id/stop`)
- ✅ Complete job control: Start, Stop, Pause, Resume
- ✅ Enhanced error handling and response consistency
- ✅ Improved CORS configuration for frontend integration

### ✅ **Frontend Improvements** 
- ✅ Complete job control UI with Pause/Resume buttons
- ✅ Enhanced type safety (removed `any` types where possible)
- ✅ Better state management with dedicated store getters
- ✅ Improved environment variable handling
- ✅ Production-ready build configuration

### ✅ **Development Tools**
- ✅ Automated build scripts with validation
- ✅ Simple deployment script
- ✅ Enhanced Vite configuration with proxy support

## 🏗️ Quick Start

### 1. **Development Setup**

```bash
# Clone and setup
git clone <your-repo>
cd go-distributed-hashcat

# Backend setup
go mod download
cp .env-example .env
# Edit .env with your settings

# Frontend setup
cd frontend
npm install
cd ..
```

### 2. **Run Development**

```bash
# Terminal 1: Start backend
go run cmd/server/main.go

# Terminal 2: Start frontend (in frontend/)
cd frontend
npm run dev
```

Frontend will be available at `http://localhost:3000` with API proxy to backend.

### 3. **Production Build**

```bash
# Build frontend with validation
./scripts/build-frontend.sh

# Or manual build
cd frontend
npm run build:prod
cd ..

# Build backend
go build -o bin/server cmd/server/main.go
```

### 4. **Simple Deployment**

```bash
# Create deployment package
./scripts/deploy.sh

# This creates ./deploy/ with:
# - Backend binary
# - Frontend static files  
# - Configuration files
```

## 🔧 Configuration

### **Environment Variables**

Create `.env` file for backend:

```bash
# Server Configuration
GIN_MODE=release
PORT=1337
HOST=0.0.0.0

# Database
DATABASE_URL=./data/hashcat.db

# File Storage
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE=100MB

# CORS Settings (for frontend)
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
```

### **Frontend Environment**

Create `frontend/.env.production`:

```bash
# API Configuration
VITE_API_BASE_URL=https://api.yourdomain.com

# Or for same-origin deployment
VITE_API_BASE_URL=/api
```

## 📦 API Endpoints Reference

### **Job Control Endpoints**
All job control endpoints are now available:

```bash
# Start job
POST /api/v1/jobs/:id/start

# Stop job (NEW - added in improvements)
POST /api/v1/jobs/:id/stop

# Pause job  
POST /api/v1/jobs/:id/pause

# Resume job
POST /api/v1/jobs/:id/resume

# Update progress
PUT /api/v1/jobs/:id/progress
{
  "progress": 75.5,
  "speed": 12500,
  "eta": "2023-12-25T10:30:00Z"
}
```

### **File Management**
```bash
# Upload hash file
POST /api/v1/hashfiles/upload
Content-Type: multipart/form-data

# Download hash file
GET /api/v1/hashfiles/:id/download

# Upload wordlist
POST /api/v1/wordlists/upload
Content-Type: multipart/form-data

# Download wordlist  
GET /api/v1/wordlists/:id/download
```

## 🌐 Production Deployment

### **Option 1: Single Server with Nginx**

1. **Deploy files**:
```bash
# Copy deployment package to server
scp -r deploy/ user@server:/opt/hashcat/
```

2. **Nginx configuration**:
```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # Serve frontend static files
    location / {
        root /opt/hashcat/frontend;
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to backend
    location /api/ {
        proxy_pass http://localhost:1337;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:1337;
    }
}
```

3. **Run backend**:
```bash
cd /opt/hashcat
./server
```

### **Option 2: Docker Deployment**

Create `Dockerfile` in project root:

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go

FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build:prod

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/dist ./frontend/dist
COPY configs/ ./configs/
EXPOSE 1337
CMD ["./server"]
```

Build and run:
```bash
docker build -t hashcat-dashboard .
docker run -p 1337:1337 -v $(pwd)/data:/app/data hashcat-dashboard
```

## 🔍 Testing the Improvements

### **Job Control Testing**
1. Create a job via frontend
2. Test all job control buttons:
   - ▶️ Start Job
   - ⏸️ Pause Job  
   - ▶️ Resume Job
   - ⏹️ Stop Job

### **File Upload Testing**
1. Upload hash files and wordlists
2. Verify download functionality
3. Test file deletion

### **Error Handling Testing**
1. Disconnect backend and verify frontend shows appropriate errors
2. Upload invalid files and check error messages
3. Try invalid API requests

## 🐛 Troubleshooting

### **Common Issues**

**Frontend shows "API connection failed"**
- Check backend is running on correct port
- Verify CORS settings in backend
- Check VITE_API_BASE_URL environment variable

**Job controls not working**
- Verify all job endpoints are available
- Check browser console for JavaScript errors
- Ensure job store is properly connected

**File uploads failing**
- Check upload directory permissions
- Verify MAX_UPLOAD_SIZE setting
- Check available disk space

### **Log Locations**
- Backend logs: stdout/stderr
- Frontend errors: Browser console
- Nginx logs: `/var/log/nginx/`

## 📊 Performance Considerations

### **Backend Optimizations**
- Gzip compression enabled
- Request timeouts configured
- Connection pooling for database
- Efficient file serving

### **Frontend Optimizations**
- Code splitting and lazy loading
- Asset optimization and compression
- Efficient API request patterns
- Production build minification

## 🔄 Updating

To update the system:

1. **Update code**:
```bash
git pull origin main
```

2. **Rebuild**:
```bash
./scripts/deploy.sh
```

3. **Deploy new version**:
```bash
# Stop current services
# Copy new deployment package
# Restart services
```

## 📝 API Integration Examples

### **JavaScript/TypeScript Client**
```typescript
// Using the improved API service
import { apiService } from './services/api.service'

// Job control
await apiService.startJob(jobId)
await apiService.pauseJob(jobId)
await apiService.resumeJob(jobId)
await apiService.stopJob(jobId)

// File operations
const hashFile = await apiService.uploadHashFile(file)
const blob = await apiService.downloadHashFile(fileId)
```

### **curl Examples**
```bash
# Start job
curl -X POST http://localhost:1337/api/v1/jobs/$JOB_ID/start

# Stop job (NEW endpoint)
curl -X POST http://localhost:1337/api/v1/jobs/$JOB_ID/stop

# Upload file
curl -X POST http://localhost:1337/api/v1/hashfiles/upload \
  -F "file=@hashfile.txt"

# Check health
curl http://localhost:1337/health
```

## ✅ Improvement Summary

This deployment includes all the improvements requested:

1. ✅ **Missing StopJob endpoint** - Added backend handler and route
2. ✅ **Better error handling** - Comprehensive error responses  
3. ✅ **Type safety** - Removed `any` types, proper interfaces
4. ✅ **Environment config** - Smart API URL detection
5. ✅ **Production build** - Optimized build scripts and configuration
6. ✅ **Complete integration** - All API endpoints working with frontend

The system is now production-ready with improved reliability, better user experience, and proper error handling! 🎉 
