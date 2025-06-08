// Alpine.js is loaded via CDN and available as window.Alpine
import { componentLoader } from './utils/component-loader'
import { componentRegistry, perf, getConfig } from './config/build.config'

// Import all services and stores
import { apiService } from './services/api.service'
import { agentStore } from './stores/agent.store'
import { jobStore } from './stores/job.store'
import { fileStore } from './stores/file.store'
import { wordlistStore } from './stores/wordlist.store'

// Alpine.js global type
declare global {
    interface Window {
        Alpine: any
        dashboardApp: DashboardApplication
        deferLoadingAlpine: (Alpine: any) => void
        alpineReady: boolean
        alpineStarted: boolean
        alpineManuallyStarted: boolean
    }
}

// Types for better type safety
interface Agent {
    id: string
    name: string
    ip_address: string
    port?: number
    status: 'online' | 'offline' | 'busy'
    capabilities?: string
    gpu_info?: string
    last_seen: string
    endpoint?: string
    created_at: string
    updated_at: string
}

interface Job {
    id: string
    name: string
    status: 'pending' | 'running' | 'completed' | 'failed'
    progress?: number
    hash_file_name?: string
    hash_file_id?: string
    wordlist_name?: string
    wordlist_id?: string
    agent_name?: string
    agent_id?: string
    created_at: string
    command?: string
}

interface HashFile {
    id: string
    name: string
    orig_name: string
    type: string
    size: number
    path?: string
    created_at: string
}

interface Wordlist {
    id: string
    name: string
    orig_name: string
    size: number
    word_count?: number
    path?: string
    created_at: string
}

// Dashboard Application Class
class DashboardApplication {
    private isInitialized = false
    private alpineDataRegistered = false
    private config = getConfig()

    // Initialize the entire application
    public async init(): Promise<void> {
        if (this.isInitialized) {
            console.log('⚠️ Application already initialized, skipping...')
            return
        }

        try {
            console.log('🚀 Initializing Dashboard Application...')
            console.log('📍 Current DOM state:', document.body.innerHTML.length, 'characters')
            perf.startTimer('app-initialization')

            // 1. Wait for Alpine.js CDN to be available
            console.log('⏳ Waiting for Alpine.js to be available...')
            await this.waitForAlpine()
            console.log('✅ Alpine.js CDN is available')

            // 2. Initialize Alpine.js data BEFORE any DOM changes
            this.initializeAlpine()
            console.log('✅ Alpine.js data registered')
            
            // 3. Setup body with Alpine data BEFORE starting Alpine
            this.setupAlpineOnBody()
            console.log('✅ Alpine.js body setup complete')

            // 4. Start Alpine.js BEFORE component injection
            if (!window.alpineManuallyStarted && window.Alpine && typeof window.Alpine.start === 'function') {
                try {
                    window.alpineManuallyStarted = true
                    console.log('🚀 Starting Alpine.js manually...')
                    window.Alpine.start()
                    console.log('✅ Alpine.js started successfully')
                } catch (error) {
                    window.alpineManuallyStarted = false
                    console.error('❌ Alpine start failed:', error instanceof Error ? error.message : String(error))
                    throw error
                }
            } else {
                console.log('ℹ️ Alpine already manually started or not available')
            }

            // 5. Initialize services and data stores
            await this.initializeServices()
            
            // 6. Register event listeners
            this.setupEventListeners()

            // 7. Load and inject HTML components AFTER Alpine is fully ready
            await this.loadComponents()

            this.isInitialized = true
            perf.endTimer('app-initialization')
            console.log('✅ Dashboard Application initialized successfully')
            console.log('📍 Final DOM state:', document.body.innerHTML.length, 'characters')
            console.log('📊 Body classes:', document.body.className)
            console.log('🔍 Main containers:', document.querySelectorAll('main').length)

        } catch (error) {
            console.error('❌ Failed to initialize dashboard:', error)
            this.showErrorState()
        }
        
        // Fallback: Show debug info if nothing renders after 5 seconds
        setTimeout(() => {
            const mainContainers = document.querySelectorAll('main')
            const hasContent = mainContainers.length > 0 && Array.from(mainContainers).some(main => main.children.length > 0)
            
            if (!hasContent) {
                console.warn('⚠️ No content rendered after 5 seconds, showing debug mode')
                const debugContainer = document.getElementById('debug-container')
                if (debugContainer) {
                    debugContainer.style.display = 'block'
                    const debugComponents = document.getElementById('debug-components')
                    const debugAlpine = document.getElementById('debug-alpine')
                    
                    if (debugComponents) debugComponents.textContent = `${mainContainers.length} main containers found`
                    if (debugAlpine) debugAlpine.textContent = window.Alpine ? 'Loaded' : 'Not loaded'
                }
            }
        }, 5000)
    }

