# Frontend - Distributed Hashcat Dashboard

## API Endpoints

This frontend application consumes the Distributed Hashcat API with the following corrected endpoints:

### Agent Management

#### Agent Registration and Management
- `GET /api/v1/agents/list` - Get all agents (with pagination support)
- `GET /api/v1/agents/by-key` - Get agent by agent key (query parameter: agent_key)
- `POST /api/v1/agents/register` - Register a new agent
- `GET /api/v1/agents/:id` - Get specific agent by ID
- `PUT /api/v1/agents/:id/status` - Update agent status
- `DELETE /api/v1/agents/:id` - Delete agent

#### Agent Operations
- `POST /api/v1/agents/generate-key` - Generate agent key
- `POST /api/v1/agents/startup` - Agent startup
- `POST /api/v1/agents/heartbeat` - Agent heartbeat
- `POST /api/v1/agents/update-data` - Update agent data
- `PUT /api/v1/agents/:id/heartbeat` - Update agent heartbeat
- `POST /api/v1/agents/:id/files` - Register agent files

#### Agent Job Management
- `GET /api/v1/agents/:id/jobs` - Get jobs by agent ID
- `GET /api/v1/agents/:id/jobs/next` - Get available job for agent

### Job Management

#### Job Operations
- `GET /api/v1/jobs/list` - Get all jobs (with pagination support)
- `POST /api/v1/jobs/create` - Create new job
- `GET /api/v1/jobs/:id` - Get specific job by ID
- `DELETE /api/v1/jobs/:id` - Delete job

#### Job Control
- `POST /api/v1/jobs/assign` - Assign jobs to agents
- `POST /api/v1/jobs/auto` - Create parallel jobs automatically
- `POST /api/v1/jobs/:id/start` - Start a job
- `PUT /api/v1/jobs/:id/progress` - Update job progress
- `PUT /api/v1/jobs/:id/data` - Update job data from agent
- `POST /api/v1/jobs/:id/complete` - Complete a job
- `POST /api/v1/jobs/:id/fail` - Mark job as failed
- `POST /api/v1/jobs/:id/pause` - Pause a job
- `POST /api/v1/jobs/:id/resume` - Resume a job
- `POST /api/v1/jobs/:id/stop` - Stop a job

#### Job Queries
- `GET /api/v1/jobs/parallel/summary` - Get parallel jobs summary
- `GET /api/v1/jobs/agent/:id` - Get available job for specific agent

### File Management

#### Hash Files
- `POST /api/v1/hashfiles/upload` - Upload hash file
- `GET /api/v1/hashfiles/` - Get all hash files
- `GET /api/v1/hashfiles/:id` - Get specific hash file
- `GET /api/v1/hashfiles/:id/download` - Download hash file
- `DELETE /api/v1/hashfiles/:id` - Delete hash file

#### Wordlists
- `POST /api/v1/wordlists/upload` - Upload wordlist
- `POST /api/v1/wordlists/upload/init` - Initialize chunked upload
- `POST /api/v1/wordlists/upload/chunk` - Upload chunk
- `POST /api/v1/wordlists/upload/finalize` - Finalize chunked upload
- `GET /api/v1/wordlists/` - Get all wordlists
- `GET /api/v1/wordlists/:id` - Get specific wordlist
- `GET /api/v1/wordlists/:id/download` - Download wordlist
- `DELETE /api/v1/wordlists/:id` - Delete wordlist

### Health Check
- `GET /health` - Health check endpoint

## Services

### API Service (`src/services/api.service.ts`)
Main service for all API calls with proper error handling and response parsing.

### API Client (`src/services/api-client.ts`)
Alternative API client implementation with simplified interface.

### Stores
- **Agent Store** (`src/stores/agent.store.ts`): Manages agent state and operations
- **Job Store** (`src/stores/job.store.ts`): Manages job state and operations

## Key Changes Made

### 1. Fixed Agent Endpoints
- **Before**: `POST /api/v1/agents/` → **After**: `POST /api/v1/agents/register`
- **Before**: `GET /api/v1/agents/` → **After**: `GET /api/v1/agents/list`

### 2. Fixed Job Endpoints
- **Before**: `POST /api/v1/jobs/` → **After**: `POST /api/v1/jobs/create`
- **Before**: `GET /api/v1/jobs/` → **After**: `GET /api/v1/jobs/list`

### 3. Updated API Client
- Added proper `/api/v1` prefix to all endpoints
- Fixed all endpoint paths to match backend changes

## Usage Examples

### Creating an Agent
```typescript
import { apiService } from '@/services/api.service'

const newAgent = await apiService.createAgent({
    name: 'GPU-Agent-01',
    ip_address: '192.168.1.100',
    port: 8080,
    capabilities: 'NVIDIA RTX 4090',
    agent_key: 'generated-key-here'
})
```

### Creating a Job
```typescript
import { apiService } from '@/services/api.service'

const newJob = await apiService.createJob({
    name: 'Hash Cracking Job',
    hash_type: 2500,
    attack_mode: 0,
    hash_file_id: 'hash-file-uuid',
    wordlist: 'rockyou.txt'
})
```

### Fetching Agents with Pagination
```typescript
import { apiService } from '@/services/api.service'

const agents = await apiService.getAgents({
    page: 1,
    page_size: 10,
    search: 'GPU'
})
```

## Error Handling

All API calls include proper error handling with:
- HTTP status code checking
- Response validation
- Error message extraction
- Fallback error messages

## State Management

The application uses a custom store pattern for state management:
- **Reactive updates**: Components automatically re-render when state changes
- **Loading states**: Built-in loading indicators for async operations
- **Error handling**: Centralized error state management
- **Pagination support**: Built-in pagination for large datasets

## Development

### Running the Frontend
```bash
cd frontend
npm install
npm run dev
```

### Building for Production
```bash
npm run build
```

### Testing
```bash
npm run test
```

## Notes

- All endpoints now use the corrected paths that match the backend
- The frontend automatically handles API versioning (`/api/v1`)
- Pagination is supported for both agents and jobs lists
- File uploads use FormData for proper multipart handling
- WebSocket support is available for real-time updates
