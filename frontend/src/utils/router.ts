// Client-side Router for Dashboard
import { authStore } from '@/stores/auth.store'

type RouteKey = '' | 'login' | 'overview' | 'agents' | 'agent-keys' | 'jobs' | 'files' | 'wordlists' | 'docs'
type RouteValue = 'login' | 'overview' | 'agents' | 'agent-keys' | 'jobs' | 'files' | 'wordlists' | 'docs'

export class Router {
    private static instance: Router
    private currentRoute: string = ''
    private listeners: Set<(route: string) => void> = new Set()
    private isLoggingOut: boolean = false

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
        // Get route from pathname instead of hash
        const path = window.location.pathname.slice(1) // Remove leading /
        const route = this.routes[path as RouteKey] || 'overview'
        
        // Always check authentication state first
        const isAuthenticated = authStore.isAuthenticated()
        
        // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY route except login
        if (!isAuthenticated && route !== 'login') {
            console.log('🔒 STRICT AUTH: Unauthenticated access to', route, '- redirecting to login')
            this.currentRoute = 'login'
            window.location.replace('/login')
            this.notifyListeners('login')
            return
        }
        
        // Check if route requires authentication (additional check)
        if (this.protectedRoutes.has(route) && !isAuthenticated) {
            // Redirect to login if not authenticated
            this.currentRoute = 'login'
            window.location.replace('/login')
            this.notifyListeners('login')
            return
        }
        
        // Redirect to overview if trying to access login while authenticated
        // But allow login page if user is logging out
        if (route === 'login' && isAuthenticated && !this.isLoggingOut) {
            this.currentRoute = 'overview'
            window.location.replace('/')
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
        
        // STRICT AUTHENTICATION: If not authenticated, redirect to login for ANY route except login
        if (!authStore.isAuthenticated() && validRoute !== 'login') {
            console.log('🔒 STRICT AUTH: Unauthenticated navigation to', validRoute, '- redirecting to login')
            this.currentRoute = 'login'
            window.location.replace('/login')
            this.notifyListeners('login')
            return
        }
        
        // Check if route requires authentication (additional check)
        if (this.protectedRoutes.has(validRoute) && !authStore.isAuthenticated()) {
            // Redirect to login if not authenticated
            this.currentRoute = 'login'
            window.location.replace('/login')
            this.notifyListeners('login')
            return
        }
        
        // Redirect to overview if trying to access login while authenticated
        if (validRoute === 'login' && authStore.isAuthenticated()) {
            this.currentRoute = 'overview'
            window.location.replace('/')
            this.notifyListeners('overview')
            return
        }
        
        // Update URL with clean path-based routing
        const newPath = validRoute === 'overview' ? '/' : `/${validRoute}`
        if (window.location.pathname !== newPath) {
            window.history.pushState({ route: validRoute }, '', newPath)
        }
        
        this.currentRoute = validRoute
        this.notifyListeners(validRoute)
        
        // Check for login success notification when navigating to overview
        if (validRoute === 'overview') {
            // Use setTimeout to ensure the route change is processed first
            setTimeout(() => {
                console.log('🔍 ===== ROUTER: CHECKING LOGIN SUCCESS NOTIFICATION =====')
                console.log('🔍 Router: Checking for login success notification after navigation to overview')
                console.log('🔍 Current pathname:', window.location.pathname)
                console.log('🔍 SessionStorage showLoginSuccess:', sessionStorage.getItem('showLoginSuccess'))
                
                // Check sessionStorage first
                const showLoginSuccess = sessionStorage.getItem('showLoginSuccess')
                if (showLoginSuccess === 'true') {
                    console.log('🔍 Router: Found login success flag, triggering notification...')
                    
                    // Try multiple ways to access Alpine component
                    let dashboardApp = null
                    
                    // Method 1: Direct Alpine data access
                    if (window.Alpine && window.Alpine.data) {
                        dashboardApp = window.Alpine.data('dashboardApp')
                        console.log('🔍 Router: Method 1 - Alpine data:', !!dashboardApp)
                    }
                    
                    // Method 2: Try to find Alpine component in DOM
                    if (!dashboardApp) {
                        const alpineElement = document.querySelector('[x-data*="dashboardApp"]')
                        if (alpineElement && (alpineElement as any)._x_dataStack) {
                            dashboardApp = (alpineElement as any)._x_dataStack[0]
                            console.log('🔍 Router: Method 2 - DOM element:', !!dashboardApp)
                        }
                    }
                    
                    // Method 3: Try global window access
                    if (!dashboardApp) {
                        dashboardApp = (window as any).dashboardApp
                        console.log('🔍 Router: Method 3 - Global window:', !!dashboardApp)
                    }
                    
                    console.log('🔍 Router: Dashboard app available:', !!dashboardApp)
                    
                    if (dashboardApp) {
                        console.log('🔍 Router: Setting showLoginSuccessNotification to true')
                        dashboardApp.showLoginSuccessNotification = true
                        
                        // Trigger notification method
                        if (typeof dashboardApp.checkLoginSuccessNotification === 'function') {
                            console.log('🔍 Router: Calling checkLoginSuccessNotification')
                            dashboardApp.checkLoginSuccessNotification()
                        } else if (typeof dashboardApp.showLoginSuccessNotificationNow === 'function') {
                            console.log('🔍 Router: Calling showLoginSuccessNotificationNow')
                            dashboardApp.showLoginSuccessNotificationNow()
                        }
                    } else {
                        console.log('❌ Router: Could not access Alpine component, trying direct DOM manipulation')
                        // Fallback: Direct DOM manipulation
                        this.triggerNotificationDirectly()
                    }
                } else {
                    console.log('🔍 Router: No login success flag found')
                }
                
                console.log('🔍 ========================================================')
            }, 100)
        }
    }

