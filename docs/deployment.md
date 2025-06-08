# 🚀 Production Deployment Guide

## Overview
This guide covers deploying Go Distributed Hashcat in production environments with proper database configuration, environment variables, and security settings.

## 📋 Prerequisites

### System Requirements
- **OS:** Linux (Ubuntu 20.04+ recommended) / CentOS 8+ / RHEL 8+
- **CPU:** Multi-core CPU (8+ cores recommended)
- **RAM:** 16GB+ recommended 
- **GPU:** NVIDIA GPU with CUDA support (optional but recommended)
- **Storage:** SSD storage for database and file uploads
- **Network:** Stable network connection for distributed agents

### Software Dependencies
- **Go:** 1.24+
- **Node.js:** 18+ (for frontend build)
- **Database:** PostgreSQL 13+ or MySQL 8+ (SQLite for small deployments)
- **Hashcat:** Latest version installed on agent machines
- **Reverse Proxy:** Nginx or Apache (recommended)

## 🗄️ Database Setup

### Option 1: SQLite (Small Deployments)
```bash
# Default - no additional setup required
# Database will be created automatically
```

### Option 2: PostgreSQL (Recommended for Production)
```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql
CREATE DATABASE hashcat_distributed;
CREATE USER hashcat WITH ENCRYPTED PASSWORD 'secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE hashcat_distributed TO hashcat;
\q

# Configure .env
DB_TYPE=postgres
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=hashcat
DB_POSTGRES_PASSWORD=secure_password_here
DB_POSTGRES_DB=hashcat_distributed
DB_POSTGRES_SSL=require
```

### Option 3: MySQL (Alternative Production Option)
```bash
# Install MySQL
sudo apt update
sudo apt install mysql-server

# Create database and user
sudo mysql
CREATE DATABASE hashcat_distributed;
CREATE USER 'hashcat'@'localhost' IDENTIFIED BY 'secure_password_here';
GRANT ALL PRIVILEGES ON hashcat_distributed.* TO 'hashcat'@'localhost';
FLUSH PRIVILEGES;
EXIT;

# Configure .env
DB_TYPE=mysql
DB_MYSQL_HOST=localhost
DB_MYSQL_PORT=3306
DB_MYSQL_USER=hashcat
DB_MYSQL_PASSWORD=secure_password_here
DB_MYSQL_DB=hashcat_distributed
```

## ⚙️ Environment Configuration

### 1. Copy Environment Template
```bash
cp .env-example .env
```

### 2. Edit Production Settings
```bash
nano .env
```

### 3. Key Production Settings
```bash
# Server
SERVER_PORT=1337
SERVER_HOST=0.0.0.0
GIN_MODE=release

# Security
CORS_ORIGINS=https://your-domain.com
RATE_LIMIT_RPM=100

# Database (choose one)
DB_TYPE=postgres  # or mysql, sqlite

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🔧 Build and Deployment

### 1. Build Backend
```bash
# Build server binary
make build-server

# Build agent binary
make build-agent

# Or build all
make build
```

### 2. Build Frontend
```bash
cd frontend
npm install
npm run build:prod
cd ..
```

### 3. Deploy Files
```bash
# Create production directories
sudo mkdir -p /opt/hashcat-distributed
sudo mkdir -p /var/lib/hashcat/{data,uploads}
sudo mkdir -p /var/log/hashcat

# Copy binaries
sudo cp bin/server /opt/hashcat-distributed/
sudo cp bin/agent /opt/hashcat-distributed/
sudo cp -r frontend/dist /opt/hashcat-distributed/

# Copy config
sudo cp .env /opt/hashcat-distributed/
sudo cp configs/config.production.yaml /opt/hashcat-distributed/config.yaml

# Set permissions
sudo chown -R hashcat:hashcat /opt/hashcat-distributed
sudo chown -R hashcat:hashcat /var/lib/hashcat
sudo chown -R hashcat:hashcat /var/log/hashcat
```

## 🐳 Docker Deployment

### 1. Create Dockerfile
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN make build-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/server .
COPY --from=builder /app/frontend/dist ./frontend/dist
COPY --from=builder /app/configs ./configs
CMD ["./server"]
```

