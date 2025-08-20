// Job Store for State Management
import { apiService, type Job } from '@/services/api.service'

interface JobState {
    jobs: Job[]
    loading: boolean
    error: string | null
    lastUpdated: Date | null
}

class JobStore {
    private state: JobState = {
        jobs: [],
        loading: false,
        error: null,
        lastUpdated: null
    }

    private listeners: Set<() => void> = new Set()

    public getState(): JobState {
        return { ...this.state }
    }

    public subscribe(listener: () => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    private notify(): void {
        this.listeners.forEach(listener => listener())
    }

    private setState(updates: Partial<JobState>): void {
        this.state = { ...this.state, ...updates }
        this.notify()
    }

    public actions = {
        fetchJobs: async (params?: { page?: number; page_size?: number; search?: string }): Promise<{ data: Job[]; total: number; page: number; page_size: number } | null> => {
            this.setState({ loading: true, error: null })
            
            try {
                const result = await apiService.getJobs(params)
                this.setState({
                    jobs: result.data,
                    loading: false,
                    lastUpdated: new Date(),
                    error: null
                })
                return result
            } catch (error) {
                this.setState({
                    loading: false,
                    error: error instanceof Error ? error.message : 'Failed to fetch jobs'
                })
                return null
            }
        },

        fetchJob: async (id: string): Promise<Job | null> => {
            try {
                const job = await apiService.getJob(id)
                if (job) {
                    const updatedJobs = this.state.jobs.map(j => 
                        j.id === job.id ? job : j
                    )
                    if (!updatedJobs.find(j => j.id === job.id)) {
                        updatedJobs.push(job)
                    }
                    this.setState({ jobs: updatedJobs })
                }
                return job
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to fetch job' 
                })
                return null
            }
        },

        createJob: async (jobData: Partial<Job>): Promise<Job | null> => {
            this.setState({ loading: true, error: null })
            
            try {
                const newJob = await apiService.createJob(jobData)
                if (newJob) {
                    // ✅ Immediately add the new job to the beginning of the list
                    const updatedJobs = [newJob, ...this.state.jobs]
                    this.setState({ 
                        jobs: updatedJobs,
                        loading: false 
                    })
                    
                    // ✅ Log for debugging
                    console.log('✅ New job created and added to store:', newJob.id, newJob.name)
                    
                    return newJob
                } else {
                    // API returned success but null data
                    this.setState({ 
                        loading: false, 
                        error: 'Server returned empty response'
                    })
                    return null
                }
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to create job'
                this.setState({ 
                    loading: false, 
                    error: errorMessage
                })
                // Re-throw the error so it can be caught by the calling code
                throw error
            }
        },

        updateJob: async (id: string, jobData: Partial<Job>): Promise<Job | null> => {
            try {
                const updatedJob = await apiService.updateJob(id, jobData)
                if (updatedJob) {
                    const updatedJobs = this.state.jobs.map(job => 
                        job.id === id ? updatedJob : job
                    )
                    this.setState({ jobs: updatedJobs })
                }
                return updatedJob
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to update job' 
                })
                return null
            }
        },

        deleteJob: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.deleteJob(id)
                if (success) {
                    const updatedJobs = this.state.jobs.filter(job => job.id !== id)
                    this.setState({ jobs: updatedJobs })
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to delete job' 
                })
                return false
            }
        },

        startJob: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.startJob(id)
                if (success) {
                    await this.actions.fetchJob(id) // Refresh job status
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to start job' 
                })
                return false
            }
        },

        stopJob: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.stopJob(id)
                if (success) {
                    await this.actions.fetchJob(id) // Refresh job status
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to stop job' 
                })
                return false
            }
        },

        pauseJob: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.pauseJob(id)
                if (success) {
                    await this.actions.fetchJob(id) // Refresh job status
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to pause job' 
                })
                return false
            }
        },

        resumeJob: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.resumeJob(id)
                if (success) {
                    await this.actions.fetchJob(id) // Refresh job status
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to resume job' 
                })
                return false
            }
        },

        clearError: (): void => {
            this.setState({ error: null })
        },

        reset: (): void => {
            this.setState({
                jobs: [],
                loading: false,
                error: null,
                lastUpdated: null
            })
        },

        // Real-time update methods
        updateJobProgress: (jobId: string, progress: number, speed?: number, eta?: string, status?: string): void => {
            const jobs = this.state.jobs
            const jobIndex = jobs.findIndex(job => job.id === jobId)
            if (jobIndex !== -1) {
                const validStatuses = ['pending', 'running', 'completed', 'failed', 'cancelled', 'paused']
                const updatedJob = { 
                    ...jobs[jobIndex], 
                    progress,
                    speed: speed || jobs[jobIndex].speed,
                    eta: eta || jobs[jobIndex].eta,
                    status: (status && validStatuses.includes(status)) ? status as Job['status'] : jobs[jobIndex].status
                }
                const newJobs = [...jobs]
                newJobs[jobIndex] = updatedJob
                this.setState({ jobs: newJobs })
            }
        },

        updateJobStatus: (jobId: string, status: string, result?: string): void => {
            const jobs = this.state.jobs
            const jobIndex = jobs.findIndex(job => job.id === jobId)
            if (jobIndex !== -1) {
                const validStatuses = ['pending', 'running', 'completed', 'failed', 'cancelled', 'paused']
                const updatedJob = { 
                    ...jobs[jobIndex], 
                    status: validStatuses.includes(status) ? status as Job['status'] : jobs[jobIndex].status,
                    result: result || jobs[jobIndex].result,
                    progress: status === 'completed' ? 100 : jobs[jobIndex].progress
                }
                const newJobs = [...jobs]
                newJobs[jobIndex] = updatedJob
                this.setState({ jobs: newJobs })
            }
        }
    }

    public getters = {
        getRunningJobs: (): Job[] => {
            return this.state.jobs.filter(job => job.status === 'running')
        },

        getCompletedJobs: (): Job[] => {
            return this.state.jobs.filter(job => job.status === 'completed')
        },

        getPendingJobs: (): Job[] => {
            return this.state.jobs.filter(job => job.status === 'pending')
        },

        getFailedJobs: (): Job[] => {
            return this.state.jobs.filter(job => job.status === 'failed')
        },

        getPausedJobs: (): Job[] => {
            return this.state.jobs.filter(job => job.status === 'paused')
        },

        getJobById: (id: string): Job | undefined => {
            return this.state.jobs.find(job => job.id === id)
        },

        getJobsByAgentId: (agentId: string): Job[] => {
            return this.state.jobs.filter(job => job.agent_id === agentId)
        },

        getTotalCount: (): number => {
            return this.state.jobs.length
        },

        getJobsWithProgress: (): Job[] => {
            return this.state.jobs.filter(job => job.progress !== undefined && job.progress > 0)
        }
    }
}

export const jobStore = new JobStore()
export type { JobState } 
