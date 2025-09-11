# Systemd Backend Service Documentation

## Overview

This document provides comprehensive information about the systemd service configuration for the Go Distributed Hashcat backend service.

## Service Information

- **Service Name:** `hashcat-backend.service`
- **Description:** Hashcat Backend Service
- **Service File Location:** `/etc/systemd/system/hashcat-backend.service`
- **Working Directory:** `/home/leonaldy/go-distributed-hashcat`
- **User:** `leonaldy`

## Service Configuration

### Systemd Unit File

The service is configured with the following systemd unit file:

```ini
[Unit]
Description=Hashcat Backend Service
After=network.target

[Service]
User=leonaldy
WorkingDirectory=/home/leonaldy/go-distributed-hashcat
EnvironmentFile=/home/leonaldy/go-distributed-hashcat/.env
ExecStartPre=/usr/bin/make build
ExecStart=/bin/bash /home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh
Restart=always

[Install]
WantedBy=multi-user.target
```

### Configuration Details

- **User:** Runs as `leonaldy` user
- **Working Directory:** `/home/leonaldy/go-distributed-hashcat`
- **Environment:** Loads variables from `.env` file
- **Pre-start:** Builds the application using `make build`
- **Start Command:** Executes the development server script
- **Restart Policy:** Always restart on failure
- **Target:** Multi-user target (starts after system boot)

## Service Management Commands

### Basic Service Control

```bash
# Check service status
systemctl status hashcat-backend.service

# Start the service
sudo systemctl start hashcat-backend.service

# Stop the service
sudo systemctl stop hashcat-backend.service

# Restart the service
sudo systemctl restart hashcat-backend.service

# Reload the service (NOT SUPPORTED - use restart instead)
# sudo systemctl reload hashcat-backend.service  # This will fail
```

### Service Lifecycle Management

```bash
# Enable service to start on boot
sudo systemctl enable hashcat-backend.service

# Disable service from starting on boot
sudo systemctl disable hashcat-backend.service

# Check if service is enabled
systemctl is-enabled hashcat-backend.service

# Check if service is active
systemctl is-active hashcat-backend.service
```

### Service Monitoring

```bash
# View real-time logs
journalctl -u hashcat-backend.service -f

# View logs with timestamps
journalctl -u hashcat-backend.service -t

# View recent logs (last 100 lines)
journalctl -u hashcat-backend.service -n 100

# View logs from specific time
journalctl -u hashcat-backend.service --since "2025-01-01 00:00:00"

# View logs with priority level
journalctl -u hashcat-backend.service -p err
```

## Development Server Script

The service uses the development server script located at `/home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh`:

### Script Features

- **Environment Loading:** Automatically loads `.env` file if present
- **Development Mode:** Sets `GIN_MODE=debug` for detailed logging
- **Port Configuration:** Uses port 1337 by default
- **Directory Creation:** Creates necessary `data` and `uploads` directories
- **Go Execution:** Runs the server using `go run cmd/server/main.go`

### Script Content

```bash
#!/bin/bash

# Development server script
echo "🚀 Starting Hashcat Distributed Server in Development Mode..."

# Change to project root directory
cd "$(dirname "$0")/.."

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "📋 Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  No .env file found, using default values"
fi

# Set development environment (override .env if needed)
export GIN_MODE=debug
export SERVER_PORT=1337

# Create necessary directories
mkdir -p data
mkdir -p uploads

# Run the server
go run cmd/server/main.go 
```

## Service Status Information

### Current Status

- **State:** Active (Running)
- **Start Time:** Tue 2025-09-09 17:21:32 WIB
- **Uptime:** 1 day 15h (as of last check)
- **Main PID:** 1905676
- **Memory Usage:** 24.6M
- **CPU Time:** 38.908s

### Process Tree

```
hashcat-backend.service
├── /bin/bash /home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh
├── go run cmd/server/main.go
└── Go build cache process
```

## Troubleshooting

### Common Issues

1. **Reload Command Fails**
   ```bash
   # This error is expected and normal:
   # Failed to reload hashcat-backend.service: Job type reload is not applicable for unit hashcat-backend.service
   
   # Solution: Use restart instead of reload
   sudo systemctl restart hashcat-backend.service
   ```

