# WireGuard Deployment Guide

Complete setup for distributed hashcat system using secure WireGuard VPN.

## 🏗️ Architecture Overview

```
VPN Network (15.15.15.0/24)
├── Node 1 - Control Server (15.15.15.1)
│   ├── Backend :1337
│   ├── Frontend :3000  
│   └── SQLite Database
└── Node 2 - GPU Worker (15.15.15.2)
    ├── Remote Agent
    └── File Cache
```

## 📋 Prerequisites

### **Node Requirements**
| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **Control Server** | 2 vCPU, 4GB RAM | 4 vCPU, 8GB RAM |
| **GPU Worker** | 2 vCPU, 4GB RAM | 4 vCPU, 8GB RAM |
| **Network** | 10 Mbps | 100 Mbps |
| **OS** | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |

### **Network Information**
```bash
SERVER_PUBLIC_IP="YOUR_SERVER_PUBLIC_IP"
VPN_NETWORK="15.15.15.0/24"
SERVER_VPN_IP="15.15.15.1"
CLIENT_VPN_IP="15.15.15.2"
```

## 🔐 VPN Setup

### **1. Install WireGuard**
```bash
# On both nodes
sudo apt update && sudo apt install wireguard -y
```

### **2. Generate Keys**
```bash
# On both nodes
sudo wg genkey | sudo tee /etc/wireguard/privatekey
sudo cat /etc/wireguard/privatekey | wg pubkey | sudo tee /etc/wireguard/publickey
sudo chmod 600 /etc/wireguard/privatekey
```

### **3. Server Configuration**
```bash
# /etc/wireguard/wg0.conf on control server
sudo tee /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $(sudo cat /etc/wireguard/privatekey)
Address = 15.15.15.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

[Peer]
PublicKey = CLIENT_PUBLIC_KEY_HERE
AllowedIPs = 15.15.15.2/32
EOF
```

### **4. Client Configuration**
```bash
# /etc/wireguard/wg0.conf on GPU worker
sudo tee /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $(sudo cat /etc/wireguard/privatekey)
Address = 15.15.15.2/24

[Peer]
PublicKey = SERVER_PUBLIC_KEY_HERE
Endpoint = $SERVER_PUBLIC_IP:51820
AllowedIPs = 15.15.15.0/24
PersistentKeepalive = 25
EOF
```

### **5. Exchange Keys & Start VPN**
```bash
# Exchange public keys between nodes
# Server public key → Client config
# Client public key → Server config

# Enable IP forwarding on server
echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Start VPN on both nodes
sudo systemctl enable wg-quick@wg0
sudo systemctl start wg-quick@wg0

# Verify connection
sudo wg show
ping 15.15.15.1  # From client to server
ping 15.15.15.2  # From server to client
```

## 🚀 Application Deployment

### **Control Server (15.15.15.1)**
```bash
# Clone and build
git clone https://github.com/purwowd/go-distributed-hashcat.git
cd go-distributed-hashcat
make build

# Configure for VPN deployment
export HOST=15.15.15.1
export PORT=1337

# Start backend
./bin/server --host 15.15.15.1 --port 1337 &

# Build and serve frontend
cd frontend
npm install && npm run build
python3 -m http.server 3000 --bind 15.15.15.1 &
```

### **GPU Worker (15.15.15.2)**
```bash
# Clone and build agent
git clone https://github.com/purwowd/go-distributed-hashcat.git  
cd go-distributed-hashcat
go build -o bin/agent cmd/agent/main.go

# Start agent
./bin/agent --server http://15.15.15.1:1337 --name gpu-worker-$(hostname)
```

## ✅ Verification & Testing

### **Network Connectivity**
```bash
# Test VPN connectivity
ping 15.15.15.1
ping 15.15.15.2

# Check VPN status
sudo wg show

# Test services
curl http://15.15.15.1:1337/health
curl http://15.15.15.1:1337/api/v1/agents/

# Access web interface
open http://15.15.15.1:3000
```

### **Performance Test**
```bash
# Test latency
ping -c 10 15.15.15.1

# Test bandwidth
iperf3 -s  # On server
iperf3 -c 15.15.15.1  # On client

# Monitor VPN traffic
sudo wg show
watch -n 1 'sudo wg show'
```

## 🔧 Performance Optimization

### **VPN Optimization**
```bash
# MTU optimization
ip link set dev wg0 mtu 1420

# TCP optimization
echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf
sudo sysctl -p
```

### **Firewall Configuration**
```bash
# Server firewall
sudo ufw allow 51820/udp     # WireGuard
sudo ufw allow from 15.15.15.0/24 to any port 1337  # API
sudo ufw allow from 15.15.15.0/24 to any port 3000  # Frontend

# Client firewall  
sudo ufw allow out 51820/udp
sudo ufw allow out on wg0
```

## 🔍 Monitoring & Maintenance

### **VPN Monitoring**
```bash
# Connection status
sudo wg show all

# Traffic statistics
sudo wg show wg0 transfer

# Connection logs
journalctl -u wg-quick@wg0 -f

# Network interface status
ip addr show wg0
```

### **Troubleshooting**
| Issue | Diagnosis | Solution |
|-------|-----------|----------|
| VPN won't start | Check config syntax | `wg-quick up wg0` |
| Can't connect | Check endpoint/firewall | `telnet IP 51820` |
| Slow performance | Check MTU size | Set MTU to 1420 |
| Frequent disconnects | Check keepalive | Set PersistentKeepalive=25 |

### **Systemd Services**
```ini
# /etc/systemd/system/hashcat-server.service
[Unit]
Description=Distributed Hashcat Server
After=wg-quick@wg0.service
Requires=wg-quick@wg0.service

[Service]
Type=simple
User=hashcat
WorkingDirectory=/opt/hashcat
ExecStart=/opt/hashcat/bin/server --host 15.15.15.1
Restart=always

[Install]
WantedBy=multi-user.target
```

## 📋 Security Best Practices

### **VPN Security**
- Use strong private keys (generated with `wg genkey`)
- Limit AllowedIPs to specific networks
- Enable firewall rules to restrict access
- Regular key rotation (monthly)

### **Network Security**
- Change default ports if needed
- Use fail2ban for brute force protection
- Monitor connection logs
- Implement network segmentation

## 🚀 Multi-Node Scaling

### **Adding More Workers**
```bash
# Generate new keys for new worker
sudo wg genkey | tee worker3_private | wg pubkey > worker3_public

# Add peer to server config
[Peer]
PublicKey = WORKER3_PUBLIC_KEY
AllowedIPs = 15.15.15.3/32

# Restart VPN
sudo systemctl restart wg-quick@wg0
```

### **Load Balancing**
```bash
# Multiple server instances (behind load balancer)
./bin/server --host 15.15.15.1 --port 1337 &
./bin/server --host 15.15.15.1 --port 1338 &

# Nginx load balancer
upstream hashcat_backend {
    server 15.15.15.1:1337;
    server 15.15.15.1:1338;
}
```

**Next Steps**: [`99-performance.md`](99-performance.md) for performance optimization
