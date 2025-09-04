// Component Loader for Dynamic HTML Components
export interface ComponentCache {
    html: Map<string, string>
    metadata: Map<string, ComponentMetadata>
}

export interface ComponentMetadata {
    path: string
    loadedAt: number
    dependencies?: string[]
    size: number
}

export class ComponentLoader {
    private static instance: ComponentLoader
    private cache: ComponentCache = {
        html: new Map(),
        metadata: new Map()
    }
    private isLoaded = false
    private readonly MAX_CACHE_SIZE = 50
    private readonly CACHE_TTL = 5 * 60 * 1000 // 5 minutes

    private constructor() {}

    public static getInstance(): ComponentLoader {
        if (!ComponentLoader.instance) {
            ComponentLoader.instance = new ComponentLoader()
        }
        return ComponentLoader.instance
    }

    // Load component from URL or cache
    public async loadComponent(name: string, path?: string): Promise<string> {
        if (this.cache.html.has(name)) {
            return this.cache.html.get(name)!
        }

        if (path) {
            try {
                // console.log(`Fetching component ${name} from: ${path}`)
                const response = await fetch(path)
                // console.log(`📡 Response for ${name}: ${response.status} ${response.statusText}`)
                
                if (response.ok) {
                    const html = await response.text()
                    // console.log(`✅ Component ${name} loaded: ${html.length} characters`)
                    this.cacheComponent(name, html)
                    
                    // Dispatch component loaded event
                    document.dispatchEvent(new CustomEvent('componentLoaded', {
                        detail: { name, path }
                    }))
                    
                    return html
                } else {
                    console.warn(`❌ Failed to fetch ${name}: ${response.status} ${response.statusText}`)
                }
            } catch (error) {
                console.warn(`❌ Network error loading component ${name}:`, error)
            }
        }

        // Return fallback placeholder
        const placeholder = `<!-- Component ${name} not found -->`
        this.cache.html.set(name, placeholder)
        return placeholder
    }

    // Initialize all dashboard components
    public async initializeDashboard(): Promise<void> {
        if (this.isLoaded) return

        try {
            // Define component mapping
            const componentMap = {
                'navigation': '/components/layout/navigation.html',
                'overview': '/components/tabs/overview.html',
                'agents': '/components/tabs/agents.html',
                'jobs': '/components/tabs/jobs.html',
                'files': '/components/tabs/files.html',
                'wordlists': '/components/tabs/wordlists.html',
                'docs': '/components/tabs/docs.html',
                'modals': '/components/modals/all-modals.html',
                'distributed-job-modal': '/components/modals/distributed-job-modal.html',
                'notifications': '/components/ui/notifications.html',
                'loading': '/components/ui/loading.html'
            }

            // Load all components in parallel
            const loadPromises = Object.entries(componentMap).map(async ([name, path]) => {
                const component = await this.loadComponent(name, path)
                return { name, component }
            })

            const results = await Promise.allSettled(loadPromises)
            
            results.forEach((result, index) => {
                if (result.status === 'fulfilled') {
                    const { name, component } = result.value
                    this.cache.html.set(name, component)
                } else {
                    console.warn(`Failed to load component at index ${index}:`, result.reason)
                }
            })

            this.isLoaded = true
            // console.log('✅ Dashboard components loaded successfully')
        } catch (error) {
            console.error('Failed to initialize dashboard components:', error)
        }
    }

    // Render complete dashboard
    public async renderDashboard(): Promise<string> {
        await this.initializeDashboard()

        const navigation = this.cache.html.get('navigation') || ''
        const overview = this.cache.html.get('overview') || ''
        const agents = this.cache.html.get('agents') || ''
        const jobs = this.cache.html.get('jobs') || ''
        const files = this.cache.html.get('files') || ''
        const wordlists = this.cache.html.get('wordlists') || ''
        const docs = this.cache.html.get('docs') || ''
        const modals = this.cache.html.get('modals') || ''
        const notifications = this.cache.html.get('notifications') || ''
        const loading = this.cache.html.get('loading') || ''

        return `
            ${navigation}
            
            <main class="container-modern">
                ${overview}
                ${agents}
                ${jobs}
                ${files}
                ${wordlists}
                ${docs}
            </main>

            ${modals}
            ${notifications}
            ${loading}
        `
    }

