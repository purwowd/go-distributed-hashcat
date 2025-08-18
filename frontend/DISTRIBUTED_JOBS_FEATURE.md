# Distributed Jobs Feature

## Overview
Fitur **Distributed Jobs** memungkinkan sistem untuk secara otomatis membagi wordlist menjadi multiple jobs yang dijalankan secara paralel oleh agent berdasarkan performa CPU/GPU. Sistem ini mengoptimalkan distribusi tugas berdasarkan resource yang tersedia.

## 🎯 **Fitur Utama**

### 1. **Automatic Wordlist Division**
- Wordlist otomatis dibagi berdasarkan jumlah agent yang tersedia
- Pembagian berdasarkan performa agent (GPU vs CPU)
- Agent dengan GPU mendapat bagian lebih besar
- Agent dengan CPU mendapat bagian lebih kecil

### 2. **Performance-Based Distribution**
- **GPU Agents**: 100% performance score, mendapat 60-80% dari total words
- **CPU Agents**: 30% performance score, mendapat 20-40% dari total words
- Distribusi proporsional berdasarkan total performance score

### 3. **Parallel Execution**
- Setiap agent menjalankan job secara bersamaan
- Master job untuk monitoring overall progress
- Sub-jobs untuk setiap agent dengan wordlist segment

## 🔧 **Arsitektur Sistem**

### **Backend Components**
```
┌─────────────────────────────────────┐
│           Domain Models             │
├─────────────────────────────────────┤
│ - AgentPerformance                 │
│ - DistributedJobRequest            │
│ - WordlistSegment                  │
│ - DistributedJobResult             │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│        Use Case Layer              │
├─────────────────────────────────────┤
│ - DistributedJobUsecase            │
│ - Agent performance calculation    │
│ - Wordlist segmentation            │
│ - Job distribution logic           │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│         HTTP Handler                │
├─────────────────────────────────────┤
│ - CreateDistributedJobs            │
│ - GetDistributedJobStatus          │
│ - StartAllSubJobs                  │
│ - GetAgentPerformance              │
└─────────────────────────────────────┘
```

### **Frontend Components**
```
┌─────────────────────────────────────┐
│        Distributed Job Modal        │
├─────────────────────────────────────┤
│ - File selection (hash + wordlist) │
│ - Distribution preview              │
│ - Performance visualization         │
│ - Command template generation      │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│         Jobs Tab                    │
├─────────────────────────────────────┤
│ - Create Distributed Job button     │
│ - Job monitoring                   │
│ - Progress tracking                 │
└─────────────────────────────────────┘
```

## 📊 **Algoritma Distribusi**

### **Performance Calculation**
```typescript
function calculateAgentPerformance(agent) {
    const capabilities = agent.capabilities.toLowerCase()
    
    if (capabilities.includes('gpu') || 
        capabilities.includes('cuda') || 
        capabilities.includes('rtx') || 
        capabilities.includes('gtx')) {
        return {
            resourceType: 'GPU',
            performance: 1.0,    // 100%
            speed: 1000000       // 1M H/s
        }
    } else {
        return {
            resourceType: 'CPU',
            performance: 0.3,    // 30%
            speed: 100000        // 100K H/s
        }
    }
}
```

### **Word Distribution Formula**
```typescript
function distributeWords(agents, totalWords) {
    const totalPerformance = agents.reduce((sum, a) => sum + a.performance, 0)
    
    return agents.map(agent => {
        const performanceRatio = agent.performance / totalPerformance
        const assignedWords = Math.round(totalWords * performanceRatio)
        
        return {
            agent: agent,
            words: assignedWords,
            percentage: (performanceRatio * 100).toFixed(1)
        }
    })
}
```

### **Example Distribution**
```
Total Words: 10,000
Agents: 3 (2 GPU + 1 CPU)

GPU Agent 1: 4,000 words (40%)
GPU Agent 2: 4,000 words (40%)  
CPU Agent 1: 2,000 words (20%)

Total Performance: 2.3 (1.0 + 1.0 + 0.3)
```

