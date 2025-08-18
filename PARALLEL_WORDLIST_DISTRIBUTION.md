# Parallel Wordlist Distribution Feature

## Overview

Fitur ini memungkinkan distribusi wordlist ke multiple agent secara proporsional berdasarkan kecepatan masing-masing agent. Setiap agent akan menerima bagian wordlist yang sesuai dengan kapabilitas dan kecepatan mereka.

## Fitur Utama

### 1. Distribusi Proporsional Berdasarkan Kecepatan Agent

- **CPU Agent**: Speed = 1 (baseline)
- **GPU Agent**: Speed = 5 (5x lebih cepat)
- **GTX Agent**: Speed = 6 (6x lebih cepat)  
- **RTX Agent**: Speed = 8 (8x lebih cepat)

### 2. Logging Detail

#### Saat Job Dibuat
```
🚀 Starting parallel job creation with 3 online agents
📝 Wordlist contains 6 valid words
📋 Words: [admin tehbotolsosro admin.admin bambang1234 Starbucks2025@@!! makanyuk]
🤖 Agent distribution plan:
   - tes-agent-cpu-01 (CPU): Speed=1, Weight=0.07, Words=0
   - tes-agent-gpu-01 (GPU): Speed=5, Weight=0.36, Words=2
   - tes-agent-cpu-02 (CPU): Speed=1, Weight=0.07, Words=0
📦 Assigned 2 words to tes-agent-gpu-01: [admin tehbotolsosro]
📦 Assigned 2 words to tes-agent-cpu-01: [admin.admin bambang1234]
📦 Assigned 2 words to tes-agent-cpu-02: [Starbucks2025@@!! makanyuk]
✅ Created job 8c8d3203-70a9-4d59-a8cb-a99b429e558d for agent 913710df-9601-4788-8fd5-ca107999cc79 with 2 words
✅ Created job 9a27523e-a704-4a93-b768-9a5152da5cee for agent 913710df-9601-4788-8fd5-ca107999cc79 with 2 words
✅ Created job f2fddccd-1755-4b40-bb79-a73013764293 for agent 913710df-9601-4788-8fd5-ca107999cc79 with 2 words
🎉 Successfully created 3 parallel jobs
```

#### Saat Job Selesai
```
🎉 SUCCESS: Agent tes-agent-gpu-01 found password for job Parallel Job - test_wordlist
   📍 Result: Starbucks2025@@!!
   🔍 Job ID: 8c8d3203-70a9-4d59-a8cb-a99b429e558d
   ⚡ Speed: 2500000 H/s
   📊 Progress: 100.00%
✅ Successfully updated agent tes-agent-gpu-01 status to online

❌ FAILED: Agent tes-agent-cpu-01 did not find password for job Parallel Job - test_wordlist
   🔍 Job ID: 9a27523e-a704-4a93-b768-9a5152da5cee
   ⚡ Speed: 800000 H/s
   📊 Progress: 100.00%
   📝 Reason: Password not found - exhausted
✅ Successfully updated agent tes-agent-cpu-01 status to online

❌ FAILED: Agent tes-agent-cpu-02 did not find password for job Parallel Job - test_wordlist
   🔍 Job ID: f2fddccd-1755-4b40-bb79-a73013764293
   ⚡ Speed: 600000 H/s
   📊 Progress: 100.00%
   📝 Reason: Password not found - exhausted
✅ Successfully updated agent tes-agent-cpu-02 status to online
```

#### Ringkasan Parallel Jobs
```
📊 Parallel Jobs Summary:
   📋 Wordlist: test_wordlist
   🎯 Overall: SUCCESS: Password found by 1 agent(s) - Starbucks2025@@!!
   🤖 Agents: 3 total (1 success, 2 failed)
      - tes-agent-gpu-01: SUCCESS: Found password (Starbucks2025@@!!)
      - tes-agent-cpu-01: FAILED: No password found
      - tes-agent-cpu-02: FAILED: No password found
```

