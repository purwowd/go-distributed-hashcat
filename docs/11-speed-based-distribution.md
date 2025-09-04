# Speed-Based Word Distribution

## Overview

This document describes the improved word distribution system that uses actual agent speed data instead of hardcoded capability-based estimates.

## Problem

Previously, the system used hardcoded speed values based on agent capabilities:
- CPU: 1 (default)
- GPU: 5
- RTX: 8
- GTX: 6

This resulted in equal distribution (33%, 33%, 34%) regardless of actual agent performance, leading to suboptimal cracking efficiency.

## Solution

The system now uses actual speed data from the `agent.speed` field in the database, with fallback to capability-based estimation for agents without speed data.

## Implementation

### Backend Changes

#### 1. Job Handler (`internal/delivery/http/handler/job_handler.go`)
```go
// Use actual speed from database, fallback to capability-based estimation if speed is 0
speed := agent.Speed
if speed == 0 {
    // Fallback to capability-based estimation for agents without speed data
    speed = 1 // Default for CPU
    if strings.Contains(strings.ToLower(agent.Capabilities), "gpu") {
        speed = 5 // GPU is faster
    } else if strings.Contains(strings.ToLower(agent.Capabilities), "rtx") {
        speed = 8 // RTX lebih cepat lagi
    } else if strings.Contains(strings.ToLower(agent.Capabilities), "gtx") {
        speed = 6 // GTX lebih cepat
    }
}
```

#### 2. Job Usecase (`internal/usecase/job_usecase.go`)
Similar changes to use actual speed data with fallback.

#### 3. Distributed Job Usecase (`internal/usecase/distributed_job_usecase.go`)
Enhanced `calculateAgentPerformance` function:
- Uses actual speed from database
- Calculates performance score based on actual speed (normalized to 0-1 scale)
- Sorts agents by actual speed (highest first)

### Frontend Changes

#### Main Application (`frontend/src/main.ts`)
Updated `getAgentPerformanceScore` function:
```javascript
getAgentPerformanceScore(agent: any): number {
    // Use actual speed from database if available, fallback to capability-based estimation
    if (agent.speed && agent.speed > 0) {
        return agent.speed // Use actual speed in H/s
    }
    
    // Fallback to capability-based estimation for agents without speed data
    // ... capability-based logic
}
```

## Example Distribution

### Input Data
- **Agent-B**: 2,404 H/s (highest)
- **Agent-A**: 1,907 H/s (medium)
- **Agent-C**: 1,294 H/s (lowest)
- **Total Words**: 758,561

### Old Distribution (Incorrect)
- All agents: 252,853 words each (33% each)

### New Distribution (Correct)
- **Agent-B**: 325,348 words (42.9%)
- **Agent-A**: 258,086 words (34.0%)
- **Agent-C**: 175,127 words (23.1%)

### Efficiency Improvement
- Agent-B (fastest) gets 72,495 more words
- Agent-C (slowest) gets 77,726 fewer words
- This optimizes overall cracking performance

## Benefits

1. **Optimal Performance**: Faster agents get more work, maximizing overall throughput
2. **Real-time Adaptation**: Uses actual benchmark data instead of estimates
3. **Backward Compatibility**: Falls back to capability-based estimation for agents without speed data
4. **Accurate Distribution**: Ensures total words are distributed correctly without loss

## Testing

Use the test script to verify distribution:
```bash
./scripts/test_speed_distribution.sh
```

This script demonstrates the correct distribution calculation and compares it with the old equal distribution method.

## Database Requirements

Agents must have the `speed` field populated with actual benchmark data (in H/s). The system will fall back to capability-based estimation if this field is 0 or empty.

## Future Enhancements

1. **Dynamic Speed Updates**: Real-time speed monitoring and updates
2. **Historical Performance**: Use average speed over time for more stable distribution
3. **Load Balancing**: Consider current agent load in addition to speed
4. **Hash Type Optimization**: Different speed profiles for different hash types
