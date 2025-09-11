# Authentication System

## Overview
Sistem authentication telah berhasil diimplementasi dengan fitur login, logout, dan role-based access control menggunakan JWT tokens.

## Endpoints

### Authentication Endpoints (Public)

#### 1. Login
- **POST** `/api/v1/auth/login`
- **Body**:
  ```json
  {
    "username": "admin",
    "password": "admin123"
  }
  ```
- **Response**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "00000000-0000-0000-0000-000000000001",
      "username": "admin",
      "email": "admin@hashcat.local",
      "role": "admin",
      "is_active": true,
      "created_at": "2025-09-11T06:27:44Z",
      "updated_at": "2025-09-11T13:34:48.57067+07:00",
      "last_login": "2025-09-11T13:34:48.57067+07:00"
    },
    "expires_at": "2025-09-12T13:34:48.565087+07:00"
  }
  ```

#### 2. Logout
- **POST** `/api/v1/auth/logout`
- **Body**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Response**:
  ```json
  {
    "message": "Successfully logged out"
  }
  ```

#### 3. Validate Token
- **POST** `/api/v1/auth/validate`
- **Body**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Response**:
  ```json
  {
    "user_id": "00000000-0000-0000-0000-000000000001",
    "username": "admin",
    "email": "admin@hashcat.local",
    "role": "admin",
    "iss": "go-distributed-hashcat",
    "sub": "00000000-0000-0000-0000-000000000001",
    "aud": ["hashcat-users"],
    "exp": 1757658888,
    "nbf": 1757572488,
    "iat": 1757572488
  }
  ```

#### 4. Refresh Token
- **POST** `/api/v1/auth/refresh`
- **Body**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Response**: Same as login response with new token

### User Management Endpoints (Admin Only)

#### 1. Get All Users
- **GET** `/api/v1/users/`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "pagination": {
      "limit": 10,
      "page": 1,
      "total": 2
    },
    "users": [...]
  }
  ```

#### 2. Create User
- **POST** `/api/v1/users/`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "role": "user"
  }
  ```

#### 3. Get User by ID
- **GET** `/api/v1/users/{id}`
- **Headers**: `Authorization: Bearer <token>`

#### 4. Update User
- **PUT** `/api/v1/users/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Body**: Same as create user (all fields optional)

#### 5. Delete User
- **DELETE** `/api/v1/users/{id}`
- **Headers**: `Authorization: Bearer <token>`

## Default User

Sistem sudah dilengkapi dengan default admin user:
- **Username**: `admin`
- **Password**: `admin123`
- **Email**: `admin@hashcat.local`
- **Role**: `admin`

## User Roles

1. **admin**: Full access to all endpoints including user management
2. **user**: Basic access (future: can access jobs and agents)
3. **guest**: Read-only access (future implementation)

## Authentication Flow

1. **Login**: User mengirim username/password ke `/api/v1/auth/login`
2. **Token Generation**: Server memvalidasi credentials dan mengembalikan JWT token
3. **API Access**: Client mengirim token di header `Authorization: Bearer <token>`
4. **Token Validation**: Middleware memvalidasi token untuk setiap protected endpoint
5. **Role Check**: Middleware memeriksa role user untuk endpoint yang memerlukan permission khusus
6. **Logout**: User mengirim token ke `/api/v1/auth/logout` untuk invalidasi

## Security Features

- Password hashing menggunakan bcrypt
- JWT tokens dengan expiration time (default 24 jam)
- Role-based access control
- Secure headers middleware
- CORS protection
- Input validation

## Environment Variables

- `JWT_SECRET_KEY`: Secret key untuk signing JWT tokens (default: fallback key)
- `JWT_TOKEN_DURATION_HOURS`: Durasi token dalam jam (default: 24 jam)

## Database Schema

Tabel `users` telah dibuat dengan migration `006_create_users_table.sql` yang mencakup:
- User information (id, username, email, password)
- Role management
- Account status (is_active)
- Timestamps (created_at, updated_at, last_login)
- Indexes untuk performance

## Testing

Semua endpoint telah ditest dan berfungsi dengan baik:
- ✅ Login dengan credentials yang benar
- ✅ Login gagal dengan credentials yang salah
- ✅ Logout dengan token valid
- ✅ Token validation
- ✅ Protected endpoints dengan authentication
- ✅ Role-based access control
- ✅ User management (create, read, update, delete)
- ✅ Security middleware (unauthorized access blocked)

Sistem authentication sudah siap untuk production use.
