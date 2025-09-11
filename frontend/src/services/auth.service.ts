// Authentication Service for Distributed Hashcat Dashboard
import { getConfig } from '@/config/build.config'
import { LoginRequest, LoginResponse, LogoutRequest, User } from '../types/index'

class AuthService {
    private baseUrl: string
    private config = getConfig()
    private tokenKey = 'hashcat_auth_token'
    private userKey = 'hashcat_auth_user'

    constructor() {
        this.baseUrl = this.config.apiBaseUrl
    }

    // Generic HTTP request method with auth support
    private async request<T>(
        endpoint: string, 
        options: RequestInit = {}
    ): Promise<{ success: boolean; data?: T; error?: string }> {
        try {
            const url = `${this.baseUrl}${endpoint}`
            const token = this.getToken()
            
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...(token && { 'Authorization': `Bearer ${token}` }),
                    ...options.headers
                },
                ...options
            })

            if (!response.ok) {
                let errorMessage = `HTTP ${response.status}: ${response.statusText}`
                try {
                    const errorData = await response.json()
                    if (errorData.error) {
                        errorMessage = errorData.error
                    }
                } catch (e) {
                    // If response isn't JSON, use default message
                }
                throw new Error(errorMessage)
            }

            const data = await response.json()
            return {
                success: true,
                data
            }
        } catch (error) {
            console.error(`Auth API Request failed for ${endpoint}:`, error)
            return {
                success: false,
                error: error instanceof Error ? error.message : 'Unknown error'
            }
        }
    }

    // Login user
    public async login(credentials: LoginRequest): Promise<{ success: boolean; data?: LoginResponse; error?: string }> {
        const response = await this.request<LoginResponse>('/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        })

        if (response.success && response.data) {
            // Store token and user data
            this.setToken(response.data.token)
            this.setUser(response.data.user)
        }

        return response
    }

    // Logout user
    public async logout(): Promise<{ success: boolean; error?: string }> {
        const token = this.getToken()
        if (!token) {
            return { success: true }
        }

        const response = await this.request('/api/v1/auth/logout', {
            method: 'POST',
            body: JSON.stringify({ token })
        })

        // Clear stored data regardless of API response
        this.clearAuth()

        return response
    }

    // Validate token
    public async validateToken(): Promise<{ success: boolean; data?: any; error?: string }> {
        const token = this.getToken()
        if (!token) {
            return { success: false, error: 'No token found' }
        }

        return await this.request('/api/v1/auth/validate', {
            method: 'POST',
            body: JSON.stringify({ token })
        })
    }

    // Refresh token
    public async refreshToken(): Promise<{ success: boolean; data?: LoginResponse; error?: string }> {
        const token = this.getToken()
        if (!token) {
            return { success: false, error: 'No token found' }
        }

        const response = await this.request<LoginResponse>('/api/v1/auth/refresh', {
            method: 'POST',
            body: JSON.stringify({ token })
        })

        if (response.success && response.data) {
            // Update stored token and user data
            this.setToken(response.data.token)
            this.setUser(response.data.user)
        }

        return response
    }

    // Token management
    public getToken(): string | null {
        return localStorage.getItem(this.tokenKey)
    }

    public setToken(token: string): void {
        localStorage.setItem(this.tokenKey, token)
    }

    public clearToken(): void {
        localStorage.removeItem(this.tokenKey)
    }

    // User management
    public getUser(): User | null {
        const userStr = localStorage.getItem(this.userKey)
        if (!userStr) return null
        
        try {
            return JSON.parse(userStr)
        } catch {
            return null
        }
    }

    public setUser(user: User): void {
        localStorage.setItem(this.userKey, JSON.stringify(user))
    }

    public clearUser(): void {
        localStorage.removeItem(this.userKey)
    }

    // Clear all auth data
    public clearAuth(): void {
        this.clearToken()
        this.clearUser()
    }

    // Check if user is authenticated
    public isAuthenticated(): boolean {
        const token = this.getToken()
        const user = this.getUser()
        return !!(token && user)
    }

    // Check if token is expired (basic check)
    public isTokenExpired(): boolean {
        const token = this.getToken()
        if (!token) return true

        try {
            // Decode JWT payload (basic implementation)
            const payload = JSON.parse(atob(token.split('.')[1]))
            const now = Math.floor(Date.now() / 1000)
            return payload.exp < now
        } catch {
            return true
        }
    }

    // Get authorization header for API requests
    public getAuthHeader(): { Authorization: string } | {} {
        const token = this.getToken()
        return token ? { Authorization: `Bearer ${token}` } : {}
    }

    // Check if user has specific role
    public hasRole(role: string): boolean {
        const user = this.getUser()
        return user?.role === role
    }

    // Check if user is admin
    public isAdmin(): boolean {
        return this.hasRole('admin')
    }

    // Check if username exists
    public async checkUsernameExists(username: string): Promise<{ success: boolean; exists?: boolean; error?: string }> {
        const response = await this.request<{ exists: boolean }>('/api/v1/auth/check-username', {
            method: 'POST',
            body: JSON.stringify({ username })
        })

        if (response.success && response.data) {
            return {
                success: true,
                exists: response.data.exists
            }
        }

        return {
            success: false,
            error: response.error || 'Failed to check username'
        }
    }
}

// Export singleton instance
export const authService = new AuthService()
