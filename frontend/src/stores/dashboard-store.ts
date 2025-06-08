import { ApiClient } from '../services/api-client'
import { Agent, Job, HashFile, Stats, Notification } from '../types/index'

export class DashboardStore {
    private apiClient: ApiClient
    private state: {
        currentTab: string
        mobileMenuOpen: boolean
        agents: Agent[]
        jobs: Job[]
        hashFiles: HashFile[]
        stats: Stats
        notification: Notification
        loading: boolean
    }

    constructor(apiClient: ApiClient) {
        this.apiClient = apiClient
        this.state = {
            currentTab: 'overview',
            mobileMenuOpen: false,
            agents: [],
            jobs: [],
            hashFiles: [],
            stats: {
                onlineAgents: 0,
                runningJobs: 0,
                completedJobs: 0,
                hashFiles: 0
            },
            notification: {
                show: false,
                message: '',
                type: 'success'
            },
            loading: false
        }
    }

    getState() {
        return {
            ...this.state,
            // Methods bound to this context
            init: this.init.bind(this),
            loadData: this.loadData.bind(this),
            setCurrentTab: this.setCurrentTab.bind(this),
            toggleMobileMenu: this.toggleMobileMenu.bind(this),
            showNotification: this.showNotification.bind(this),
            hideNotification: this.hideNotification.bind(this)
        }
    }

    async init() {
        await this.loadData()
        // Auto-refresh every 10 seconds
        setInterval(() => {
            this.loadData()
        }, 10000)
    }

    async loadData() {
        this.state.loading = true
        try {
            const [agents, jobs, hashFiles] = await Promise.all([
                this.apiClient.getAgents(),
                this.apiClient.getJobs(),
                this.apiClient.getHashFiles()
            ])

            this.state.agents = agents
            this.state.jobs = jobs
            this.state.hashFiles = hashFiles
            this.updateStats()
        } catch (error) {
            console.error('Error loading data:', error)
            this.showNotification('Failed to load data', 'error')
        } finally {
            this.state.loading = false
        }
    }

    setCurrentTab(tab: string) {
        this.state.currentTab = tab
        this.state.mobileMenuOpen = false
    }

    toggleMobileMenu() {
        this.state.mobileMenuOpen = !this.state.mobileMenuOpen
    }

    showNotification(message: string, type: 'success' | 'error' = 'success') {
        this.state.notification = {
            show: true,
            message,
            type
        }
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            this.hideNotification()
        }, 5000)
    }

    hideNotification() {
        this.state.notification.show = false
    }

    private updateStats() {
        this.state.stats = {
            onlineAgents: this.state.agents.filter(a => a.status === 'online').length,
            runningJobs: this.state.jobs.filter(j => j.status === 'running').length,
            completedJobs: this.state.jobs.filter(j => j.status === 'completed').length,
            hashFiles: this.state.hashFiles.length
        }
    }
} 
