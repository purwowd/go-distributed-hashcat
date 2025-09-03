#!/bin/bash

# Hashcat Agent Setup Script
# This script sets up the local file structure for the distributed hashcat agent

set -e

UPLOAD_DIR="/root/uploads"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Setting up Hashcat Agent local files..."
echo "📁 Upload directory: $UPLOAD_DIR"

# Create directory structure
echo "📂 Creating directory structure..."
sudo mkdir -p "$UPLOAD_DIR"/{wordlists,hash-files,temp}
sudo mkdir -p "$UPLOAD_DIR/wordlists"/{common,leaked,custom}
sudo mkdir -p "$UPLOAD_DIR/hash-files"/{wifi,other}

# Set proper permissions
echo "🔒 Setting permissions..."
sudo chown -R root:root "$UPLOAD_DIR"
sudo chmod -R 755 "$UPLOAD_DIR"

echo "✅ Directory structure created:"
tree "$UPLOAD_DIR" 2>/dev/null || ls -la "$UPLOAD_DIR"

# Create example files and documentation
echo "📝 Creating example files..."

# Wordlists README
sudo tee "$UPLOAD_DIR/wordlists/README.md" > /dev/null << 'EOF'
# Wordlists Directory

Place your wordlists in the appropriate subdirectories:

## Common Wordlists (`common/`)
- `rockyou.txt` - Most popular wordlist (14M passwords)
- `common-passwords.txt` - Common passwords collection
- `top-1000.txt` - Top 1000 most common passwords

## Leaked Data (`leaked/`)
- `linkedin.txt` - LinkedIn breach data
- `adobe.txt` - Adobe breach data
- `facebook.txt` - Facebook breach data

## Custom/Target-Specific (`custom/`)
- `company-wordlist.txt` - Company-specific terms
- `location-words.txt` - Location-based passwords
- `dates-years.txt` - Important dates and years

## Download Popular Wordlists

```bash
# Download rockyou.txt (most popular)
wget https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt
sudo mv rockyou.txt /root/uploads/wordlists/common/

# Download SecLists wordlists
git clone https://github.com/danielmiessler/SecLists.git
sudo cp SecLists/Passwords/Common-Credentials/10k-most-common.txt /root/uploads/wordlists/common/
sudo cp SecLists/Passwords/Leaked-Databases/*.txt /root/uploads/wordlists/leaked/
```

## Upload via API (Alternative)

```bash
# Upload via API v1 endpoints
curl -X POST -F "file=@rockyou.txt" http://your-server:1337/api/v1/wordlists/upload
curl -X POST -F "file=@custom.txt" http://your-server:1337/api/v1/wordlists/upload

# List uploaded wordlists
curl http://your-server:1337/api/v1/wordlists/
```

## Upload via Web Dashboard

Access the modern web dashboard at: http://your-server:3000
- Navigate to "Wordlists" tab
- Upload files via drag & drop interface
- Real-time sync to all agents

## File Formats Supported
- `.txt` - Plain text wordlists
- `.lst` - List files
- `.dic` - Dictionary files  
- `.wordlist` - Wordlist files

Files are automatically detected and registered with the server.
EOF

# Hash Files README
sudo tee "$UPLOAD_DIR/hash-files/README.md" > /dev/null << 'EOF'
# Hash Files Directory

Place your hash files in the appropriate subdirectories:

## WiFi Captures (`wifi/`)
- `*.hccapx` - Modern WiFi handshake format
- `*.hccap` - Legacy WiFi handshake format
- `*.cap` - Raw packet captures
- `*.pcap` - Packet capture files

## Other Hash Types (`other/`)
- `*.hash` - Various hash formats
- `*.txt` - Text-based hash files

## Capturing WiFi Handshakes

```bash
# Using aircrack-ng
airmon-ng start wlan0
airodump-ng wlan0mon --write capture
aireplay-ng -0 10 -a [BSSID] wlan0mon

# Convert to hccapx
cap2hccapx.bin capture-01.cap capture.hccapx

# Copy to agent
sudo cp capture.hccapx /root/uploads/hash-files/wifi/
```