    // Get individual component
    public getComponent(name: string): string {
        return this.cache.html.get(name) || `<!-- Component ${name} not found -->`
    }

    // Force reload component
    public async reloadComponent(name: string, path: string): Promise<string> {
        this.cache.html.delete(name)
        return await this.loadComponent(name, path)
    }

    // Check if component is loaded
    public isComponentLoaded(name: string): boolean {
        return this.cache.html.has(name)
    }

    // Get all loaded component names
    public getLoadedComponents(): string[] {
        return Array.from(this.cache.html.keys())
    }

    // Clear all components (useful for testing)
    public clearComponents(): void {
        this.cache.html.clear()
        this.isLoaded = false
    }

    private cacheComponent(path: string, html: string): void {
        // Implement LRU cache eviction if needed
        if (this.cache.html.size >= this.MAX_CACHE_SIZE) {
            this.evictOldestCache()
        }

        this.cache.html.set(path, html)
        this.cache.metadata.set(path, {
            path,
            loadedAt: Date.now(),
            size: html.length
        })
    }

    private isCacheValid(metadata: ComponentMetadata): boolean {
        return Date.now() - metadata.loadedAt < this.CACHE_TTL
    }

    private evictOldestCache(): void {
        let oldestPath = ''
        let oldestTime = Date.now()

        this.cache.metadata.forEach((metadata, path) => {
            if (metadata.loadedAt < oldestTime) {
                oldestTime = metadata.loadedAt
                oldestPath = path
            }
        })

        if (oldestPath) {
            this.cache.html.delete(oldestPath)
            this.cache.metadata.delete(oldestPath)
        }
    }

    /**
     * Clear cache
     */
    clearCache(): void {
        this.cache.html.clear()
        this.cache.metadata.clear()
    }

    /**
     * Preload multiple components for better performance
     */
    async preloadComponents(componentPaths: string[]): Promise<void> {
        const loadPromises = componentPaths.map(path => this.loadComponent(path))
        await Promise.all(loadPromises)
    }

    /**
     * Get cache statistics
     */
    getCacheStats(): {
        size: number
        totalMemory: number
        components: string[]
    } {
        const totalMemory = Array.from(this.cache.metadata.values())
            .reduce((total, metadata) => total + metadata.size, 0)

        return {
            size: this.cache.html.size,
            totalMemory,
            components: Array.from(this.cache.html.keys())
        }
    }
}

// Export singleton instance
export const componentLoader = ComponentLoader.getInstance()

// Utility functions for component management
export const componentUtils = {
    // Inject component into specific element
    async injectComponent(elementId: string, componentName: string, componentPath?: string): Promise<boolean> {
        try {
            const element = document.getElementById(elementId)
            if (!element) {
                console.warn(`Element with id '${elementId}' not found`)
                return false
            }

            const html = await componentLoader.loadComponent(componentName, componentPath)
            element.innerHTML = html
            
            // Reinitialize Alpine.js if available
            if ((window as any).Alpine && typeof (window as any).Alpine.initTree === 'function') {
                (window as any).Alpine.initTree(element)
            }
            
            return true
        } catch (error) {
            console.error(`Failed to inject component ${componentName}:`, error)
            return false
        }
    },

    // Replace component content
    async replaceComponent(oldComponentName: string, newComponentName: string, newComponentPath?: string): Promise<boolean> {
        try {
            const newHtml = await componentLoader.loadComponent(newComponentName, newComponentPath)
            
            // Find elements containing the old component
            const elements = document.querySelectorAll(`[data-component="${oldComponentName}"]`)
            
            elements.forEach(element => {
                element.innerHTML = newHtml
                element.setAttribute('data-component', newComponentName)
            })
            
            return elements.length > 0
        } catch (error) {
            console.error(`Failed to replace component ${oldComponentName}:`, error)
            return false
        }
    }
}

// Global types for better TypeScript support
declare global {
    interface Window {
        componentLoader: ComponentLoader
        componentUtils: typeof componentUtils
    }
}

// Make available globally for debugging
if (typeof window !== 'undefined') {
    window.componentLoader = componentLoader
    window.componentUtils = componentUtils
} 
