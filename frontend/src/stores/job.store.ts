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
        fetchJobs: async (): Promise<void> => {
            this.setState({ loading: true, error: null })
            
            try {
                const jobs = await apiService.getJobs()
                this.setState({ 
                    jobs, 
                    loading: false, 
                    lastUpdated: new Date(),
                    error: null 
                })
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to fetch jobs' 
                })
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
                    this.setState({ 
                        jobs: [...this.state.jobs, newJob],
                        loading: false 
                    })
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