## File Formats Supported
- `.hccapx` - Hashcat WiFi format (preferred)
- `.hccap` - Legacy hashcat WiFi format
- `.cap` - Raw 802.11 captures
- `.pcap` - Packet capture files

Files are automatically detected and registered with the server.
EOF

# Create example download script
sudo tee "$UPLOAD_DIR/download-wordlists.sh" > /dev/null << 'EOF'
#!/bin/bash

# Example script to download popular wordlists
# Run as root: sudo bash download-wordlists.sh

WORDLIST_DIR="/root/uploads/wordlists"

echo "📥 Downloading popular wordlists..."

# Create directories
mkdir -p "$WORDLIST_DIR"/{common,leaked,custom}

cd "$WORDLIST_DIR/common"

# Download rockyou.txt (most important)
if [ ! -f "rockyou.txt" ]; then
    echo "Downloading rockyou.txt..."
    wget -q --show-progress https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt
    echo "✅ rockyou.txt downloaded (14M passwords)"
fi

# Download common passwords
if [ ! -f "10k-most-common.txt" ]; then
    echo "Downloading 10k most common passwords..."
    wget -q --show-progress https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/10k-most-common.txt
    echo "✅ 10k-most-common.txt downloaded"
fi

# Download leaked passwords
cd "$WORDLIST_DIR/leaked"
if [ ! -f "linkedin.txt" ]; then
    echo "Downloading leaked passwords..."
    wget -q --show-progress https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Leaked-Databases/rockyou-75.txt -O linkedin.txt
    echo "✅ LinkedIn passwords downloaded"
fi

echo "🎉 Wordlist download complete!"
echo "📊 Summary:"
echo "  - rockyou.txt: $(wc -l < $WORDLIST_DIR/common/rockyou.txt 2>/dev/null || echo '0') passwords"
echo "  - 10k-most-common.txt: $(wc -l < $WORDLIST_DIR/common/10k-most-common.txt 2>/dev/null || echo '0') passwords"
echo "  - leaked data: $(wc -l < $WORDLIST_DIR/leaked/linkedin.txt 2>/dev/null || echo '0') passwords"

echo ""
echo "🔍 To verify files are detected by agent:"
echo "  tail -f /var/log/hashcat-agent.log | grep 'Scanning local files'"
EOF

sudo chmod +x "$UPLOAD_DIR/download-wordlists.sh"

# Create systemd service example
sudo tee /tmp/hashcat-agent.service > /dev/null << 'EOF'
[Unit]
Description=Hashcat Distributed Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/hashcat-agent
ExecStart=/opt/hashcat-agent/bin/agent --server http://your-server:1337 --upload-dir /root/uploads
Restart=always
RestartSec=10
Environment=AGENT_NAME=gpu-server-01
Environment=CAPABILITIES=RTX 4090, OpenCL

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=hashcat-agent

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo "🎯 Next Steps:"
echo ""
echo "1. 📥 Download wordlists:"
echo "   sudo bash $UPLOAD_DIR/download-wordlists.sh"
echo ""
echo "2. 📁 Add your own files:"
echo "   sudo cp /path/to/rockyou.txt $UPLOAD_DIR/wordlists/common/"
echo "   sudo cp /path/to/handshake.hccapx $UPLOAD_DIR/hash-files/wifi/"
echo ""
echo "3. 🔧 Install agent service:"
echo "   sudo cp /tmp/hashcat-agent.service /etc/systemd/system/"
echo "   sudo systemctl daemon-reload"
echo "   sudo systemctl enable hashcat-agent"
echo "   sudo systemctl start hashcat-agent"
echo ""
echo "4. 📊 Monitor agent logs:"
echo "   sudo journalctl -u hashcat-agent -f"
echo ""
echo "5. 🌐 Access dashboard:"
echo "   http://your-server:1337"
echo ""
echo "✅ Agent setup complete! Files in $UPLOAD_DIR will be auto-detected."
echo "Agent scans for new files every 5 minutes." 
