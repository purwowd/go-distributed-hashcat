// Alpine.js is loaded via CDN and available as window.Alpine
import { componentLoader } from './utils/component-loader'
import { componentRegistry, perf, getConfig } from './config/build.config'
import { router } from './utils/router'

// Import all services and stores
import { apiService } from './services/api.service'
import { webSocketService } from './services/websocket.service'
// import './services/websocket-mock.service' // Auto-start mock in development (DISABLED)
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
    port?: number | string
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
    status: 'pending' | 'running' | 'completed' | 'failed' | 'paused' | 'cancelled'
    progress?: number
    hash_file_name?: string
    hash_file_id?: string
    wordlist_name?: string
    wordlist_id?: string
    agent_name?: string
    agent_id?: string
    created_at: string
    updated_at?: string
    started_at?: string
    completed_at?: string
    // NEW: Missing backend fields
    hash_type?: number
    attack_mode?: number
    rules?: string
    speed?: number
    eta?: string
    result?: string
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
    private lastAgentStatuses = new Map<string, string>() // Track previous status to prevent duplicate notifications

    // Initialize the entire application
    public async init(): Promise<void> {
        if (this.isInitialized) {
            // console.log('⚠️ Application already initialized, skipping...')
            return
        }

        try {
            // console.log('🚀 Initializing Dashboard Application...')
            // console.log('📍 Current DOM state:', document.body.innerHTML.length, 'characters')
            perf.startTimer('app-initialization')

            // 1. Wait for Alpine.js CDN to be available
            // console.log('⏳ Waiting for Alpine.js to be available...')
            await this.waitForAlpine()
            // console.log('✅ Alpine.js CDN is available')

            // 2. Initialize Alpine.js data BEFORE any DOM changes
            this.initializeAlpine()
            // console.log('✅ Alpine.js data registered')
            
            // 3. Setup body with Alpine data BEFORE starting Alpine
            this.setupAlpineOnBody()
            // console.log('✅ Alpine.js body setup complete')

            // 4. Start Alpine.js BEFORE component injection
            if (!window.alpineManuallyStarted && window.Alpine && typeof window.Alpine.start === 'function') {
                try {
                    window.alpineManuallyStarted = true
                    // console.log('🚀 Starting Alpine.js manually...')
                    window.Alpine.start()
                    // console.log('✅ Alpine.js started successfully')
                } catch (error) {
                    window.alpineManuallyStarted = false
                    // console.error('❌ Alpine start failed:', error instanceof Error ? error.message : String(error))
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
            // console.log('✅ Dashboard Application initialized successfully')
            // console.log('📍 Final DOM state:', document.body.innerHTML.length, 'characters')
            // console.log('📊 Body classes:', document.body.className)
            // console.log('🔍 Main containers:', document.querySelectorAll('main').length)

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
            // console.log('📦 Loading HTML components...')
            perf.startTimer('component-loading')
            
            // Load all components in parallel with proper paths
            const componentMap = [
                { name: 'layout/navigation', path: '/components/layout/navigation.html' },
                { name: 'ui/breadcrumb', path: '/components/ui/breadcrumb.html' },
                { name: 'tabs/overview', path: '/components/tabs/overview.html' },
                { name: 'tabs/agents', path: '/components/tabs/agents.html' },
                { name: 'tabs/agent-keys', path: '/components/tabs/agent-keys.html' },
                { name: 'tabs/jobs', path: '/components/tabs/jobs.html' },
                { name: 'tabs/files', path: '/components/tabs/files.html' },
                { name: 'tabs/wordlists', path: '/components/tabs/wordlists.html' },
                { name: 'tabs/docs', path: '/components/tabs/docs.html' },
                { name: 'modals/agent-modal', path: '/components/modals/agent-modal.html' },
                { name: 'modals/agent-key-modal', path: '/components/modals/agent-key-modal.html' },
                { name: 'modals/delete-confirm-modal', path: '/components/modals/delete-confirm-modal.html' },
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
            // console.log('✅ HTML components loaded successfully')
        } catch (error) {
            console.error('❌ Failed to load components:', error)
            throw error
        }
    }

    /**
     * Inject loaded components into the DOM
     */
    private async injectComponents(): Promise<void> {
        // console.log('🔧 Starting component injection...')
        
        // Find main container or create it
        let mainContainer = document.querySelector('main.container-modern')
        if (!mainContainer) {
            // console.log('📦 Creating main container...')
            // Create main container if it doesn't exist
            mainContainer = document.createElement('main')
            mainContainer.className = 'container-modern'
            document.body.appendChild(mainContainer)
        } else {
            console.log('📦 Found existing main container')
        }

        // Load navigation
        // console.log('🧭 Loading navigation component...')
        const navigation = await componentLoader.loadComponent('layout/navigation')
        // console.log('✅ Navigation loaded:', navigation.length, 'characters')
        
        const navigationContainer = document.createElement('div')
        navigationContainer.innerHTML = navigation
        
        // Find the actual nav element (skip script tags)
        const navElement = navigationContainer.querySelector('nav')
        if (navElement) {
            document.body.insertBefore(navElement, mainContainer)
            // console.log('✅ Navigation injected into DOM')
        } else {
            console.error('❌ Failed to find nav element in navigation component')
            // console.log('📝 Navigation content preview:', navigation.substring(0, 200) + '...')
        }

        // Load breadcrumb component
        // console.log('🗂️ Loading breadcrumb component...')
        const breadcrumb = await componentLoader.loadComponent('ui/breadcrumb')
        // console.log('✅ Breadcrumb loaded:', breadcrumb.length, 'characters')
        
        const breadcrumbContainer = document.createElement('div')
        breadcrumbContainer.innerHTML = breadcrumb
        
        // Find the actual nav element for breadcrumb
        const breadcrumbElement = breadcrumbContainer.querySelector('nav')
        if (breadcrumbElement) {
            document.body.insertBefore(breadcrumbElement, mainContainer)
            // console.log('✅ Breadcrumb injected into DOM')
        } else {
            console.error('❌ Failed to find nav element in breadcrumb component')
        }

        // Load tab content
        const tabComponents = [
            'tabs/overview',
            'tabs/agents',
            'tabs/agent-keys',
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
                // console.log(`✅ Injected ${component} component`)
            } else {
                console.warn(`❌ Failed to inject ${component} component`)
            }
        }

        // Load modals
        const modalComponents = [
            'modals/agent-modal',
            'modals/agent-key-modal',
            'modals/delete-confirm-modal',
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
                // console.log(`✅ Injected ${component} modal`)
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
                // console.log(`✅ Injected ${component} UI component`)
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
            // Reactive state
            currentTab: router.getCurrentRoute(),
            isLoading: false,
            isAlpineInitialized: false,
            notifications: [] as any[],
            
            // Modal states
            showAgentModal: false,
            showAgentKeyModal: false,
            showDeleteModal: false,
            showAgentKeys: false,
            showJobModal: false,
            currentStep: 1, // 1: Basic Config, 2: Distribution Preview
            showFileModal: false,
            showWordlistModal: false,
            showDistributedJobModal: false,
            
            // Compact mode for agent selection
            isCompactMode: false,
            
            // Form states
            agentForm: { ip_address: '', port: null as number | null, capabilities: '', agent_key: '' },
            agentKeyForm: { name: '', agent_key: '' },
            createdAgent: null as any,
            createdAgentKey: null as any,
            deleteModalConfig: { entityType: '', entityName: '', description: '', warning: '', entityId: '', confirmAction: null as any },
            jobForm: { name: '', hash_file_id: '', wordlist_id: '', agent_ids: [] as string[], hash_type: '', attack_mode: '' },
            distributedJobForm: { name: '', hash_file_id: '', wordlist_id: '', hash_type: '', attack_mode: '', auto_distribute: true },
            fileForm: { file: null },
            wordlistForm: { file: null },
            
            // Command template for job creation
            commandTemplate: '',
            distributedCommandTemplate: '',
            
            // Manual loading reset (fallback)
            forceStopLoading() {
                this.isLoading = false
                console.log('🛑 Loading force stopped by user')
            },
            
            // Cache stats (if needed)
            cacheStats: null as any,
            
            // WebSocket connection status
            wsConnected: false,
            wsConnectionAttempts: 0,
            
            // Track agent status changes to prevent duplicate notifications
            lastAgentStatuses: new Map(),
            
            // Reactive data arrays - these will be updated by store subscriptions
            reactiveAgents: [] as any[],
            reactiveAgentKeys: [] as any[],
            reactiveJobs: [] as any[],
            reactiveHashFiles: [] as any[],
            reactiveWordlists: [] as any[],

            // Server-side table state for Agents/Agent-Keys
            agentTable: {
                page: 1,
                pageSize: 10,
                search: '',
                total: 0
            },
            
            // Server-side table state for Jobs
            jobTable: {
                page: 1,
                pageSize: 10,
                search: '',
                total: 0
            },

            // Getters that return reactive data
            get agents() {
                return this.reactiveAgents || []
            },
            get agentKeys() {
                return this.reactiveAgentKeys || []
            },
            get jobs() {
                return this.reactiveJobs || []
            },
            get hashFiles() {
                return this.reactiveHashFiles || []
            },
            get wordlists() {
                return this.reactiveWordlists || []
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
            get pendingJobs() {
                const jobs = this.jobs
                return Array.isArray(jobs) ? jobs.filter((job: any) => job.status === 'pending') : []
            },
            
            // Computed property for selected wordlist count
            get selectedWordlistCount() {
                if (!this.jobForm.wordlist_id) {
                    return '0'
                }
                
                const selectedWordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                if (!selectedWordlist) {
                    return '0'
                }
                
                // Return word count if available, otherwise return size in KB
                if (selectedWordlist.word_count && selectedWordlist.word_count > 0) {
                    return selectedWordlist.word_count.toLocaleString()
                } else if (selectedWordlist.size) {
                    const sizeKB = Math.round(selectedWordlist.size / 1024)
                    return `${sizeKB} KB`
                }
                
                return '0'
            },


            
            // Computed properties for hash type based on selected file
            get selectedHashTypeValue() {
                if (!this.jobForm.hash_file_id) {
                    return '2500' // Default for WPA/WPA2
                }
                
                const selectedFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                if (!selectedFile) {
                    return '2500'
                }
                
                // Determine hash type based on file type
                const fileType = selectedFile.type?.toLowerCase() || ''
                switch (fileType) {
                    case 'hccapx':
                    case 'hccap':
                    case 'cap':
                    case 'pcap':
                        return '2500' // WPA/WPA2
                    case 'hash':
                    default:
                        return '0' // Generic hash
                }
            },
            
            get selectedHashTypeName() {
                if (!this.jobForm.hash_file_id) {
                    return 'WPA/WPA2'
                }
                
                const selectedFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                if (!selectedFile) {
                    return 'WPA/WPA2'
                }
                
                // Determine hash type name based on file type
                const fileType = selectedFile.type?.toLowerCase() || ''
                switch (fileType) {
                    case 'hccapx':
                    case 'hccap':
                    case 'cap':
                    case 'pcap':
                        return 'WPA/WPA2'
                    case 'hash':
                    default:
                        return 'Generic Hash'
                }
            },
            
            get selectedHashTypeDescription() {
                if (!this.jobForm.hash_file_id) {
                    return 'This system is specialized for WPA/WPA2 cracking'
                }
                
                const selectedFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                if (!selectedFile) {
                    return 'This system is specialized for WPA/WPA2 cracking'
                }
                
                // Determine description based on file type
                const fileType = selectedFile.type?.toLowerCase() || ''
                switch (fileType) {
                    case 'hccapx':
                        return 'WiFi handshake file (.hccapx) - WPA/WPA2 cracking'
                    case 'hccap':
                        return 'Legacy WiFi handshake file (.hccap) - WPA/WPA2 cracking'
                    case 'cap':
                    case 'pcap':
                        return 'WiFi packet capture file - WPA/WPA2 cracking'
                    case 'hash':
                    default:
                        return 'Generic hash file - various hash types supported'
                }
            },

            // Methods
            async init() {
                // console.log('🔄 Initializing Alpine.js dashboard data...')
                this.isAlpineInitialized = true
                
                // Setup router listener
                router.subscribe((route: string) => {
                    this.currentTab = route
                })
                
                // Setup store subscriptions for reactivity
                this.setupStoreSubscriptions()
                
                // Setup WebSocket for real-time updates
                this.setupWebSocketSubscriptions()
                
                try {
                    await this.loadInitialData()
                    // console.log('🎉 Dashboard initialization complete')
                } catch (error) {
                    console.error('❌ Dashboard initialization failed:', error)
                    this.showNotification('Failed to initialize dashboard', 'error')
                }
                
                this.setupPolling()
                
                // Safety timeout to prevent infinite loading
                setTimeout(() => {
                    if (this.isLoading) {
                        console.warn('⚠️ Loading timeout reached, forcing stop')
                        this.forceStopLoading()
                        this.showNotification('Loading took too long. Data may be incomplete.', 'warning')
                    }
                }, 15000) // 15 second safety timeout
            },

            // NEW: Setup store subscriptions for reactive updates
            setupStoreSubscriptions() {
                // Subscribe to agent store changes
                agentStore.subscribe(() => {
                    const state = agentStore.getState()
                    
                    // ✅ Implement stable sorting to maintain card positions
                    const agents = state.agents || []
                    // Sort by created_at DESC, then by ID ASC for stable ordering
                    const stableSortedAgents = agents.sort((a, b) => {
                        // First sort by created_at DESC (newest first)
                        const dateComparison = new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
                        if (dateComparison !== 0) {
                            return dateComparison
                        }
                        // If dates are equal, sort by ID ASC for stable ordering
                        return a.id.localeCompare(b.id)
                    })
                    
                    // ✅ Force Alpine.js reactivity by creating new array reference
                    this.reactiveAgents = [...stableSortedAgents]
                    // Also update agentKeys with agents that have no IP address (these are just keys)
                    this.reactiveAgentKeys = [...stableSortedAgents].filter(agent => !agent.ip_address || agent.ip_address === '')
                    console.log('🔄 Agent store updated:', this.reactiveAgents.length, 'agents')

                    // Sync pagination (if available)
                    if (state.pagination) {
                        this.agentTable.total = state.pagination.total
                        // Keep page size and page if already set
                        if (state.pagination.pageSize && this.agentTable.pageSize !== state.pagination.pageSize) {
                            this.agentTable.pageSize = state.pagination.pageSize
                        }
                        if (state.pagination.page && this.agentTable.page !== state.pagination.page) {
                            this.agentTable.page = state.pagination.page
                        }
                    }
                })
                
                // Subscribe to job store changes
                jobStore.subscribe(() => {
                    const state = jobStore.getState()
                    // ✅ Force Alpine.js reactivity by creating new array reference
                    this.reactiveJobs = [...(state.jobs || [])]
                    console.log('🔄 Job store updated:', this.reactiveJobs.length, 'jobs')
                    
                    // Sync pagination (if available)
                    // Note: This would need to be updated when we implement job pagination in the backend
                })
                
                // Subscribe to file store changes
                fileStore.subscribe(() => {
                    const state = fileStore.getState()
                    this.reactiveHashFiles = [...(state.hashFiles || [])]
                })
                
                // Subscribe to wordlist store changes
                wordlistStore.subscribe(() => {
                    const state = wordlistStore.getState()
                    this.reactiveWordlists = [...(state.wordlists || [])]
                })
                
                // console.log('📡 Store subscriptions setup for reactive UI updates')
            },

            // Server-side table helpers
            async refreshAgentsTable() {
                await agentStore.actions.fetchAgents({
                    page: this.agentTable.page,
                    page_size: this.agentTable.pageSize,
                    search: this.agentTable.search
                })
            },
            async setAgentTablePageSize(event: any) {
                const val = parseInt(event?.target?.value || '10')
                this.agentTable.pageSize = isNaN(val) ? 10 : val
                this.agentTable.page = 1
                await this.refreshAgentsTable()
            },
            async setAgentTableSearch(event: any) {
                this.agentTable.search = event?.target?.value || ''
                this.agentTable.page = 1
                await this.refreshAgentsTable()
            },
            async goPrevAgentsPage() {
                if (this.agentTable.page > 1) {
                    this.agentTable.page -= 1
                    await this.refreshAgentsTable()
                }
            },
            async goNextAgentsPage() {
                const canNext = this.agentTable.page * this.agentTable.pageSize < (this.agentTable.total || 0)
                if (canNext) {
                    this.agentTable.page += 1
                    await this.refreshAgentsTable()
                }
            },

            // Server-side table helpers for Jobs
            async refreshJobsTable() {
                const result = await jobStore.actions.fetchJobs({
                    page: this.jobTable.page,
                    page_size: this.jobTable.pageSize,
                    search: this.jobTable.search
                })
                if (result) {
                    this.jobTable.total = result.total
                }
            },
            async setJobTablePageSize(event: any) {
                const val = parseInt(event?.target?.value || '10')
                this.jobTable.pageSize = isNaN(val) ? 10 : val
                this.jobTable.page = 1
                await this.refreshJobsTable()
            },
            async setJobTableSearch(event: any) {
                this.jobTable.search = event?.target?.value || ''
                this.jobTable.page = 1
                await this.refreshJobsTable()
            },
            async goPrevJobsPage() {
                if (this.jobTable.page > 1) {
                    this.jobTable.page -= 1
                    await this.refreshJobsTable()
                }
            },
            async goNextJobsPage() {
                const canNext = this.jobTable.page * this.jobTable.pageSize < (this.jobTable.total || 0)
                if (canNext) {
                    this.jobTable.page += 1
                    await this.refreshJobsTable()
                }
            },
            async goToJobsPage(page: number) {
                const totalPages = Math.max(1, Math.ceil((this.jobTable.total || 0) / (this.jobTable.pageSize || 10)))
                const target = Math.min(Math.max(1, page), totalPages)
                if (target !== this.jobTable.page) {
                    this.jobTable.page = target
                    await this.refreshJobsTable()
                }
            },
            async goToAgentsPage(page: number) {
                const totalPages = Math.max(1, Math.ceil((this.agentTable.total || 0) / (this.agentTable.pageSize || 10)))
                const target = Math.min(Math.max(1, page), totalPages)
                if (target !== this.agentTable.page) {
                    this.agentTable.page = target
                    await this.refreshAgentsTable()
                }
            },

            // NEW: Setup WebSocket subscriptions for real-time updates
            setupWebSocketSubscriptions() {
                // Connection status monitoring
                webSocketService.onConnection((status) => {
                    this.wsConnected = status.connected
                    if (status.connected) {
                        this.showNotification('🔗 Real-time updates connected', 'success')
                        // Subscribe to all updates when connected
                        webSocketService.subscribeToJobs()
                        webSocketService.subscribeToAgents()
                    } else {
                        this.showNotification('🔌 Real-time updates disconnected', 'warning')
                    }
                })
                
                // Job progress updates
                webSocketService.onJobProgress((update) => {
                    // Update specific job instead of fetching all
                    this.updateJobProgress(update)
                })
                
                // Job status changes (start, stop, complete, etc.)
                webSocketService.onJobStatus((update) => {
                    // Update specific job status instead of fetching all
                    this.updateJobStatus(update)
                })
                
                // Agent status updates - Real-time without API call
                webSocketService.onAgentStatus((update) => {
                    // ✅ Update agent status directly in store (no API call needed!)
                    if (update.agent_id && update.status) {
                        // Check if this is actually a status change
                        const previousStatus = this.lastAgentStatuses.get(update.agent_id)
                        const isStatusChange = previousStatus && previousStatus !== update.status
                        
                        // Update store
                        agentStore.actions.updateAgentStatus(
                            update.agent_id, 
                            update.status, 
                            update.last_seen
                        )
                        
                        // Update tracking
                        this.lastAgentStatuses.set(update.agent_id, update.status)
                        
                        // Show notification ONLY for actual status changes (not initial load)
                        if (isStatusChange) {
                            const agent = this.agents.find(a => a.id === update.agent_id)
                            const agentName = agent?.name || 'Agent'
                            
                            if (update.status === 'online') {
                                this.showNotification(`🟢 ${agentName} is now online`, 'success')
                            } else if (update.status === 'offline') {
                                this.showNotification(`🔴 ${agentName} went offline`, 'warning')
                            } else if (update.status === 'busy') {
                                this.showNotification(`🟡 ${agentName} is now busy`, 'info')
                            }
                        }
                    }
                })
                
                // Real-time notifications
                webSocketService.onNotification((notification) => {
                    this.showNotification(notification.message, notification.type || 'info')
                })
                
                // console.log('🌐 WebSocket subscriptions setup for real-time updates')
            },

            // Real-time job updates
            updateJobProgress(update: any) {
                // Update individual job via store action
                if (update.job_id) {
                    jobStore.actions.updateJobProgress(update.job_id, update.progress, update.speed, update.eta, update.status)
                }
            },

            updateJobStatus(update: any) {
                // Update individual job via store action
                if (update.job_id) {
                    jobStore.actions.updateJobStatus(update.job_id, update.status, update.result)
                    
                    // Show notification for important status changes
                    const job = this.jobs.find(j => j.id === update.job_id)
                    if (job) {
                        if (update.status === 'completed') {
                            this.showNotification(`🎉 Job "${job.name}" completed!`, 'success')
                        } else if (update.status === 'failed') {
                            this.showNotification(`❌ Job "${job.name}" failed`, 'error')
                        }
                    }
                }
            },

            async loadInitialData() {
                try {
                    this.isLoading = true
                    
                    // Add timeout to prevent infinite loading
                    const timeout = new Promise((_, reject) => 
                        setTimeout(() => reject(new Error('Loading timeout')), 10000)
                    )
                    
                    // Load with timeout protection
                    await Promise.race([
                        timeout,
                        Promise.all([
                            agentStore.actions.fetchAgents().catch(err => {
                                console.warn('Failed to load agents:', err)
                                return []
                            }),
                            fileStore.actions.fetchHashFiles().catch(err => {
                                console.warn('Failed to load hash files:', err)
                                return []
                            }),
                            wordlistStore.actions.fetchWordlists().catch(err => {
                                console.warn('Failed to load wordlists:', err)
                                return []
                            })
                        ])
                    ])
                    
                    // ✅ Initialize agent status tracking after agents are loaded
                    this.agents.forEach(agent => {
                        this.lastAgentStatuses.set(agent.id, agent.status)
                    })
                    
                    // Load jobs with separate timeout
                    const jobResult = await Promise.race([
                        timeout,
                        jobStore.actions.fetchJobs({
                            page: this.jobTable.page,
                            page_size: this.jobTable.pageSize,
                            search: this.jobTable.search
                        }).catch(err => {
                            console.warn('Failed to load jobs:', err)
                            return null
                        })
                    ])
                    
                    // Sync job pagination data
                    if (jobResult) {
                        this.jobTable.total = (jobResult as any).total
                    }
                    
                    // Load cache stats
                    await this.refreshCacheStats().catch(err => {
                        console.warn('Failed to load cache stats:', err)
                    })
                    
                } catch (error) {
                    console.error('❌ Failed to load initial data:', error)
                    this.showNotification('Failed to load data. Please refresh the page.', 'error')
                } finally {
                    this.isLoading = false
                }
            },

            // Tab management with router integration
            async switchTab(tab: string) {
                if (this.currentTab === tab) return
                
                perf.startTimer(`tab-switch-${tab}`)
                
                // Use router to navigate (this will update URL and current tab)
                router.navigate(tab)
                
                // Lazy load tab component if needed
                if (self.config.features.lazyLoading) {
                    const component = componentRegistry.getComponent(tab)
                    if (component?.lazy) {
                        await componentLoader.loadComponent(component.name, component.path)
                    }
                }
                
                perf.endTimer(`tab-switch-${tab}`)
            },

            // NEW: Get current route URL for linking
            getRouteUrl(route: string) {
                return router.getRouteUrl(route)
            },

            // NEW: Check if route is current (for active states)
            isCurrentRoute(route: string) {
                return router.isCurrentRoute(route)
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

            // Extract password from job result safely
            extractPassword(result: string | null | undefined): string {
                if (!result || typeof result !== 'string') {
                    return ''
                }
                if (result.includes('Password found:')) {
                    return result.replace('Password found: ', '').trim()
                }
                return result
            },

            // Check if job has found password
            hasFoundPassword(result: string | null | undefined): boolean {
                return !!(result && typeof result === 'string' && result.includes('Password found:'))
            },

            // Helper function to check if job can be started
            canStartJob(job: any): boolean {
                return job.status !== 'running' && 
                       job.status !== 'completed' && 
                       job.agent_name && 
                       job.agent_name !== 'Unassigned'
            },

            // Get start button tooltip
            getStartButtonTooltip(job: any): string {
                if (job.status === 'completed') return 'Job already completed'
                if (!job.agent_name || job.agent_name === 'Unassigned') return 'No agent assigned'
                if (job.status === 'running') return 'Job is running'
                if (job.status === 'paused') return 'Resume Job'
                if (job.status === 'failed') return 'Retry Job'
                return 'Start Job'
            },

            // Get start button text based on job status
            getStartButtonText(job: any): string {
                if (job.status === 'paused') return 'Resume'
                if (job.status === 'failed') return 'Retry'
                return 'Start'
            },

            // Get start button icon based on job status
            getStartButtonIcon(job: any): string {
                if (job.status === 'paused') return 'fas fa-play'
                if (job.status === 'failed') return 'fas fa-redo'
                return 'fas fa-play'
            },

            showNotification(message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info') {
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

            copyToClipboard(text: string, element?: HTMLElement) {
                // Clean the text - remove language indicators and copy prompts
                const cleanText = text
                    .replace(/^📋.*$/gm, '')  // Remove copy indicators
                    .replace(/^✅.*$/gm, '')  // Remove success indicators  
                    .replace(/^\s*bash\s*$/gm, '') // Remove language labels
                    .replace(/^\s*shell\s*$/gm, '')
                    .replace(/^\s*yaml\s*$/gm, '')
                    .trim()

                navigator.clipboard.writeText(cleanText).then(() => {
                    // Enhanced visual feedback
                    if (element) {
                        element.classList.add('copied', 'copy-success')
                        setTimeout(() => {
                            element.classList.remove('copied')
                        }, 2000)
                        setTimeout(() => {
                            element.classList.remove('copy-success')
                        }, 600)
                    }
                    this.showNotification('Code copied to clipboard!', 'success')
                }).catch(() => {
                    this.showNotification('Failed to copy to clipboard', 'error')
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
                        const [_, jobResult] = await Promise.all([
                            agentStore.actions.fetchAgents(),
                            jobStore.actions.fetchJobs({
                                page: this.jobTable.page,
                                page_size: this.jobTable.pageSize,
                                search: this.jobTable.search
                            })
                        ])
                        
                        // Sync job pagination data
                        if (jobResult) {
                            this.jobTable.total = (jobResult as any).total
                        }
                    }
                }, 30000)
            },

            // Modal actions
            async openAgentModal() {
                this.showAgentModal = true
                this.agentForm = { ip_address: '', port: null as number | null, capabilities: '', agent_key: '' }
                this.createdAgent = null
            },
            
            closeAgentModal() {
                this.showAgentModal = false
                this.createdAgent = null
            },
            
            async openAgentKeyModal() {
                this.showAgentKeyModal = true
                // Pre-generate an 8-char hex key on the client for instant UX; server will accept or generate if absent
                const pregenerated = Math.random().toString(16).slice(2, 10).padEnd(8, '0').slice(0,8)
                this.agentKeyForm = { name: '', agent_key: pregenerated }
                this.createdAgentKey = null
            },
            
            closeAgentKeyModal() {
                this.showAgentKeyModal = false
                this.createdAgentKey = null
            },

            // Generic Delete Modal actions
            openDeleteModal(entityType: string, entity: any, confirmAction: any) {
                const configs = {
                    'agent': {
                        entityType: 'Agent',
                        entityName: entity?.name || 'agent',
                        description: 'This action will remove the agent from the list. This operation cannot be undone.',
                        warning: 'This will permanently remove the agent and all associated data.'
                    },
                    'job': {
                        entityType: 'Job',
                        entityName: entity?.name || 'job',
                        description: 'This action will permanently delete the job and all associated data. This operation cannot be undone.',
                        warning: 'This will remove the job from the system permanently.'
                    },
                    'file': {
                        entityType: 'Hash File',
                        entityName: entity?.orig_name || entity?.name || 'file',
                        description: 'This action will permanently delete the hash file. This operation cannot be undone.',
                        warning: 'This will remove the file and all associated data permanently.'
                    },
                    'wordlist': {
                        entityType: 'Wordlist',
                        entityName: entity?.orig_name || entity?.name || 'wordlist',
                        description: 'This action will permanently delete the wordlist. This operation cannot be undone.',
                        warning: 'This will remove the wordlist and all associated data permanently.'
                    }
                }
                
                this.deleteModalConfig = {
                    ...configs[entityType as keyof typeof configs],
                    entityId: entity?.id,
                    confirmAction: confirmAction
                }
                this.showDeleteModal = true
            },
            closeDeleteModal() {
                this.showDeleteModal = false
                this.deleteModalConfig = { entityType: '', entityName: '', description: '', warning: '', entityId: '', confirmAction: null }
            },
            async confirmDelete() {
                if (this.deleteModalConfig.confirmAction) {
                    await this.deleteModalConfig.confirmAction()
                }
                this.closeDeleteModal()
            },

            async createAgent(agentData: any) {
                try {
                    this.isLoading = true
                    
                    // Validate required fields - only agent_key is required now
                    if (!agentData.agent_key) {
                    this.showNotification('Agent key is required. Please enter an agent key.', 'error')
                    return
                }
                
                    // Set default port 8080 if port is empty or null
                    const processedData = { ...agentData }
                    if (!processedData.port || processedData.port === '' || processedData.port === null || processedData.port === undefined) {
                        processedData.port = 8080
                        console.log('🔧 Setting default port 8080 for agent')
                    }
                
                // Use updateAgentData instead of createAgent to only update data without changing status
                const result = await agentStore.actions.updateAgentData(processedData)
                if (result.success) {
                    // Agent data updated successfully, show success notification and close modal
                    this.showNotification(result.message || 'Agent data updated successfully!', 'success')
                    this.closeAgentModal() // Close modal after success
                } else {
                    // Show specific error message
                    if (result.error) {
                        // Remove HTTP status prefix if present for better user experience
                        let userMessage = result.error
                        
                        // Remove HTTP status prefix if present
                        if (userMessage.includes('HTTP 400: Bad Request - ')) {
                            userMessage = userMessage.replace('HTTP 400: Bad Request - ', '')
                        } else if (userMessage.includes('HTTP 409: Conflict - ')) {
                            userMessage = userMessage.replace('HTTP 409: Conflict - ', '')
                        } else if (userMessage.includes('HTTP 500: Internal Server Error - ')) {
                            userMessage = userMessage.replace('HTTP 500: Internal Server Error - ', '')
                        }
                        
                        this.showNotification(userMessage, 'error')
                    } else {
                        this.showNotification('Failed to update agent data', 'error')
                    }
                }
                } catch (error) {
                    console.error('Error creating agent:', error)
                    this.showNotification('Failed to register agent', 'error')
                } finally {
                    this.isLoading = false
                }
            },


            async copyAgentKey(agentKey: string) {
                try {
                    await navigator.clipboard.writeText(agentKey)
                    this.showNotification('Agent key copied to clipboard!', 'success')
                } catch (err) {
                    console.error('Failed to copy agent key: ', err)
                    this.showNotification('Failed to copy agent key', 'error')
                }
            },
            
            async createAgentKey(agentKeyData: any) {
                // Use the new generateAgentKey action for creating agent keys
                const result = await agentStore.actions.generateAgentKey(agentKeyData.name)
                if (result) {
                    this.showNotification('Agent Key Generated Successfully!', 'success')
                    this.createdAgentKey = result
                    if (result.agent_key) {
                        this.agentKeyForm.agent_key = result.agent_key
                    }
                    // Auto-close modal after short delay
                    setTimeout(() => {
                        this.closeAgentKeyModal()
                    }, 400)
                } else {
                    const state = agentStore.getState()
                    const rawError = String(state.error || '')
                    const err = rawError.toLowerCase()
                    if (err.includes('already exists')) {
                        // Extract agent name if present: "already exists <name>"
                        const namePart = rawError.split('already exists')[1]?.trim() || agentKeyData.name
                        this.showNotification(`Agent name '${namePart}' already exists`, 'error')
                    } else if (state.error != null && String(state.error).trim() !== '') {
                        // Try to extract user-friendly message from error
                        let userMessage = String(state.error);
                        
                        // Remove HTTP status prefix if present
                        if (userMessage.includes('HTTP 400: Bad Request - ')) {
                            userMessage = userMessage.replace('HTTP 400: Bad Request - ', '');
                        } else if (userMessage.includes('HTTP 409: Conflict - ')) {
                            userMessage = userMessage.replace('HTTP 409: Conflict - ', '');
                        } else if (userMessage.includes('HTTP 500: Internal Server Error - ')) {
                            userMessage = userMessage.replace('HTTP 500: Internal Server Error - ', '');
                        }
                        
                        this.showNotification(userMessage, 'error')
                    } else {
                        this.showNotification('Failed to generate agent key', 'error')
                    }
                }
            },

            async openJobModal() {
                // Check if there are online agents
                if (this.onlineAgents.length === 0) {
                    this.showNotification('No online agents available. Please start an agent first before creating jobs.', 'warning')
                    return
                }
                
                this.showJobModal = true
                this.currentStep = 1
                this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', agent_ids: [], hash_type: '2500', attack_mode: '0' }
            },
            
            closeJobModal() {
                this.showJobModal = false
                this.currentStep = 1
            },

            // Step management functions
            goToStep(step: number) {
                if (step === 2 && !this.canProceedToStep2()) {
                    this.showNotification('Please complete all required fields before proceeding', 'warning')
                    return
                }
                this.currentStep = step
                if (step === 2) {
                    this.updateCommandTemplate()
                }
            },

            canProceedToStep2(): boolean {
                return !!(this.jobForm.name && 
                         this.jobForm.hash_file_id && 
                         this.jobForm.wordlist_id &&
                         this.jobForm.agent_ids && 
                         this.jobForm.agent_ids.length > 0)
            },

            getSelectedHashFileName(): string {
                if (!this.jobForm.hash_file_id) return 'Not selected'
                const hashFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                return hashFile ? (hashFile.orig_name || hashFile.name) : 'Not selected'
            },

            getSelectedWordlistName(): string {
                if (!this.jobForm.wordlist_id) return 'Not selected'
                const wordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                return wordlist ? (wordlist.orig_name || wordlist.name) : 'Not selected'
            },

            async createJob(jobData: any) {
                try {
                    this.isLoading = true
                    
                    // Get wordlist name for backend requirement
                    const selectedWordlist = this.wordlists.find((w: any) => w.id === jobData.wordlist_id)
                    const wordlistName = selectedWordlist ? (selectedWordlist.orig_name || selectedWordlist.name) : 'unknown.txt'
                    
                    // Enhanced job creation with agent assignment
                    const jobPayload = {
                        name: jobData.name,
                        hash_type: parseInt(jobData.hash_type),
                        attack_mode: parseInt(jobData.attack_mode),
                        hash_file_id: jobData.hash_file_id,
                        wordlist: wordlistName,                    // Required field for backend
                        wordlist_id: jobData.wordlist_id,         // Optional reference ID
                        agent_ids: jobData.agent_ids || []        // Include multiple agent assignments
                    }
                    
                    // Validate required fields before sending
                    if (!jobPayload.name || !jobPayload.hash_file_id || !jobPayload.wordlist || 
                        jobPayload.hash_type === undefined || jobPayload.attack_mode === undefined) {
                        this.showNotification('Please fill in all required fields', 'error')
                        return
                    }
                    
                    // Validate agent assignment is required
                    if (!jobPayload.agent_ids || jobPayload.agent_ids.length === 0) {
                        this.showNotification('Please select at least one agent to run this job', 'error')
                        return
                    }
                    
                    // Validate all selected agents are online
                    const selectedAgents = this.agents.filter((a: any) => jobPayload.agent_ids.includes(a.id))
                    if (selectedAgents.length !== jobPayload.agent_ids.length) {
                        this.showNotification('Some selected agents were not found', 'error')
                        return
                    }
                    
                    const offlineAgents = selectedAgents.filter((a: any) => a.status !== 'online')
                    if (offlineAgents.length > 0) {
                        const agentNames = offlineAgents.map((a: any) => a.name).join(', ')
                        this.showNotification(`Cannot create job: Agents "${agentNames}" are offline`, 'error')
                        return
                    }
                    

                    
                    const result = await jobStore.actions.createJob(jobPayload)
                    if (result) {
                        this.showNotification('Job created successfully!', 'success')
                        this.showJobModal = false
                        this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', agent_ids: [], hash_type: '2500', attack_mode: '0' }
                        
                        // Refresh jobs list to show the new job
                        await this.refreshJobsTable()
                    } else {
                        this.showNotification('Failed to create job - server returned null', 'error')
                    }
                } catch (error) {
                    console.error('Job creation error:', error)
                    const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred'
                    this.showNotification(`Failed to create job: ${errorMessage}`, 'error')
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

            // Update command template based on form inputs
            updateCommandTemplate() {
                // Set default values if not set
                if (!this.jobForm.hash_type) {
                    this.jobForm.hash_type = '2500'
                }
                if (!this.jobForm.attack_mode) {
                    this.jobForm.attack_mode = '0'
                }
                
                if (!this.jobForm.hash_file_id || !this.jobForm.wordlist_id || !this.jobForm.agent_ids || this.jobForm.agent_ids.length === 0) {
                    this.commandTemplate = 'hashcat command will appear here...'
                    return
                }

                // Get file names for display
                const hashFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                const wordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                
                const hashFileName = hashFile ? (hashFile.orig_name || hashFile.name) : 'hashfile'
                const wordlistName = wordlist ? (wordlist.orig_name || wordlist.name) : 'wordlist'

                if (this.jobForm.agent_ids.length === 1) {
                    // Single agent - simple command
                    this.commandTemplate = `hashcat -m ${this.jobForm.hash_type || 2500} -a ${this.jobForm.attack_mode || 0} ${hashFileName} ${wordlistName}`
                    
                    // Add WPA/WPA2 specific optimizations
                    this.commandTemplate += ' -O --force --status --status-timer=5'
                    
                    // Add session name
                    if (this.jobForm.name) {
                        const sessionName = this.jobForm.name.toLowerCase().replace(/[^a-z0-9]/g, '_')
                        this.commandTemplate += ` --session=${sessionName}`
                    }
                    
                    // Add outfile for WPA/WPA2
                    this.commandTemplate += ' --outfile=cracked.txt --outfile-format=2'
                } else {
                    // Multiple agents - distributed commands
                    this.commandTemplate = `# Distributed Hashcat Commands\n`
                    this.commandTemplate += `# Hash File: ${hashFileName}\n`
                    this.commandTemplate += `# Wordlist: ${wordlistName}\n`
                    this.commandTemplate += `# Total Agents: ${this.jobForm.agent_ids.length}\n\n`
                    
                    this.jobForm.agent_ids.forEach((agentId: string, index: number) => {
                        const agent = this.agents.find((a: any) => a.id === agentId)
                        if (agent) {
                            const assignedWords = this.getAssignedWordCountForSelected(agent)
                            const resourceType = this.isGPUAgent(agent) ? 'GPU' : 'CPU'
                            const performance = this.getAgentPerformanceScore(agent)
                            
                            this.commandTemplate += `# Job ${index + 1}: ${this.jobForm.name} - ${agent.name}\n`
                            this.commandTemplate += `# Agent: ${agent.name} (${resourceType}, ${performance}%)\n`
                            this.commandTemplate += `# Assigned: ${assignedWords.toLocaleString()} words\n`
                            this.commandTemplate += `hashcat -m ${this.jobForm.hash_type || 2500} -a ${this.jobForm.attack_mode || 0} \\\n`
                            this.commandTemplate += `  ${hashFileName} \\\n`
                            this.commandTemplate += `  ${wordlistName}_part_${index + 1} \\\n`
                            this.commandTemplate += `  -O --force --status --status-timer=5 \\\n`
                            this.commandTemplate += `  --session=${(this.jobForm.name || 'job').toLowerCase().replace(/[^a-z0-9]/g, '_')}_part_${index + 1} \\\n`
                            this.commandTemplate += `  --outfile=cracked_part_${index + 1}.txt --outfile-format=2\n\n`
                        }
                    })
                }
            },



            // Toggle select all agents
            toggleSelectAllAgents(checked: boolean) {
                if (checked) {
                    // Select all online agents
                    this.jobForm.agent_ids = this.onlineAgents.map((agent: any) => agent.id)
                } else {
                    // Deselect all agents
                    this.jobForm.agent_ids = []
                }
            },

            // Check if all agents are selected
            areAllAgentsSelected(): boolean {
                return this.onlineAgents.length > 0 && 
                       this.jobForm.agent_ids.length === this.onlineAgents.length &&
                       this.onlineAgents.every((agent: any) => this.jobForm.agent_ids.includes(agent.id))
            },

            // Toggle compact mode for agent selection
            toggleCompactMode() {
                this.isCompactMode = !this.isCompactMode
            },

            // Distributed Job Functions
            openDistributedJobModal() {
                this.showDistributedJobModal = true
                this.distributedJobForm = { 
                    name: '', 
                    hash_file_id: '', 
                    wordlist_id: '', 
                    hash_type: '', 
                    attack_mode: '', 
                    auto_distribute: true 
                }
                this.updateDistributedCommandTemplate()
            },

            closeDistributedJobModal() {
                this.showDistributedJobModal = false
            },

            // Check if agent is GPU-based
            isGPUAgent(agent: any): boolean {
                const capabilities = (agent.capabilities || '').toLowerCase()
                return capabilities.includes('gpu') || 
                       capabilities.includes('cuda') || 
                       capabilities.includes('opencl') ||
                       capabilities.includes('rtx') ||
                       capabilities.includes('gtx') ||
                       capabilities.includes('radeon')
            },

            // Get selected agents objects
            getSelectedAgents() {
                if (!this.jobForm.agent_ids || this.jobForm.agent_ids.length === 0) {
                    return []
                }
                return this.onlineAgents.filter(agent => 
                    this.jobForm.agent_ids.includes(agent.id)
                )
            },

            // Get agent performance score (0-100)
            getAgentPerformanceScore(agent: any): number {
                if (this.isGPUAgent(agent)) {
                    return 100 // GPU gets full performance
                }
                return 30 // CPU gets 30% performance
            },

            // Calculate assigned word count for agent
            getAssignedWordCount(agent: any): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const wordlist = this.wordlists.find((w: any) => w.id === this.distributedJobForm.wordlist_id)
                if (!wordlist || !wordlist.word_count) return 0
                
                const totalWords = wordlist.word_count
                const onlineAgents = this.onlineAgents
                
                if (onlineAgents.length === 0) return 0
                
                // Calculate performance-based distribution
                const totalPerformance = onlineAgents.reduce((sum, a) => sum + this.getAgentPerformanceScore(a), 0)
                const agentPerformance = this.getAgentPerformanceScore(agent)
                const performanceRatio = agentPerformance / totalPerformance
                
                return Math.round(totalWords * performanceRatio)
            },

            // Get distribution method description
            getDistributionMethod(): string {
                const gpuAgents = this.onlineAgents.filter(agent => this.isGPUAgent(agent))
                const cpuAgents = this.onlineAgents.filter(agent => !this.isGPUAgent(agent))
                
                if (gpuAgents.length > 0 && cpuAgents.length > 0) {
                    return `Performance-based (${gpuAgents.length} GPU + ${cpuAgents.length} CPU)`
                } else if (gpuAgents.length > 0) {
                    return `GPU-only distribution (${gpuAgents.length} agents)`
                } else {
                    return `CPU-only distribution (${cpuAgents.length} agents)`
                }
            },

            // Get distribution method for selected agents
            getDistributionMethodForSelected(): string {
                const selectedAgents = this.getSelectedAgents()
                const gpuAgents = selectedAgents.filter(agent => this.isGPUAgent(agent))
                const cpuAgents = selectedAgents.filter(agent => !this.isGPUAgent(agent))
                
                if (gpuAgents.length > 0 && cpuAgents.length > 0) {
                    return `Performance-based (${gpuAgents.length} GPU + ${cpuAgents.length} CPU)`
                } else if (gpuAgents.length > 0) {
                    return `GPU-only distribution (${gpuAgents.length} agents)`
                } else {
                    return `CPU-only distribution (${cpuAgents.length} agents)`
                }
            },

            // Calculate assigned word count for selected agent
            getAssignedWordCountForSelected(agent: any): number {
                if (!this.jobForm.wordlist_id) return 0
                
                const wordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                if (!wordlist || !wordlist.word_count) return 0
                
                const totalWords = wordlist.word_count
                const selectedAgents = this.getSelectedAgents()
                
                if (selectedAgents.length === 0) return 0
                
                // Calculate performance-based distribution
                const totalPerformance = selectedAgents.reduce((sum, a) => sum + this.getAgentPerformanceScore(a), 0)
                const agentPerformance = this.getAgentPerformanceScore(agent)
                const performanceRatio = agentPerformance / totalPerformance
                
                return Math.round(totalWords * performanceRatio)
            },

            // Get assigned percentage for selected agents (ensures total = 100%)
            getAssignedPercentageForSelected(agent: any): number {
                const selectedAgents = this.getSelectedAgents();
                if (selectedAgents.length === 0) return 0;
                
                // Calculate performance scores
                const agentScores = selectedAgents.map(a => ({
                    agent: a,
                    score: this.getAgentPerformanceScore(a)
                }));
                
                // Sort by performance (highest first)
                agentScores.sort((a, b) => b.score - a.score);
                
                // Find this agent's position
                const agentIndex = agentScores.findIndex(item => item.agent.id === agent.id);
                if (agentIndex === -1) return 0;
                
                // Calculate distribution based on performance
                const totalScore = agentScores.reduce((sum, item) => sum + item.score, 0);
                const agentScore = agentScores[agentIndex].score;
                
                // Calculate percentage (ensures total = 100%)
                const percentage = (agentScore / totalScore) * 100;
                
                return Math.round(percentage);
            },

            // Get assigned percentage with exact 100% total (no rounding errors)
            getAssignedPercentageExact(agent: any): number {
                const selectedAgents = this.getSelectedAgents();
                if (selectedAgents.length === 0) return 0;
                
                // Calculate performance scores
                const agentScores = selectedAgents.map(a => ({
                    agent: a,
                    score: this.getAgentPerformanceScore(a)
                }));
                
                // Sort by performance (highest first)
                agentScores.sort((a, b) => b.score - a.score);
                
                // Find this agent's position
                const agentIndex = agentScores.findIndex(item => item.agent.id === agent.id);
                if (agentIndex === -1) return 0;
                
                // Calculate distribution based on performance
                const totalScore = agentScores.reduce((sum, item) => sum + item.score, 0);
                const agentScore = agentScores[agentIndex].score;
                
                // Calculate percentage without rounding first
                const exactPercentage = (agentScore / totalScore) * 100;
                
                // For the last agent, ensure total = 100%
                if (agentIndex === selectedAgents.length - 1) {
                    // Calculate what the total should be for previous agents
                    let previousTotal = 0;
                    for (let i = 0; i < agentIndex; i++) {
                        const prevScore = agentScores[i].score;
                        const prevPercent = Math.round((prevScore / totalScore) * 100);
                        previousTotal += prevPercent;
                    }
                    
                    // Last agent gets the remaining percentage to make total = 100%
                    return 100 - previousTotal;
                }
                
                // For other agents, round normally
                return Math.round(exactPercentage);
            },

            // Get assigned percentage based on word count distribution (more accurate)
            getAssignedPercentageByWords(agent: any): number {
                const selectedAgents = this.getSelectedAgents();
                if (selectedAgents.length === 0) return 0;
                
                // Get word count for this agent
                const agentWords = this.getAssignedWordCountForSelected(agent);
                if (agentWords === 0) return 0;
                
                // Get total words for all selected agents
                const totalWords = selectedAgents.reduce((sum, a) => sum + this.getAssignedWordCountForSelected(a), 0);
                if (totalWords === 0) return 0;
                
                // Calculate percentage based on actual word count
                const percentage = (agentWords / totalWords) * 100;
                
                return Math.round(percentage);
            },

            // Get assigned percentage with word-based distribution (most accurate)
            getAssignedPercentageWordBased(agent: any): number {
                const selectedAgents = this.getSelectedAgents();
                if (selectedAgents.length === 0) return 0;
                
                // Get word counts for all agents
                const agentWordCounts = selectedAgents.map(a => ({
                    agent: a,
                    words: this.getAssignedWordCountForSelected(a)
                }));
                
                // Sort by word count (highest first)
                agentWordCounts.sort((a, b) => b.words - a.words);
                
                // Find this agent's position
                const agentIndex = agentWordCounts.findIndex(item => item.agent.id === agent.id);
                if (agentIndex === -1) return 0;
                
                // Get total words
                const totalWords = agentWordCounts.reduce((sum, item) => sum + item.words, 0);
                if (totalWords === 0) return 0;
                
                // For the last agent, ensure total = 100%
                if (agentIndex === selectedAgents.length - 1) {
                    // Calculate what the total should be for previous agents
                    let previousTotal = 0;
                    for (let i = 0; i < agentIndex; i++) {
                        const prevWords = agentWordCounts[i].words;
                        const prevPercent = Math.round((prevWords / totalWords) * 100);
                        previousTotal += prevPercent;
                    }
                    
                    // Last agent gets the remaining percentage to make total = 100%
                    return 100 - previousTotal;
                }
                
                // For other agents, calculate based on word count
                const agentWords = agentWordCounts[agentIndex].words;
                const percentage = (agentWords / totalWords) * 100;
                
                return Math.round(percentage);
            },

            // Update distributed command template
            updateDistributedCommandTemplate() {
                if (!this.distributedJobForm.hash_file_id || !this.distributedJobForm.wordlist_id) {
                    this.distributedCommandTemplate = 'Distributed hashcat commands will appear here...'
                    return
                }

                const hashFile = this.hashFiles.find((f: any) => f.id === this.distributedJobForm.hash_file_id)
                const wordlist = this.wordlists.find((w: any) => w.id === this.distributedJobForm.wordlist_id)
                
                if (!hashFile || !wordlist) return

                const hashFileName = hashFile.orig_name || hashFile.name
                const wordlistName = wordlist.orig_name || wordlist.name
                const totalWords = wordlist.word_count || 0

                let template = `# Distributed Hashcat Commands\n`
                template += `# Hash File: ${hashFileName}\n`
                template += `# Wordlist: ${wordlistName} (${totalWords.toLocaleString()} words)\n`
                template += `# Agents: ${this.onlineAgents.length}\n\n`

                this.onlineAgents.forEach((agent, index) => {
                    const assignedWords = this.getAssignedWordCount(agent)
                    const resourceType = this.isGPUAgent(agent) ? 'GPU' : 'CPU'
                    const performance = this.getAgentPerformanceScore(agent)
                    
                    template += `# Agent ${index + 1}: ${agent.name} (${resourceType}, ${performance}%)\n`
                    template += `# Assigned: ${assignedWords.toLocaleString()} words\n`
                    template += `hashcat -m ${this.distributedJobForm.hash_type || 2500} -a ${this.distributedJobForm.attack_mode || 0} \\\n`
                    template += `  ${hashFileName} \\\n`
                    template += `  ${wordlistName}_part_${index + 1} \\\n`
                    template += `  -O --force --status --status-timer=5 \\\n`
                    template += `  --session=${this.distributedJobForm.name?.toLowerCase().replace(/[^a-z0-9]/g, '_') || 'distributed'}_part_${index + 1} \\\n`
                    template += `  --outfile=cracked_part_${index + 1}.txt --outfile-format=2\n\n`
                })

                this.distributedCommandTemplate = template
            },

            // Update distribution preview
            updateDistributionPreview() {
                // This will trigger Alpine.js reactivity for the preview section
                setTimeout(() => {
                    console.log('Distribution preview updated')
                }, 0)
            },

            // Check if can create distributed job
            canCreateDistributedJob(): boolean {
                return !!(this.distributedJobForm.name && 
                         this.distributedJobForm.hash_file_id && 
                         this.distributedJobForm.wordlist_id &&
                         this.onlineAgents.length > 0)
            },

            // Preview distribution
            previewDistribution() {
                this.updateDistributionPreview()
                this.showNotification('Distribution preview updated', 'success')
            },

            // Create distributed job
            async createDistributedJob() {
                if (!this.canCreateDistributedJob()) {
                    this.showNotification('Please fill all required fields and ensure agents are available', 'error')
                    return
                }

                this.isLoading = true
                try {
                    // For now, show success message
                    // In real implementation, this would call the API
                    this.showNotification('Distributed job creation feature not yet implemented', 'info')
                    this.closeDistributedJobModal()
                } catch (error) {
                    this.showNotification('Failed to create distributed job', 'error')
                } finally {
                    this.isLoading = false
                }
            },

            async startJob(jobId: string) {
                // Find the job to check conditions
                const job = this.jobs.find(j => j.id === jobId)
                if (!job) {
                    this.showNotification('Job not found', 'error')
                    return
                }

                // Check if job can be started
                if (job.status === 'completed') {
                    this.showNotification('Cannot start completed job', 'warning')
                    return
                }

                if (!job.agent_name || job.agent_name === 'Unassigned') {
                    this.showNotification('Cannot start job: No agent assigned', 'warning')
                    return
                }

                if (job.status === 'running') {
                    this.showNotification('Job is already running', 'warning')
                    return
                }

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

            async pauseJob(jobId: string) {
                const success = await jobStore.actions.pauseJob(jobId)
                if (success) {
                    this.showNotification('Job paused successfully!', 'success')
                } else {
                    this.showNotification('Failed to pause job', 'error')
                }
            },

            async resumeJob(jobId: string) {
                const success = await jobStore.actions.resumeJob(jobId)
                if (success) {
                    this.showNotification('Job resumed successfully!', 'success')
                } else {
                    this.showNotification('Failed to resume job', 'error')
                }
            },

            // File actions with enhanced loading states
            async uploadHashFile(file: File) {
                if (!file) {
                    this.showNotification('Please select a file to upload', 'error')
                    return
                }
                
                try {
                    this.isLoading = true
                    this.showNotification(`Uploading ${file.name}...`, 'info')
                    
                    const result = await fileStore.actions.uploadHashFile(file)
                    if (result) {
                        // Immediate UI update with uploaded file data
                        await fileStore.actions.fetchHashFiles()
                        
                        this.showNotification('Hash file uploaded successfully!', 'success')
                        this.closeFileModal()
                        
                        // Scroll to Files tab if not already there
                        if (this.currentTab !== 'files') {
                            await this.switchTab('files')
                        }
                    } else {
                        this.showNotification('Failed to upload hash file', 'error')
                    }
                } catch (error) {
                    console.error('Upload error:', error)
                    this.showNotification('Upload failed due to network error', 'error')
                } finally {
                    this.isLoading = false
                }
            },

            async uploadWordlist(file: File) {
                if (!file) {
                    this.showNotification('Please select a file to upload', 'error')
                    return
                }
                
                try {
                    this.isLoading = true
                    this.showNotification(`Uploading ${file.name}...`, 'info')
                    
                    const result = await wordlistStore.actions.uploadWordlist(file)
                    if (result) {
                        // Immediate UI update with uploaded file data
                        await wordlistStore.actions.fetchWordlists()
                        
                        this.showNotification('Wordlist uploaded successfully!', 'success')
                        this.closeWordlistModal()
                        
                        // Scroll to Wordlists tab if not already there
                        if (this.currentTab !== 'wordlists') {
                            await this.switchTab('wordlists')
                        }
                    } else {
                        this.showNotification('Failed to upload wordlist', 'error')
                    }
                } catch (error) {
                    console.error('Upload error:', error)
                    this.showNotification('Upload failed due to network error', 'error')
                } finally {
                    this.isLoading = false
                }
            },

            // Delete actions
            async deleteAgent(id: string) {
                if (!id) {
                    this.showNotification('Error: No agent ID provided', 'error')
                    return
                }
                const agent = this.agents.find((a: any) => a.id === id)
                if (agent) {
                    this.openDeleteModal('agent', agent, async () => {
                    const success = await agentStore.actions.deleteAgent(id)
                    if (success) {
                        this.showNotification('Agent deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete agent', 'error')
                    }
                    })
                }
            },

            async deleteJob(id: string) {
                if (!id) {
                    this.showNotification('Error: No job ID provided', 'error')
                    return
                }
                const job = this.jobs.find((j: any) => j.id === id)
                if (job) {
                    this.openDeleteModal('job', job, async () => {
                    const success = await jobStore.actions.deleteJob(id)
                    if (success) {
                        this.showNotification('Job deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete job', 'error')
                    }
                    })
                }
            },

            async deleteFile(id: string) {
                if (!id) {
                    this.showNotification('Error: No file ID provided', 'error')
                    return
                }
                const file = this.hashFiles.find((f: any) => f.id === id)
                if (file) {
                    this.openDeleteModal('file', file, async () => {
                    const success = await fileStore.actions.deleteHashFile(id)
                    if (success) {
                        this.showNotification('Hash file deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete hash file', 'error')
                    }
                    })
                }
            },

            async deleteWordlist(id: string) {
                if (!id) {
                    this.showNotification('Error: No wordlist ID provided', 'error')
                    return
                }
                const wordlist = this.wordlists.find((w: any) => w.id === id)
                if (wordlist) {
                    this.openDeleteModal('wordlist', wordlist, async () => {
                    const success = await wordlistStore.actions.deleteWordlist(id)
                    if (success) {
                        this.showNotification('Wordlist deleted successfully!', 'success')
                    } else {
                        this.showNotification('Failed to delete wordlist', 'error')
                    }
                    })
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
            },

            // NEW: Cache Management Methods
            async refreshCacheStats() {
                try {
                    const stats = await apiService.getCacheStats()
                    if (stats) {
                        this.cacheStats = stats
                        // console.log('📊 Cache stats refreshed:', stats)
                        this.showNotification('Cache stats refreshed', 'info')
                    } else {
                        console.warn('No cache stats received from API')
                        // Keep existing stats or set defaults if null
                        if (!this.cacheStats) {
                            this.cacheStats = {
                                hitRate: 0,
                                missRate: 0,
                                queryReduction: 0,
                                responseSpeedImprovement: 0,
                                totalRequests: 0,
                                agents: 0,
                                wordlists: 0,
                                hashFiles: 0
                            }
                        }
                    }
                } catch (error) {
                    console.error('Failed to refresh cache stats:', error)
                    this.showNotification('Failed to refresh cache stats', 'error')
                    // Set default values if stats don't exist
                    if (!this.cacheStats) {
                        this.cacheStats = {
                            hitRate: 0,
                            missRate: 0,
                            queryReduction: 0,
                            responseSpeedImprovement: 0,
                            totalRequests: 0,
                            agents: 0,
                            wordlists: 0,
                            hashFiles: 0
                        }
                    }
                }
            },

            async clearCache() {
                if (confirm('Are you sure you want to clear all cache? This will temporarily reduce performance.')) {
                    try {
                        const success = await apiService.clearCache()
                        if (success) {
                            this.showNotification('Cache cleared successfully!', 'success')
                            await this.refreshCacheStats() // Refresh stats after clearing
                            await this.loadInitialData() // Reload data to repopulate cache
                        } else {
                            this.showNotification('Failed to clear cache', 'error')
                        }
                    } catch (error) {
                        this.showNotification('Failed to clear cache', 'error')
                    }
                }
            },


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
