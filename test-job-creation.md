# Job Creation Fix - Test Documentation

## 🐛 **Problem Fixed**

**Error**: `400 Bad Request` when creating jobs via frontend

**Root Cause**: Frontend was missing the required `wordlist` field that backend expects.

## 🔧 **Changes Made**

### 1. **Frontend Payload Fix** (`main.ts`)
```javascript
// OLD (missing wordlist field)
const jobPayload = {
    name: jobData.name,
    hash_type: parseInt(jobData.hash_type),
    attack_mode: parseInt(jobData.attack_mode),
    hash_file_id: jobData.hash_file_id,
    wordlist_id: jobData.wordlist_id,  // Only ID, missing actual filename
    agent_id: jobData.agent_id || undefined
}

// NEW (includes required wordlist field)
const jobPayload = {
    name: jobData.name,
    hash_type: parseInt(jobData.hash_type),
    attack_mode: parseInt(jobData.attack_mode),
    hash_file_id: jobData.hash_file_id,
    wordlist: wordlistName,           // ✅ Required field for backend
    wordlist_id: jobData.wordlist_id, // ✅ Optional reference ID
    agent_id: jobData.agent_id || undefined
}
```

### 2. **Enhanced Error Handling**
- Added frontend validation before API call
- Improved error messages with actual server response
- Better debug logging
- Proper error propagation from stores

### 3. **Server Response Format**
Backend expects:
```json
{
    "name": "WiFi Crack Job",
    "hash_type": 2500,
    "attack_mode": 0,
    "hash_file_id": "uuid-here",
    "wordlist": "rockyou.txt",        // Required filename
    "wordlist_id": "uuid-here",       // Optional
    "agent_id": "uuid-here"           // Optional
}
```

## 🧪 **Testing Steps**

### Prerequisites
1. Start the server: `make run-server`
2. Start the frontend: `cd frontend && npm run dev`
3. Upload at least one hash file and one wordlist

### Test Case 1: Basic Job Creation
1. Navigate to `http://localhost:3000`
2. Click "Create New Job" button
3. Fill out the form:
   - **Job Name**: "Test WiFi Crack"
   - **Hash File**: Select any uploaded hash file
   - **Wordlist**: Select any uploaded wordlist
   - **Hash Type**: 2500 (WPA/WPA2)
   - **Attack Mode**: 0 (Dictionary Attack)
4. Click "Create Job"
5. **Expected**: Success notification + job appears in jobs list

### Test Case 2: Error Validation
1. Try to create a job with missing fields
2. **Expected**: "Please fill in all required fields" error

### Test Case 3: Network Error Handling
1. Stop the server
2. Try to create a job
3. **Expected**: Detailed error message about connection failure

### Debug Information
Check browser console for these logs:
```
Creating job with payload: {
  name: "Test WiFi Crack",
  hash_type: 2500,
  attack_mode: 0,
  hash_file_id: "uuid-here",
  wordlist: "rockyou.txt",
  wordlist_id: "uuid-here"
}
```

## 🔍 **Backend Validation**

The backend validates these fields as required:
- `name`: string, required
- `hash_type`: integer >= 0
- `attack_mode`: integer >= 0
- `hash_file_id`: string, required (must be valid UUID)
- `wordlist`: string, required

Optional fields:
- `wordlist_id`: string (UUID)
- `agent_id`: string (UUID)
- `rules`: string

## 📋 **Verification Checklist**

- [ ] No more 400 errors when creating jobs
- [ ] Jobs appear in the list after creation
- [ ] Error messages are descriptive
- [ ] Console shows debug logs
- [ ] Form validation works before submission
- [ ] Auto-refresh of jobs list after creation

## 🚀 **Next Steps**

After testing, you may want to:
1. Remove debug console.log statements for production
2. Add more specific validation for hash types and attack modes
3. Implement real-time job status updates via WebSocket 
