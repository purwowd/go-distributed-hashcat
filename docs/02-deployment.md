# Production Deployment Guide

Deploy distributed hashcat system to production with security and performance.

## 🚀 Overview

Distributed password cracking system with:
- **Clean Architecture**: Go backend + TypeScript frontend
- **Real-time Dashboard**: Live monitoring
- **Horizontal Scaling**: Multiple GPU agents
- **Security**: WireGuard VPN support

## 🏗️ Deployment Options

### **Single Server**
```bash
# Build and deploy
make build
cd frontend && npm run build
./bin/server --host 0.0.0.0 --port 1337 &

# Serve frontend
sudo cp frontend/dist/* /var/www/html/
```

### **Multi-Server**
```bash
# Control server
./bin/server --host 15.15.15.1

# GPU workers
./bin/agent --server http://15.15.15.1:1337 --name gpu-worker-01
./bin/agent --server http://15.15.15.1:1337 --name gpu-worker-02
```

### **Docker**
```bash
make docker-build
docker run -p 1337:1337 -v $(pwd)/data:/app/data hashcat-server
docker run --gpus all hashcat-agent --server http://server:1337
```

## 🔧 Configuration

### **Backend Environment**
```bash
# .env
GIN_MODE=release
PORT=1337
HOST=0.0.0.0
DATABASE_URL=./data/hashcat.db
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE=100MB
```

### **Frontend Environment**
```bash
# frontend/.env.production
VITE_API_BASE_URL=https://api.yourdomain.com
```

## 🌐 Nginx Configuration

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # Frontend
    location / {
        root /opt/hashcat/frontend;
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://localhost:1337;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔐 Security Hardening

### **System Security**
```bash
# Firewall
sudo ufw allow 1337/tcp    # API
sudo ufw allow 51820/udp   # WireGuard

# User permissions
sudo useradd -r -s /bin/false hashcat
sudo chown -R hashcat:hashcat /opt/hashcat
```

### **SSL Certificate**
```bash
# Let's Encrypt
sudo certbot --nginx -d yourdomain.com
```

## 📊 Performance Optimization

### **Database**
```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=10000;
```

### **System Tuning**
```bash
# File limits
echo "hashcat soft nofile 65536" >> /etc/security/limits.conf

# GPU optimization
echo 'GRUB_CMDLINE_LINUX="iommu=soft"' >> /etc/default/grub
sudo update-grub
```

## 🔍 Monitoring

### **Health Checks**
```bash
# Server health
curl http://localhost:1337/health

# Agent status
curl http://localhost:1337/api/v1/agents/

# System resources
htop
nvidia-smi
```

### **Systemd Service**
```ini
# /etc/systemd/system/hashcat.service
[Unit]
Description=Distributed Hashcat Server
After=network.target

[Service]
Type=simple
User=hashcat
WorkingDirectory=/opt/hashcat
ExecStart=/opt/hashcat/bin/server
Restart=always

[Install]
WantedBy=multi-user.target
```

## 📈 Scaling

### **Horizontal Scaling**
- Add more GPU agents for performance
- Load balancer for multiple server instances
- Database sharding for large workloads

### **Vertical Scaling**
- **CPU**: 8+ cores optimal
- **RAM**: 32GB+ for large wordlists
- **Storage**: NVMe SSD for database
- **Network**: Gigabit minimum

## 🐛 Troubleshooting

| Issue | Solution |
|-------|----------|
| High memory usage | Check file caching, reduce batch sizes |
| Slow API responses | Enable database indexes and WAL mode |
| Agent disconnections | Check network stability and firewall |
| GPU underutilization | Optimize hashcat parameters |

## 📋 Production Checklist

- [ ] **Security**: Firewall, SSL, user permissions
- [ ] **Performance**: Database optimization, system tuning
- [ ] **Monitoring**: Health checks, logging, alerts
- [ ] **Backup**: Database backup strategy
- [ ] **Scaling**: Load balancer, multiple agents

**Next Steps**: [`06-wireguard-deployment.md`](06-wireguard-deployment.md) for secure VPN setup
