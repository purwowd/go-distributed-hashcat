# Capabilities Preservation Fix

## Problem
When agent was stopped (Ctrl+C), capabilities field was being cleared to empty string

## Root Cause
`restoreOriginalPort()` function was calling `updateAgentInfo` with empty capabilities:
```go
err := a.updateAgentInfo(a.ID, a.ServerIP, a.OriginalPort, "", "offline")
//                                                           ^^ empty capabilities!
```

## Solution
1. Fixed `restoreOriginalPort()` to preserve current capabilities
2. Enhanced shutdown logging to show capabilities preservation
3. Ensured capabilities are never cleared during agent lifecycle

## Expected Result
Capabilities will remain 'CPU' during entire agent lifecycle:
- Startup: capabilities = 'CPU' (detected from hashcat -I)
- Running: capabilities = 'CPU' (maintained)
- Shutdown: capabilities = 'CPU' (PRESERVED - not changed!)

## Test Command
```bash
sudo ./bin/agent --server http://172.15.2.76:1337 --name GPU-Agent --ip "30.30.30.39" --agent-key "3730b5d6"
```
