// Build Configuration for Production-Ready System
export interface ComponentConfig {
    name: string
    path: string
    lazy?: boolean
    preload?: boolean
    dependencies?: string[]
}

export interface BuildConfig {
    mode: 'development' | 'production'
    apiBaseUrl: string
    components: ComponentConfig[]
    features: {
        hotReload: boolean
        lazyLoading: boolean
        componentCaching: boolean
        performanceMonitoring: boolean
    }
    optimization: {
        bundleSplitting: boolean
        treeshaking: boolean
        minification: boolean
        compression: boolean
    }
}

// Environment-specific configurations
export const configs: Record<string, BuildConfig> = {
    development: {
        mode: 'development',
        apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://30.30.30.39:1337',
        components: [
            { name: 'navigation', path: '/components/layout/navigation.html', preload: true },
            { name: 'overview', path: '/components/tabs/overview.html', preload: true },
            { name: 'agents', path: '/components/tabs/agents.html', lazy: true },
            { name: 'jobs', path: '/components/tabs/jobs.html', lazy: true },
            { name: 'files', path: '/components/tabs/files.html', lazy: true },
            { name: 'wordlists', path: '/components/tabs/wordlists.html', lazy: true },
            { name: 'docs', path: '/components/tabs/docs.html', lazy: true },
            { name: 'modals', path: '/components/modals/all-modals.html', lazy: true },
            { name: 'notifications', path: '/components/ui/notifications.html', preload: true },
            { name: 'loading', path: '/components/ui/loading.html', preload: true }
        ],
        features: {
            hotReload: import.meta.env.VITE_ENABLE_HOT_RELOAD === 'true',
            lazyLoading: import.meta.env.VITE_ENABLE_LAZY_LOADING === 'true',
            componentCaching: import.meta.env.VITE_ENABLE_COMPONENT_CACHING === 'true',
            performanceMonitoring: import.meta.env.VITE_ENABLE_PERFORMANCE_MONITORING === 'true'
        },
        optimization: {
            bundleSplitting: import.meta.env.VITE_ENABLE_BUNDLE_SPLITTING === 'true',
            treeshaking: import.meta.env.VITE_ENABLE_TREESHAKING === 'true',
            minification: import.meta.env.VITE_ENABLE_MINIFICATION === 'true',
            compression: import.meta.env.VITE_ENABLE_COMPRESSION === 'true'
        }
    },
    production: {
        mode: 'production',
        apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://30.30.30.39:1337',
        components: [
            { name: 'navigation', path: '/components/layout/navigation.html', preload: true },
            { name: 'overview', path: '/components/tabs/overview.html', preload: true },
            { name: 'agents', path: '/components/tabs/agents.html', lazy: true },
            { name: 'jobs', path: '/components/tabs/jobs.html', lazy: true },
            { name: 'files', path: '/components/tabs/files.html', lazy: true },
            { name: 'wordlists', path: '/components/tabs/wordlists.html', lazy: true },
            { name: 'docs', path: '/components/tabs/docs.html', lazy: true },
            { name: 'modals', path: '/components/modals/all-modals.html', lazy: true },
            { name: 'notifications', path: '/components/ui/notifications.html', preload: true },
            { name: 'loading', path: '/components/ui/loading.html', preload: true }
        ],
        features: {
            hotReload: import.meta.env.VITE_ENABLE_HOT_RELOAD === 'true',
            lazyLoading: import.meta.env.VITE_ENABLE_LAZY_LOADING === 'true',
            componentCaching: import.meta.env.VITE_ENABLE_COMPONENT_CACHING === 'true',
            performanceMonitoring: import.meta.env.VITE_ENABLE_PERFORMANCE_MONITORING === 'true'
        },
        optimization: {
            bundleSplitting: import.meta.env.VITE_ENABLE_BUNDLE_SPLITTING === 'true',
            treeshaking: import.meta.env.VITE_ENABLE_TREESHAKING === 'true',
            minification: import.meta.env.VITE_ENABLE_MINIFICATION === 'true',
            compression: import.meta.env.VITE_ENABLE_COMPRESSION === 'true'
        }
    }
}

// Get current configuration
export function getConfig(): BuildConfig {
    const mode = import.meta.env.MODE || 'development'
    return configs[mode] || configs.development
}

// Component registration system
export class ComponentRegistry {
    private static instance: ComponentRegistry
    private config: BuildConfig
    private registeredComponents: Map<string, ComponentConfig> = new Map()

    private constructor() {
        this.config = getConfig()
        this.registerComponents()
    }

    public static getInstance(): ComponentRegistry {
        if (!ComponentRegistry.instance) {
            ComponentRegistry.instance = new ComponentRegistry()
        }
        return ComponentRegistry.instance
    }

    private registerComponents(): void {
        this.config.components.forEach(component => {
            this.registeredComponents.set(component.name, component)
        })
        // console.log(`📦 Registered ${this.registeredComponents.size} components for ${this.config.mode} mode`)
    }

    public getComponent(name: string): ComponentConfig | undefined {
        return this.registeredComponents.get(name)
    }

    public getAllComponents(): ComponentConfig[] {
        return Array.from(this.registeredComponents.values())
    }

    public getPreloadComponents(): ComponentConfig[] {
        return this.getAllComponents().filter(c => c.preload)
    }

    public getLazyComponents(): ComponentConfig[] {
        return this.getAllComponents().filter(c => c.lazy)
    }

    public getConfig(): BuildConfig {
        return this.config
    }

    // Dynamic component registration
    public registerComponent(component: ComponentConfig): void {
        this.registeredComponents.set(component.name, component)
        // console.log(`🔧 Dynamically registered component: ${component.name}`)
    }

    // Unregister component
    public unregisterComponent(name: string): boolean {
        return this.registeredComponents.delete(name)
    }
}

// Export singleton
export const componentRegistry = ComponentRegistry.getInstance()

// Performance monitoring utilities
export class PerformanceMonitor {
    private static metrics: Map<string, number> = new Map()

    public static startTimer(name: string): void {
        this.metrics.set(name, performance.now())
    }

    public static endTimer(name: string): number {
        const start = this.metrics.get(name)
        if (!start) return 0
        
        const duration = performance.now() - start
        this.metrics.delete(name)
        
        if (getConfig().features.performanceMonitoring) {
            // console.log(`⏱️ ${name}: ${duration.toFixed(2)}ms`)
        }
        
        return duration
    }

    public static measure<T>(name: string, fn: () => T | Promise<T>): T | Promise<T> {
        this.startTimer(name)
        const result = fn()
        
        if (result instanceof Promise) {
            return result.finally(() => this.endTimer(name))
        } else {
            this.endTimer(name)
            return result
        }
    }
}

// Export utilities
export { PerformanceMonitor as perf } 
