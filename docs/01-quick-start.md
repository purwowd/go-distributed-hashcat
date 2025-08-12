# Quick Start Guide

Setup distributed hashcat system in **15 minutes**.

## 🚀 Prerequisites

- **OS**: Ubuntu 22.04 LTS
- **Control Server**: 2 vCPU, 4GB RAM  
- **GPU Worker**: 2 vCPU, 4GB RAM

## ⚡ 1. Local Setup

```bash
# Install dependencies
sudo apt update && sudo apt install git nodejs npm golang-go hashcat -y

# Clone and build
git clone https://github.com/purwowd/go-distributed-hashcat.git
cd go-distributed-hashcat
make build

# Start backend
./bin/server &

# Start frontend  
cd frontend && npm install && npm run dev
```
**Access**: http://localhost:3000

### 🗝️ Agent Key Setup

1. **Buka dashboard server di browser:**  
   http://localhost:3000
2. **Masuk ke menu "Agent Key".**
3. **Buat agent name baru** (misal: `gpu-worker-01`), lalu copy agent key yang di-generate.
4. **Simpan agent key untuk digunakan pada worker.**

**GPU Worker** (separate machine):
```bash
cd go-distributed-hashcat
./bin/agent --server http://SERVER_IP:1337 --name gpu-worker-01 --ip "AGENT_IP" --agent-key "AGENT_KEY"
```
- Ganti `SERVER_IP` dengan IP server.
- Ganti `AGENT_IP` dengan IP worker.
- Ganti `AGENT_KEY` dengan agent key yang sudah di-copy dari dashboard.

## 🔒 2. Production Setup (with VPN)

**VPN Setup:**
```bash
# Install WireGuard
sudo apt install wireguard -y

# Generate keys
sudo wg genkey | tee privatekey | wg pubkey > publickey

# Server config (/etc/wireguard/wg0.conf)
[Interface]
PrivateKey = YOUR_PRIVATE_KEY
Address = 15.15.15.1/24
ListenPort = 51820

[Peer]
PublicKey = CLIENT_PUBLIC_KEY  
AllowedIPs = 15.15.15.2/32

# Client config
[Interface]  
PrivateKey = CLIENT_PRIVATE_KEY
Address = 15.15.15.2/24

[Peer]
PublicKey = SERVER_PUBLIC_KEY
Endpoint = SERVER_PUBLIC_IP:51820
AllowedIPs = 15.15.15.0/24

# Start VPN
sudo systemctl enable wg-quick@wg0 && sudo systemctl start wg-quick@wg0
```

**Deploy Application:**
```bash
# Server (15.15.15.1)
./bin/server --host 15.15.15.1 &
cd frontend && npm run build && python3 -m http.server 3000 --bind 15.15.15.1 &

# Worker (15.15.15.2) 
./bin/agent --server http://15.15.15.1:1337 --name remote-gpu --ip "15.15.15.2" --agent-key "AGENT_KEY"
```
- Pastikan agent key sudah dibuat dan di-copy dari dashboard server.

## ✅ 3. Verification

```bash
# Check services
curl http://localhost:1337/health                    # Backend health
curl http://localhost:1337/api/v1/agents/           # List agents  
sudo wg show                                         # VPN status

# Web dashboard
open http://localhost:3000  # Local
open http://15.15.15.1:3000 # VPN
```

## 🧪 Quick Test

```bash
# Test hash: MD5 "password"
echo "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8" > test.txt

# Upload hash file
curl -X POST -F "file=@test.txt" http://localhost:1337/api/v1/hash-files/

# Create test job
curl -X POST http://localhost:1337/api/v1/jobs/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Quick Test", "hash_file_id": 1, "attack_mode": 0, "hash_type": 0}'
```

## 🐛 Troubleshooting

| Issue | Solution |
|-------|----------|
| Port 1337 busy | `sudo lsof -i :1337 && sudo kill -9 PID` |
| Agent won't connect | `sudo ufw allow 1337` |
| No GPU detected | `sudo ubuntu-drivers autoinstall` |
| VPN fails | `sudo ufw allow 51820/udp` |

## 📈 Next Steps

- **Production**: [`02-deployment.md`](02-deployment.md)
- **VPN Setup**: [`06-wireguard-deployment.md`](06-wireguard-deployment.md)
- **API Docs**: [`03-api-reference.md`](03-api-reference.md)
