// Authentication Service for Frontend
import { getConfig } from '@/config/build.config'

export interface User {
    id: string
    username: string
    email: string
    role: 'admin' | 'user'
    is_active: boolean
    created_at: string
    updated_at: string
    last_login?: string
}

export interface LoginRequest {
    username: string
    password: string
}

export interface LoginResponse {
    token: string
    user: User
    expires_at: string
}

export interface RegisterRequest {
    username: string
    email: string
    password: string
    role?: 'admin' | 'user'
}

export interface ChangePasswordRequest {
    old_password: string
    new_password: string
}

export interface ApiResponse<T> {
    success: boolean
    data?: T
    error?: string
}

class AuthService {
    private baseUrl: string
    private config = getConfig()
    private token: string | null = null
    private user: User | null = null

    constructor() {
        this.baseUrl = this.config.apiBaseUrl
        this.loadFromStorage()
    }

    // Load auth data from localStorage
    private loadFromStorage(): void {
        if (typeof window !== 'undefined' && window.localStorage) {
            const token = localStorage.getItem('auth_token')
            const user = localStorage.getItem('auth_user')
            
            if (token && user) {
                try {
                    this.token = token
                    this.user = JSON.parse(user)
                } catch (error) {
                    console.warn('Failed to parse stored user data:', error)
                    this.clearStorage()
                }
            }
        }
    }

    // Save auth data to localStorage
    private saveToStorage(token: string, user: User): void {
        if (typeof window !== 'undefined' && window.localStorage) {
            localStorage.setItem('auth_token', token)
            localStorage.setItem('auth_user', JSON.stringify(user))
        }
    }

    // Clear auth data from localStorage
    private clearStorage(): void {
        if (typeof window !== 'undefined' && window.localStorage) {
            localStorage.removeItem('auth_token')
            localStorage.removeItem('auth_user')
        }
    }

    // HTTP request with auth header
    private async request<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<ApiResponse<T>> {
        try {
            const url = `${this.baseUrl}${endpoint}`
            const headers: Record<string, string> = {
                'Content-Type': 'application/json',
                ...options.headers as Record<string, string>
            }

            // Add auth header if token is available
            if (this.token) {
                headers['Authorization'] = `Bearer ${this.token}`
            }

            const response = await fetch(url, {
                ...options,
                headers
            })

            if (!response.ok) {
                // Handle 401 Unauthorized
                if (response.status === 401) {
                    this.logout()
                }

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
                data: data.data || data
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
    public async login(credentials: LoginRequest): Promise<LoginResponse | null> {
        const response = await this.request<LoginResponse>('/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        })

        if (response.success && response.data) {
            const { token, user } = response.data
            this.token = token
            this.user = user
            this.saveToStorage(token, user)
            return response.data
        }

        throw new Error(response.error || 'Login failed')
    }

    // Register new user
    public async register(userData: RegisterRequest): Promise<User | null> {
        const response = await this.request<User>('/api/v1/auth/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        })

        if (response.success && response.data) {
            return response.data
        }

        throw new Error(response.error || 'Registration failed')
    }

    // Logout user
    public async logout(): Promise<void> {
        try {
            // Call logout endpoint if token exists
            if (this.token) {
                await this.request('/api/v1/auth/logout', {
                    method: 'POST'
                })
            }
        } catch (error) {
            console.warn('Logout API call failed:', error)
        } finally {
            // Always clear local state
            this.token = null
            this.user = null
            this.clearStorage()
        }
    }

    // Get current user from API
    public async getCurrentUser(): Promise<User | null> {
        if (!this.token) {
            return null
        }

        const response = await this.request<User>('/api/v1/auth/me')
        
        if (response.success && response.data) {
            this.user = response.data
            this.saveToStorage(this.token, response.data)
            return response.data
        }

        // If request fails, clear auth state
        this.logout()
        return null
    }

    // Change password
    public async changePassword(passwordData: ChangePasswordRequest): Promise<boolean> {
        const response = await this.request('/api/v1/auth/change-password', {
            method: 'POST',
            body: JSON.stringify(passwordData)
        })

        if (response.success) {
            return true
        }

        throw new Error(response.error || 'Password change failed')
    }

    // Check if user is authenticated
    public isAuthenticated(): boolean {
        return this.token !== null && this.user !== null
    }

    // Get current user
    public getUser(): User | null {
        return this.user
    }

    // Get current token
    public getToken(): string | null {
        return this.token
    }

    // Check if user has admin role
    public isAdmin(): boolean {
        return this.user?.role === 'admin'
    }

    // Check if user has specific role
    public hasRole(role: string): boolean {
        return this.user?.role === role
    }

    // Refresh auth state (useful for checking if token is still valid)
    public async refreshAuth(): Promise<boolean> {
        if (!this.token) {
            return false
        }

        try {
            const user = await this.getCurrentUser()
            return user !== null
        } catch (error) {
            console.warn('Auth refresh failed:', error)
            return false
        }
    }
}

// Export singleton instance
export const authService = new AuthService() 
