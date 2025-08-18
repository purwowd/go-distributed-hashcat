# Job Progress Tracking Feature

## Overview
This document describes the job progress tracking feature that allows real-time monitoring of hashcat job execution across distributed agents.

## Features

### Real-time Progress Updates
- Jobs send progress updates every 2 seconds during execution
- Progress percentage, speed (H/s), and ETA are tracked
- WebSocket broadcasts enable real-time frontend updates

### Agent Data Integration
- Agents send complete job data including attack_mode, rules, speed, and eta
- Data is immediately stored in database upon receipt from agent
- Agent ID is automatically linked to job data

## API Endpoints

### Update Job Progress (Legacy)
```
PUT /api/v1/jobs/:id/progress
```
Updates only progress and speed.

### Update Job Data from Agent (New)
```
PUT /api/v1/jobs/:id/data
```
Receives complete job data from agent and updates database immediately.

**Request Body:**
```json
{
  "agent_id": "uuid",
  "attack_mode": 0,
  "rules": "hashcat_rules",
  "speed": 1000000,
  "eta": "2024-01-01T12:00:00Z",
  "progress": 50.5
}
```

**Response:**
```json
{
  "message": "Job data updated successfully"
}
```

## Data Flow

### 1. Job Initialization
When a job starts:
1. Agent calls `sendInitialJobData()` with job configuration
2. Server receives data via `PUT /api/v1/jobs/:id/data`
3. Database is updated immediately with attack_mode, rules, and agent_id

### 2. Progress Updates
During job execution:
1. Agent monitors hashcat output for progress, speed, and ETA
2. Agent calls `updateJobDataFromAgent()` with real-time data
3. Server updates database immediately
4. WebSocket broadcasts update to frontend

### 3. Job Completion
When job completes:
1. Agent extracts password from hashcat output
2. Agent calls `completeJob()` with result
3. Job status is set to "completed" with final result

## Database Schema

### Jobs Table
```sql
CREATE TABLE jobs (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  hash_type INTEGER NOT NULL,
  attack_mode INTEGER NOT NULL,
  hash_file TEXT,
  hash_file_id UUID,
  wordlist TEXT,
  wordlist_id UUID,
  rules TEXT,                    -- Hashcat rules or cracked password
  agent_id UUID,                 -- Linked to agent executing the job
  progress REAL DEFAULT 0,
  speed BIGINT DEFAULT 0,        -- Hash rate in H/s
  eta TIMESTAMP,                 -- Estimated time of completion
  result TEXT,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  started_at TIMESTAMP,
  completed_at TIMESTAMP
);
```

## Agent Implementation

### Data Extraction
Agents extract the following data from hashcat output:
- **Progress**: `Progress\.+:\s*(\d+)/(\d+)\s*\((\d+\.\d+)%\)`
- **Speed**: `Speed\.+:\s*(\d+)\s*H/s`
- **ETA**: `ETA\.+:\s*(\d+):(\d+):(\d+)`

### Functions
- `sendInitialJobData()`: Sends job configuration when job starts
- `updateJobDataFromAgent()`: Sends real-time progress updates
- `monitorHashcatOutput()`: Parses hashcat output for data extraction

## Backend Implementation

### New Usecase Method
```go
func (u *jobUsecase) UpdateJobData(ctx context.Context, job *domain.Job) error
```
Updates complete job data in database immediately.

### Handler Method
```go
func (h *JobHandler) UpdateJobDataFromAgent(c *gin.Context)
```
Receives complete job data from agent and updates database.

## Benefits

1. **Immediate Data Storage**: Job data is stored in database as soon as it's received from agent
2. **Complete Information**: All relevant job data (attack_mode, rules, speed, eta) is captured
3. **Agent Association**: Each job is properly linked to the agent executing it
4. **Real-time Updates**: Frontend receives immediate updates via WebSocket
5. **Backward Compatibility**: Legacy progress endpoint still works

## Migration Notes

- Existing jobs will continue to work with the legacy progress endpoint
- New jobs will use the enhanced data endpoint for better tracking
- Database schema remains compatible with existing data
- WebSocket broadcasts work with both endpoints

## Future Enhancements

1. **Distributed Job Support**: Extend to support multiple agents per job
2. **Performance Metrics**: Track agent performance over time
3. **Job Queuing**: Implement intelligent job distribution
4. **Resource Monitoring**: Track CPU/GPU usage during execution
