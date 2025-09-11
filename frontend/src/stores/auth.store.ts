// Authentication Store for Distributed Hashcat Dashboard
import { authService } from '@/services/auth.service'
import { LoginRequest, User, AuthState } from '../types/index'

class AuthStore {
    private state: AuthState = {
        isAuthenticated: false,
        user: null,
        token: null,
        isLoading: false,
        error: null
    }
    private listeners: Set<(state: AuthState) => void> = new Set()

    constructor() {
        this.initializeAuth()
    }

    // Initialize authentication state from localStorage
    private initializeAuth(): void {
        const token = authService.getToken()
        const user = authService.getUser()
        
        if (token && user && !authService.isTokenExpired()) {
            this.state = {
                isAuthenticated: true,
                user,
                token,
                isLoading: false,
                error: null
            }
        } else {
            // Clear invalid auth data
            authService.clearAuth()
            this.state = {
                isAuthenticated: false,
                user: null,
                token: null,
                isLoading: false,
                error: null
            }
        }
        
        this.notifyListeners()
    }

    // Login user
    public async login(credentials: LoginRequest): Promise<boolean> {
        this.setState({ isLoading: true, error: null })
        
        const response = await authService.login(credentials)
        
        if (response.success && response.data) {
            this.setState({
                isAuthenticated: true,
                user: response.data.user,
                token: response.data.token,
                isLoading: false,
                error: null
            })
            return true
        } else {
            this.setState({
                isAuthenticated: false,
                user: null,
                token: null,
                isLoading: false,
                error: response.error || 'Login failed'
            })
            return false
        }
    }

    // Logout user
    public async logout(): Promise<void> {
        this.setState({ isLoading: true })
        
        await authService.logout()
        
        this.setState({
            isAuthenticated: false,
            user: null,
            token: null,
            isLoading: false,
            error: null
        })
    }

    // Validate current token
    public async validateToken(): Promise<boolean> {
        if (!this.state.token) return false
        
        const response = await authService.validateToken()
        
        if (response.success) {
            return true
        } else {
            // Token is invalid, logout user
            await this.logout()
            return false
        }
    }

    // Refresh token
    public async refreshToken(): Promise<boolean> {
        if (!this.state.token) return false
        
        const response = await authService.refreshToken()
        
        if (response.success && response.data) {
            this.setState({
                user: response.data.user,
                token: response.data.token,
                error: null
            })
            return true
        } else {
            // Refresh failed, logout user
            await this.logout()
            return false
        }
    }

    // Clear error
    public clearError(): void {
        this.setState({ error: null })
    }

    // Get current state
    public getState(): AuthState {
        return { ...this.state }
    }

    // Check if user is authenticated
    public isAuthenticated(): boolean {
        return this.state.isAuthenticated
    }

    // Get current user
    public getUser(): User | null {
        return this.state.user
    }

    // Get current token
    public getToken(): string | null {
        return this.state.token
    }

    // Check if user has specific role
    public hasRole(role: string): boolean {
        return this.state.user?.role === role
    }

    // Check if user is admin
    public isAdmin(): boolean {
        return this.hasRole('admin')
    }

    // Check if loading
    public isLoading(): boolean {
        return this.state.isLoading
    }

    // Get error
    public getError(): string | null {
        return this.state.error
    }

    // Subscribe to state changes
    public subscribe(listener: (state: AuthState) => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    // Private method to update state
    private setState(updates: Partial<AuthState>): void {
        this.state = { ...this.state, ...updates }
        this.notifyListeners()
    }

    // Notify all listeners
    private notifyListeners(): void {
        this.listeners.forEach(listener => listener(this.getState()))
    }

    // Auto-refresh token before expiration
    public startTokenRefresh(): void {
        const refreshInterval = 5 * 60 * 1000 // 5 minutes
        
        setInterval(async () => {
            if (this.isAuthenticated() && authService.isTokenExpired()) {
                await this.refreshToken()
            }
        }, refreshInterval)
    }
}

// Export singleton instance
export const authStore = new AuthStore()
