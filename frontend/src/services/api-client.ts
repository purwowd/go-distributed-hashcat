import { Agent, Job, HashFile } from '../types/index'

export class ApiClient {
    private baseUrl: string
    private xToken: string

    constructor(baseUrl: string) {
        this.baseUrl = baseUrl
        this.xToken = import.meta.env.VITE_X_TOKEN || ''
    }

    private buildHeaders(extra?: HeadersInit): HeadersInit {
        return {
            'Content-Type': 'application/json',
            ...(this.xToken ? { 'X-Token': this.xToken } : {}),
            ...extra
        }
    }

    private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`

        const response = await fetch(url, {
            headers: this.buildHeaders(options?.headers), // <<< pakai helper
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
        return this.request<Agent[]>('/agents')
    }

    async getAgent(id: string): Promise<Agent> {
        return this.request<Agent>(`/agents/${id}`)
    }

    async createAgent(agent: Partial<Agent>): Promise<Agent> {
        return this.request<Agent>('/agents', {
            method: 'POST',
            body: JSON.stringify(agent)
        })
    }

    async updateAgentStatus(id: string, status: string): Promise<void> {
        return this.request<void>(`/agents/${id}/status`, {
            method: 'PUT',
            body: JSON.stringify({ status })
        })
    }

    async deleteAgent(id: string): Promise<void> {
        return this.request<void>(`/agents/${id}`, {
            method: 'DELETE'
        })
    }

    // Job methods
    async getJobs(): Promise<Job[]> {
        return this.request<Job[]>('/jobs')
    }

    async getJob(id: string): Promise<Job> {
        return this.request<Job>(`/jobs/${id}`)
    }

    async createJob(job: Partial<Job>): Promise<Job> {
        return this.request<Job>('/jobs', {
            method: 'POST',
            body: JSON.stringify(job)
        })
    }

    async startJob(id: string): Promise<void> {
        return this.request<void>(`/jobs/${id}/start`, { method: 'POST' })
    }

    async pauseJob(id: string): Promise<void> {
        return this.request<void>(`/jobs/${id}/pause`, { method: 'POST' })
    }

    async resumeJob(id: string): Promise<void> {
        return this.request<void>(`/jobs/${id}/resume`, { method: 'POST' })
    }

    async deleteJob(id: string): Promise<void> {
        return this.request<void>(`/jobs/${id}`, { method: 'DELETE' })
    }

    // Hash file methods
    async getHashFiles(): Promise<HashFile[]> {
        return this.request<HashFile[]>('/hashfiles')
    }

    async uploadHashFile(file: File): Promise<HashFile> {
        const formData = new FormData()
        formData.append('file', file)

        const response = await fetch(`${this.baseUrl}/hashfiles/upload`, {
            method: 'POST',
            headers: {
                ...(this.xToken ? { 'X-Token': this.xToken } : {})
            },
            body: formData
        })

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.statusText}`)
        }

        const data = await response.json()
        return data.data || data
    }

    async deleteHashFile(id: string): Promise<void> {
        return this.request<void>(`/hashfiles/${id}`, { method: 'DELETE' })
    }
}
