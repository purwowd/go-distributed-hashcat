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
import { authStore, createAuthData } from './stores/auth.store'

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
                { name: 'pages/login', path: '/components/pages/login.html' },
                { name: 'layout/navigation', path: '/components/layout/navigation.html' },
                { name: 'ui/breadcrumb', path: '/components/ui/breadcrumb.html' },
                { name: 'tabs/overview', path: '/components/tabs/overview.html' },
                { name: 'tabs/agents', path: '/components/tabs/agents.html' },
                { name: 'tabs/jobs', path: '/components/tabs/jobs.html' },
                { name: 'tabs/files', path: '/components/tabs/files.html' },
                { name: 'tabs/wordlists', path: '/components/tabs/wordlists.html' },
                { name: 'pages/agent-keys', path: '/components/pages/agent-keys.html' },
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
        // console.log('🔧 Starting authentication-aware component injection...')
        
        // Load login component first (always available)
        const loginHtml = await componentLoader.loadComponent('pages/login')
        const loginContainer = document.getElementById('login-container')
        if (loginContainer && loginHtml) {
            loginContainer.innerHTML = loginHtml
            // console.log('✅ Login component injected')
        }

        // Find or create dashboard container
        let dashboardContainer = document.getElementById('dashboard-container')
        if (!dashboardContainer) {
            dashboardContainer = document.getElementById('dashboard-container')
        }

        if (!dashboardContainer) {
            console.error('❌ Dashboard container not found')
            return
        }

        // Create main container for dashboard components
        let mainContainer = document.querySelector('main.container-modern')
        if (!mainContainer) {
            mainContainer = document.createElement('main')
            mainContainer.className = 'container-modern'
            dashboardContainer.appendChild(mainContainer)
        }

        // Load navigation
        const navigation = await componentLoader.loadComponent('layout/navigation')
        const navigationContainer = document.createElement('div')
        navigationContainer.innerHTML = navigation
        
        const navElement = navigationContainer.querySelector('nav')
        if (navElement) {
            dashboardContainer.insertBefore(navElement, mainContainer)
            // console.log('✅ Navigation injected into dashboard')
        }

        // Load breadcrumb component
        const breadcrumb = await componentLoader.loadComponent('ui/breadcrumb')
        const breadcrumbContainer = document.createElement('div')
        breadcrumbContainer.innerHTML = breadcrumb
        
        const breadcrumbElement = breadcrumbContainer.querySelector('nav')
        if (breadcrumbElement) {
            dashboardContainer.insertBefore(breadcrumbElement, mainContainer)
            // console.log('✅ Breadcrumb injected into dashboard')
        }

        // Load tab content with proper routing visibility
        const tabComponents = [
            { name: 'tabs/overview', route: 'overview' },
            { name: 'tabs/agents', route: 'agents' }, 
            { name: 'tabs/jobs', route: 'jobs' },
            { name: 'tabs/files', route: 'files' },
            { name: 'tabs/wordlists', route: 'wordlists' },
            { name: 'pages/agent-keys', route: 'agent-keys' },
            { name: 'tabs/docs', route: 'docs' }
        ]

        for (const component of tabComponents) {
            const html = await componentLoader.loadComponent(component.name)
            const container = document.createElement('div')
            container.innerHTML = html
            const element = container.querySelector('section, div, article') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                // Add route-based visibility attributes
                element.setAttribute('x-show', `currentTab === '${component.route}'`)
                element.setAttribute('x-transition:enter', 'transition ease-out duration-300')
                element.setAttribute('x-transition:enter-start', 'opacity-0 transform scale-95')
                element.setAttribute('x-transition:enter-end', 'opacity-100 transform scale-100')
                element.setAttribute('x-transition:leave', 'transition ease-in duration-200')
                element.setAttribute('x-transition:leave-start', 'opacity-100 transform scale-100')
                element.setAttribute('x-transition:leave-end', 'opacity-0 transform scale-95')
                
                mainContainer.appendChild(element)
                // console.log(`✅ Injected ${component.name} component with route visibility`)
            }
        }

        // Load modals (available globally)
        const modalComponents = [
            'modals/agent-modal',
            'modals/job-modal',
            'modals/file-modal', 
            'modals/wordlist-modal',
            
        ]

        for (const component of modalComponents) {
            const html = await componentLoader.loadComponent(component)
            const container = document.createElement('div')
            container.innerHTML = html
            const element = container.querySelector('div[x-data], [x-show]') || container.querySelector('div') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                document.body.appendChild(element)
                // console.log(`✅ Injected ${component} modal`)
            }
        }

        // Load UI components (available globally)
        const uiComponents = [
            'ui/notification',
            'ui/loading'
        ]

        for (const component of uiComponents) {
            const html = await componentLoader.loadComponent(component)
            const container = document.createElement('div')
            container.innerHTML = html
            const element = container.querySelector('div, section') || container.firstElementChild
            if (element && element.tagName !== 'SCRIPT') {
                document.body.appendChild(element)
                // console.log(`✅ Injected ${component} UI component`)
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
            currentTab: router.getCurrentRoute() || 'overview',
            isLoading: false,
            isAlpineInitialized: false,
            notifications: [] as any[],
            
            // Modal states
            showAgentModal: false,
            showJobModal: false,
            showFileModal: false,
            showWordlistModal: false,
            showGenerateModal: false,
            showKeyModal: false,
            
            // Form states
            agentForm: { name: '', ip_address: '', port: 8080, capabilities: '' },
            jobForm: { name: '', hash_file_id: '', wordlist_id: '', agent_id: '', hash_type: '', attack_mode: '' },
            fileForm: { file: null },
            wordlistForm: { file: null },
            generateForm: { name: '', description: '' },
            generatedKey: null as any,
            isGenerating: false,
            
            // Command template for job creation
            commandTemplate: '',
            
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
            reactiveJobs: [] as any[],
            reactiveHashFiles: [] as any[],
            reactiveWordlists: [] as any[],
            agentKeys: [] as any[],
            
            // Auth state for easy access
            auth: createAuthData(),

            // Getters that return reactive data
            get agents() { 
                // Filter out agents with status "key_generated" (these are just placeholder records for agent keys)
                const allAgents = this.reactiveAgents || []
                return allAgents.filter((agent: any) => agent.status !== 'key_generated')
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
            
            // Get all agents including key_generated ones (for agent keys page)
            get allAgents() {
                return this.reactiveAgents || []
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

            // Methods
            async init() {
                // console.log('🔄 Initializing Alpine.js dashboard data...')
                this.isAlpineInitialized = true
                
                // Ensure currentTab is set correctly on init
                this.currentTab = router.getCurrentRoute() || 'overview'
                
                // Setup router listener
                router.subscribe((route: string) => {
                    this.currentTab = route
                })
                
                // Setup store subscriptions for reactivity
                this.setupStoreSubscriptions()
                
                // Setup auth store subscription for UI visibility control
                authStore.subscribe(() => {
                    this.updateUIVisibility()
                })
                
                // Setup WebSocket for real-time updates
                this.setupWebSocketSubscriptions()
                
                // Initial auth check and UI setup
                this.updateUIVisibility()
                
                try {
                    await this.loadInitialData()
                    // console.log('🎉 Dashboard initialization complete')
                } catch (error) {
                    console.error('❌ Dashboard initialization failed:', error)
                    this.showNotification('Failed to initialize dashboard', 'error')
                }
                
                this.setupPolling()
                
                // Hide initial loader
                this.hideInitialLoader()
                
                // Safety timeout to prevent infinite loading
                setTimeout(() => {
                    if (this.isLoading) {
                        console.warn('⚠️ Loading timeout reached, forcing stop')
                        this.forceStopLoading()
                        this.showNotification('Loading took too long. Data may be incomplete.', 'warning')
                    }
                }, 15000) // 15 second safety timeout
            },

            // Update UI visibility based on auth state
            updateUIVisibility() {
                const isAuthenticated = authStore.isAuthenticated()
                const currentUser = authStore.getUser()
                const loginContainer = document.getElementById('login-container')
                const dashboardContainer = document.getElementById('dashboard-container')
                
                console.log('🔄 Auth state update:', { isAuthenticated, currentUser })
                
                if (isAuthenticated) {
                    // Show dashboard, hide login
                    if (loginContainer) loginContainer.style.display = 'none'
                    if (dashboardContainer) dashboardContainer.style.display = 'block'
                    console.log('✅ Showing dashboard, hiding login')
                    
                    // Force Alpine.js to re-evaluate auth data
                    this.auth = createAuthData()
                    
                    // Force all Alpine components to re-render
                    if (window.Alpine && window.Alpine.nextTick) {
                        window.Alpine.nextTick(() => {
                            // Force reactive update for all auth-related components
                            const authElements = document.querySelectorAll('[x-show*="auth"]')
                            authElements.forEach((el: any) => {
                                if (el._x_dataStack) {
                                    // Update Alpine data stack for auth
                                    el._x_dataStack.forEach((data: any) => {
                                        if (data.auth) {
                                            data.auth = createAuthData()
                                        }
                                    })
                                }
                            })
                        })
                    }
                } else {
                    // Show login, hide dashboard
                    if (loginContainer) loginContainer.style.display = 'block'
                    if (dashboardContainer) dashboardContainer.style.display = 'none'
                    console.log('🔐 Showing login, hiding dashboard')
                }
            },

            // Hide initial loader
            hideInitialLoader() {
                const loader = document.getElementById('initial-loader')
                if (loader) {
                    setTimeout(() => {
                        loader.style.opacity = '0'
                        loader.style.transition = 'opacity 0.3s ease-out'
                        setTimeout(() => loader.remove(), 300)
                    }, 1000)
                }
            },

            // NEW: Setup store subscriptions for reactive updates
            setupStoreSubscriptions() {
                // Subscribe to agent store changes
                agentStore.subscribe(() => {
                    const state = agentStore.getState()
                    // ✅ Force Alpine.js reactivity by creating new array reference
                    this.reactiveAgents = [...(state.agents || [])]
                    console.log('🔄 Agent store updated:', this.reactiveAgents.length, 'agents')
                })
                
                // Subscribe to job store changes
                jobStore.subscribe(() => {
                    const state = jobStore.getState()
                    // ✅ Force Alpine.js reactivity by creating new array reference
                    this.reactiveJobs = [...(state.jobs || [])]
                    console.log('🔄 Job store updated:', this.reactiveJobs.length, 'jobs')
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
                    await Promise.race([
                        timeout,
                        jobStore.actions.fetchJobs().catch(err => {
                            console.warn('Failed to load jobs:', err)
                            return []
                        })
                    ])
                    
                    // Load cache stats
                    await this.refreshCacheStats().catch(err => {
                        console.warn('Failed to load cache stats:', err)
                    })
                    
                    // Load agent keys
                    await this.loadAgentKeys().catch(err => {
                        console.warn('Failed to load agent keys:', err)
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
                
                // Load data for specific tabs
                if (tab === 'agent-keys') {
                    await this.loadAgentKeys()
                }
                
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
                // Check if there are online agents
                if (this.onlineAgents.length === 0) {
                    this.showNotification('No online agents available. Please start an agent first before creating jobs.', 'warning')
                    return
                }
                
                this.showJobModal = true
                this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', agent_id: '', hash_type: '2500', attack_mode: '0' }
            },
            
            closeJobModal() {
                this.showJobModal = false
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
                        agent_id: jobData.agent_id || undefined   // Include agent assignment if specified
                    }
                    
                    // Validate required fields before sending
                    if (!jobPayload.name || !jobPayload.hash_file_id || !jobPayload.wordlist || 
                        jobPayload.hash_type === undefined || jobPayload.attack_mode === undefined) {
                        this.showNotification('Please fill in all required fields', 'error')
                        return
                    }
                    
                    // Validate agent assignment is required
                    if (!jobPayload.agent_id) {
                        this.showNotification('Please select an agent to run this job', 'error')
                        return
                    }
                    
                    // Validate selected agent is online
                    const selectedAgent = this.agents.find((a: any) => a.id === jobPayload.agent_id)
                    if (!selectedAgent) {
                        this.showNotification('Selected agent not found', 'error')
                        return
                    }
                    
                    if (selectedAgent.status !== 'online') {
                        this.showNotification(`Cannot create job: Agent "${selectedAgent.name}" is ${selectedAgent.status}`, 'error')
                        return
                    }
                    

                    
                    const result = await jobStore.actions.createJob(jobPayload)
                    if (result) {
                        this.showNotification('Job created successfully!', 'success')
                        this.showJobModal = false
                        this.jobForm = { name: '', hash_file_id: '', wordlist_id: '', agent_id: '', hash_type: '2500', attack_mode: '0' }
                        
                        // Refresh jobs list to show the new job
                        await jobStore.actions.fetchJobs()
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
                
                if (!this.jobForm.hash_file_id || !this.jobForm.wordlist_id) {
                    this.commandTemplate = 'hashcat command will appear here...'
                    return
                }

                // Get file names for display
                const hashFile = this.hashFiles.find((f: any) => f.id === this.jobForm.hash_file_id)
                const wordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
                
                const hashFileName = hashFile ? (hashFile.orig_name || hashFile.name) : 'hashfile'
                const wordlistName = wordlist ? (wordlist.orig_name || wordlist.name) : 'wordlist'

                // Build hashcat command for WPA/WPA2 Dictionary Attack
                this.commandTemplate = `hashcat -m 2500 -a 0 ${hashFileName} ${wordlistName}`
                
                // Add WPA/WPA2 specific optimizations
                this.commandTemplate += ' -O --force --status --status-timer=5'
                
                // Add session name
                if (this.jobForm.name) {
                    const sessionName = this.jobForm.name.toLowerCase().replace(/[^a-z0-9]/g, '_')
                    this.commandTemplate += ` --session=${sessionName}`
                }
                
                // Add outfile for WPA/WPA2
                this.commandTemplate += ' --outfile=cracked.txt --outfile-format=2'
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

            // Agent Keys Management
            async loadAgentKeys() {
                try {
                    const { AgentKeyService } = await import('./services/agent-key.service')
                    const agentKeyService = new AgentKeyService()
                    const keys = await agentKeyService.listAgentKeys()
                    // ✅ Force Alpine.js reactivity by creating new array reference
                    this.agentKeys = [...keys]
                } catch (error) {
                    console.error('Failed to load agent keys:', error)
                    this.showNotification('Failed to load agent keys', 'error')
                }
            },

            async generateKey() {
                if (!this.generateForm.name) {
                    this.showNotification('Key name is required', 'error')
                    return
                }

                this.isGenerating = true
                try {
                    const { AgentKeyService } = await import('./services/agent-key.service')
                    const agentKeyService = new AgentKeyService()
                    
                    this.generatedKey = await agentKeyService.generateAgentKey({
                        name: this.generateForm.name,
                        description: this.generateForm.description
                    })
                    
                    this.showGenerateModal = false
                    this.showKeyModal = true
                    this.generateForm = { name: '', description: '' }
                    this.showNotification('Agent key generated successfully!', 'success')
                    
                    // Refresh the agent keys list
                    await this.loadAgentKeys()
                } catch (error) {
                    console.error('Failed to generate agent key:', error)
                    this.showNotification('Failed to generate agent key', 'error')
                } finally {
                    this.isGenerating = false
                }
            },

            async revokeKey(agentKey: string) {
                if (!confirm('Are you sure you want to revoke this agent key? This will disable the key but keep it for audit purposes.')) {
                    return
                }

                try {
                    const { AgentKeyService } = await import('./services/agent-key.service')
                    const agentKeyService = new AgentKeyService()
                    
                    await agentKeyService.revokeAgentKey(agentKey)
                    await this.loadAgentKeys() // Refresh the list
                    this.showNotification('Agent key revoked successfully', 'success')
                } catch (error) {
                    console.error('Failed to revoke agent key:', error)
                    this.showNotification('Failed to revoke agent key', 'error')
                }
            },

            async deleteKey(agentKey: string) {
                if (!confirm('Are you sure you want to permanently delete this agent key? This action cannot be undone and will remove all history.')) {
                    return
                }

                try {
                    const { AgentKeyService } = await import('./services/agent-key.service')
                    const agentKeyService = new AgentKeyService()
                    
                    await agentKeyService.deleteAgentKey(agentKey)
                    await this.loadAgentKeys() // Refresh the list
                    this.showNotification('Agent key deleted successfully', 'success')
                } catch (error) {
                    console.error('Failed to delete agent key:', error)
                    const errorMessage = error instanceof Error ? error.message : 'Failed to delete agent key'
                    this.showNotification(errorMessage, 'error')
                }
            },

            // Get agent name by agent key
            getAgentNameByKey(key: any) {
                if (!key.agent_id) return 'Not connected'
                
                const agent = this.allAgents.find((agent: any) => agent.id === key.agent_id)
                return agent ? agent.name : 'Unknown Agent'
            },


        }))

        // Register auth data
        const authData = createAuthData()
        window.Alpine.data('auth', () => authData)

        // Subscribe to auth store changes to force Alpine reactivity
        authStore.subscribe(() => {
            console.log('🔔 Auth store changed, triggering UI update')
            
            // Simple approach: just update container visibility
            // Alpine.js will handle reactivity through its own getters
            try {
                const isAuthenticated = authStore.isAuthenticated()
                const loginContainer = document.getElementById('login-container')
                const dashboardContainer = document.getElementById('dashboard-container')
                
                if (isAuthenticated) {
                    if (loginContainer) loginContainer.style.display = 'none'
                    if (dashboardContainer) dashboardContainer.style.display = 'block'
                    console.log('✅ Showing dashboard, hiding login')
                } else {
                    if (loginContainer) loginContainer.style.display = 'block'
                    if (dashboardContainer) dashboardContainer.style.display = 'none'
                    console.log('🔐 Showing login, hiding dashboard')
                }
                
                // Let Alpine.js handle its own reactivity through getters
                // No need to manually update proxy objects
                console.log('🔄 UI containers updated, Alpine.js will handle reactivity')
                
            } catch (error) {
                console.error('Error updating auth UI:', error)
            }
        })

        // Make auth functions globally available for login component
        window.Alpine.data('loginPageData', () => ({
            showPassword: false,
            isLoading: false,
            error: null as string | null,
            form: {
                username: '',
                password: ''
            },

            // Form validation
            get isFormValid() {
                return this.form.username && this.form.password
            },

            // Handle form submission
            async handleSubmit() {
                if (!this.isFormValid) return

                this.isLoading = true
                this.error = null

                try {
                    // Login user
                    await this.login()
                } catch (error: any) {
                    this.error = error.message || 'An error occurred'
                } finally {
                    this.isLoading = false
                }
            },

            // Login user
            async login() {
                await authStore.login({
                    username: this.form.username,
                    password: this.form.password
                })
                
                // Force immediate UI update after successful login
                console.log('🎉 Login successful!')
                
                // Manually trigger updateUIVisibility to ensure immediate UI update
                setTimeout(() => {
                    // Force auth store subscription trigger
                    authStore.forceUpdate()
                }, 100)
            },

            // Reset form
            resetForm() {
                this.form = {
                    username: '',
                    password: ''
                }
                this.showPassword = false
            },

            // Clear messages
            clearMessages() {
                this.error = null
            },

            // Initialize
            init() {
                // No special initialization needed for login-only mode
            },

            // Get auth state
            get isAuthenticated() {
                return authStore.isAuthenticated()
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