## API Endpoints

### 1. Create Parallel Jobs
```http
POST /api/v1/jobs/auto
Content-Type: application/json

{
  "hash_file_id": "8c8d3203-70a9-4d59-a8cb-a99b429e558d",
  "wordlist_id": "9a27523e-a704-4a93-b768-9a5152da5cee"
}
```

**Response:**
```json
{
  "message": "Parallel jobs created successfully",
  "data": {
    "total_jobs": 3,
    "total_words": 6,
    "agents_used": 3,
    "jobs": [...]
  }
}
```

### 2. Get Parallel Jobs Summary
```http
GET /api/v1/jobs/parallel/summary
```

**Response:**
```json
{
  "message": "Parallel jobs summary retrieved successfully",
  "data": {
    "total_parallel_jobs": 1,
    "summaries": [
      {
        "wordlist_name": "test_wordlist",
        "total_agents": 3,
        "success_count": 1,
        "failure_count": 2,
        "overall_result": "SUCCESS: Password found by 1 agent(s) - Starbucks2025@@!!",
        "agent_results": [
          {
            "agent_name": "tes-agent-gpu-01",
            "agent_id": "913710df-9601-4788-8fd5-ca107999cc79",
            "job_id": "8c8d3203-70a9-4d59-a8cb-a99b429e558d",
            "status": "success",
            "result": "SUCCESS: Found password (Starbucks2025@@!!)",
            "speed": 2500000,
            "progress": 100.0,
            "started_at": "2024-08-18T14:30:00Z",
            "completed_at": "2024-08-18T14:35:00Z"
          }
        ],
        "created_at": "2024-08-18T14:30:00Z"
      }
    ]
  }
}
```

## Error Handling

### Agent Status Update Errors
```
⚠️ Failed to update agent status to online for agent tes-agent-cpu-01: agent not found
❌ Failed to send job completion to server: connection refused
❌ Job completion failed with status 500: internal server error
```

### Job Data Update Errors
```
❌ Failed to send initial job data to server: connection refused
❌ Initial job data failed with status 400: invalid request body
❌ Failed to send job data update to server: timeout
❌ Job data update failed with status 404: job not found
```

## Contoh Penggunaan

### 1. Buat Wordlist dengan 6 Kata
```
admin
tehbotolsosro
admin.admin
bambang1234
Starbucks2025@@!!
makanyuk
```

### 2. Jalankan 3 Agents
- **tes-agent-cpu-01**: CPU agent (Speed: 1)
- **tes-agent-gpu-01**: GPU agent (Speed: 5)  
- **tes-agent-cpu-02**: CPU agent (Speed: 1)

### 3. Distribusi Otomatis
- **tes-agent-gpu-01**: 2 kata (36% dari total)
- **tes-agent-cpu-01**: 2 kata (36% dari total)
- **tes-agent-cpu-02**: 2 kata (28% dari total)

### 4. Hasil
- **tes-agent-gpu-01**: ✅ Berhasil menemukan password "Starbucks2025@@!!"
- **tes-agent-cpu-01**: ❌ Tidak menemukan password
- **tes-agent-cpu-02**: ❌ Tidak menemukan password

## Keuntungan

1. **Efisiensi**: Agent dengan kecepatan tinggi mendapat lebih banyak kata
2. **Load Balancing**: Beban kerja terdistribusi secara proporsional
3. **Monitoring**: Logging detail untuk tracking progress dan hasil
4. **Error Handling**: Penanganan error yang lebih baik dengan logging jelas
5. **Real-time Updates**: Data job diupdate secara real-time dari agent

## Troubleshooting

### Masalah Umum

1. **Agent tidak menerima distribusi**
   - Pastikan agent status "online"
   - Periksa capabilities agent

2. **Error status update agent**
   - Periksa koneksi network
   - Periksa log server untuk detail error

3. **Job tidak selesai**
   - Periksa log agent untuk error hashcat
   - Periksa status agent di database
