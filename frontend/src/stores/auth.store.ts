// Authentication Store
import { authService, type User, type LoginRequest, type RegisterRequest } from '@/services/auth.service'

interface AuthState {
    user: User | null
    isAuthenticated: boolean
    isLoading: boolean
    error: string | null
}

class AuthStore {
    private listeners: Set<() => void> = new Set()
    private state: AuthState = {
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null
    }

    constructor() {
        this.initializeAuth()
    }

    // Initialize auth state from stored data
    private async initializeAuth(): Promise<void> {
        this.state.isLoading = true
        this.notifyListeners()

        try {
            if (authService.isAuthenticated()) {
                // Refresh user data from API to ensure token is still valid
                const user = await authService.getCurrentUser()
                if (user) {
                    this.state.user = user
                    this.state.isAuthenticated = true
                    this.state.error = null
                } else {
                    this.state.isAuthenticated = false
                    this.state.user = null
                }
            }
        } catch (error) {
            console.warn('Failed to initialize auth:', error)
            this.state.isAuthenticated = false
            this.state.user = null
            this.state.error = null
        } finally {
            this.state.isLoading = false
            this.notifyListeners()
            
            // Debug logging for auth state
            console.log('🔍 Auth initialization complete:', {
                isAuthenticated: this.state.isAuthenticated,
                user: this.state.user,
                listeners: this.listeners.size
            })
        }
    }

    // Login user
    public async login(credentials: LoginRequest): Promise<void> {
        this.state.isLoading = true
        this.state.error = null
        this.notifyListeners()

        try {
            const response = await authService.login(credentials)
            if (response) {
                this.state.user = response.user
                this.state.isAuthenticated = true
                this.state.error = null
                console.log('✅ Login successful, updating auth state:', {
                    user: response.user,
                    isAuthenticated: true
                })
            }
        } catch (error) {
            this.state.error = error instanceof Error ? error.message : 'Login failed'
            this.state.isAuthenticated = false
            this.state.user = null
            throw error
        } finally {
            this.state.isLoading = false
            this.notifyListeners()
            
            // Debug logging for login
            console.log('🔍 Login attempt complete:', {
                isAuthenticated: this.state.isAuthenticated,
                user: this.state.user,
                error: this.state.error,
                listeners: this.listeners.size
            })
        }
    }

    // Register new user
    public async register(userData: RegisterRequest): Promise<User | null> {
        this.state.isLoading = true
        this.state.error = null
        this.notifyListeners()

        try {
            const user = await authService.register(userData)
            this.state.error = null
            return user
        } catch (error) {
            this.state.error = error instanceof Error ? error.message : 'Registration failed'
            throw error
        } finally {
            this.state.isLoading = false
            this.notifyListeners()
        }
    }

    // Logout user
    public async logout(): Promise<void> {
        this.state.isLoading = true
        this.notifyListeners()

        try {
            await authService.logout()
        } catch (error) {
            console.warn('Logout error:', error)
        } finally {
            this.state.user = null
            this.state.isAuthenticated = false
            this.state.error = null
            this.state.isLoading = false
            this.notifyListeners()
        }
    }

    // Change password
    public async changePassword(oldPassword: string, newPassword: string): Promise<void> {
        this.state.isLoading = true
        this.state.error = null
        this.notifyListeners()

        try {
            await authService.changePassword({
                old_password: oldPassword,
                new_password: newPassword
            })
            this.state.error = null
        } catch (error) {
            this.state.error = error instanceof Error ? error.message : 'Password change failed'
            throw error
        } finally {
            this.state.isLoading = false
            this.notifyListeners()
        }
    }

    // Refresh current user data
    public async refreshUser(): Promise<void> {
        if (!this.state.isAuthenticated) return

        try {
            const user = await authService.getCurrentUser()
            if (user) {
                this.state.user = user
                this.notifyListeners()
            } else {
                // Token expired or user not found
                this.logout()
            }
        } catch (error) {
            console.warn('Failed to refresh user:', error)
            this.logout()
        }
    }

    // Clear error state
    public clearError(): void {
        this.state.error = null
        this.notifyListeners()
    }

    // Get current state
    public getState(): AuthState {
        return { ...this.state }
    }

    // Get user
    public getUser(): User | null {
        return this.state.user
    }

    // Check if authenticated
    public isAuthenticated(): boolean {
        return this.state.isAuthenticated
    }

    // Check if loading
    public isLoading(): boolean {
        return this.state.isLoading
    }

    // Get error
    public getError(): string | null {
        return this.state.error
    }

    // Check if user is admin
    public isAdmin(): boolean {
        return this.state.user?.role === 'admin' || false
    }

    // Check if user has specific role
    public hasRole(role: string): boolean {
        return this.state.user?.role === role || false
    }

    // Subscribe to state changes
    public subscribe(listener: () => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    // Notify all listeners
    private notifyListeners(): void {
        this.listeners.forEach(listener => {
            try {
                listener()
            } catch (error) {
                console.error('Auth store listener error:', error)
            }
        })
    }

    // Force update state (useful for debugging)
    public forceUpdate(): void {
        this.notifyListeners()
    }
}

// Export singleton instance
export const authStore = new AuthStore()

// For Alpine.js integration - create reactive data object
export function createAuthData() {
    // Create a reactive data object that will be updated by subscription
    const authData = {
        // State properties (will be updated by subscription)
        user: authStore.getUser(),
        currentUser: authStore.getUser(),
        isAuthenticated: authStore.isAuthenticated(),
        isLoading: authStore.isLoading(),
        error: authStore.getError(),
        isAdmin: authStore.isAdmin(),
        
        // UI state
        showChangePasswordModal: false,

        // Actions
        async login(credentials: LoginRequest) {
            return authStore.login(credentials)
        },

        async register(userData: RegisterRequest) {
            return authStore.register(userData)
        },

        async logout() {
            return authStore.logout()
        },

        async changePassword(oldPassword: string, newPassword: string) {
            return authStore.changePassword(oldPassword, newPassword)
        },

        clearError() {
            authStore.clearError()
        },

        // Utility methods
        hasRole(role: string) {
            return authStore.hasRole(role)
        },

        getUserDisplayName() {
            const user = this.user || authStore.getUser()
            return user ? user.username : 'Guest'
        },

        getUserRole() {
            const user = this.user || authStore.getUser()
            return user ? user.role : null
        },

        // Method to update reactive properties
        _updateFromStore() {
            this.user = authStore.getUser()
            this.currentUser = authStore.getUser()
            this.isAuthenticated = authStore.isAuthenticated()
            this.isLoading = authStore.isLoading()
            this.error = authStore.getError()
            this.isAdmin = authStore.isAdmin()
        }
    }

    // Subscribe to auth store changes to update reactive properties
    authStore.subscribe(() => {
        authData._updateFromStore()
    })

    return authData
} 
