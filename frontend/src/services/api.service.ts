// API Service for Distributed Hashcat Dashboard
import { getConfig } from '@/config/build.config'

// Types for API responses
export interface Agent {
    id: string
    name: string
    ip_address: string
    port?: number
    status: 'online' | 'offline' | 'busy'
    capabilities?: string
    agent_key: string
    last_seen: string
    created_at: string
    updated_at: string
}

export interface Job {
    id: string
    name: string
    hash_file_id?: string    // Changed from hash_file
    wordlist_id?: string     // Changed from wordlist
    hash_type: number
    attack_mode: number
    status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused'
    created_at: string
    started_at?: string
    completed_at?: string
    progress?: number
    speed?: number
    eta?: string
    result?: string
    agent_id?: string
    assigned_agents?: string[]
    
    // NEW: Enriched fields from backend
    agent_name?: string      // Human-readable agent name
    wordlist_name?: string   // Original wordlist filename
    hash_file_name?: string  // Original hash file filename
}

export interface HashFile {
    id: string
    name: string
    orig_name: string
    path?: string
    size: number
    type: string
    created_at: string
}

export interface Wordlist {
    id: string
    name: string
    orig_name: string
    path?: string
    size: number
    word_count?: number
    created_at: string
}

export interface ApiResponse<T> {
    success: boolean
    data?: T
    message?: string
    error?: string
}

export interface HealthResponse {
    status: string
    timestamp: string
    version: string
    uptime: number
}

class ApiService {
    private baseUrl: string
    private config = getConfig()

    constructor() {
        this.baseUrl = this.config.apiBaseUrl
    }

    // Generic HTTP request method
    private async request<T>(
        endpoint: string, 
        options: RequestInit = {}
    ): Promise<ApiResponse<T>> {
        try {
            const url = `${this.baseUrl}${endpoint}`
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            })

            if (!response.ok) {
                let errorMessage = `HTTP ${response.status}: ${response.statusText}`
                try {
                    const errorData = await response.json()
                    if (errorData.error) {
                        // For duplicate agent errors, use only the backend error message
                        if (errorData.error.startsWith('already exists')) {
                            errorMessage = errorData.error
                        } else if (errorData.error.includes('is already registered with agent name')) {
                            // For agent key already registered errors, use only the backend error message
                            errorMessage = errorData.error
                        } else if (errorData.error.includes('is not registered in database')) {
                            // For agent key not registered errors, use only the backend error message
                            errorMessage = errorData.error
                        } else if (errorData.error.includes('is already registered with IP address')) {
                            // For already registered agent errors, use only the backend error message
                            errorMessage = errorData.error
                        } else {
                            errorMessage = `${errorMessage} - ${errorData.error}`
                        }
                    }
                } catch (e) {
                    // If response isn't JSON, use default message
                }
                throw new Error(errorMessage)
            }