### 2. Docker Compose
```yaml
version: '3.8'
services:
  hashcat-server:
    build: .
    ports:
      - "1337:1337"
    environment:
      - DB_TYPE=postgres
      - DB_POSTGRES_HOST=db
    depends_on:
      - db
    volumes:
      - ./uploads:/uploads
      - ./data:/data

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: hashcat_distributed
      POSTGRES_USER: hashcat
      POSTGRES_PASSWORD: secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 🔒 Security Configuration

### 1. Firewall Setup
```bash
# Allow only necessary ports
sudo ufw allow 22/tcp     # SSH
sudo ufw allow 1337/tcp   # Application
sudo ufw allow 80/tcp     # HTTP (for redirect)
sudo ufw allow 443/tcp    # HTTPS
sudo ufw enable
```

### 2. SSL/TLS Setup with Let's Encrypt
```bash
# Install Certbot
sudo apt install certbot

# Get certificate
sudo certbot certonly --standalone -d your-domain.com

# Configure .env for TLS
ENABLE_TLS=true
TLS_CERT_FILE=/etc/letsencrypt/live/your-domain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/your-domain.com/privkey.pem
```

### 3. Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:1337;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🚀 Systemd Service

### 1. Create Service File
```bash
sudo nano /etc/systemd/system/hashcat-server.service
```

```ini
[Unit]
Description=Hashcat Distributed Server
After=network.target

[Service]
Type=simple
User=hashcat
Group=hashcat
WorkingDirectory=/opt/hashcat-distributed
ExecStart=/opt/hashcat-distributed/server
Restart=always
RestartSec=5
Environment=PATH=/usr/local/bin:/usr/bin:/bin
EnvironmentFile=/opt/hashcat-distributed/.env

[Install]
WantedBy=multi-user.target
```

### 2. Enable and Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable hashcat-server
sudo systemctl start hashcat-server
sudo systemctl status hashcat-server
```

## 📊 Monitoring and Logs

### 1. Log Management
```bash
# View logs
sudo journalctl -u hashcat-server -f

# Rotate logs
sudo logrotate -f /etc/logrotate.d/hashcat
```

### 2. Metrics Endpoint
```bash
# Prometheus metrics available at
curl http://localhost:9090/metrics
```

### 3. Health Check
```bash
# Application health
curl http://localhost:1337/health
```

## 🔄 Database Migrations

### Run Migrations
```bash
# Migrations run automatically on startup
# Manual migration (if needed):
./server --migrate-only
```

### Backup Database
```bash
# PostgreSQL
pg_dump -h localhost -U hashcat hashcat_distributed > backup.sql

# MySQL
mysqldump -u hashcat -p hashcat_distributed > backup.sql

# SQLite
cp /var/lib/hashcat/data/hashcat.db backup.db
```

## 🛠️ Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database status
   sudo systemctl status postgresql
   # Check connection
   psql -h localhost -U hashcat -d hashcat_distributed
   ```

2. **Permission Denied**
   ```bash
   # Fix file permissions
   sudo chown -R hashcat:hashcat /opt/hashcat-distributed
   sudo chmod +x /opt/hashcat-distributed/server
   ```

3. **Port Already in Use**
   ```bash
   # Check what's using port 1337
   sudo netstat -tulpn | grep 1337
   # Kill process if needed
   sudo kill -9 <PID>
   ```

## 📈 Performance Tuning

### Database Optimization
```bash
# PostgreSQL
# Edit /etc/postgresql/*/main/postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
```

### System Limits
```bash
# Edit /etc/security/limits.conf
hashcat soft nofile 65535
hashcat hard nofile 65535
```

## 🎯 Production Checklist

- [ ] Database properly configured and secured
- [ ] Environment variables set correctly
- [ ] SSL/TLS certificates installed
- [ ] Firewall configured
- [ ] Systemd service enabled
- [ ] Backup strategy implemented
- [ ] Monitoring setup
- [ ] Log rotation configured
- [ ] Performance tuning applied
- [ ] Security hardening complete

---

For additional support, check the main [README.md](../README.md) or open an issue on GitHub. 
