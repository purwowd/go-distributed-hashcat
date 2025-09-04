export interface Agent {
    id: string
    name: string
    ip_address: string
    port: number
    status: 'online' | 'offline' | 'busy'
    capabilities: string
    last_seen: string
    created_at: string
    updated_at: string
}

export interface Job {
    id: string
    name: string
    status: 'pending' | 'running' | 'completed' | 'failed' | 'paused'
    hash_type: number
    attack_mode: number
    hash_file: string
    hash_file_id?: string
    wordlist: string
    rules?: string
    agent_id?: string
    progress: number
    speed: number
    eta?: string
    result?: string
    created_at: string
    updated_at: string
    started_at?: string
    completed_at?: string
}

export interface HashFile {
    id: string
    name: string
    orig_name: string
    path: string
    size: number
    type: string
    created_at: string
}

export interface Stats {
    onlineAgents: number
    runningJobs: number
    completedJobs: number
    hashFiles: number
}

export interface Notification {
    show: boolean
    message: string
    type: 'success' | 'error' | 'info' | 'warning'
}

export interface ApiResponse<T> {
    success: boolean
    data: T
    error?: string
} 