2. **Service Won't Start**
   ```bash
   # Check service status for errors
   systemctl status hashcat-backend.service
   
   # Check logs for detailed error messages
   journalctl -u hashcat-backend.service -n 50
   ```

2. **Build Failures**
   ```bash
   # Check if Go is installed
   go version
   
   # Check if Make is available
   make --version
   
   # Try building manually
   cd /home/leonaldy/go-distributed-hashcat
   make build
   ```

3. **Permission Issues**
   ```bash
   # Check file permissions
   ls -la /home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh
   
   # Make script executable if needed
   chmod +x /home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh
   ```

4. **Environment Issues**
   ```bash
   # Check if .env file exists and is readable
   ls -la /home/leonaldy/go-distributed-hashcat/.env
   
   # Check environment variables
   systemctl show hashcat-backend.service --property=Environment
   ```

### Log Analysis

```bash
# Filter logs by level
journalctl -u hashcat-backend.service -p warning
journalctl -u hashcat-backend.service -p error

# Search for specific patterns
journalctl -u hashcat-backend.service | grep "ERROR"
journalctl -u hashcat-backend.service | grep "panic"

# Monitor logs in real-time
journalctl -u hashcat-backend.service -f | grep -E "(ERROR|WARN|panic)"
```

## Service Dependencies

### Required Dependencies

- **Go Runtime:** Go 1.19+ installed and in PATH
- **Make:** GNU Make for building the application
- **Bash:** For executing the startup script
- **Environment File:** `.env` file with configuration (optional)

### Network Dependencies

- **Port 1337:** Backend API server port
- **Database:** SQLite database file in `data/` directory
- **File System:** Read/write access to project directory

## Performance Monitoring

### Resource Usage

```bash
# Monitor CPU and memory usage
systemctl status hashcat-backend.service

# Detailed resource monitoring
systemd-cgtop

# Check service resource limits
systemctl show hashcat-backend.service --property=MemoryLimit,CPUQuota
```

### Health Checks

```bash
# Check if service is responding
curl -f http://localhost:1337/health || echo "Service not responding"

# Check service health endpoint
curl -s http://localhost:1337/api/health | jq .

# Monitor service uptime
systemctl show hashcat-backend.service --property=ActiveEnterTimestamp
```

## Security Considerations

### File Permissions

```bash
# Ensure proper permissions on service file
sudo chmod 644 /etc/systemd/system/hashcat-backend.service

# Ensure script is executable but not writable by others
chmod 755 /home/leonaldy/go-distributed-hashcat/scripts/dev-server.sh

# Secure .env file
chmod 600 /home/leonaldy/go-distributed-hashcat/.env
```

### User Isolation

- Service runs as `leonaldy` user (non-root)
- Limited to project directory access
- Environment variables loaded from secure `.env` file

## Backup and Recovery

### Service Configuration Backup

```bash
# Backup service file
sudo cp /etc/systemd/system/hashcat-backend.service /home/leonaldy/backups/

# Backup environment file
cp /home/leonaldy/go-distributed-hashcat/.env /home/leonaldy/backups/
```

### Service Recovery

```bash
# Restore service file
sudo cp /home/leonaldy/backups/hashcat-backend.service /etc/systemd/system/

# Reload systemd configuration
sudo systemctl daemon-reload

# Restart service
sudo systemctl restart hashcat-backend.service
```

## Maintenance

### Regular Maintenance Tasks

1. **Log Rotation:** Monitor log size and implement rotation if needed
2. **Resource Monitoring:** Check CPU and memory usage regularly
3. **Update Management:** Update Go dependencies and rebuild when needed
4. **Health Checks:** Regular health endpoint monitoring

### Update Procedure

```bash
# Stop service
sudo systemctl stop hashcat-backend.service

# Update code (git pull, etc.)
cd /home/leonaldy/go-distributed-hashcat
git pull origin main

# Rebuild application
make build

# Start service
sudo systemctl start hashcat-backend.service

# Verify service is running
systemctl status hashcat-backend.service
```

## Related Documentation

- [Quick Start Guide](01-quick-start.md)
- [Deployment Guide](02-deployment.md)
- [API Reference](03-api-reference.md)
- [Architecture Overview](04-architecture.md)
- [Environment Configuration](ENVIRONMENT_CONFIG.md)