## 🚀 **API Endpoints**

### **Create Distributed Jobs**
```http
POST /api/v1/distributed-jobs
Content-Type: application/json

{
    "name": "WiFi Crack - Distributed",
    "hash_type": 2500,
    "attack_mode": 0,
    "hash_file_id": "uuid-here",
    "wordlist_id": "uuid-here",
    "rules": "",
    "auto_distribute": true
}
```

### **Get Job Status**
```http
GET /api/v1/distributed-jobs/{master_job_id}/status
```

### **Start All Sub-Jobs**
```http
POST /api/v1/distributed-jobs/{master_job_id}/start-all
```

### **Get Agent Performance**
```http
GET /api/v1/distributed-jobs/performance
```

### **Distribution Preview**
```http
GET /api/v1/distributed-jobs/preview?wordlist_id={uuid}
```

## 💻 **Frontend Implementation**

### **Modal Structure**
```html
<div x-show="showDistributedJobModal" class="modal-modern">
    <!-- File Selection -->
    <select x-model="distributedJobForm.hash_file_id">...</select>
    <select x-model="distributedJobForm.wordlist_id">...</select>
    
    <!-- Distribution Preview -->
    <div class="distribution-preview">
        <template x-for="agent in onlineAgents">
            <div class="agent-card">
                <span x-text="agent.name"></span>
                <span x-text="getAgentPerformanceScore(agent) + '%'"></span>
                <span x-text="getAssignedWordCount(agent) + ' words'"></span>
            </div>
        </template>
    </div>
    
    <!-- Command Template -->
    <pre x-text="distributedCommandTemplate"></pre>
</div>
```

### **JavaScript Functions**
```typescript
// Check if agent is GPU-based
isGPUAgent(agent: any): boolean {
    const capabilities = (agent.capabilities || '').toLowerCase()
    return capabilities.includes('gpu') || 
           capabilities.includes('cuda') || 
           capabilities.includes('rtx')
}

// Calculate performance score
getAgentPerformanceScore(agent: any): number {
    return this.isGPUAgent(agent) ? 100 : 30
}

// Calculate assigned words
getAssignedWordCount(agent: any): number {
    const totalWords = this.selectedWordlist.word_count
    const totalPerformance = this.onlineAgents.reduce((sum, a) => 
        sum + this.getAgentPerformanceScore(a), 0)
    const agentPerformance = this.getAgentPerformanceScore(agent)
    
    return Math.round(totalWords * (agentPerformance / totalPerformance))
}
```

## 📈 **Monitoring & Progress**

### **Master Job Status**
- **Status**: `distributed` (master), `pending` (sub-jobs)
- **Progress**: Aggregated dari semua sub-jobs
- **Result**: Combined results dari semua agent

### **Sub-Job Tracking**
- Individual progress untuk setiap agent
- Speed dan ETA per agent
- Word count yang diproses per agent

### **Real-time Updates**
- WebSocket updates untuk progress
- Live performance metrics
- Agent status monitoring

## 🔍 **Use Cases**

### **1. Large Wordlist Processing**
```
Input: rockyou.txt (14M words)
Agents: 5 (3 GPU + 2 CPU)

Distribution:
- GPU Agent 1: 4.2M words
- GPU Agent 2: 4.2M words  
- GPU Agent 3: 4.2M words
- CPU Agent 1: 0.7M words
- CPU Agent 2: 0.7M words

Result: 5x faster processing
```

### **2. Mixed Resource Environment**
```
Input: 100K wordlist
Agents: 3 (1 RTX 4090 + 2 CPU)

Distribution:
- RTX 4090: 60K words (60%)
- CPU 1: 20K words (20%)
- CPU 2: 20K words (20%)

Optimization: GPU handles heavy lifting
```

