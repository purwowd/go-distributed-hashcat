# Quick Start - WireGuard Deployment

Panduan cepat untuk setup distributed hashcat system dengan WireGuard VPN dalam 2 nodes.

## 🚀 TL;DR Setup

### **Prerequisites**
- Node 1: Mini PC dengan GPU + Public IP
- Node 2: DigitalOcean GPU server
- Ubuntu 22.04 LTS pada kedua nodes

## ⚡ Rapid Deployment

### **1. Setup WireGuard (5 menit)**

**Node 1 (Server):**
```bash
# Install WireGuard
sudo apt update && sudo apt install wireguard -y

# Generate keys
cd /etc/wireguard
sudo wg genkey | sudo tee node1-private | sudo wg pubkey | sudo tee node1-public

# Create config
sudo tee /etc/wireguard/wg0.conf << 'EOF'
[Interface]
PrivateKey = YOUR_NODE1_PRIVATE_KEY
Address = 15.15.15.1/24
ListenPort = 51820
PostUp = echo 1 > /proc/sys/net/ipv4/ip_forward; iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

[Peer]
PublicKey = YOUR_NODE2_PUBLIC_KEY
AllowedIPs = 15.15.15.2/32
PersistentKeepalive = 25
EOF

# Start VPN
sudo ufw allow 51820/udp
sudo systemctl enable wg-quick@wg0 && sudo systemctl start wg-quick@wg0
```

**Node 2 (Client):**
```bash
# Install WireGuard
sudo apt update && sudo apt install wireguard -y

# Generate keys
cd /etc/wireguard
sudo wg genkey | sudo tee node2-private | sudo wg pubkey | sudo tee node2-public

# Create config (replace NODE1_PUBLIC_IP)
sudo tee /etc/wireguard/wg0.conf << 'EOF'
[Interface]
PrivateKey = YOUR_NODE2_PRIVATE_KEY
Address = 15.15.15.2/24

[Peer]
PublicKey = YOUR_NODE1_PUBLIC_KEY
Endpoint = NODE1_PUBLIC_IP:51820
AllowedIPs = 15.15.15.0/24
PersistentKeepalive = 25
EOF

# Start VPN
sudo systemctl enable wg-quick@wg0 && sudo systemctl start wg-quick@wg0

# Test
ping 15.15.15.1
```

### **2. Deploy Application (10 menit)**

**Node 1 (Full Stack):**
```bash
# Install dependencies
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install nodejs hashcat git -y
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc

# Clone and build
git clone https://github.com/your-repo/go-distributed-hashcat.git
cd go-distributed-hashcat
make build
cd frontend && npm install && npm run build && cd ..

# Quick start script
tee start-quick.sh << 'EOF'
#!/bin/bash
cd /home/user/go-distributed-hashcat
./bin/server &
sleep 3
./bin/agent --server http://15.15.15.1:1337 --name node1-gpu &
cd frontend && npm run preview -- --host 15.15.15.1 --port 3000 &
echo "🎉 Node 1 ready: http://15.15.15.1:3000"
EOF

chmod +x start-quick.sh && ./start-quick.sh
```

**Node 2 (Worker Only):**
```bash
# Install dependencies
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
sudo apt install hashcat git -y
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc

# Deploy agent
git clone https://github.com/your-repo/go-distributed-hashcat.git
cd go-distributed-hashcat
go build -o bin/agent cmd/agent/main.go

# Start agent
./bin/agent --server http://15.15.15.1:1337 --name node2-do-gpu --gpu-info Tesla-T4 --max-jobs 4
```

### **3. Verification (1 menit)**

```bash
# Check VPN
sudo wg show

# Check agents (dari Node 1)
curl http://15.15.15.1:1337/api/v1/agents/

# Access dashboard
open http://15.15.15.1:3000
```

## 🔧 Essential Commands

```bash
# VPN Control
sudo systemctl {start|stop|restart} wg-quick@wg0
sudo wg show

# Service Control (Node 1)
pkill -f "bin/server|bin/agent|npm"  # Stop all
./start-quick.sh                     # Start all

# Monitor (Node 2)
journalctl -f | grep agent
nvidia-smi
```

## 🐛 Common Issues & Fixes

| Problem | Solution |
|---------|----------|
| VPN won't connect | Check firewall: `sudo ufw allow 51820/udp` |
| Agent can't register | Test: `ping 15.15.15.1` dari Node 2 |
| No GPU detected | Install drivers: `sudo ubuntu-drivers autoinstall` |
| High latency | Check MTU: Add `MTU = 1420` di WireGuard config |

## 📊 Network Overview

```
Node 1 (15.15.15.1) ←→ Node 2 (15.15.15.2)
     ↓                        ↓
  Full Stack              Worker Only
- Backend :1337         - Agent only
- Frontend :3000        - GPU processing
- Local Agent           - File cache
- Database & Storage    
```

## 🎯 Performance Tips

```bash
# GPU optimization
sudo nvidia-smi -pm 1        # Persistence mode
sudo nvidia-smi -pl 300      # Power limit

# Network optimization (both nodes)
echo 'net.core.default_qdisc=fq' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_congestion_control=bbr' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

Total setup time: **~15 minutes** untuk basic working setup! 🚀

Untuk setup production dan monitoring yang advanced, lihat [dokumentasi lengkap](wireguard-deployment.md). 
