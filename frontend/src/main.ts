// Alpine.js is loaded via CDN and available as window.Alpine
import { componentLoader } from './utils/component-loader'
import { componentRegistry, perf, getConfig } from './config/build.config'
import { router } from './utils/router'

// Import all services and stores
import { apiService } from './services/api.service'
import { authService } from './services/auth.service'
import { webSocketService } from './services/websocket.service'
// import './services/websocket-mock.service' // Auto-start mock in development (DISABLED)
import { agentStore } from './stores/agent.store'
import { authStore } from './stores/auth.store'
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
        alpineStartCount: number
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
                    // console.log('🚀 Starting Alpine.js manually...')
                    window.Alpine.start()
                    window.alpineManuallyStarted = true
                    // console.log('✅ Alpine.js started successfully')
                } catch (error) {
                    // console.error('❌ Alpine start failed:', error instanceof Error ? error.message : String(error))
                    throw error
                }
            } else {
                // console.log('ℹ️ Alpine already manually started or not available')
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
            // console.error('❌ Failed to initialize dashboard:', error)
            this.showErrorState()
        }
        
        // Fallback: Show debug info if nothing renders after 5 seconds
        setTimeout(() => {
            const mainContainers = document.querySelectorAll('main')
            const hasContent = mainContainers.length > 0 && Array.from(mainContainers).some(main => main.children.length > 0)
            
            if (!hasContent) {
                // console.warn('⚠️ No content rendered after 5 seconds, showing debug mode')
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
                { name: 'auth/login', path: '/components/auth/login.html' },
                { name: 'auth/logout', path: '/components/auth/logout.html' },
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
                { name: 'modals/large-file-warning-modal', path: '/components/modals/large-file-warning-modal.html' },
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
            // console.error('❌ Failed to load components:', error)
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
            // console.log('📦 Found existing main container')
        }

        // Load navigation (only if not on login page)
        const currentRoute = router.getCurrentRoute()
        if (currentRoute !== 'login') {
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
                // console.error('❌ Failed to find nav element in navigation component')
                // console.log('📝 Navigation content preview:', navigation.substring(0, 200) + '...')
            }

            // Load breadcrumb component (only if not on login page)
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
                // console.error('❌ Failed to find nav element in breadcrumb component')
            }
        }

        // Load content based on current route
        if (currentRoute === 'login') {
            // Load login page
            const loginHtml = await componentLoader.loadComponent('auth/login')
            const loginContainer = document.createElement('div')
            loginContainer.innerHTML = loginHtml
            // Find actual content element (skip script tags)
            const loginElement = loginContainer.querySelector('div') || loginContainer.firstElementChild
            if (loginElement && loginElement.tagName !== 'SCRIPT') {
                // Add x-data directive to connect with Alpine.js
                loginElement.setAttribute('x-data', 'dashboardApp')
                mainContainer.appendChild(loginElement)
                // console.log('✅ Injected login component with x-data')
            } else {
                // console.warn('❌ Failed to inject login component')
            }
        } else {
            // Load tab content for other routes
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
                    // Silently handle failed component injection
                    // console.warn(`❌ Failed to inject ${component} component`)
                }
            }
        }

        // Load modals
        const modalComponents = [
            'modals/agent-modal',
            'modals/agent-key-modal',
            'modals/delete-confirm-modal',
            'modals/job-modal',
            'modals/file-modal',
            'modals/wordlist-modal',
            'modals/large-file-warning-modal'
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
                // Silently handle failed modal injection
                // console.warn(`❌ Failed to inject ${component} modal`)
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
                // console.warn(`❌ Failed to inject ${component} UI component`)
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
            // console.warn('Alpine not available during initialization')
            return
        }

        // Prevent duplicate data registration
        if (this.alpineDataRegistered) {
            // console.log('Alpine data already registered, skipping...')
            return
        }
        this.alpineDataRegistered = true

        // Global dashboard data and methods  
        const self = this
        window.Alpine.data('dashboardApp', () => ({
            // Reactive state
            currentTab: router.getCurrentRoute(),
            mobileMenuOpen: false,
            isLoading: false,
            isAlpineInitialized: false,
            notifications: [] as Array<{id: number, message: string, type: 'success' | 'error' | 'info' | 'warning', timestamp: Date}>,
            
            // Authentication state
            isAuthenticated: authStore.isAuthenticated(),
            user: authStore.getUser(),
            authLoading: authStore.isLoading(),
            showLoginSuccessNotification: false,
            
            // Modal states
            showAgentModal: false,
            showAgentKeyModal: false,
            showDeleteModal: false,
            showAgentKeys: false,
            showJobModal: false,
            currentStep: 1, // 1: Basic Config, 2: Distribution Preview
            showFileModal: false,
            showWordlistModal: false,
            showLargeFileWarningModal: false,
            showLargeFileWarning: false,
            largeFileInfo: null as { name: string, size: string } | null,
            showDistributedJobModal: false,
            
            // Compact mode for agent selection
            isCompactMode: false,
            
            // Form states
            agentForm: { ip_address: '', port: null as number | null, capabilities: '', agent_key: '' },
            agentKeyForm: { name: '', agent_key: '' },
            createdAgent: null as any,
            createdAgentKey: null as any,
            showAgentNameError: false,
            deleteModalConfig: { entityType: '', entityName: '', description: '', warning: '', entityId: '', confirmAction: null as any },
            jobForm: { name: '', hash_file_id: '', wordlist_id: '', agent_ids: [] as string[], hash_type: '', attack_mode: '' },
            distributedJobForm: { name: '', hash_file_id: '', wordlist_id: '', hash_type: '', attack_mode: '', auto_distribute: true },
            fileForm: { file: null },
            wordlistForm: { file: null as File | null },
            loginForm: { username: '', password: '' },
            showPassword: false,
            usernameError: null as string | null,
            passwordError: null as string | null,
            showValidationErrors: false,
            agentDistributionData: [] as Array<{agent: any, wordLimit: number, percentage: number, speed: number}>,
            
            // Command template for job creation
            commandTemplate: '',
            distributedCommandTemplate: '',
            
            // Manual loading reset (fallback)
            forceStopLoading() {
                this.isLoading = false
                // console.log('🛑 Loading force stopped by user')
            },

            // Generate agent key on frontend
            generateAgentKey() {
                // Generate 8-character hex key
                const bytes = new Uint8Array(4)
                crypto.getRandomValues(bytes)
                return Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('')
            },
            
            // Cache stats (if needed)
            cacheStats: null as any,
            showCacheStatsNotification: false,
            
            // WebSocket connection status
            wsConnected: false,
            wsConnectionAttempts: 0,
            wsSubscriptionsSetup: false,
            
            // Track agent status changes to prevent duplicate notifications
            lastAgentStatuses: new Map(),
            
            // Reactive data arrays - these will be updated by store subscriptions
            reactiveAgents: [] as any[],
            reactiveAgentKeys: [] as any[],
            reactiveJobs: [] as any[],
            reactiveHashFiles: [] as any[],
            reactiveWordlists: [] as any[],

            // Store references for Alpine.js expressions
            agentStore: agentStore,
            jobStore: jobStore,
            fileStore: fileStore,
            wordlistStore: wordlistStore,
            authStore: authStore,

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
                // Filter out master jobs (distributed coordinators) and invalid jobs from UI display
                return (this.reactiveJobs || []).filter(job => 
                    job.status !== 'distributed' && 
                    job.name && 
                    job.name.trim() !== '' && 
                    job.name !== '-' &&
                    job.name !== 'null'
                )
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
                if (!this.distributedJobForm.wordlist_id) {
                    return '0'
                }
                
                const selectedWordlist = this.wordlists.find((w: any) => w.id === this.distributedJobForm.wordlist_id)
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

            // Get selected wordlist count as number for calculations
            get selectedWordlistCountNumber() {
                if (!this.distributedJobForm.wordlist_id) {
                    return 0
                }
                
                const selectedWordlist = this.wordlists.find((w: any) => w.id === this.distributedJobForm.wordlist_id)
                if (!selectedWordlist || !selectedWordlist.word_count) {
                    return 0
                }
                
                return selectedWordlist.word_count
            },

            // Computed property for selected wordlist count in regular job form
            get selectedWordlistCountRegular() {
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

            // Get selected wordlist count as number for regular job form
            get selectedWordlistCountRegularNumber() {
                if (!this.jobForm.wordlist_id) {
                    return 0
                }
                
                const selectedWordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                if (!selectedWordlist) {
                    return 0
                }
                
                return selectedWordlist.word_count || 0
            },

            // Check if selected wordlist has less than 1000 words
            get isWordlistTooSmall() {
                return this.selectedWordlistCountRegularNumber > 0 && this.selectedWordlistCountRegularNumber < 1000
            },

            // Check if multi-agent selection should be disabled
            get isMultiAgentSelectionDisabled() {
                return this.isWordlistTooSmall
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
                // console.log('Initializing Alpine.js dashboard data...')
                this.isAlpineInitialized = true
                
                // Check authentication status on init
                const currentRoute = router.getCurrentRoute()
                // console.log('🔍 Initial route:', currentRoute)
                // console.log('🔍 Authentication status:', this.isAuthenticated)
                // console.log('🔍 Auth store state:', authStore.getState())
                // console.log('🔍 Current URL:', window.location.href)
                // console.log('🔍 Current hash:', window.location.hash)
                
                // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY route except login
                if (!this.isAuthenticated && currentRoute !== 'login') {
                    // console.log('🔒 STRICT AUTH: Unauthenticated access to', currentRoute, '- redirecting to login')
                    this.showNotification('Please login to access this page', 'warning')
                    // Force redirect to login with URL change
                    window.location.replace('/login')
                    return
                }
                
                // Setup router listener
                router.subscribe(async (route: string) => {
                    this.currentTab = route
                    // Only clear login form when navigating away from login page
                    // But don't clear if we're in loading state (successful login)
                    if (this.currentTab !== 'login' && route !== 'login' && !this.isLoading) {
                        this.clearLoginForm()
                    }
                    
                    // Load content based on route
                    await this.loadContentForRoute(route)
                })
                
                // Additional protection: Listen for direct URL changes
                window.addEventListener('popstate', () => {
                    const currentRoute = router.getCurrentRoute()
                    // console.log('🔍 URL changed to:', currentRoute)
                    
                    // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY route except login
                    if (!this.isAuthenticated && currentRoute !== 'login') {
                        // console.log('🔒 STRICT AUTH: Direct URL access to', currentRoute, '- redirecting to login')
                        this.showNotification('Please login to access this page', 'warning')
                        window.location.replace('/login')
                    }
                })
                
                // Setup store subscriptions for reactivity
                this.setupStoreSubscriptions()
                
                // Don't clear login form if currently on login page to preserve user input
                // Only clear on successful login or when navigating away
                
                // Setup WebSocket for real-time updates (only if not on login page)
                if (this.currentTab !== 'login') {
                    this.setupWebSocketSubscriptions()
                }
                
                try {
                    await this.loadInitialData()
                    
                    // Load content for current route
                    const currentRoute = router.getCurrentRoute()
                    await this.loadContentForRoute(currentRoute)
                    
                    // console.log('🎉 Dashboard initialization complete')
                } catch (error) {
                    // console.error('❌ Dashboard initialization failed:', error)
                    this.showNotification('Failed to initialize dashboard', 'error')
                }
                
                this.setupPolling()
                
                // Safety timeout to prevent infinite loading
                setTimeout(() => {
                    if (this.isLoading) {
                        // console.warn('⚠️ Loading timeout reached, forcing stop')
                        this.forceStopLoading()
                        this.showNotification('Loading took too long. Data may be incomplete.', 'warning')
                    }
                }, 15000) // 15 second safety timeout
            },

            // NEW: Setup store subscriptions for reactive updates
            setupStoreSubscriptions() {
                // Subscribe to auth store changes
                authStore.subscribe((state) => {
                    this.isAuthenticated = state.isAuthenticated
                    this.user = state.user
                    this.authLoading = state.isLoading
                    
                    // Refresh router when auth state changes
                    router.refresh()
                })
                
                // Subscribe to agent store changes with throttling
                let agentUpdateTimeout: number | null = null
                agentStore.subscribe(() => {
                    // Throttle updates to prevent excessive re-renders
                    if (agentUpdateTimeout) {
                        clearTimeout(agentUpdateTimeout)
                    }
                    
                    agentUpdateTimeout = setTimeout(() => {
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
                        // console.log('Agent store updated:', this.reactiveAgents.length, 'agents')

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
                    }, 100) // Throttle to 100ms
                })
                
                // Subscribe to job store changes with throttling
                let jobUpdateTimeout: number | null = null
                jobStore.subscribe(() => {
                    // Throttle updates to prevent excessive re-renders
                    if (jobUpdateTimeout) {
                        clearTimeout(jobUpdateTimeout)
                    }
                    
                    jobUpdateTimeout = setTimeout(() => {
                        const state = jobStore.getState()
                        // ✅ Force Alpine.js reactivity by creating new array reference
                        this.reactiveJobs = [...(state.jobs || [])]
                        // console.log('Job store updated:', this.reactiveJobs.length, 'jobs')
                        
                        // Sync pagination (if available)
                        // Note: This would need to be updated when we implement job pagination in the backend
                    }, 100) // Throttle to 100ms
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
            
            // Wait for modal to close and then show notification
            waitForModalClose(modalElement: Element) {
                // console.log('⏳ Waiting for modal to close...')
                
                let checkCount = 0
                const maxChecks = 100 // Maximum 10 seconds (100 * 100ms)
                
                const checkModal = () => {
                    checkCount++
                    
                    // Check if modal is still present
                    const isModalPresent = document.contains(modalElement) && 
                                         ((modalElement as HTMLElement).offsetParent !== null || 
                                          (modalElement as HTMLElement).style.display !== 'none')
                    
                    if (!isModalPresent || checkCount >= maxChecks) {
                        // console.log('✅ Modal closed or timeout reached, showing notification...')
                        // Add small delay to ensure modal is completely gone
                        setTimeout(() => {
                            this.showLoginSuccessNotificationNow()
                        }, 500)
                    } else {
                        // Check again in 100ms
                        setTimeout(checkModal, 100)
                    }
                }
                
                // Start checking
                setTimeout(checkModal, 100)
            },

            // Show notification immediately
            showLoginSuccessNotificationNow() {
                // console.log('🚀 ===== SHOWING LOGIN SUCCESS NOTIFICATION NOW =====')
                
                // Check if notification has already been shown for this session
                const notificationShown = sessionStorage.getItem('loginSuccessNotificationShown')
                if (notificationShown === 'true') {
                    // console.log('🚫 Login success notification already shown in this session, skipping...')
                    return
                }
                
                // Mark notification as shown
                sessionStorage.setItem('loginSuccessNotificationShown', 'true')
                
                this.showLoginSuccessNotification = true
                // console.log('🚀 State set to:', this.showLoginSuccessNotification)
                
                // Force Alpine reactivity
                if (window.Alpine) {
                    window.Alpine.nextTick(() => {
                        // console.log('🚀 After nextTick, state is:', this.showLoginSuccessNotification)
                    })
                }
                
                // Show notifications
                this.showNotification('🎉 Welcome to Hashcat Dashboard!', 'success')
                setTimeout(() => {
                    this.showNotification('✅ Login successful! You are now logged in.', 'success')
                }, 1000)
                setTimeout(() => {
                    this.showNotification('🚀 Ready to start cracking passwords!', 'info')
                }, 2000)
                
                // console.log('🚀 =====================================================')
            },


            // Check for login success notification
            checkLoginSuccessNotification() {
                // console.log('🔍 ===== CHECKING LOGIN SUCCESS NOTIFICATION =====')
                // console.log('🔍 Current tab in checkLoginSuccessNotification:', this.currentTab)
                // console.log('🔍 Current route:', window.location.pathname)
                // console.log('🔍 Alpine available:', !!window.Alpine)
                // console.log('🔍 Dashboard app available:', !!(window as any).Alpine?.data('dashboardApp'))
                
                const showLoginSuccess = sessionStorage.getItem('showLoginSuccess')
                const loginSuccessTime = sessionStorage.getItem('loginSuccessTime')
                // console.log('📝 showLoginSuccess flag:', showLoginSuccess)
                // console.log('⏰ loginSuccessTime:', loginSuccessTime)
                // console.log('🔍 showLoginSuccess === "true"?', showLoginSuccess === 'true')
                // console.log('🔍 SessionStorage keys:', Object.keys(sessionStorage))
                // console.log('🔍 ================================================')
                
                // Check if Google Password Manager modal is present
                const googleModal = document.querySelector('[role="dialog"]') || 
                                 document.querySelector('.modal') || 
                                 document.querySelector('[data-testid="modal"]') ||
                                 document.querySelector('[aria-modal="true"]') ||
                                 document.querySelector('[class*="modal"]') ||
                                 document.querySelector('[class*="dialog"]') ||
                                 document.querySelector('[class*="overlay"]')
                
                if (googleModal) {
                    // console.log('⚠️ Google Password Manager modal detected, delaying notification...')
                    // console.log('⚠️ Modal element:', googleModal)
                    // Wait for modal to be closed
                    this.waitForModalClose(googleModal)
                    return
                } else {
                    // console.log('✅ No modal detected, showing notification immediately')
                }
                
                // Make this method globally accessible
                if (typeof window !== 'undefined') {
                    (window as any).checkLoginSuccessNotification = this.checkLoginSuccessNotification.bind(this)
                }
                
                if (showLoginSuccess === 'true') {
                    // console.log('✅ Showing login success notification!')
                    // console.log('🔍 About to call showNotification...')
                    
                    // Check if notification has already been shown for this session
                    const notificationShown = sessionStorage.getItem('loginSuccessNotificationShown')
                    if (notificationShown === 'true') {
                        console.log('🚫 Login success notification already shown in this session, skipping...')
                        // Clear the flags since we're not showing notification
                        sessionStorage.removeItem('showLoginSuccess')
                        sessionStorage.removeItem('loginSuccessTime')
                        return
                    }
                    
                    // Show notification immediately
                    this.showLoginSuccessNotificationNow()
                    
                    // Clear the flags after a delay to ensure notification is shown
                    setTimeout(() => {
                        sessionStorage.removeItem('showLoginSuccess')
                        sessionStorage.removeItem('loginSuccessTime')
                        // console.log('🧹 Cleared login success flags from sessionStorage')
                    }, 5000) // Clear after 5 seconds
                    
                    // Auto-hide notification after 8 seconds
                    setTimeout(() => {
                        console.log('⏰ Auto-hiding login success notification')
                        this.showLoginSuccessNotification = false
                    }, 8000)
                } else {
                    console.log('❌ No login success flag found')
                    console.log('❌ showLoginSuccess value:', showLoginSuccess)
                    console.log('❌ typeof showLoginSuccess:', typeof showLoginSuccess)
                }
            },

            // Authentication methods
            clearLoginForm() {
                this.loginForm = { username: '', password: '' }
                this.usernameError = null
                this.passwordError = null
                this.showPassword = false
            },
            
            // Clear only errors, keep form data
            clearLoginErrors() {
                this.usernameError = null
                this.passwordError = null
            },
            
            async handleLogin() {
                // Prevent multiple clicks
                if (this.isLoading) {
                    console.log('🛑 Login already in progress, ignoring click')
                    return
                }
                
                // Set loading state immediately for instant feedback
                this.isLoading = true
                
                console.log('🔍 handleLogin called')
                console.log('📝 Username:', this.loginForm.username)
                console.log('🔐 Password:', this.loginForm.password)
                
                // Clear previous errors first, but keep form data
                this.clearLoginErrors()
                
                // Validate fields - only show errors when button is clicked
                let hasErrors = false
                
                // Check username/email
                if (!this.loginForm.username || this.loginForm.username.trim() === '') {
                    console.log('❌ Username validation failed')
                    this.usernameError = 'Username or email is required'
                    hasErrors = true
                }
                
                // Check password
                if (!this.loginForm.password || this.loginForm.password.trim() === '') {
                    console.log('❌ Password validation failed')
                    this.passwordError = 'Password is required'
                    hasErrors = true
                }
                
                // If there are validation errors, don't proceed with login
                if (hasErrors) {
                    console.log('🛑 Form validation failed, stopping login process')
                    this.isLoading = false // Reset loading state
                    return
                }
                
                console.log('🔍 Has errors:', hasErrors)
                console.log('📝 Username error:', this.usernameError)
                console.log('🔐 Password error:', this.passwordError)
                
                // If validation passes, proceed with login
                let loginSuccessful = false
                try {
                    console.log('🚀 Proceeding with login...')
                    const success = await authStore.login({
                        username: this.loginForm.username.trim(),
                        password: this.loginForm.password
                    })
                    
                    if (success) {
                        loginSuccessful = true
                        console.log('🎉 ===== LOGIN SUCCESSFUL! =====')
                        console.log('🎉 Setting notification flag...')
                        console.log('📝 Form data before loading:', { username: this.loginForm.username, password: this.loginForm.password })
                        
                        // Set loading state to show loading component instead of form
                        this.isLoading = true
                        console.log('📝 Form data after loading set:', { username: this.loginForm.username, password: this.loginForm.password })
                        
                        // Set multiple flags for success notification
                        sessionStorage.setItem('showLoginSuccess', 'true')
                        sessionStorage.setItem('loginSuccessTime', Date.now().toString())
                        // console.log('💾 showLoginSuccess flag set in sessionStorage')
                        console.log('💾 SessionStorage contents:', {
                            showLoginSuccess: sessionStorage.getItem('showLoginSuccess'),
                            loginSuccessTime: sessionStorage.getItem('loginSuccessTime')
                        })
                        // console.log('💾 All sessionStorage keys:', Object.keys(sessionStorage))
                        console.log('🎉 ================================')
                        
                        // Force trigger notification check immediately
                        setTimeout(() => {
                            console.log('🔄 Triggering immediate notification check...')
                            this.checkLoginSuccessNotification()
                        }, 100)
                        
                        // Also trigger after redirect
                        setTimeout(() => {
                            console.log('🔄 Triggering notification after redirect...')
                            this.showLoginSuccessNotificationNow()
                        }, 1000)
                        
                        // Clear form and redirect to dashboard
                        setTimeout(() => {
                            console.log('🔄 Redirecting to dashboard...')
                            console.log('📝 Form data before clear:', { username: this.loginForm.username, password: this.loginForm.password })
                            // Clear form before redirect
                            this.clearLoginForm()
                            // Use router navigation instead of reload to preserve sessionStorage
                            router.navigate('overview')
                            
                            // Trigger notification after navigation
                            setTimeout(() => {
                                console.log('🔄 Triggering notification after navigation...')
                                this.showLoginSuccessNotificationNow()
                            }, 500)
                        }, 800) // Reduced delay for faster redirect
                    } else {
                        // Show authentication errors as notifications instead of form errors
                        const error = authStore.getError()
                        if (error && error.includes('Invalid username or password')) {
                            // Show error messages on both fields for wrong credentials
                            this.usernameError = 'Invalid username or email'
                            this.passwordError = 'Invalid password'
                            // Also show notification for authentication failure
                            this.showNotification('Authentication Failed: Invalid username or password', 'error')
                        } else {
                            // Show other errors as notifications
                            this.showNotification(error || 'Login failed. Please try again.', 'error')
                        }
                        
                        // Reset loading state when login fails
                        this.isLoading = false
                    }
                } catch (error) {
                    // Handle network or other errors
                    console.error('❌ Login error:', error)
                    this.showNotification('Login failed. Please check your connection and try again.', 'error')
                    this.isLoading = false
                } finally {
                    // Only reset loading state if login failed (not successful)
                    // For successful login, keep loading state until redirect
                    if (this.isLoading && !loginSuccessful) {
                        this.isLoading = false
                    }
                    console.log('🏁 Login process completed')
                }
            },
            
            // Clear error methods (only used when user types)
            clearUsernameError() {
                this.usernameError = null
            },
            
            clearPasswordError() {
                this.passwordError = null
            },
            
            // Logout functionality
            async handleLogout() {
                try {
                    // Show loading state
                    this.isLoading = true
                    
                    // Set logout flag to allow navigation to login page
                    router.setLoggingOut(true)
                    
                    // Call auth store logout
                    await authStore.logout()
                    
                    // Clear all form data
                    this.clearLoginForm()
                    
                    // Clear all reactive data
                    this.reactiveAgents = []
                    this.reactiveJobs = []
                    this.reactiveHashFiles = []
                    this.reactiveWordlists = []
                    
                    // Clear login success notification flag to allow it to show again on next login
                    sessionStorage.removeItem('loginSuccessNotificationShown')
                    
                    // Reset user info
                    this.user = null
                    this.isAuthenticated = false
                    
                    // Show success notification
                    this.showNotification('Successfully logged out', 'success')
                    
                    // Redirect to login page
                    router.navigate('login')
                    
                    // Auto reload after logout to ensure complete state reset
                    setTimeout(() => {
                        window.location.reload()
                    }, 1000) // 1 second delay to show notification
                    
                } catch (error) {
                    console.error('Logout error:', error)
                    this.showNotification('Logout failed. Please try again.', 'error')
                    this.isLoading = false
                } finally {
                    // Don't reset loading state here - let it stay until page reload
                    // Reset logout flag after navigation
                    setTimeout(() => {
                        router.setLoggingOut(false)
                    }, 100)
                }
            },
            
            clearAuthError() {
                authStore.clearError()
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
                // Don't setup WebSocket on login page
                if (this.currentTab === 'login') {
                    return
                }
                
                // Prevent multiple subscriptions
                if (this.wsSubscriptionsSetup) {
                    return
                }
                this.wsSubscriptionsSetup = true
                
                // Connection status monitoring
                webSocketService.onConnection((status) => {
                    this.wsConnected = status.connected
                    if (status.connected) {
                        // Only show notification if not on login page
                        if (this.currentTab !== 'login') {
                            // Add delay to ensure login success notification shows first
                            setTimeout(() => {
                                this.showNotification('🔗 Real-time updates connected', 'success')
                            }, 1500) // 1.5 second delay after login success
                        }
                        // Subscribe to all updates when connected
                        webSocketService.subscribeToJobs()
                        webSocketService.subscribeToAgents()
                    } else {
                        // Only show notification if not on login page
                        if (this.currentTab !== 'login') {
                            this.showNotification('🔌 Real-time updates disconnected', 'warning')
                        }
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
                
                // Listen for auto-fetched job results
                window.addEventListener('jobStore:jobResultFetched', (event: any) => {
                    const { jobId, result } = event.detail
                    const job = this.jobs.find(j => j.id === jobId)
                    if (job) {
                        this.showNotification(`🎉 Job "${job.name}" result automatically retrieved!`, 'success')
                    }
                })
                
                // console.log('🌐 WebSocket subscriptions setup for real-time updates')
            },

            // Real-time job updates
            async updateJobProgress(update: any) {
                // Update individual job via store action
                if (update.job_id) {
                    await jobStore.actions.updateJobProgress(update.job_id, update.progress, update.speed, update.eta, update.status)
                }
            },

            async updateJobStatus(update: any) {
                // Update individual job via store action
                if (update.job_id) {
                    await jobStore.actions.updateJobStatus(update.job_id, update.status, update.result)
                    
                    // Show notification for important status changes
                    const job = this.jobs.find(j => j.id === update.job_id)
                    if (job) {
                        if (update.status === 'completed') {
                            this.showNotification(`Job "${job.name}" completed!`, 'success')
                        } else if (update.status === 'failed') {
                            this.showNotification(`Job "${job.name}" failed`, 'error')
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
                    console.error('Failed to load initial data:', error)
                    this.showNotification('Failed to load data. Please refresh the page.', 'error')
                } finally {
                    this.isLoading = false
                }
            },

            // Load content based on route
            async loadContentForRoute(route: string) {
                // console.log('🔄 loadContentForRoute called with route:', route)
                const mainContainer = document.getElementById('main-content')
                if (!mainContainer) return
                
                // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY route except login
                if (!this.isAuthenticated && route !== 'login') {
                    console.log('🔒 STRICT AUTH: Unauthenticated access to', route, '- redirecting to login')
                    this.showNotification('Please login to access this page', 'warning')
                    // Force redirect to login with URL change
                    window.location.replace('/login')
                    return
                }
                
                // If already authenticated and trying to access login page, redirect to overview
                if (route === 'login' && this.isAuthenticated) {
                    console.log('🔒 Already authenticated, redirecting to overview')
                    router.navigate('overview')
                    return
                }
                
                // Clear existing content
                mainContainer.innerHTML = ''
                
                if (route === 'login') {
                    // Load login page
                    const loginHtml = await componentLoader.loadComponent('auth/login')
                    const loginContainer = document.createElement('div')
                    loginContainer.innerHTML = loginHtml
                    // Find actual content element (skip script tags)
                    const loginElement = loginContainer.querySelector('div') || loginContainer.firstElementChild
                    if (loginElement && loginElement.tagName !== 'SCRIPT') {
                        // Add x-data directive to connect with Alpine.js
                        loginElement.setAttribute('x-data', 'dashboardApp')
                        mainContainer.appendChild(loginElement)
                        // console.log('✅ Injected login component with x-data')
                    } else {
                        console.warn('❌ Failed to inject login component')
                    }
                } else {
                    console.log('🔍 Entering else block for non-login routes')
                    // Load tab content for other routes
                    const tabComponents = [
                        'tabs/overview',
                        'tabs/agents', 
                        'tabs/agent-keys',
                        'tabs/jobs',
                        'tabs/files',
                        'tabs/wordlists',
                        'tabs/docs'
                    ]
                    
                    for (const componentName of tabComponents) {
                        const componentHtml = await componentLoader.loadComponent(componentName)
                        const componentContainer = document.createElement('div')
                        componentContainer.innerHTML = componentHtml
                        const componentElement = componentContainer.querySelector('div') || componentContainer.firstElementChild
                        if (componentElement && componentElement.tagName !== 'SCRIPT') {
                            componentElement.setAttribute('x-data', 'dashboardApp')
                            mainContainer.appendChild(componentElement)
                        }
                    }
                    
                    console.log('🔍 About to check for login success notification...')
                    // Check for login success notification after loading overview
                    console.log('🔍 Checking route for login success notification:', route)
                    console.log('🔍 Route === overview?', route === 'overview')
                    if (route === 'overview') {
                        console.log('🏠 Overview loaded, checking for login success notification...')
                        console.log('🔍 Current tab when checking:', this.currentTab)
                        console.log('🔍 Route when checking:', route)
                        console.log('🔍 About to call checkLoginSuccessNotification...')
                        this.checkLoginSuccessNotification()
                        console.log('🔍 checkLoginSuccessNotification called')
                    } else {
                        console.log('❌ Not overview route, skipping login success check. Route:', route)
                    }
                }
            },

            // Tab management with router integration
            async switchTab(tab: string) {
                if (this.currentTab === tab) return
                
                // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY tab except login
                if (!this.isAuthenticated && tab !== 'login') {
                    console.log('🔒 STRICT AUTH: Unauthenticated access to', tab, '- redirecting to login')
                    this.showNotification('Please login to access this page', 'warning')
                    // Force redirect to login with URL change
                    window.location.replace('/login')
                    return
                }
                
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

            showNotification(
                message: string,
                type: 'success' | 'error' | 'info' | 'warning' = 'info'
            ) {
                // console.log(`🔔 showNotification called: [${type.toUpperCase()}] ${message}`)
                // console.log(`🔔 Current tab: ${this.currentTab}`)
            
                const isLoginSuccess = message.includes('Login successful')
                const isLoginError =
                    message.includes('Authentication Failed') ||
                    message.includes('Login failed')
            
                // Jika login sukses → paksa jadi success
                if (isLoginSuccess) {
                    type = 'success'
                }
            
                // Jangan tampilkan notifikasi di halaman login kecuali login error
                if (this.currentTab === 'login' && !isLoginError) {
                    // console.log(`🔔 Notification blocked on login page: [${type.toUpperCase()}] ${message}`)
                    return
                }
            
                // Hindari duplikat notif dengan pesan & type yang sama
                const exists = this.notifications.some(
                    (n) => n.message === message && n.type === type
                )
                if (exists) {
                    // console.log(`🔔 Duplicate notification blocked: [${type.toUpperCase()}] ${message}`)
                    return
                }
            
                // console.log(`🔔 Showing notification: [${type.toUpperCase()}] ${message}`)
            
                const notification = {
                    id: Date.now(),
                    message,
                    type,
                    timestamp: new Date()
                }
            
                this.notifications.unshift(notification)
            
                // Auto-remove setelah 5 detik
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
                    .replace(/^.*$/gm, '')  // Remove copy indicators
                    .replace(/^.*$/gm, '')  // Remove success indicators  
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
                // Generate key on frontend for immediate display
                const generatedKey = this.generateAgentKey()
                console.log('Generated key on modal open:', generatedKey)
                this.agentKeyForm = { name: '', agent_key: generatedKey }
                this.createdAgentKey = null
                this.showAgentNameError = false
            },
            
            closeAgentKeyModal() {
                this.showAgentKeyModal = false
                this.createdAgentKey = null
                this.showAgentNameError = false
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
                console.log('copyAgentKey called with:', agentKey)
                
                // Validate agent key
                if (!agentKey || agentKey.trim() === '') {
                    console.log('Agent key is empty or invalid')
                    this.showNotification('No agent key to copy', 'error')
                    return
                }

                try {
                    // Check if clipboard API is available
                    if (navigator.clipboard && window.isSecureContext) {
                        console.log('Using modern clipboard API')
                        await navigator.clipboard.writeText(agentKey)
                        this.showNotification('Agent key copied to clipboard!', 'success')
                    } else {
                        console.log('Using fallback copy method')
                        // Fallback for older browsers or non-HTTPS contexts
                        this.fallbackCopyTextToClipboard(agentKey)
                    }
                } catch (err) {
                    console.error('Failed to copy agent key: ', err)
                    // Try fallback method
                    this.fallbackCopyTextToClipboard(agentKey)
                }
            },

            fallbackCopyTextToClipboard(text: string) {
                console.log('fallbackCopyTextToClipboard called with:', text)
                
                const textArea = document.createElement("textarea")
                textArea.value = text
                
                // Avoid scrolling to bottom
                textArea.style.top = "0"
                textArea.style.left = "0"
                textArea.style.position = "fixed"
                textArea.style.opacity = "0"
                
                document.body.appendChild(textArea)
                textArea.focus()
                textArea.select()
                
                try {
                    const successful = document.execCommand('copy')
                    console.log('execCommand copy result:', successful)
                    if (successful) {
                        this.showNotification('Agent key copied to clipboard!', 'success')
                    } else {
                        this.showNotification('Failed to copy agent key', 'error')
                    }
                } catch (err) {
                    console.error('Fallback copy failed: ', err)
                    this.showNotification('Failed to copy agent key', 'error')
                }
                
                document.body.removeChild(textArea)
            },
            
            async validateAndCreateAgentKey(agentKeyData: any) {
                // Reset error state
                this.showAgentNameError = false

                // Validate required fields
                if (!agentKeyData.name || agentKeyData.name.trim() === '') {
                    this.showAgentNameError = true
                    this.showNotification('Agent name is required', 'error')
                    return
                }

                // Call the original createAgentKey function
                await this.createAgentKey(agentKeyData)
            },

            async createAgentKey(agentKeyData: any) {
                console.log('Creating agent key with data:', agentKeyData)
                console.log('Frontend key before sending to server:', agentKeyData.agent_key)
                
                // Use the new generateAgentKey action for creating agent keys
                const result = await agentStore.actions.generateAgentKey(agentKeyData.name, agentKeyData.agent_key)
                if (result) {
                    console.log('Server returned agent:', result)
                    console.log('Server key:', result.agent_key)
                    console.log('Frontend key after server response:', agentKeyData.agent_key)
                    
                    this.showNotification('Agent Key Generated Successfully!', 'success')
                    this.createdAgentKey = result
                    
                    // Refresh the agents table to show the new agent
                    await this.refreshAgentsTable()
                    
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

            openJobModal() {
                // Check if there are online agents
                if (this.onlineAgents.length === 0) {
                    this.showNotification('No online agents available. Please start an agent first before creating jobs.', 'warning')
                    return
                }
                
                this.showJobModal = true
                this.currentStep = 1
                // Reset form to initial state
                this.jobForm = { 
                    name: '', 
                    hash_file_id: '', 
                    wordlist_id: '', 
                    agent_ids: [], 
                    hash_type: '2500', 
                    attack_mode: '0' 
                }
                // Clear command template
                this.commandTemplate = 'hashcat command will appear here...'
                console.log('🔓 Job modal opened with fresh form')
            },
            
            closeJobModal() {
                console.log('🔒 Closing job modal...')
                
                try {
                    // Close the modal
                    this.showJobModal = false
                    
                    // Reset step to first step
                    this.currentStep = 1
                    
                    // Reset form to initial state
                    this.jobForm = { 
                        name: '', 
                        hash_file_id: '', 
                        wordlist_id: '', 
                        agent_ids: [], 
                        hash_type: '2500', 
                        attack_mode: '0' 
                    }
                    
                    // Clear command template
                    this.commandTemplate = 'hashcat command will appear here...'
                    
                    // Reset validation errors
                    this.showValidationErrors = false
                    
                    // Clear distribution data
                    this.agentDistributionData = []
                    
                    // Clear loading state
                    this.isLoading = false
                    
                    // Force UI update
                    setTimeout(() => {
                        console.log('✅ Job modal closed and form reset successfully')
                    }, 0)
                } catch (error) {
                    console.error('Error closing job modal:', error)
                    // Force close even if there's an error
                    this.showJobModal = false
                }
            },

            // Step management functions
            goToStep(step: number) {
                if (step === 2 && !this.canProceedToStep2()) {
                    this.showValidationErrors = true
                    this.showNotification('Please complete all required fields before proceeding', 'warning')
                    return
                }
                this.showValidationErrors = false
                this.currentStep = step
                if (step === 2) {
                    this.updateCommandTemplate()
                    this.updateDistributionPreview()
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
                    console.log('🚀 Starting job creation process...')
                    
                    // Get wordlist details for backend requirement
                    const selectedWordlist = this.wordlists.find((w: any) => w.id === jobData.wordlist_id)
                    const wordlistName = selectedWordlist ? (selectedWordlist.orig_name || selectedWordlist.name) : 'unknown.txt'
                    
                    // Get wordlist content for distribution
                    let wordlistContent = ''
                    if (selectedWordlist && selectedWordlist.content) {
                        wordlistContent = selectedWordlist.content
                    } else if (selectedWordlist && selectedWordlist.path) {
                        // Try to fetch wordlist content from server
                        try {
                            const response = await fetch(`/api/v1/wordlists/${selectedWordlist.id}/content`)
                            if (response.ok) {
                                wordlistContent = await response.text()
                            }
                        } catch (error) {
                            console.warn('Failed to fetch wordlist content:', error)
                        }
                    }
                    
                    // Enhanced job creation with agent assignment
                    const jobPayload = {
                        name: jobData.name,
                        hash_type: parseInt(jobData.hash_type),
                        attack_mode: parseInt(jobData.attack_mode),
                        hash_file_id: jobData.hash_file_id,
                        wordlist: wordlistContent || wordlistName,  // Send content if available, otherwise name
                        wordlist_id: jobData.wordlist_id,         // Optional reference ID
                        agent_ids: jobData.agent_ids || [],       // Include multiple agent assignments
                        distribution_data: this.agentDistributionData // Include calculated distribution data
                    }
                    
                    // Log distribution information for debugging
                    if (this.agentDistributionData && this.agentDistributionData.length > 0) {
                        console.log('📊 Job distribution data:', this.agentDistributionData)
                        this.agentDistributionData.forEach((dist: any) => {
                            console.log(`Agent ${dist.agent.name}: ${dist.wordLimit} words (${dist.percentage}%) - Speed: ${dist.speed} H/s`)
                        })
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
                    
                    console.log('🚀 Creating job with payload:', jobPayload)
                    
                    const result = await jobStore.actions.createJob(jobPayload)
                    
                    // Always check job state for success/warning messages
                    const jobState = jobStore.getState()
                    
                    console.log('📊 Job creation result:', { result, jobState })
                    
                    // Simplified success check - if we get a result or no error, consider it successful
                    const isSuccess = result !== null || (!jobState.error || jobState.error.includes('Warning:'))
                    
                    if (isSuccess) {
                        // Job creation successful (either single or distributed)
                        if (jobState.error && jobState.error.includes('Warning:')) {
                            // Distributed job with some failed agents
                            this.showNotification('Jobs created with warnings - some agents failed', 'warning')
                        } else {
                            // Full success
                            this.showNotification('Distributed jobs created successfully', 'success')
                        }
                        
                        console.log('✅ Job creation successful, closing modal...')
                        
                        // Always close modal and reset form on success
                        // Use setTimeout to ensure notification is shown before modal closes
                        setTimeout(() => {
                            this.closeJobModal()
                        }, 100)
                        
                        // Refresh jobs list to show the new job
                        await this.refreshJobsTable()
                        
                        // Small delay to ensure UI updates
                        setTimeout(() => {
                            console.log('✅ Job creation completed successfully')
                        }, 500)
                    } else {
                        // Job creation failed
                        const errorMsg = jobState.error || 'Failed to create job - server returned null'
                        this.showNotification(errorMsg, 'error')
                        console.error('❌ Job creation failed:', errorMsg)
                    }
                } catch (error) {
                    console.error('Job creation error:', error)
                    const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred'
                    this.showNotification(`Failed to create job: ${errorMessage}`, 'error')
                } finally {
                    this.isLoading = false
                    console.log('🏁 Job creation process finished')
                    
                    // Fallback: If modal is still open after 3 seconds, close it anyway
                    // This ensures the modal doesn't get stuck open
                    setTimeout(() => {
                        if (this.showJobModal) {
                            console.log('⚠️ Modal still open after 3 seconds, forcing close...')
                            this.closeJobModal()
                        }
                    }, 3000)
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
                this.wordlistForm = { file: null }
                // Also close large file warning if open
                this.closeLargeFileWarning()
                this.closeLargeFileWarningModal()
            },
            
            openLargeFileWarningModal() {
                this.showLargeFileWarningModal = true
            },
            
            closeLargeFileWarningModal() {
                this.showLargeFileWarningModal = false
            },

            // Handle wordlist file selection
            handleWordlistFileChange(event: Event) {
                const target = event.target as HTMLInputElement
                if (target.files && target.files[0]) {
                    const file = target.files[0]
                    this.wordlistForm.file = file
                    // console.log('File selected:', file.name, 'Size:', file.size)
                    
                    // File size check will be done when user clicks upload
                }
            },

            // Handle wordlist change and validate agent selection
            handleWordlistChange() {
                // Check if wordlist is too small and reset agent selection if needed
                if (this.isWordlistTooSmall) {
                    // If more than 1 agent is selected, reset to single agent
                    if (this.jobForm.agent_ids && this.jobForm.agent_ids.length > 1) {
                        this.jobForm.agent_ids = [this.jobForm.agent_ids[0]] // Keep only first agent
                        this.showNotification('Wordlist dengan kurang dari 1000 kata hanya dapat menggunakan 1 agent. Agent selection telah di-reset.', 'warning')
                    }
                }
                this.updateCommandTemplate()
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
                // Check if wordlist is too small for multi-agent selection
                if (this.isWordlistTooSmall) {
                    this.showNotification('Wordlists with less than 1000 words can only use one agent. Select agents individually.', 'warning')
                    return
                }

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
                // If wordlist is too small, disable select all functionality
                if (this.isWordlistTooSmall) {
                    return false
                }
                
                return this.onlineAgents.length > 0 && 
                       this.jobForm.agent_ids.length === this.onlineAgents.length &&
                       this.onlineAgents.every((agent: any) => this.jobForm.agent_ids.includes(agent.id))
            },

            // Handle individual agent selection with wordlist validation
            toggleAgentSelection(agentId: string) {
                if (!this.jobForm.agent_ids) {
                    this.jobForm.agent_ids = []
                }

                const isSelected = this.jobForm.agent_ids.includes(agentId)
                
                if (isSelected) {
                    // Remove agent from selection
                    this.jobForm.agent_ids = this.jobForm.agent_ids.filter((id: string) => id !== agentId)
                } else {
                    // Check if wordlist is too small
                    if (this.isWordlistTooSmall) {
                        // If wordlist is too small, only allow single agent selection
                        if (this.jobForm.agent_ids.length > 0) {
                            this.showNotification('Wordlists with less than 1000 words can only use one agent. Deselect other agents first.', 'warning')
                            return
                        }
                    }
                    
                    // Add agent to selection
                    this.jobForm.agent_ids.push(agentId)
                }
            },

            // Toggle compact mode for agent selection
            toggleCompactMode() {
                this.isCompactMode = !this.isCompactMode
            },

            // Distributed Job Functions
            openDistributedJobModal() {
                console.log('🚀 Opening distributed job modal')
                console.log('📊 Initial state:', {
                    agents: this.agents.length,
                    onlineAgents: this.onlineAgents.length,
                    hashFiles: this.hashFiles.length,
                    wordlists: this.wordlists.length
                })
                
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
                
                console.log('✅ Modal opened and form initialized')
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

            // Get agent performance score (0-100) with enhanced detection
            getAgentPerformanceScore(agent: any): number {
                // Use actual speed from database if available, fallback to capability-based estimation
                if (agent.speed && agent.speed > 0) {
                    return agent.speed // Use actual speed in H/s
                }
                
                // Fallback to capability-based estimation for agents without speed data
                const capabilities = (agent.capabilities || '').toLowerCase()
                
                if (capabilities.includes('rtx 4090') || capabilities.includes('rtx 4080')) {
                    return 5000000 // 5M H/s for high-end RTX
                } else if (capabilities.includes('rtx 4070') || capabilities.includes('rtx 3060')) {
                    return 4000000 // 4M H/s for mid-range RTX
                } else if (capabilities.includes('gtx 1660') || capabilities.includes('gtx 1070')) {
                    return 3000000 // 3M H/s for GTX series
                } else if (capabilities.includes('gpu') || capabilities.includes('cuda') || capabilities.includes('opencl')) {
                    return 3500000 // 3.5M H/s for generic GPU
                } else if (capabilities.includes('ryzen 9') || capabilities.includes('i9')) {
                    return 200000 // 200K H/s for high-end CPU
                } else if (capabilities.includes('ryzen 7') || capabilities.includes('i7')) {
                    return 150000 // 150K H/s for mid-range CPU
                } else {
                    return 100000 // 100K H/s for standard CPU
                }
            },

            // Calculate assigned word count for agent with improved distribution
            getAssignedWordCount(agent: any): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const totalWords = this.selectedWordlistCountNumber
                if (totalWords === 0) return 0
                
                const onlineAgents = this.onlineAgents
                if (onlineAgents.length === 0) return 0
                
                // Find the index of this agent
                const agentIndex = onlineAgents.findIndex(a => a.id === agent.id)
                if (agentIndex === -1) return 0
                
                // Calculate performance-based distribution with guaranteed accuracy
                const totalPerformance = onlineAgents.reduce((sum, a) => sum + this.getAgentPerformanceScore(a), 0)
                const agentPerformance = this.getAgentPerformanceScore(agent)
                const performanceRatio = agentPerformance / totalPerformance
                
                let wordsForAgent: number
                if (agentIndex === onlineAgents.length - 1) {
                    // Last agent gets remaining words to ensure total equals original
                    // Calculate how many words are already distributed
                    let distributedWords = 0
                    for (let i = 0; i < onlineAgents.length - 1; i++) {
                        const otherAgent = onlineAgents[i]
                        const otherPerformance = this.getAgentPerformanceScore(otherAgent)
                        const otherRatio = otherPerformance / totalPerformance
                        const otherWords = Math.floor(totalWords * otherRatio)
                        distributedWords += Math.max(otherWords, 2) // Minimum 2 words
                    }
                    wordsForAgent = totalWords - distributedWords
                } else {
                    // Calculate words for this agent
                    wordsForAgent = Math.floor(totalWords * performanceRatio)
                    // Ensure minimum 2 words per agent
                    wordsForAgent = Math.max(wordsForAgent, 2)
                }
                
                return wordsForAgent
            },

            // Get total distributed words to verify accuracy
            getTotalDistributedWords(): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const totalWords = this.selectedWordlistCountNumber
                if (totalWords === 0) return 0
                
                const onlineAgents = this.onlineAgents
                if (onlineAgents.length === 0) return 0
                
                // Calculate total distributed words with guaranteed accuracy
                let totalDistributed = 0
                let remainingWords = totalWords
                
                for (let i = 0; i < onlineAgents.length; i++) {
                    const agent = onlineAgents[i]
                const totalPerformance = onlineAgents.reduce((sum, a) => sum + this.getAgentPerformanceScore(a), 0)
                const agentPerformance = this.getAgentPerformanceScore(agent)
                const performanceRatio = agentPerformance / totalPerformance
                
                    let wordsForAgent: number
                    if (i === onlineAgents.length - 1) {
                        // Last agent gets remaining words to ensure total equals original
                        wordsForAgent = remainingWords
                    } else {
                        wordsForAgent = Math.floor(totalWords * performanceRatio)
                        wordsForAgent = Math.max(wordsForAgent, 2) // Minimum 2 words
                        remainingWords -= wordsForAgent
                    }
                    
                    totalDistributed += wordsForAgent
                }
                
                return totalDistributed
            },

            // NEW: Accurate word distribution algorithm
            getAccurateAssignedWordCount(agent: any): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const totalWords = this.selectedWordlistCountNumber
                if (totalWords === 0) return 0
                
                const onlineAgents = this.onlineAgents
                if (onlineAgents.length === 0) return 0
                
                // Find the index of this agent
                const agentIndex = onlineAgents.findIndex(a => a.id === agent.id)
                if (agentIndex === -1) return 0
                
                // Calculate performance scores for all agents
                const agentScores = onlineAgents.map(a => ({
                    agent: a,
                    performance: this.getAgentPerformanceScore(a)
                }))
                
                const totalPerformance = agentScores.reduce((sum, item) => sum + item.performance, 0)
                
                // Calculate words for each agent with guaranteed accuracy
                let distributedWords = 0
                let wordsForThisAgent = 0
                
                for (let i = 0; i < onlineAgents.length; i++) {
                    const currentAgent = agentScores[i]
                    const performanceRatio = currentAgent.performance / totalPerformance
                    
                    let wordsForCurrentAgent: number
                    if (i === onlineAgents.length - 1) {
                        // Last agent gets remaining words
                        wordsForCurrentAgent = totalWords - distributedWords
                    } else {
                        // Calculate words based on performance ratio
                        wordsForCurrentAgent = Math.floor(totalWords * performanceRatio)
                        // Ensure minimum 2 words
                        wordsForCurrentAgent = Math.max(wordsForCurrentAgent, 2)
                        distributedWords += wordsForCurrentAgent
                    }
                    
                    // If this is the agent we're looking for, store the result
                    if (i === agentIndex) {
                        wordsForThisAgent = wordsForCurrentAgent
                    }
                }
                
                return wordsForThisAgent
            },

            // NEW: Get accurate total distributed words
            getAccurateTotalDistributedWords(): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const totalWords = this.selectedWordlistCountNumber
                if (totalWords === 0) return 0
                
                const onlineAgents = this.onlineAgents
                if (onlineAgents.length === 0) return 0
                
                // Calculate total distributed words with guaranteed accuracy
                let totalDistributed = 0
                let remainingWords = totalWords
                
                for (let i = 0; i < onlineAgents.length; i++) {
                    const agent = onlineAgents[i]
                    const totalPerformance = onlineAgents.reduce((sum, a) => sum + this.getAgentPerformanceScore(a), 0)
                    const agentPerformance = this.getAgentPerformanceScore(agent)
                    const performanceRatio = agentPerformance / totalPerformance
                    
                    let wordsForAgent: number
                    if (i === onlineAgents.length - 1) {
                        // Last agent gets remaining words to ensure total equals original
                        wordsForAgent = remainingWords
                    } else {
                        wordsForAgent = Math.floor(totalWords * performanceRatio)
                        wordsForAgent = Math.max(wordsForAgent, 2) // Minimum 2 words
                        remainingWords -= wordsForAgent
                    }
                    
                    totalDistributed += wordsForAgent
                }
                
                return totalDistributed
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

            // Update distribution preview with detailed information
            updateDistributionPreview() {
                // This will trigger Alpine.js reactivity for the preview section
                console.log('🔄 updateDistributionPreview called')
                console.log('📊 Current state:', {
                    wordlistId: this.distributedJobForm.wordlist_id,
                    onlineAgentsCount: this.onlineAgents.length,
                    agents: this.onlineAgents,
                    wordlists: this.wordlists,
                    selectedWordlist: this.wordlists.find(w => w.id === this.distributedJobForm.wordlist_id)
                })
                
                setTimeout(() => {
                    // Force Alpine.js to re-evaluate computed properties
                    console.log('✅ Distribution preview updated')
                }, 100)
            },

            // Get detailed distribution information for debugging
            getDistributionDetails() {
                if (!this.distributedJobForm.wordlist_id || this.onlineAgents.length === 0) {
                    return null
                }
                
                const totalWords = this.selectedWordlistCountNumber
                const agents = this.onlineAgents
                
                const distribution = agents.map(agent => {
                    const performance = this.getAgentPerformanceScore(agent)
                    const assignedWords = this.getAssignedWordCount(agent)
                    const percentage = totalWords > 0 ? ((assignedWords / totalWords) * 100).toFixed(1) : '0.0'
                    
                    return {
                        name: agent.name,
                        capabilities: agent.capabilities,
                        performance: performance,
                        assignedWords: assignedWords,
                        percentage: percentage,
                        resourceType: this.isGPUAgent(agent) ? 'GPU' : 'CPU'
                    }
                })
                
                return {
                    totalWords: totalWords,
                    totalAgents: agents.length,
                    distribution: distribution
                }
            },

            // Check if can create distributed job
            canCreateDistributedJob(): boolean {
                return !!(this.distributedJobForm.name && 
                         this.distributedJobForm.hash_file_id && 
                         this.distributedJobForm.wordlist_id &&
                         this.onlineAgents.length > 0)
            },

            // Check if can preview distribution (same validation as create)
            canPreviewDistribution(): boolean {
                return !!(this.distributedJobForm.name && 
                         this.distributedJobForm.hash_file_id && 
                         this.distributedJobForm.wordlist_id &&
                         this.onlineAgents.length > 0)
            },

            // Preview distribution
            previewDistribution() {
                if (!this.canPreviewDistribution()) {
                    this.showNotification('Please complete all required fields: Job name, WiFi Handshake File, Wordlist, and ensure agents are available', 'warning')
                    return
                }
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
                    // Prepare job data for distributed creation
                    const jobData = {
                        name: this.distributedJobForm.name,
                        hash_type: parseInt(this.distributedJobForm.hash_type) || 2500,
                        attack_mode: parseInt(this.distributedJobForm.attack_mode) || 0,
                        hash_file_id: this.distributedJobForm.hash_file_id,
                        wordlist_id: this.distributedJobForm.wordlist_id,
                        agent_ids: this.onlineAgents.map(agent => agent.id) // Use all online agents
                    }

                    // Create distributed job using API
                    const result = await jobStore.actions.createJob(jobData)
                    
                    if (result) {
                        this.showNotification('Distributed job created successfully!', 'success')
                    this.closeDistributedJobModal()
                        // Refresh jobs list
                        await this.refreshJobsTable()
                    } else {
                    this.showNotification('Failed to create distributed job', 'error')
                    }
                } catch (error) {
                    console.error('Distributed job creation error:', error)
                    this.showNotification('Failed to create distributed job: ' + (error as Error).message, 'error')
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

            handleWordlistUpload() {
                // console.log('handleWordlistUpload called, wordlistForm.file:', this.wordlistForm.file)
                if (!this.wordlistForm.file) {
                    this.showNotification('Please select a file to upload', 'error')
                    return
                }
                
                // File validation will be done in uploadWordlist
                this.uploadWordlist(this.wordlistForm.file)
            },

            async uploadWordlist(file: File) {
                if (!file) {
                    this.showNotification('Please select a file to upload', 'error')
                    return
                }
                
                // Check file size (1GB = 1073741824 bytes)
                const maxSize = 1073741824 // 1GB in bytes
                if (file.size > maxSize) {
                    this.openLargeFileWarningModal()
                    return
                }
                
                await this.performWordlistUpload(file)
            },

            async performWordlistUpload(file: File) {
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


            closeLargeFileWarning() {
                this.showLargeFileWarning = false
                this.largeFileInfo = null
            },

            async proceedWithLargeFileUpload() {
                this.closeLargeFileWarning()
                this.closeLargeFileWarningModal()
                if (this.wordlistForm.file) {
                    await this.performWordlistUpload(this.wordlistForm.file)
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
                        // Only show notification if explicitly requested, not on auto-refresh
                        if (this.showCacheStatsNotification) {
                            this.showNotification('Cache stats refreshed', 'info')
                            this.showCacheStatsNotification = false
                        }
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

            // Get distribution percentage for agent
            getDistributionPercentage(agent: any): number {
                if (!this.distributedJobForm.wordlist_id) return 0
                
                const totalWords = this.selectedWordlistCountNumber
                if (totalWords === 0) return 0
                
                const assignedWords = this.getAccurateAssignedWordCount(agent)
                return Math.round((assignedWords / totalWords) * 100)
            },

            // Verify distribution accuracy
            isDistributionAccurate(): boolean {
                const totalWords = this.selectedWordlistCountNumber
                const totalDistributed = this.getAccurateTotalDistributedWords()
                return totalWords === totalDistributed
            },

            // Get distribution summary for debugging
            getDistributionSummary(): any {
                if (!this.distributedJobForm.wordlist_id) return null
                
                const totalWords = this.selectedWordlistCountNumber
                const agents = this.onlineAgents
                
                const summary = agents.map(agent => {
                    const performance = this.getAgentPerformanceScore(agent)
                    const assignedWords = this.getAccurateAssignedWordCount(agent)
                    const percentage = this.getDistributionPercentage(agent)
                    const resourceType = this.isGPUAgent(agent) ? 'GPU' : 'CPU'
                    
                    return {
                        name: agent.name,
                        capabilities: agent.capabilities,
                        resourceType: resourceType,
                        performance: performance,
                        assignedWords: assignedWords,
                        percentage: percentage
                    }
                })
                
                const totalDistributed = this.getAccurateTotalDistributedWords()
                const isAccurate = this.isDistributionAccurate()
                
                return {
                    totalWords: totalWords,
                    totalDistributed: totalDistributed,
                    isAccurate: isAccurate,
                    agents: summary
                }
            },

            // Get total speed of all selected agents
            getTotalSelectedAgentSpeed(): number {
                if (!this.jobForm.agent_ids || this.jobForm.agent_ids.length === 0) return 0
                
                return this.jobForm.agent_ids.reduce((total: number, agentId: string) => {
                    const agent = this.agents.find((a: any) => a.id === agentId)
                    return total + (agent?.speed || 0)
                }, 0)
            },



        }))

        // Add fallback check for login success notification after Alpine initialization
        setTimeout(() => {
            const showLoginSuccess = sessionStorage.getItem('showLoginSuccess')
            // console.log('🔍 Fallback: Checking sessionStorage:', showLoginSuccess)
            // console.log('🔍 Fallback: All sessionStorage keys:', Object.keys(sessionStorage))
            // console.log('🔍 Fallback: Alpine available:', !!window.Alpine)
            
            if (showLoginSuccess === 'true') {
                // console.log('🔍 Fallback: Found login success flag, checking notification...')
                
                // Try multiple ways to access Alpine component
                let dashboardApp = null
                
                // Method 1: Direct Alpine data access
                if (window.Alpine && window.Alpine.data) {
                    dashboardApp = window.Alpine.data('dashboardApp')
                    // console.log('🔍 Fallback: Method 1 - Alpine data:', !!dashboardApp)
                }
                
                // Method 2: Try to find Alpine component in DOM
                if (!dashboardApp) {
                    const alpineElement = document.querySelector('[x-data*="dashboardApp"]')
                    if (alpineElement && (alpineElement as any)._x_dataStack) {
                        dashboardApp = (alpineElement as any)._x_dataStack[0]
                        // console.log('🔍 Fallback: Method 2 - DOM element:', !!dashboardApp)
                    }
                }
                
                // Method 3: Try global window access
                if (!dashboardApp) {
                    dashboardApp = (window as any).dashboardApp
                    // console.log('🔍 Fallback: Method 3 - Global window:', !!dashboardApp)
                }
                
                // Method 4: Try to access via Alpine store
                if (!dashboardApp && window.Alpine && window.Alpine.store) {
                    try {
                        dashboardApp = window.Alpine.store('dashboardApp')
                        // console.log('🔍 Fallback: Method 4 - Alpine store:', !!dashboardApp)
                    } catch (e) {
                        // console.log('🔍 Fallback: Method 4 failed:', e)
                    }
                }
                
                // console.log('🔍 Fallback: Dashboard app available:', !!dashboardApp)
                
                if (dashboardApp) {
                    // console.log('🔍 Fallback: Setting showLoginSuccessNotification to true')
                    dashboardApp.showLoginSuccessNotification = true
                    // console.log('📊 Fallback: State set to:', dashboardApp.showLoginSuccessNotification)
                    
                    // Force trigger notification immediately
                    if (typeof dashboardApp.showLoginSuccessNotificationNow === 'function') {
                        // console.log('🔍 Fallback: Calling showLoginSuccessNotificationNow')
                        dashboardApp.showLoginSuccessNotificationNow()
                    } else if (typeof dashboardApp.checkLoginSuccessNotification === 'function') {
                        // console.log('🔍 Fallback: Calling checkLoginSuccessNotification')
                        dashboardApp.checkLoginSuccessNotification()
                    }
                } else {
                    // console.log('❌ Fallback: Could not access Alpine component, trying direct DOM manipulation')
                    // Fallback: Direct DOM manipulation
                    this.triggerNotificationDirectly()
                }
            } else {
                // console.log('🔍 Fallback: No login success flag found')
            }
        }, 500)

        perf.endTimer('alpine-initialization')
    }

    // Direct DOM manipulation fallback
    private triggerNotificationDirectly() {
        console.log('🚀 ===== DIRECT DOM MANIPULATION FALLBACK =====')
        
        // Try to find and show notification elements
        const fixedNotification = document.querySelector('[x-show="showLoginSuccessNotification"]')
        const pageNotification = document.querySelector('[x-show="showLoginSuccessNotification"]')
        
        if (fixedNotification) {
            console.log('🚀 Found fixed notification element, showing...')
            const element = fixedNotification as HTMLElement
            element.style.display = 'block'
            element.style.opacity = '1'
            element.style.transform = 'translateY(0)'
        }
        
        if (pageNotification) {
            console.log('🚀 Found page notification element, showing...')
            const element = pageNotification as HTMLElement
            element.style.display = 'block'
            element.style.opacity = '1'
            element.style.transform = 'scale(1) translateY(0)'
        }
        
        // Check if notification has already been shown for this session
        const notificationShown = sessionStorage.getItem('loginSuccessNotificationShown')
        if (notificationShown === 'true') {
            console.log('🚫 Login success notification already shown in this session, skipping direct toast...')
            return
        }
        
        // Also try to show toast notifications
        this.showDirectToastNotification('🎉 Welcome to Hashcat Dashboard!', 'success')
        setTimeout(() => {
            this.showDirectToastNotification('✅ Login successful! You are now logged in.', 'success')
        }, 1000)
        setTimeout(() => {
            this.showDirectToastNotification('🚀 Ready to start cracking passwords!', 'info')
        }, 2000)
        
        console.log('🚀 ================================================')
    }

    // Direct toast notification fallback
    private showDirectToastNotification(message: string, type: string) {
        // console.log(`🔔 Direct toast: [${type.toUpperCase()}] ${message}`)
        
        // Check if we're on login page and block login success notifications
        const isLoginSuccess = message.includes('Login successful')
        const isLoginError = message.includes('Authentication Failed') || message.includes('Login failed')
        const currentPath = window.location.pathname
        
        if (currentPath === '/login' && !isLoginError) {
            // console.log(`🔔 Direct toast blocked on login page: [${type.toUpperCase()}] ${message}`)
            return
        }
        
        // Check if login success notification has already been shown for this session
        if (isLoginSuccess) {
            const notificationShown = sessionStorage.getItem('loginSuccessNotificationShown')
            if (notificationShown === 'true') {
                console.log('🚫 Login success notification already shown in this session, skipping direct toast...')
                return
            }
        }
        
        // Create notification element
        const notification = document.createElement('div')
        notification.className = 'fixed top-4 right-4 z-[99999] p-4 rounded-lg shadow-lg max-w-sm'
        
        // Set colors based on type
        if (type === 'success') {
            notification.className += ' bg-green-50 border-l-4 border-green-400'
        } else if (type === 'error') {
            notification.className += ' bg-red-50 border-l-4 border-red-400'
        } else if (type === 'warning') {
            notification.className += ' bg-yellow-50 border-l-4 border-yellow-400'
        } else {
            notification.className += ' bg-blue-50 border-l-4 border-blue-400'
        }
        
        notification.innerHTML = `
            <div class="flex items-start">
                <div class="flex-shrink-0">
                    <div class="w-6 h-6 rounded-full flex items-center justify-center text-white ${
                        type === 'success' ? 'bg-green-400' : 
                        type === 'error' ? 'bg-red-400' : 
                        type === 'warning' ? 'bg-yellow-400' : 'bg-blue-400'
                    }">
                        <i class="fas text-xs ${
                            type === 'success' ? 'fa-check' : 
                            type === 'error' ? 'fa-exclamation-triangle' : 
                            type === 'warning' ? 'fa-exclamation' : 'fa-info'
                        }"></i>
                    </div>
                </div>
                <div class="ml-3 flex-1">
                    <p class="text-sm font-medium ${
                        type === 'success' ? 'text-green-800' : 
                        type === 'error' ? 'text-red-800' : 
                        type === 'warning' ? 'text-yellow-800' : 'text-blue-800'
                    }">${message}</p>
                </div>
                <div class="ml-3 flex-shrink-0">
                    <button onclick="this.parentElement.parentElement.parentElement.remove()" 
                            class="inline-flex rounded-md p-1.5 transition-colors hover:bg-black/5 ${
                                type === 'success' ? 'text-green-400' : 
                                type === 'error' ? 'text-red-400' : 
                                type === 'warning' ? 'text-yellow-400' : 'text-blue-400'
                            }">
                        <i class="fas fa-times text-xs"></i>
                    </button>
                </div>
            </div>
        `
        
        // Add to DOM
        document.body.appendChild(notification)
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove()
            }
        }, 5000)
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
