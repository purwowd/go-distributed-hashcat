// Client-side Router for Dashboard
type RouteKey = '' | 'overview' | 'agents' | 'jobs' | 'files' | 'wordlists' | 'agent-keys' | 'docs'
type RouteValue = 'overview' | 'agents' | 'jobs' | 'files' | 'wordlists' | 'agent-keys' | 'docs'

export class Router {
    private static instance: Router
    private currentRoute: string = ''
    private listeners: Set<(route: string) => void> = new Set()

    // Route definitions
    private routes: Record<RouteKey, RouteValue> = {
        '': 'overview',
        'overview': 'overview', 
        'agents': 'agents',
        'jobs': 'jobs',
        'files': 'files',
        'wordlists': 'wordlists',
        'agent-keys': 'agent-keys',
        'docs': 'docs'
    }

    private constructor() {
        this.init()
    }

    public static getInstance(): Router {
        if (!Router.instance) {
            Router.instance = new Router()
        }
        return Router.instance
    }

    private init(): void {
        // Handle browser back/forward navigation
        window.addEventListener('popstate', () => {
            this.handleRouteChange()
        })

        // Handle initial route
        this.handleRouteChange()
    }

    private handleRouteChange(): void {
        const hash = window.location.hash.slice(1) // Remove #
        const route = this.routes[hash as RouteKey] || 'overview'
        
        if (route !== this.currentRoute) {
            this.currentRoute = route
            this.notifyListeners(route)
        }
    }

    public navigate(route: string): void {
        const validRoute = this.routes[route as RouteKey] || 'overview'
        const hash = route === 'overview' ? '' : route
        
        // Update URL without page reload
        const newUrl = hash ? `#${hash}` : window.location.pathname
        if (window.location.hash !== `#${hash}`) {
            window.history.pushState({ route }, '', newUrl)
        }
        
        this.currentRoute = validRoute
        this.notifyListeners(validRoute)
    }

    public getCurrentRoute(): string {
        return this.currentRoute
    }

    public subscribe(listener: (route: string) => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    private notifyListeners(route: string): void {
        this.listeners.forEach(listener => listener(route))
    }

    public getRouteUrl(route: string): string {
        const hash = route === 'overview' ? '' : route
        return hash ? `#${hash}` : '/'
    }

    public isCurrentRoute(route: string): boolean {
        return this.currentRoute === route
    }
}

// Export singleton instance
export const router = Router.getInstance() 
