// Client-side Router for Dashboard
import { authStore } from '@/stores/auth.store'

type RouteKey = '' | 'login' | 'overview' | 'agents' | 'agent-keys' | 'jobs' | 'files' | 'wordlists' | 'docs'
type RouteValue = 'login' | 'overview' | 'agents' | 'agent-keys' | 'jobs' | 'files' | 'wordlists' | 'docs'

export class Router {
    private static instance: Router
    private currentRoute: string = ''
    private listeners: Set<(route: string) => void> = new Set()

    // Route definitions
    private routes: Record<RouteKey, RouteValue> = {
        '': 'overview',
        'login': 'login',
        'overview': 'overview', 
        'agents': 'agents',
        'agent-keys': 'agent-keys',
        'jobs': 'jobs',
        'files': 'files',
        'wordlists': 'wordlists',
        'docs': 'docs'
    }

    // Protected routes (require authentication)
    private protectedRoutes: Set<RouteValue> = new Set([
        'overview', 'agents', 'agent-keys', 'jobs', 'files', 'wordlists', 'docs'
    ])

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
        
        // Check if route requires authentication
        if (this.protectedRoutes.has(route) && !authStore.isAuthenticated()) {
            // Redirect to login if not authenticated
            this.currentRoute = 'login'
            this.notifyListeners('login')
            return
        }
        
        // Redirect to overview if trying to access login while authenticated
        if (route === 'login' && authStore.isAuthenticated()) {
            this.currentRoute = 'overview'
            this.notifyListeners('overview')
            return
        }
        
        if (route !== this.currentRoute) {
            this.currentRoute = route
            this.notifyListeners(route)
        }
    }

    public navigate(route: string): void {
        const validRoute = this.routes[route as RouteKey] || 'overview'
        
        // Check if route requires authentication
        if (this.protectedRoutes.has(validRoute) && !authStore.isAuthenticated()) {
            // Redirect to login if not authenticated
            this.navigate('login')
            return
        }
        
        // Redirect to overview if trying to access login while authenticated
        if (validRoute === 'login' && authStore.isAuthenticated()) {
            this.navigate('overview')
            return
        }
        
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

    // Check if route is protected
    public isProtectedRoute(route: string): boolean {
        return this.protectedRoutes.has(route as RouteValue)
    }

    // Force refresh current route (useful after login/logout)
    public refresh(): void {
        this.handleRouteChange()
    }
}

// Export singleton instance
export const router = Router.getInstance() 
