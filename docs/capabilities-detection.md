# Hashcat Capabilities Detection for Agent

## Overview

Agent sekarang memiliki fitur capabilities detection yang canggih menggunakan command `hashcat -I` untuk mendeteksi device types (CPU/GPU) yang tersedia di server lokal secara akurat.

## Features

### ✅ **Smart Device Detection**
- Menggunakan `hashcat -I` untuk mendapatkan informasi device yang akurat
- Parsing output hashcat untuk mengekstrak device types
- Prioritas GPU over CPU jika kedua device tersedia

### ✅ **Intelligent Database Updates**
- Hanya update database jika capabilities berubah
- Tidak melakukan update yang tidak perlu
- Log yang jelas untuk setiap perubahan

### ✅ **Fallback Mechanism**
- Fallback ke basic detection jika hashcat tidak tersedia
- Fallback ke basic detection jika parsing gagal
- Graceful degradation untuk berbagai skenario

## How It Works

### 1. **Capabilities Detection Flow**

```
Agent Startup
     ↓
Check if --capabilities parameter provided
     ↓
If capabilities not provided or "auto":
     ↓
Run hashcat -I command
     ↓
Parse output for device types
     ↓
Determine capabilities (GPU priority over CPU)
     ↓
Compare with database capabilities
     ↓
Update database only if changed
     ↓
Continue with detected capabilities
```

### 2. **Hashcat -I Output Parsing**

Agent mem-parse output `hashcat -I` untuk mencari:

```
Backend Device ID #1
  Type...........: CPU
  Vendor.ID......: 128
  Vendor.........: GenuineIntel
  Name...........: pthread-11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
  Version........: OpenCL 1.2 pocl HSTR: pthread-x86_64-pc-linux-gnu-goldmont
  Processor(s)...: 4
  Clock..........: 2803
  Memory.Total...: 2936 MB (limited to 1024 MB allocatable in one block)
  Memory.Free....: 2872 MB
  OpenCL.Version.: OpenCL C 1.2 pocl
  Driver.Version.: 1.8
```

**Key Parsing Logic:**
- Mencari section `Backend Device ID #`
- Mengekstrak line `Type...........:`
- Mengumpulkan semua device types yang ditemukan
- Prioritas: GPU > CPU

### 3. **Device Type Priority**

```
1. GPU devices (highest priority)
   - NVIDIA GPU
   - AMD GPU
   - Intel GPU
   - Any device with "GPU" in type

2. CPU devices (fallback)
   - Intel CPU
   - AMD CPU
   - Any device with "CPU" in type

3. Fallback to basic detection
   - If hashcat unavailable
   - If parsing fails
   - If no device types found
```

## Implementation Details

### **Main Function: `detectCapabilitiesWithHashcat()`**

```go
func detectCapabilitiesWithHashcat() string {
    // Check if hashcat is available
    if _, err := exec.LookPath("hashcat"); err != nil {
        log.Printf("⚠️ Warning: hashcat not found, falling back to basic detection")
        return detectCapabilitiesBasic()
    }

    // Run hashcat -I to get device information
    cmd := exec.Command("hashcat", "-I")
    output, err := cmd.Output()
    if err != nil {
        log.Printf("⚠️ Warning: Failed to run hashcat -I: %v", err)
        return detectCapabilitiesBasic()
    }

    // Parse output to find device types
    outputStr := string(output)
    lines := strings.Split(outputStr, "\n")
    
    var deviceTypes []string
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        // Look for Type line
        if strings.HasPrefix(line, "Type...........:") {
            parts := strings.Split(line, ":")
            if len(parts) >= 2 {
                deviceType := strings.TrimSpace(parts[1])
                if deviceType != "" {
                    deviceTypes = append(deviceTypes, deviceType)
                    log.Printf("🔍 Detected device type: %s", deviceType)
                }
            }
        }
    }

    // Determine capabilities based on detected devices
    if len(deviceTypes) == 0 {
        return detectCapabilitiesBasic()
    }

    // Check if any GPU devices are found
    for _, deviceType := range deviceTypes {
        if strings.Contains(strings.ToUpper(deviceType), "GPU") {
            log.Printf("✅ GPU device detected: %s", deviceType)
            return "GPU"
        }
    }

    // If no GPU, check for CPU
    for _, deviceType := range deviceTypes {
        if strings.Contains(strings.ToUpper(deviceType), "CPU") {
            log.Printf("✅ CPU device detected: %s", deviceType)
            return "CPU"
        }
    }

    // Fallback to basic detection
    return detectCapabilitiesBasic()
}
```

### **Fallback Function: `detectCapabilitiesBasic()`**

```go
func detectCapabilitiesBasic() string {
    // Try to detect GPU first
    if hasGPU() {
        return "GPU"
    }

    // Fallback to CPU
    return "CPU"
}
```

### **Database Update Logic**