    /**
     * Load HTML components dynamically
     */
    private async loadComponents(): Promise<void> {
        try {
            console.log('📦 Loading HTML components...')
            perf.startTimer('component-loading')
            
            // Load all components in parallel with proper paths
            const componentMap = [
                { name: 'layout/navigation', path: '/components/layout/navigation.html' },
                { name: 'tabs/overview', path: '/components/tabs/overview.html' },
                { name: 'tabs/agents', path: '/components/tabs/agents.html' },
                { name: 'tabs/jobs', path: '/components/tabs/jobs.html' },
                { name: 'tabs/files', path: '/components/tabs/files.html' },
                { name: 'tabs/wordlists', path: '/components/tabs/wordlists.html' },
                { name: 'tabs/docs', path: '/components/tabs/docs.html' },
                { name: 'modals/agent-modal', path: '/components/modals/agent-modal.html' },
                { name: 'modals/job-modal', path: '/components/modals/job-modal.html' },
                { name: 'modals/file-modal', path: '/components/modals/file-modal.html' },
                { name: 'modals/wordlist-modal', path: '/components/modals/wordlist-modal.html' },
                { name: 'ui/notification', path: '/components/ui/notification.html' },
                { name: 'ui/loading', path: '/components/ui/loading.html' }
            ]

            await Promise.all(
                componentMap.map(({ name, path }) => 
                    componentLoader.loadComponent(name, path)
                )
            )

            // Inject components into the page
            await this.injectComponents()
            
            perf.endTimer('component-loading')
            console.log('✅ HTML components loaded successfully')
        } catch (error) {
            console.error('❌ Failed to load components:', error)
            throw error
        }
    }

    /**
     * Inject loaded components into the DOM
     */
    private async injectComponents(): Promise<void> {
        console.log('🔧 Starting component injection...')
        
        // Find main container or create it
        let mainContainer = document.querySelector('main.container-modern')
        if (!mainContainer) {
            console.log('📦 Creating main container...')
            // Create main container if it doesn't exist
            mainContainer = document.createElement('main')
            mainContainer.className = 'container-modern'
            document.body.appendChild(mainContainer)
        } else {
            console.log('📦 Found existing main container')
        }

        // Load navigation
        console.log('🧭 Loading navigation component...')
        const navigation = await componentLoader.loadComponent('layout/navigation')
        console.log('✅ Navigation loaded:', navigation.length, 'characters')
        
        const navigationContainer = document.createElement('div')
        navigationContainer.innerHTML = navigation
        
        // Find the actual nav element (skip script tags)
        const navElement = navigationContainer.querySelector('nav')
        if (navElement) {
            document.body.insertBefore(navElement, mainContainer)
            console.log('✅ Navigation injected into DOM')
        } else {
            console.error('❌ Failed to find nav element in navigation component')
            console.log('📝 Navigation content preview:', navigation.substring(0, 200) + '...')
        }

        // Load tab content
        const tabComponents = [
            'tabs/overview',
            'tabs/agents', 
            'tabs/jobs',
            'tabs/files',
            'tabs/wordlists',
            'tabs/docs'
        ]

        for (const component of tabComponents) {
            const html = await componentLoader.loadComponent(component)
            const container = document.createElement('div')
            container.innerHTML = html
            // Find actual content element (skip script tags)
            const element = container.querySelector('section, div, article') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                mainContainer.appendChild(element)
                console.log(`✅ Injected ${component} component`)
            } else {
                console.warn(`❌ Failed to inject ${component} component`)
            }
        }

        // Load modals
        const modalComponents = [
            'modals/agent-modal',
            'modals/job-modal',
            'modals/file-modal', 
            'modals/wordlist-modal'
        ]