    public getCurrentRoute(): string {
        return this.currentRoute
    }

    public subscribe(listener: (route: string) => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    // Direct DOM manipulation fallback for router
    private triggerNotificationDirectly() {
        console.log('🚀 ===== ROUTER: DIRECT DOM MANIPULATION FALLBACK =====')
        
        // Try to find and show notification elements
        const fixedNotification = document.querySelector('[x-show="showLoginSuccessNotification"]')
        const pageNotification = document.querySelector('[x-show="showLoginSuccessNotification"]')
        
        if (fixedNotification) {
            console.log('🚀 Router: Found fixed notification element, showing...')
            const element = fixedNotification as HTMLElement
            element.style.display = 'block'
            element.style.opacity = '1'
            element.style.transform = 'translateY(0)'
        }
        
        if (pageNotification) {
            console.log('🚀 Router: Found page notification element, showing...')
            const element = pageNotification as HTMLElement
            element.style.display = 'block'
            element.style.opacity = '1'
            element.style.transform = 'scale(1) translateY(0)'
        }
        
        // Also try to show toast notifications (only once per session)
        const welcomeShown = localStorage.getItem('welcomeNotificationsShown')
        if (!welcomeShown) {
            this.showDirectToastNotification('🎉 Welcome to Hashcat Dashboard!', 'success')
            setTimeout(() => {
                this.showDirectToastNotification('✅ Login successful! You are now logged in.', 'success')
            }, 2000)
            setTimeout(() => {
                this.showDirectToastNotification('🚀 Ready to start cracking passwords!', 'info')
            }, 2000)
            
            // Mark as shown to prevent showing again on reload
            localStorage.setItem('welcomeNotificationsShown', 'true')
        }
        
        console.log('🚀 ========================================================')
    }

    // Direct toast notification fallback for router
    private showDirectToastNotification(message: string, type: string) {
        console.log(`🔔 Router Direct toast: [${type.toUpperCase()}] ${message}`)
        
        // Check if we're on login page and block login success notifications
        const isLoginSuccess = message.includes('Login successful')
        const isLoginError = message.includes('Authentication Failed') || message.includes('Login failed')
        const currentPath = window.location.pathname
        
        if (currentPath === '/login' && !isLoginError) {
            console.log(`🔔 Router Direct toast blocked on login page: [${type.toUpperCase()}] ${message}`)
            return
        }
        
        // Check if login success notification has already been shown for this session
        if (isLoginSuccess) {
            const notificationShown = sessionStorage.getItem('loginSuccessNotificationShown')
            if (notificationShown === 'true') {
                console.log('🚫 Login success notification already shown in this session, skipping router direct toast...')
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

    private notifyListeners(route: string): void {
        this.listeners.forEach(listener => listener(route))
    }

    public getRouteUrl(route: string): string {
        return route === 'overview' ? '/' : `/${route}`
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

    // Set logout flag to allow navigation to login page
    public setLoggingOut(value: boolean): void {
        this.isLoggingOut = value
    }
}

// Export singleton instance
export const router = Router.getInstance() 