            const data = await response.json()
            return {
                success: true,
                data
            }
        } catch (error) {
            console.error(`API Request failed for ${endpoint}:`, error)
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Unknown error'
            }
        }
    }

    // GET request
    private async get<T>(endpoint: string): Promise<ApiResponse<T>> {
        return this.request<T>(endpoint, { method: 'GET' })
    }

    // POST request
    private async post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
        return this.request<T>(endpoint, {
            method: 'POST',
            body: data ? JSON.stringify(data) : undefined
        })
    }

    // PUT request
    private async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
        return this.request<T>(endpoint, {
            method: 'PUT',
            body: data ? JSON.stringify(data) : undefined
        })
    }

    // DELETE request
    private async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
        return this.request<T>(endpoint, { method: 'DELETE' })
    }

    // Health Check
    public async checkHealth(): Promise<HealthResponse | null> {
        try {
            const response = await this.get<HealthResponse>('/health')
            return response.success ? response.data! : null
        } catch (error) {
            console.warn('Health check failed:', error)
            return null
        }
    }

    // Agent Management
    public async getAgents(params?: { page?: number; page_size?: number; search?: string }): Promise<{ data: Agent[]; total: number; page: number; page_size: number }> {
        const query: string[] = []
        if (params?.page) query.push(`page=${params.page}`)
        if (params?.page_size) query.push(`page_size=${params.page_size}`)
        if (params?.search) query.push(`search=${encodeURIComponent(params.search)}`)
        const qs = query.length ? `?${query.join('&')}` : ''
        const response = await this.get<{ data: Agent[]; total: number; page: number; page_size: number }>(`/api/v1/agents/list${qs}`)
        if (response.success && response.data) {
            return {
                data: (response.data as any).data || [],
                total: (response.data as any).total || 0,
                page: (response.data as any).page || 1,
                page_size: (response.data as any).page_size || 10
            }
        }
        return { data: [], total: 0, page: 1, page_size: 10 }
    }

    public async getAgent(id: string): Promise<Agent | null> {
        const response = await this.get<{data: Agent}>(`/api/v1/agents/${id}`)
        return response.success ? response.data!.data : null
    }

    public async createAgent(agentData: Partial<Agent>): Promise<{agent: Agent | null, error: string | null}> {
        const response = await this.post<{data: Agent}>('/api/v1/agents/register', agentData)
        return {
            agent: response.success ? response.data!.data : null,
            error: response.error || null
        }
    }

    public async updateAgentData(agentData: { agent_key: string; ip_address?: string; port?: number; capabilities?: string }): Promise<{success: boolean, message?: string, code?: string, error?: string}> {
        const response = await this.post<{message: string, code: string}>('/api/v1/agents/update-data', agentData)
        return {
            success: response.success,
            message: response.data?.message,
            code: response.data?.code,
            error: response.error
        }
    }

    public async generateAgentKey(name: string): Promise<{agent: Agent | null, error: string | null}> {
        const response = await this.post<{data: Agent}>('/api/v1/agents/generate-key', { name })
        return {
            agent: response.success ? response.data!.data : null,
            error: response.error || null
        }
    }

    public async updateAgent(id: string, agentData: Partial<Agent>): Promise<Agent | null> {
        const response = await this.put<Agent>(`/api/v1/agents/${id}`, agentData)
        return response.success ? response.data! : null
    }

    public async deleteAgent(id: string): Promise<boolean> {
        const response = await this.delete(`/api/v1/agents/${id}`)
        return response.success
    }

    // NEW: Get jobs assigned to specific agent
    public async getAgentJobs(agentId: string): Promise<Job[]> {
        const response = await this.get<{data: Job[]}>(`/api/v1/agents/${agentId}/jobs`)
        if (response.success && response.data && response.data.data) {
            return response.data.data
        }
        return []
    }

    // NEW: Get next available job for agent (polling)
    public async getNextJobForAgent(agentId: string): Promise<Job | null> {
        const response = await this.get<{data: Job | null}>(`/api/v1/agents/${agentId}/jobs/next`)
        if (response.success && response.data && response.data.data) {
            return response.data.data
        }
        return null
    }

    // NEW: Enhanced job progress update with ETA
    public async updateJobProgress(id: string, progress: number, speed: number, eta?: string): Promise<boolean> {
        const response = await this.put(`/api/v1/jobs/${id}/progress`, {
            progress,
            speed,
            eta
        })
        return response.success
    }

    // Job Management
    public async getJobs(params?: { page?: number; page_size?: number; search?: string }): Promise<{ data: Job[]; total: number; page: number; page_size: number }> {
        const query: string[] = []
        if (params?.page) query.push(`page=${params.page}`)
        if (params?.page_size) query.push(`page_size=${params.page_size}`)
        if (params?.search) query.push(`search=${encodeURIComponent(params.search)}`)
        const qs = query.length ? `?${query.join('&')}` : ''
        const response = await this.get<{ data: Job[]; total: number; page: number; page_size: number }>(`/api/v1/jobs/list${qs}`)
        if (response.success && response.data) {
            return {
                data: (response.data as any).data || [],
                total: (response.data as any).total || 0,
                page: (response.data as any).page || 1,
                page_size: (response.data as any).page_size || 10
            }
        }
        return { data: [], total: 0, page: 1, page_size: 10 }
    }

    public async getJob(id: string): Promise<Job | null> {
        const response = await this.get<Job>(`/api/v1/jobs/${id}`)
        return response.success ? response.data! : null
    }

    public async createJob(jobData: Partial<Job>): Promise<Job | null> {
        const response = await this.post<Job>('/api/v1/jobs/create', jobData)
        return response.success ? response.data! : null
    }

    public async updateJob(id: string, jobData: Partial<Job>): Promise<Job | null> {
        const response = await this.put<Job>(`/api/v1/jobs/${id}`, jobData)
        return response.success ? response.data! : null
    }

    public async deleteJob(id: string): Promise<boolean> {
        const response = await this.delete(`/api/v1/jobs/${id}`)
        return response.success
    }

    public async startJob(id: string): Promise<boolean> {
        const response = await this.post(`/api/v1/jobs/${id}/start`)
        return response.success
    }

    public async stopJob(id: string): Promise<boolean> {
        const response = await this.post(`/api/v1/jobs/${id}/stop`)
        return response.success
    }

    public async pauseJob(id: string): Promise<boolean> {
        const response = await this.post(`/api/v1/jobs/${id}/pause`)
        return response.success
    }

    public async resumeJob(id: string): Promise<boolean> {
        const response = await this.post(`/api/v1/jobs/${id}/resume`)
        return response.success
    }

    // File Management
    public async getHashFiles(): Promise<HashFile[]> {
        const response = await this.get<{data: HashFile[]}>('/api/v1/hashfiles/')
        if (response.success && response.data && response.data.data) {
            return response.data.data // Extract array from wrapper
        }
        return []
    }

    public async getHashFile(id: string): Promise<HashFile | null> {
        const response = await this.get<HashFile>(`/api/v1/hashfiles/${id}`)
        return response.success ? response.data! : null
    }

    public async uploadHashFile(file: File): Promise<HashFile | null> {
        try {
            const formData = new FormData()
            formData.append('file', file)

            const response = await fetch(`${this.baseUrl}/api/v1/hashfiles/upload`, {
                method: 'POST',
                body: formData
            })

            if (!response.ok) {
                // Try to get the error message from response body
                let errorMessage = `Upload failed: ${response.statusText}`
                try {
                    const errorData = await response.json()
                    if (errorData.error) {
                        errorMessage = errorData.error
                    }
                } catch (parseError) {
                    // If we can't parse the error response, use the default message
                    console.warn('Could not parse error response:', parseError)
                }
                throw new Error(errorMessage)
            }

            const data = await response.json()
            return data.data || data // Handle both wrapped and unwrapped responses
        } catch (error) {
            console.error('Hash file upload failed:', error)
            throw error // Re-throw the error so the caller can handle it
        }
    }

    public async deleteHashFile(id: string): Promise<boolean> {
        const response = await this.delete(`/api/v1/hashfiles/${id}`)
        return response.success
    }

    public async downloadHashFile(id: string): Promise<Blob | null> {
        try {
            const response = await fetch(`${this.baseUrl}/api/v1/hashfiles/${id}/download`)
            if (!response.ok) {
                throw new Error(`Download failed: ${response.statusText}`)
            }
            return await response.blob()
        } catch (error) {
            console.error('Hash file download failed:', error)
            return null
        }
    }

    // Wordlist Management
    public async getWordlists(): Promise<Wordlist[]> {
        const response = await this.get<{data: Wordlist[]}>('/api/v1/wordlists/')
        if (response.success && response.data && response.data.data) {
            return response.data.data // Extract array from wrapper
        }
        return []
    }

    public async getWordlist(id: string): Promise<Wordlist | null> {
        const response = await this.get<Wordlist>(`/api/v1/wordlists/${id}`)
        return response.success ? response.data! : null
    }

    public async uploadWordlist(file: File): Promise<Wordlist | null> {
        try {
            const formData = new FormData()
            formData.append('file', file)

            // Create AbortController for timeout management
            const controller = new AbortController()
            const timeoutId = setTimeout(() => controller.abort(), 30 * 60 * 1000) // 30 minutes timeout

            try {
                const response = await fetch(`${this.baseUrl}/api/v1/wordlists/upload`, {
                    method: 'POST',
                    body: formData,
                    signal: controller.signal
                })

                clearTimeout(timeoutId)

                if (!response.ok) {
                    // Try to get the error message from response body
                    let errorMessage = `Upload failed: ${response.statusText}`
                    try {
                        const errorData = await response.json()
                        if (errorData.error) {
                            errorMessage = errorData.error
                        }
                    } catch (parseError) {
                        // If we can't parse the error response, use the default message
                        console.warn('Could not parse error response:', parseError)
                    }
                    throw new Error(errorMessage)
                }

                const data = await response.json()
                return data.data || data // Handle both wrapped and unwrapped responses
            } catch (error: any) {
                clearTimeout(timeoutId)
                if (error.name === 'AbortError') {
                    throw new Error('Upload timeout: File too large or connection too slow. Please use CLI upload for files larger than 1GB.')
                }
                throw error
            }
        } catch (error) {
            console.error('Wordlist upload failed:', error)
            throw error // Re-throw the error so the caller can handle it
        }
    }

    public async deleteWordlist(id: string): Promise<boolean> {
        const response = await this.delete(`/api/v1/wordlists/${id}`)
        return response.success
    }

    public async downloadWordlist(id: string): Promise<Blob | null> {
        try {
            const response = await fetch(`${this.baseUrl}/api/v1/wordlists/${id}/download`)
            if (!response.ok) {
                throw new Error(`Download failed: ${response.statusText}`)
            }
            return await response.blob()
        } catch (error) {
            console.error('Wordlist download failed:', error)
            return null
        }
    }

    // Utility methods
    public async ping(): Promise<boolean> {
        try {
            const response = await fetch(`${this.baseUrl}/health`, {
                method: 'GET',
                timeout: 5000
            } as any)
            return response.ok
        } catch {
            return false
        }
    }

    // NEW: Cache Management
    public async getCacheStats(): Promise<any> {
        const response = await this.get('/api/v1/cache/stats')
        return response.success ? response.data : null
    }

    public async clearCache(): Promise<boolean> {
        const response = await this.delete('/api/v1/cache/clear')
        return response.success
    }

    public getBaseUrl(): string {
        return this.baseUrl
    }

    public updateBaseUrl(newUrl: string): void {
        this.baseUrl = newUrl
    }
}

// Export singleton instance
export const apiService = new ApiService() 
