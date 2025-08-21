# Capabilities Preservation Final Fix

## Problem
Capabilities were still becoming empty when agent went offline, despite previous fixes

## Root Cause Analysis
**Multiple `updateAgentInfo` calls during shutdown:**
1. **First call**: `agent.updateAgentInfo(agent.ID, ip, 8080, capabilities, "offline")` - preserves capabilities='CPU'
2. **Second call**: `agent.restoreOriginalPort()` - calls `updateAgentInfo` again, potentially overriding capabilities

## Solution Applied
**Removed duplicate `updateAgentInfo` call:**
- ✅ Single `updateAgentInfo` call handles status, port, and capabilities
- ❌ Removed `restoreOriginalPort()` call to avoid capabilities override
- ✅ Capabilities are now preserved during shutdown

## Expected Results
**Database state changes:**
- **Online**: `capabilities='CPU'`, `status='online'`, `port=8081`
- **Offline**: `capabilities='CPU'`, `status='offline'`, `port=8080` (capabilities PRESERVED!)

## Test Command
```bash
sudo ./bin/agent --server http://172.15.2.76:1337 --name GPU-Agent --ip "30.30.30.39" --agent-key "3730b5d6"
```

## Expected Logs
**Shutdown logs:**
```
🔄 Updating agent status to offline and restoring port to 8080...
🔄 Preserving capabilities: CPU
✅ Agent status updated to offline with port 8080 and capabilities preserved
ℹ️ Skipping restoreOriginalPort() to avoid capabilities override
```