### **3. Scalable Infrastructure**
```
Input: 1M wordlist
Agents: 10 (5 GPU + 5 CPU)

Distribution:
- Each GPU: 150K words (15%)
- Each CPU: 50K words (5%)

Scalability: Linear performance increase
```

## 🛠 **Configuration & Tuning**

### **Performance Thresholds**
```yaml
# Agent performance configuration
agent_performance:
  gpu:
    performance_score: 1.0
    speed_multiplier: 10
    word_ratio: 0.6-0.8
  
  cpu:
    performance_score: 0.3
    speed_multiplier: 1
    word_ratio: 0.2-0.4

# Distribution settings
distribution:
  min_words_per_agent: 100
  max_segments: 20
  balance_threshold: 0.1
```

### **Resource Detection**
```go
func detectResourceType(capabilities string) string {
    capabilities = strings.ToLower(capabilities)
    
    if strings.Contains(capabilities, "rtx") ||
       strings.Contains(capabilities, "gtx") ||
       strings.Contains(capabilities, "cuda") {
        return "GPU"
    }
    
    if strings.Contains(capabilities, "cpu") ||
       strings.Contains(capabilities, "intel") ||
       strings.Contains(capabilities, "amd") {
        return "CPU"
    }
    
    return "UNKNOWN"
}
```

## 🧪 **Testing & Validation**

### **Unit Tests**
```go
func TestAgentPerformanceCalculation(t *testing.T) {
    agent := domain.Agent{
        Capabilities: "RTX 4090",
    }
    
    performance := calculateAgentPerformance(agent)
    
    assert.Equal(t, "GPU", performance.ResourceType)
    assert.Equal(t, 1.0, performance.Performance)
    assert.Equal(t, int64(1000000), performance.Speed)
}
```

### **Integration Tests**
```go
func TestWordlistDistribution(t *testing.T) {
    agents := []domain.Agent{
        {Capabilities: "RTX 4090"},
        {Capabilities: "Intel i9"},
    }
    
    wordlist := &domain.Wordlist{WordCount: &[]int64{1000}[0]}
    
    segments := divideWordlistByPerformance(wordlist, agents)
    
    assert.Len(t, segments, 2)
    assert.Equal(t, int64(700), segments[0].WordCount)  // GPU: 70%
    assert.Equal(t, int64(300), segments[1].WordCount)  // CPU: 30%
}
```

## 🚀 **Future Enhancements**

### **1. Dynamic Performance Adjustment**
- Real-time performance monitoring
- Adaptive word distribution
- Load balancing based on current agent load

### **2. Advanced Resource Detection**
- GPU memory detection
- CPU core count consideration
- Network bandwidth analysis

### **3. Machine Learning Optimization**
- Historical performance data
- Predictive distribution algorithms
- Auto-tuning based on success rates

### **4. Fault Tolerance**
- Agent failure handling
- Job redistribution
- Progress preservation

## 📚 **References**

- [Hashcat Documentation](https://hashcat.net/wiki/)
- [WPA/WPA2 Cracking Guide](https://hashcat.net/wiki/doku.php?id=cracking_wpawpa2)
- [Distributed Computing Patterns](https://en.wikipedia.org/wiki/Distributed_computing)
- [Performance Optimization](https://hashcat.net/wiki/doku.php?id=frequently_asked_questions#how_to_optimize_performance)

## 🎉 **Conclusion**

Fitur **Distributed Jobs** memberikan solusi yang powerful untuk:
- **Efisiensi**: Memanfaatkan semua resource yang tersedia
- **Skalabilitas**: Linear performance increase dengan jumlah agent
- **Optimasi**: Smart distribution berdasarkan capability agent
- **Monitoring**: Centralized control dengan distributed execution

Sistem ini mengubah pendekatan tradisional single-job menjadi distributed parallel processing yang jauh lebih efisien! 🚀
