import { Agent, Job, HashFile, ApiResponse } from '../types/index'

export class ApiClient {
    private baseUrl: string

    constructor(baseUrl: string) {
        this.baseUrl = baseUrl
    }

    private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`
        
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...options?.headers
            },
            ...options
        })

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data = await response.json()
        return data.data || data
    }

    // Agent methods
    async getAgents(): Promise<Agent[]> {
        return this.request<Agent[]>('/api/v1/agents/list')
    }

    async getAgent(id: string): Promise<Agent> {
        return this.request<Agent>(`/api/v1/agents/${id}`)
    }

    async createAgent(agent: Partial<Agent>): Promise<Agent> {
        return this.request<Agent>('/api/v1/agents/register', {
            method: 'POST',
            body: JSON.stringify(agent)
        })
    }

    async updateAgentStatus(id: string, status: string): Promise<void> {
        return this.request<void>(`/api/v1/agents/${id}/status`, {
            method: 'PUT',
            body: JSON.stringify({ status })
        })
    }

    async deleteAgent(id: string): Promise<void> {
        return this.request<void>(`/api/v1/agents/${id}`, {
            method: 'DELETE'
        })
    }

    // Job methods
    async getJobs(): Promise<Job[]> {
        return this.request<Job[]>('/api/v1/jobs/list')
    }

    async getJob(id: string): Promise<Job> {
        return this.request<Job>(`/api/v1/jobs/${id}`)
    }

    async createJob(job: Partial<Job>): Promise<Job> {
        return this.request<Job>('/api/v1/jobs/create', {
            method: 'POST',
            body: JSON.stringify(job)
        })
    }

    async startJob(id: string): Promise<void> {
        return this.request<void>(`/api/v1/jobs/${id}/start`, {
            method: 'POST'
        })
    }

    async pauseJob(id: string): Promise<void> {
        return this.request<void>(`/api/v1/jobs/${id}/pause`, {
            method: 'POST'
        })
    }

    async resumeJob(id: string): Promise<void> {
        return this.request<void>(`/api/v1/jobs/${id}/resume`, {
            method: 'POST'
        })
    }

    async deleteJob(id: string): Promise<void> {
        return this.request<void>(`/api/v1/jobs/${id}`, {
            method: 'DELETE'
        })
    }

    // Hash file methods
    async getHashFiles(): Promise<HashFile[]> {
        return this.request<HashFile[]>('/api/v1/hashfiles/')
    }

    async uploadHashFile(file: File): Promise<HashFile> {
        const formData = new FormData()
        formData.append('file', file)

        const response = await fetch(`${this.baseUrl}/api/v1/hashfiles/upload`, {
            method: 'POST',
            body: formData
        })

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.statusText}`)
        }

        const data = await response.json()
        return data.data
    }

    async deleteHashFile(id: string): Promise<void> {
        return this.request<void>(`/api/v1/hashfiles/${id}`, {
            method: 'DELETE'
        })
    }
} 