        for (const component of modalComponents) {
            const html = await componentLoader.loadComponent(component)
            const container = document.createElement('div')
            container.innerHTML = html
            // Find actual content element (skip script tags)
            const element = container.querySelector('div[x-data], [x-show]') || container.querySelector('div') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                document.body.appendChild(element)
                console.log(`✅ Injected ${component} modal`)
            } else {
                console.warn(`❌ Failed to inject ${component} modal`)
            }
        }

        // Load UI components
        const uiComponents = [
            'ui/notification',
            'ui/loading'
        ]

        for (const component of uiComponents) {
            const html = await componentLoader.loadComponent(component)
            const container = document.createElement('div')
            container.innerHTML = html
            // Find actual content element (skip script tags)
            const element = container.querySelector('div, section') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                document.body.appendChild(element)
                console.log(`✅ Injected ${component} UI component`)
            } else {
                console.warn(`❌ Failed to inject ${component} UI component`)
            }
        }
    }

    /**
     * Set up Alpine.js on the body element
     */
    private setupAlpineOnBody(): void {
        document.body.setAttribute('x-data', 'dashboardApp()')
        document.body.setAttribute('x-init', 'init()')
    }

    /**
     * Wait for Alpine.js to be available from CDN
     */
    private waitForAlpine(): Promise<void> {
        return new Promise((resolve) => {
            if (window.Alpine || window.alpineReady) {
                resolve()
                return
            }
            
            const checkAlpine = () => {
                if (window.Alpine || window.alpineReady) {
                    resolve()
                } else {
                    setTimeout(checkAlpine, 50)
                }
            }
            checkAlpine()
        })
    }

    // Initialize Alpine.js with global dashboard state
    private initializeAlpine(): void {
        perf.startTimer('alpine-initialization')

        // Ensure Alpine is available (it should be from our wait)
        if (!window.Alpine) {
            console.warn('Alpine not available during initialization')
            return
        }

        // Prevent duplicate data registration
        if (this.alpineDataRegistered) {
            console.log('Alpine data already registered, skipping...')
            return
        }
        this.alpineDataRegistered = true

        // Global dashboard data and methods  
        const self = this
        window.Alpine.data('dashboardApp', () => ({
            // State
            currentTab: 'overview',
            isLoading: false,
            notifications: [] as Array<{id: number, message: string, type: string, timestamp: Date}>,
            isAlpineInitialized: false,
            
            // Modal states
            showAgentModal: false,
            showJobModal: false,
            showFileModal: false,
            showWordlistModal: false,
            
            // Form data
            agentForm: {
                name: '',
                ip_address: '',
                port: 8080,
                capabilities: ''
            },
            jobForm: {
                name: '',
                hash_file_id: '',
                wordlist_id: '',
                hash_type: '',
                attack_mode: '',
                agent_id: ''
            },
            fileForm: {
                file: null as File | null
            },
            wordlistForm: {
                file: null as File | null
            },
            
            // Other data
            commandTemplate: '',
            notification: {
                show: false,
                type: 'info',
                message: ''
            },
            
            // Data from stores with safe defaults
            get agents() { 
                return agentStore.getState().agents || []
            },
            get jobs() { 
                return jobStore.getState().jobs || []
            },
            get hashFiles() { 
                return fileStore.getState().hashFiles || []
            },
            get wordlists() { 
                const state = wordlistStore.getState()
                return state.wordlists || []
            },
            
            // Computed properties with safe checks
            get onlineAgents() {
                const agents = this.agents
                return Array.isArray(agents) ? agents.filter((agent: any) => agent.status === 'online') : []
            },
            get runningJobs() {
                const jobs = this.jobs
                return Array.isArray(jobs) ? jobs.filter((job: any) => job.status === 'running') : []
            },

            // Methods
            async init() {
                if (this.isAlpineInitialized) return
                this.isAlpineInitialized = true
                
                console.log('🎯 Initializing Alpine.js dashboard state...')
                await this.loadInitialData()
                this.setupPolling()
            },

            async loadInitialData() {
                this.isLoading = true
                try {
                    // Load all initial data in parallel with individual error handling
                    const results = await Promise.allSettled([
                        agentStore.actions.fetchAgents(),
                        jobStore.actions.fetchJobs(),
                        fileStore.actions.fetchHashFiles(),
                        wordlistStore.actions.fetchWordlists()
                    ])
                    
                    // Log any failures without throwing
                    results.forEach((result, index) => {
                        const endpoints = ['agents', 'jobs', 'hash files', 'wordlists']
                        if (result.status === 'rejected') {
                            console.warn(`Failed to load ${endpoints[index]}:`, result.reason)
                        }
                    })
                } catch (error) {
                    console.error('Failed to load initial data:', error)
                } finally {
                    this.isLoading = false
                }
            },

            // Tab management
            async switchTab(tab: string) {
                if (this.currentTab === tab) return
                
                perf.startTimer(`tab-switch-${tab}`)
                this.currentTab = tab
                
                // Lazy load tab component if needed
                if (self.config.features.lazyLoading) {
                    const component = componentRegistry.getComponent(tab)
                    if (component?.lazy) {
                        await componentLoader.loadComponent(component.name, component.path)
                    }
                }
                
                perf.endTimer(`tab-switch-${tab}`)
            },

            // Utility methods
            formatDate(dateString: string) {
                if (!dateString) return 'No date'
                try {
                    const date = new Date(dateString)
                    if (isNaN(date.getTime())) return 'Invalid date'
                    return date.toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit'
                    })
                } catch (error) {
                    return 'Invalid date'
                }
            },
            
            formatFileSize(bytes: number | string) {
                const numBytes = Number(bytes)
                if (!numBytes || numBytes === 0) return '0 Bytes'
                const k = 1024
                const sizes = ['Bytes', 'KB', 'MB', 'GB']
                const i = Math.floor(Math.log(numBytes) / Math.log(k))
                return parseFloat((numBytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
            },
            
            formatNumber(num: number | string) {
                const numValue = Number(num)
                if (!numValue || isNaN(numValue)) return '0'
                return numValue.toLocaleString()
            },
            
            getDisplayName(item: any) {
                // Try different name fields
                return item?.orig_name || item?.name || 'Unknown File'
            },
            
            getFileExtension(filename: string) {
                if (!filename) return 'file'
                const lastDot = filename.lastIndexOf('.')
                return lastDot > 0 ? filename.substring(lastDot + 1).toLowerCase() : 'file'
            },
            
            // Safe ID truncation
            getTruncatedId(item: any) {
                const id = item?.id
                if (!id || typeof id !== 'string') return 'No ID'
                return id.length > 8 ? id.substring(0, 8) + '...' : id
            },
            
            // Safe array slice
            getSlicedArray(arr: any, start: number = 0, end: number = 5) {
                return Array.isArray(arr) ? arr.slice(start, end) : []
            },
            
            // Safe array check and length
            getArrayLength(arr: any) {
                return Array.isArray(arr) ? arr.length : 0
            },
            
            // Safe object property access
            getSafeProperty(obj: any, prop: string, defaultValue: any = 'N/A') {
                return obj && obj[prop] !== undefined && obj[prop] !== null ? obj[prop] : defaultValue
            },

            // Agent-specific helpers
            getAgentGpuInfo(agent: any) {
                return agent?.gpu_info || agent?.capabilities || 'No GPU info available'
            },

            getAgentCapabilities(agent: any) {
                return agent.capabilities || 'General Purpose'
            },

            // NEW: Get job count for agent
            getAgentJobCount(agent: any): string {
                if (!agent.id) return '0 jobs'
                const assignedJobs = this.jobs.filter((job: any) => job.agent_id === agent.id)
                const runningJobs = assignedJobs.filter((job: any) => job.status === 'running')
                const pendingJobs = assignedJobs.filter((job: any) => job.status === 'pending')
                
                if (runningJobs.length > 0) {
                    return `${runningJobs.length} running, ${pendingJobs.length} pending`
                } else if (pendingJobs.length > 0) {
                    return `${pendingJobs.length} pending jobs`
                } else {
                    return 'No active jobs'
                }
            },

            // File type badge helpers
            getFileTypeBadge(filename: string) {
                const ext = this.getFileExtension(filename)
                switch (ext) {
                    case 'txt': return 'text'
                    case 'hash': return 'hash'
                    case 'lst': return 'list'
                    default: return ext || 'file'
                }
            },

            showNotification(message: string, type: 'success' | 'error' | 'info' = 'info') {
                const notification = {
                    id: Date.now(),
                    message,
                    type,
                    timestamp: new Date()
                }
                this.notifications.unshift(notification)
                
                // Auto-remove after 5 seconds
                setTimeout(() => {
                    this.removeNotification(notification.id)
                }, 5000)
            },

            removeNotification(id: number) {
                this.notifications = this.notifications.filter((n: any) => n.id !== id)
            },

            copyToClipboard(text: string) {
                navigator.clipboard.writeText(text).then(() => {
                    this.showNotification('Copied to clipboard!', 'success')
                }).catch(() => {
                    this.showNotification('Failed to copy', 'error')
                })
            },

            scrollToSection(id: string) {
                const element = document.getElementById(id)
                if (element) {
                    element.scrollIntoView({ behavior: 'smooth' })
                }
            },

            setupPolling() {
                // Poll for updates every 30 seconds
                setInterval(async () => {
                    if (!this.isLoading && this.currentTab === 'overview') {
                        await Promise.all([
                            agentStore.actions.fetchAgents(),
                            jobStore.actions.fetchJobs()
                        ])
                    }
                }, 30000)
            },

            // Modal actions
            async openAgentModal() {
                this.showAgentModal = true
                this.agentForm = { name: '', ip_address: '', port: 8080, capabilities: '' }
            },
            
            closeAgentModal() {
                this.showAgentModal = false
            },

            async createAgent(agentData: any) {
                const result = await agentStore.actions.createAgent(agentData)
                if (result) {
                    this.showNotification('Agent created successfully!', 'success')
                    this.closeAgentModal()
                } else {
                    this.showNotification('Failed to create agent', 'error')
                }
            },

            async openJobModal() {
                this.showJobModal = true
                this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', hash_type: '', attack_mode: '', agent_id: '' }
            },
            
            closeJobModal() {
                this.showJobModal = false
            },

            async createJob(jobData: any) {
                try {
                    this.isLoading = true
                    // Enhanced job creation with agent assignment
                    const jobPayload = {
                        name: jobData.name,
                        hash_type: parseInt(jobData.hash_type),
                        attack_mode: parseInt(jobData.attack_mode),
                        hash_file_id: jobData.hash_file_id,
                        wordlist_id: jobData.wordlist_id,
                        agent_id: jobData.agent_id || undefined // Include agent assignment if specified
                    }
                    
                    const result = await jobStore.actions.createJob(jobPayload)
                    if (result) {
                        this.showNotification('Job created successfully!', 'success')
                        this.showJobModal = false
                        this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', hash_type: '', attack_mode: '', agent_id: '' }
                    }
                } catch (error) {
                    this.showNotification('Failed to create job', 'error')
                } finally {
                    this.isLoading = false
                }
            },
            
            async openFileModal() {
                this.showFileModal = true
                this.fileForm = { file: null }
            },
            
            closeFileModal() {
                this.showFileModal = false
            },
            
            async openWordlistModal() {
                this.showWordlistModal = true
                this.wordlistForm = { file: null }
            },
            
            closeWordlistModal() {
                this.showWordlistModal = false
            },

            async startJob(jobId: string) {
                const success = await jobStore.actions.startJob(jobId)
                if (success) {
                    this.showNotification('Job started successfully!', 'success')
                } else {
                    this.showNotification('Failed to start job', 'error')
                }
            },

            async stopJob(jobId: string) {
                const success = await jobStore.actions.stopJob(jobId)
                if (success) {
                    this.showNotification('Job stopped successfully!', 'success')
                } else {
                    this.showNotification('Failed to stop job', 'error')
                }
            },

            // File actions
            async uploadHashFile(file: File) {
                const result = await fileStore.actions.uploadHashFile(file)
                if (result) {
                    this.showNotification('Hash file uploaded successfully!', 'success')
                    this.closeFileModal()
                } else {
                    this.showNotification('Failed to upload hash file', 'error')
                }
            },

            async uploadWordlist(file: File) {
                const result = await wordlistStore.actions.uploadWordlist(file)
                if (result) {
                    this.showNotification('Wordlist uploaded successfully!', 'success')
                    this.closeWordlistModal()
                } else {
                    this.showNotification('Failed to upload wordlist', 'error')
                }
            },

            // Delete actions
            async deleteAgent(id: string) {
                if (!id) {
                    this.showNotification('Error: No agent ID provided', 'error')
                    return
                }
                if (confirm('Are you sure you want to delete this agent?')) {
                    const success = await agentStore.actions.deleteAgent(id)
                    if (success) {
                        this.showNotification('Agent deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete agent', 'error')
                    }
                }
            },

            async deleteJob(id: string) {
                if (!id) {
                    this.showNotification('Error: No job ID provided', 'error')
                    return
                }
                if (confirm('Are you sure you want to delete this job?')) {
                    const success = await jobStore.actions.deleteJob(id)
                    if (success) {
                        this.showNotification('Job deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete job', 'error')
                    }
                }
            },

            async deleteFile(id: string) {
                if (!id) {
                    this.showNotification('Error: No file ID provided', 'error')
                    return
                }
                if (confirm('Are you sure you want to delete this hash file?')) {
                    const success = await fileStore.actions.deleteHashFile(id)
                    if (success) {
                        this.showNotification('Hash file deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete hash file', 'error')
                    }
                }
            },

            async deleteWordlist(id: string) {
                if (!id) {
                    this.showNotification('Error: No wordlist ID provided', 'error')
                    return
                }
                if (confirm('Are you sure you want to delete this wordlist?')) {
                    const success = await wordlistStore.actions.deleteWordlist(id)
                    if (success) {
                        this.showNotification('Wordlist deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete wordlist', 'error')
                    }
                }
            },

            // Download actions 
            async downloadFile(id: string, filename: string) {
                if (!id) {
                    this.showNotification('Error: No file ID provided', 'error')
                    return
                }
                try {
                    const blob = await apiService.downloadHashFile(id)
                    if (blob) {
                        const url = window.URL.createObjectURL(blob)
                        const a = document.createElement('a')
                        a.href = url
                        a.download = filename || 'file'
                        document.body.appendChild(a)
                        a.click()
                        window.URL.revokeObjectURL(url)
                        document.body.removeChild(a)
                        this.showNotification('File downloaded successfully!', 'success')
                    } else {
                        this.showNotification('Failed to download file', 'error')
                    }
                } catch (error) {
                    this.showNotification('Failed to download file', 'error')
                }
            },

            async downloadWordlist(id: string, filename: string) {
                if (!id) {
                    this.showNotification('Error: No wordlist ID provided', 'error')
                    return
                }
                try {
                    const blob = await apiService.downloadWordlist(id)
                    if (blob) {
                        const url = window.URL.createObjectURL(blob)
                        const a = document.createElement('a')
                        a.href = url
                        a.download = filename || 'wordlist'
                        document.body.appendChild(a)
                        a.click()
                        window.URL.revokeObjectURL(url)
                        document.body.removeChild(a)
                        this.showNotification('Wordlist downloaded successfully!', 'success')
                    } else {
                        this.showNotification('Failed to download wordlist', 'error')
                    }
                } catch (error) {
                    this.showNotification('Failed to download wordlist', 'error')
                }
            }
        }))

        perf.endTimer('alpine-initialization')
    }

    // Initialize services and API connections
    private async initializeServices(): Promise<void> {
        perf.startTimer('service-initialization')
        
        try {
            // Test API connection
            const health = await apiService.checkHealth()
            if (!health) {
                console.warn('⚠️ API connection failed - running in offline mode')
            }

        } catch (error) {
            console.warn('Service initialization warning:', error)
        } finally {
            perf.endTimer('service-initialization')
        }
    }

    // Setup global event listeners
    private setupEventListeners(): void {
        // Handle errors
        window.addEventListener('error', (event) => {
            console.error('Global error:', event.error)
        })

        // Handle unhandled promise rejections
        window.addEventListener('unhandledrejection', (event) => {
            console.error('Unhandled promise rejection:', event.reason)
        })
    }

    // Show error state if initialization fails
    private showErrorState(): void {
        document.body.innerHTML = `
            <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-red-50 to-red-100">
                <div class="text-center p-8">
                    <div class="inline-flex items-center justify-center w-20 h-20 rounded-3xl bg-red-500 text-white shadow-2xl mb-6">
                        <i class="fas fa-exclamation-triangle text-2xl"></i>
                    </div>
                    <h1 class="text-2xl font-bold text-red-900 mb-2">Failed to Load Dashboard</h1>
                    <p class="text-red-700 mb-4">Please check the console for more details</p>
                    <button onclick="location.reload()" class="px-6 py-3 bg-red-600 text-white rounded-xl hover:bg-red-700 transition-colors">
                        <i class="fas fa-redo mr-2"></i>
                        Retry
                    </button>
                </div>
            </div>
        `
    }
}

// Initialize application when DOM is ready
const dashboardApp = new DashboardApplication()

// Export for global access
declare global {
    interface Window {
        dashboardApp: DashboardApplication
    }
}

window.dashboardApp = dashboardApp

// Prevent multiple initialization calls
let initializationStarted = false

// Auto-initialize when DOM is loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        if (!initializationStarted) {
            initializationStarted = true
            dashboardApp.init()
        }
    })
} else {
    if (!initializationStarted) {
        initializationStarted = true
        dashboardApp.init()
    }
}

// Export for modules
export { dashboardApp } 
