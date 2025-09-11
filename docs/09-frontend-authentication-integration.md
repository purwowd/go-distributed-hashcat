# Frontend Authentication Integration

## Overview
Frontend telah berhasil diintegrasikan dengan backend authentication system. Implementasi mencakup halaman login, logout, route protection, dan state management yang lengkap.

## Fitur yang Diimplementasi

### 1. **Authentication Service** (`src/services/auth.service.ts`)
- **Login**: POST `/api/v1/auth/login`
- **Logout**: POST `/api/v1/auth/logout`
- **Token Validation**: POST `/api/v1/auth/validate`
- **Token Refresh**: POST `/api/v1/auth/refresh`
- **Auto-include Authorization header** untuk semua API requests
- **Token storage** menggunakan localStorage
- **JWT token parsing** dan expiration check

### 2. **Authentication Store** (`src/stores/auth.store.ts`)
- **State management** untuk authentication state
- **Reactive updates** dengan subscription system
- **Auto-refresh token** sebelum expiration
- **Error handling** dan loading states
- **Role-based access control** helpers

### 3. **Login Page** (`src/components/auth/login.html`)
- **Modern UI** dengan Tailwind CSS
- **Form validation** dengan Alpine.js
- **Error display** untuk login failures
- **Loading states** dan disabled states
- **Demo credentials** display
- **Responsive design**

### 4. **Logout Component** (`src/components/auth/logout.html`)
- **User dropdown menu** dengan user info
- **Logout button** dengan confirmation
- **Role display** (admin, user, guest)
- **Loading states** untuk logout process

### 5. **Route Protection** (`src/utils/router.ts`)
- **Protected routes** yang memerlukan authentication
- **Automatic redirect** ke login jika tidak authenticated
- **Redirect ke overview** jika sudah login
- **Route refresh** saat auth state berubah

### 6. **API Integration** (`src/services/api.service.ts`)
- **Auto-include Authorization header** untuk semua requests
- **Token management** otomatis
- **Error handling** untuk unauthorized requests

## Authentication Flow

### 1. **Initial Load**
```
1. App loads → Check localStorage for token
2. If token exists → Validate with backend
3. If valid → Set authenticated state
4. If invalid/expired → Clear auth data
5. Route protection → Redirect to login if needed
```

### 2. **Login Process**
```
1. User enters credentials → Login form
2. Submit → Call authService.login()
3. Backend validates → Returns JWT token + user data
4. Store token + user data in localStorage
5. Update auth store state
6. Router refresh → Redirect to overview
```

### 3. **Logout Process**
```
1. User clicks logout → Call authService.logout()
2. Send logout request to backend
3. Clear localStorage (token + user data)
4. Update auth store state
5. Router refresh → Redirect to login
```

### 4. **API Requests**
```
1. Make API request → Auto-include Authorization header
2. If 401 Unauthorized → Auto logout
3. If token expired → Auto refresh (if possible)
4. If refresh fails → Auto logout
```

## State Management

### Auth State Structure
```typescript
interface AuthState {
    isAuthenticated: boolean
    user: User | null
    token: string | null
    isLoading: boolean
    error: string | null
}
```

### User Interface
```typescript
interface User {
    id: string
    username: string
    email: string
    role: 'admin' | 'user' | 'guest'
    is_active: boolean
    created_at: string
    updated_at: string
    last_login?: string
}
```

## Route Protection

### Protected Routes
- `/overview` - Dashboard overview
- `/agents` - Agent management
- `/agent-keys` - Agent key management
- `/jobs` - Job management
- `/files` - File management
- `/wordlists` - Wordlist management
- `/docs` - Documentation

### Public Routes
- `/login` - Login page

### Route Behavior
- **Not authenticated** + **Protected route** → Redirect to `/login`
- **Authenticated** + **Login route** → Redirect to `/overview`
- **Route change** → Check authentication status

## UI Components