```go
// ✅ Update capabilities di database jika berbeda dengan yang terdeteksi
if info.Capabilities == "" || info.Capabilities != capabilities {
    log.Printf("🔄 Updating capabilities from '%s' to '%s'", info.Capabilities, capabilities)
    if err := updateAgentCapabilities(tempAgent, agentKey, capabilities); err != nil {
        log.Printf("⚠️ Warning: Failed to update capabilities: %v", err)
    } else {
        log.Printf("✅ Capabilities updated successfully")
    }
} else {
    log.Printf("ℹ️ Capabilities already up-to-date: %s", capabilities)
}
```

## Usage Examples

### 1. **Auto-detect Capabilities (Recommended)**

```bash
sudo ./bin/agent \
  --server http://172.15.2.76:1337 \
  --name GPU-Agent \
  --agent-key "3730b5d6"
```

**Expected Output:**
```
🔍 Auto-detected capabilities using hashcat -I: GPU
🔍 Detected device type: GPU
✅ GPU device detected: GPU
✅ Capabilities updated successfully
```

### 2. **Force Specific Capabilities**

```bash
sudo ./bin/agent \
  --server http://172.15.2.76:1337 \
  --name GPU-Agent \
  --capabilities "GPU" \
  --agent-key "3730b5d6"
```

**Expected Output:**
```
ℹ️ Capabilities already up-to-date: GPU
```

### 3. **Force CPU Capabilities**

```bash
sudo ./bin/agent \
  --server http://172.15.2.76:1337 \
  --name GPU-Agent \
  --capabilities "CPU" \
  --agent-key "3730b5d6"
```

**Expected Output:**
```
🔄 Updating capabilities from 'GPU' to 'CPU'
✅ Capabilities updated successfully
```

## Error Handling

### **Hashcat Not Available**

```
⚠️ Warning: hashcat not found, falling back to basic detection
🔍 Auto-detected capabilities: CPU
```

### **Hashcat Command Failed**

```
⚠️ Warning: Failed to run hashcat -I: exit status 1
⚠️ Falling back to basic detection
🔍 Auto-detected capabilities: CPU
```

### **No Device Types Found**

```
⚠️ No device types found in hashcat -I output, falling back to basic detection
🔍 Auto-detected capabilities: CPU
```

### **Unrecognized Device Types**

```
⚠️ Could not determine capabilities from device types: [CustomDevice OpenCLDevice]
⚠️ Falling back to basic detection
🔍 Auto-detected capabilities: CPU
```

## Testing

### **Run Test Script**

```bash
./scripts/test_capabilities_detection.sh
```

### **Manual Testing**

1. **Test with hashcat available:**
   ```bash
   # Install hashcat if not available
   sudo apt-get install hashcat
   
   # Test hashcat -I
   hashcat -I
   
   # Run agent
   sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key"
   ```

2. **Test without hashcat:**
   ```bash
   # Remove hashcat temporarily
   sudo mv /usr/bin/hashcat /usr/bin/hashcat.backup
   
   # Run agent (should fallback to basic detection)
   sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key"
   
   # Restore hashcat
   sudo mv /usr/bin/hashcat.backup /usr/bin/hashcat
   ```

## Benefits

1. **🎯 Accurate Detection**: Menggunakan hashcat -I untuk deteksi yang akurat
2. **🔄 Smart Updates**: Hanya update database jika capabilities berubah
3. **🛡️ Robust Fallback**: Multiple fallback mechanisms untuk reliability
4. **📊 Device Priority**: Prioritas GPU over CPU sesuai kebutuhan hashcat
5. **🔍 Detailed Logging**: Log yang jelas untuk debugging dan monitoring
6. **⚡ Performance**: Tidak ada update database yang tidak perlu

## Troubleshooting

### **Common Issues**

1. **Hashcat command not found:**
   - Install hashcat: `sudo apt-get install hashcat`
   - Agent will fallback to basic detection

2. **Hashcat -I fails:**
   - Check hashcat installation
   - Verify permissions
   - Agent will fallback to basic detection

3. **No device types detected:**
   - Check hashcat -I output format
   - Verify parsing logic
   - Agent will fallback to basic detection

### **Debug Mode**

Enable debug logging by checking agent logs:

```bash
sudo ./bin/agent --server http://localhost:1337 --name test-agent --agent-key "test-key" 2>&1 | grep -E "(🔍|✅|⚠️|❌)"
```

## Conclusion

Fitur capabilities detection yang baru memberikan:

- **Accuracy**: Deteksi device types yang akurat menggunakan hashcat -I
- **Efficiency**: Update database hanya ketika diperlukan
- **Reliability**: Multiple fallback mechanisms
- **Transparency**: Log yang jelas untuk monitoring dan debugging

Agent sekarang dapat mendeteksi capabilities server lokal dengan akurat dan mengupdate database secara intelligent, memberikan pengalaman yang lebih baik untuk distributed hashcat environment.
