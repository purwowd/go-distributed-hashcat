# Capabilities Detection Fix

## Problem
Agent defaulted to 'GPU' even when hashcat -I showed 'CPU'

## Root Cause
Hardcoded default: `capabilities = "GPU"` prevented auto-detection

## Solution
1. Changed default from 'GPU' to 'auto'
2. Auto-detection now always triggered for 'auto'
3. Better logging during hashcat -I detection
4. Raw output preview for debugging

## Expected Result
Agent will now auto-detect 'CPU' from hashcat -I output correctly!

## Test Command
```bash
sudo ./bin/agent --server http://172.15.2.76:1337 --name GPU-Agent --ip "172.15.1.94" --agent-key "3730b5d6"
```