### Login Page Features
- **Username/Email input** dengan validation
- **Password input** dengan show/hide toggle
- **Submit button** dengan loading state
- **Error display** untuk login failures
- **Demo credentials** untuk testing
- **Responsive design** untuk mobile/desktop

### Logout Component Features
- **User avatar** dengan initial
- **User info display** (username, email, role)
- **Dropdown menu** dengan options
- **Logout button** dengan confirmation
- **Loading states** untuk logout process

## Security Features

### Token Management
- **JWT tokens** dengan expiration
- **Automatic refresh** sebelum expiration
- **Secure storage** di localStorage
- **Token validation** sebelum setiap request

### Route Security
- **Client-side protection** untuk semua protected routes
- **Automatic redirects** berdasarkan auth status
- **Route refresh** saat auth state berubah

### API Security
- **Authorization header** otomatis untuk semua requests
- **401 handling** dengan auto logout
- **Token refresh** otomatis jika diperlukan

## Testing

### Backend Integration
- ✅ Login endpoint berfungsi
- ✅ Logout endpoint berfungsi
- ✅ Token validation berfungsi
- ✅ Protected endpoints memerlukan authentication
- ✅ Role-based access control berfungsi

### Frontend Features
- ✅ Login form berfungsi
- ✅ Logout component berfungsi
- ✅ Route protection berfungsi
- ✅ State management berfungsi
- ✅ API integration berfungsi
- ✅ Error handling berfungsi

## Default Credentials

### Admin User
- **Username**: `admin`
- **Password**: `admin123`
- **Email**: `admin@hashcat.local`
- **Role**: `admin`

## Environment Variables

### Frontend
- `VITE_API_URL`: Backend API URL (default: `http://localhost:1337`)

### Backend
- `JWT_SECRET_KEY`: Secret key untuk JWT signing
- `JWT_TOKEN_DURATION_HOURS`: Token expiration dalam jam (default: 24)

## File Structure

```
frontend/src/
├── components/
│   └── auth/
│       ├── login.html          # Login page component
│       └── logout.html         # Logout dropdown component
├── services/
│   ├── auth.service.ts         # Authentication API service
│   └── api.service.ts          # Updated with auth headers
├── stores/
│   └── auth.store.ts           # Authentication state management
├── types/
│   └── index.ts                # Authentication type definitions
├── utils/
│   └── router.ts               # Updated with route protection
└── main.ts                     # Updated with auth integration
```

## Usage

### 1. **Access Application**
- Buka `http://localhost:3000`
- Akan otomatis redirect ke login page jika tidak authenticated

### 2. **Login**
- Masukkan username: `admin`
- Masukkan password: `admin123`
- Klik "Sign in"
- Akan redirect ke dashboard overview

### 3. **Logout**
- Klik avatar/username di top-right
- Klik "Sign out" di dropdown menu
- Akan redirect ke login page

### 4. **Navigation**
- Semua protected routes otomatis memerlukan authentication
- Jika token expired, akan auto logout
- Jika sudah login, tidak bisa akses login page

## Error Handling

### Login Errors
- **Invalid credentials**: "Invalid username or password"
- **Network errors**: "Connection failed"
- **Server errors**: "Server error occurred"

### Logout Errors
- **Network errors**: Logout tetap berhasil (clear local data)
- **Server errors**: Logout tetap berhasil (clear local data)

### Token Errors
- **Expired token**: Auto logout
- **Invalid token**: Auto logout
- **Network errors**: Retry dengan refresh token

## Performance

### Optimizations
- **Token refresh** hanya jika diperlukan
- **Route protection** dengan minimal overhead
- **State updates** hanya saat diperlukan
- **API requests** dengan auto-retry

### Monitoring
- **Auth state changes** di console
- **Token refresh** events
- **Route changes** dengan auth status
- **API errors** dengan auto handling

Sistem authentication frontend-backend telah terintegrasi dengan sempurna dan siap untuk production use.
