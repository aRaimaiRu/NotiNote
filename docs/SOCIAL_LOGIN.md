# Social Login Implementation Guide

This document provides a comprehensive guide to the social login implementation in NotiNoteApp, supporting Google OAuth, Facebook OAuth, and traditional email/password authentication.

## Overview

The authentication system supports three authentication methods:
1. **Email/Password**: Traditional authentication with bcrypt password hashing
2. **Google OAuth 2.0**: Social login via Google
3. **Facebook OAuth 2.0**: Social login via Facebook

All authentication methods use JWT tokens for session management.

---

## Architecture

### Hexagonal Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers (Primary Adapter)           │
│  /api/v1/auth/register, /login, /google, /facebook, etc.    │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                 Auth Service (Application Layer)             │
│  Business logic for registration, login, OAuth flows        │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer (Core)                       │
│  User entity, validation, domain errors                     │
└─────────────────────────┬───────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Secondary Adapters (Infrastructure)             │
│  - UserRepository (PostgreSQL)                               │
│  - GoogleProvider (OAuth)                                    │
│  - FacebookProvider (OAuth)                                  │
│  - JWTService (Token generation)                             │
│  - BcryptPasswordHasher (Password hashing)                   │
│  - RedisStateGenerator (CSRF protection)                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Configuration

### Environment Variables (.env)

```bash
# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# OAuth - Google
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# OAuth - Facebook
FACEBOOK_APP_ID=your-facebook-app-id
FACEBOOK_APP_SECRET=your-facebook-app-secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/v1/auth/facebook/callback

# OAuth - General
OAUTH_STATE_SECRET=your-random-state-secret-for-csrf-protection
```

### YAML Configuration (config.yaml)

```yaml
jwt:
  secret: ${JWT_SECRET}
  expiration: 24h
  refresh_expiration: 168h
  issuer: notinoteapp

oauth:
  google:
    client_id: ${GOOGLE_CLIENT_ID}
    client_secret: ${GOOGLE_CLIENT_SECRET}
    redirect_url: ${GOOGLE_REDIRECT_URL}
    scopes:
      - https://www.googleapis.com/auth/userinfo.email
      - https://www.googleapis.com/auth/userinfo.profile
  facebook:
    app_id: ${FACEBOOK_APP_ID}
    app_secret: ${FACEBOOK_APP_SECRET}
    redirect_url: ${FACEBOOK_REDIRECT_URL}
    scopes:
      - email
      - public_profile
  state_secret: ${OAUTH_STATE_SECRET}
```

---

## API Endpoints

### 1. Email/Password Registration

**Endpoint**: `POST /api/v1/auth/register`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response (201 Created)**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "provider": "email",
      "created_at": "2025-12-30T10:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

**Validation Rules**:
- Email: Valid email format, unique
- Password: Minimum 8 characters, must contain:
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
  - At least one special character
- Name: 1-255 characters

---

### 2. Email/Password Login

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK)**: Same as registration response

**Error Responses**:
- `401 Unauthorized`: Invalid email or password
- `403 Forbidden`: Account is inactive

---

### 3. Google OAuth Login

**Step 1: Get Authorization URL**

**Endpoint**: `GET /api/v1/auth/google`

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&scope=...&state=..."
  }
}
```

**Step 2: User authorizes on Google**

User is redirected to the `auth_url`, authorizes the app, and Google redirects back to the callback URL.

**Step 3: Handle Callback**

**Endpoint**: `GET /api/v1/auth/google/callback?code=...&state=...`

**Response (200 OK)**: Same as registration response

**Flow Diagram**:
```
Client                 NotiNoteApp              Google
  │                         │                      │
  │── GET /auth/google ────>│                      │
  │<── auth_url ────────────│                      │
  │                         │                      │
  │────────────── Redirect to auth_url ──────────>│
  │<────────── User authorizes ───────────────────│
  │                         │                      │
  │── Callback with code ──>│                      │
  │                         │── Exchange code ────>│
  │                         │<── Access token ─────│
  │                         │── Get user info ────>│
  │                         │<── User data ────────│
  │<── JWT tokens ──────────│                      │
```

---

### 4. Facebook OAuth Login

**Step 1: Get Authorization URL**

**Endpoint**: `GET /api/v1/auth/facebook`

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "auth_url": "https://www.facebook.com/v18.0/dialog/oauth?client_id=...&redirect_uri=...&scope=...&state=..."
  }
}
```

**Step 2: Handle Callback**

**Endpoint**: `GET /api/v1/auth/facebook/callback?code=...&state=...`

**Response (200 OK)**: Same as registration response

**Flow**: Same as Google OAuth flow

---

### 5. Refresh Token

**Endpoint**: `POST /api/v1/auth/refresh`

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "user": { ... },
    "access_token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

---

### 6. Logout

**Endpoint**: `POST /api/v1/auth/logout`

**Headers**: `Authorization: Bearer <token>`

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Note**: In a stateless JWT system, logout is handled client-side by removing the token. For additional security, implement token blacklisting using Redis.

---

## Database Schema

### Users Table

```sql
CREATE TYPE auth_provider AS ENUM ('email', 'google', 'facebook');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),                    -- NULL for OAuth users
    provider auth_provider NOT NULL DEFAULT 'email',
    provider_id VARCHAR(255),                      -- OAuth provider user ID
    avatar_url VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE            -- Soft delete
);

-- Indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_provider_id ON users(provider, provider_id) WHERE deleted_at IS NULL;
```

**Field Descriptions**:
- `provider`: Authentication method (email, google, facebook)
- `provider_id`: OAuth provider's user ID (e.g., Google's sub claim)
- `password_hash`: Bcrypt hash, NULL for OAuth users
- `avatar_url`: Profile picture from OAuth provider

---

## Security Features

### 1. Password Security
- **Bcrypt hashing** with default cost (10)
- **Password validation**: Enforces strong passwords
- Passwords are never logged or exposed in API responses

### 2. OAuth CSRF Protection
- **State parameter** generated using cryptographically secure random bytes
- State stored in Redis with 10-minute expiration
- One-time use: State is deleted after validation
- Prevents CSRF attacks during OAuth flow

### 3. JWT Token Security
- **HS256 signing algorithm**
- Tokens include user ID and email claims
- Access tokens expire in 24 hours
- Refresh tokens expire in 7 days
- Tokens include issuer and timestamps for validation

### 4. Account Separation
- Users cannot mix authentication methods
- If email exists with provider A, cannot login with provider B
- Clear error messages guide users to correct login method

### 5. Data Validation
- Email format validation with regex
- Password strength validation
- Name length validation
- Input sanitization via Gin validators

---

## Implementation Details

### Domain Layer

**File**: [internal/core/domain/user.go](../internal/core/domain/user.go)

Key entities:
- `User`: Core user entity with business logic
- `AuthProvider`: Enum for authentication methods
- `OAuthUserInfo`: DTO for OAuth user data
- Validation functions: `ValidateEmail`, `ValidatePassword`, `ValidateName`

### Application Layer

**File**: [internal/application/services/auth_service.go](../internal/application/services/auth_service.go)

Key methods:
- `Register()`: Email/password registration
- `Login()`: Email/password login
- `GetOAuthURL()`: Generate OAuth authorization URL
- `HandleOAuthCallback()`: Process OAuth callback
- `RefreshToken()`: Refresh JWT tokens

### Adapters

#### OAuth Providers

**Google**: [internal/adapters/secondary/oauth/google_provider.go](../internal/adapters/secondary/oauth/google_provider.go)
- Uses `golang.org/x/oauth2` library
- Implements OAuth 2.0 authorization code flow
- Fetches user info from Google's userinfo API

**Facebook**: [internal/adapters/secondary/oauth/facebook_provider.go](../internal/adapters/secondary/oauth/facebook_provider.go)
- Custom HTTP client implementation
- Facebook Graph API v18.0
- Requires verified email from user

#### Utilities

**Password Hashing**: [pkg/utils/password.go](../pkg/utils/password.go)
- Bcrypt implementation
- Default cost factor: 10

**JWT Service**: [pkg/utils/jwt.go](../pkg/utils/jwt.go)
- Token generation and validation
- Refresh token support
- Expiration handling

**State Generator**: [pkg/utils/state.go](../pkg/utils/state.go)
- Redis-based state storage
- Cryptographically secure random generation
- One-time use pattern

#### Repository

**User Repository**: [internal/adapters/secondary/database/postgres/repositories/user_repository.go](../internal/adapters/secondary/database/postgres/repositories/user_repository.go)

Methods:
- `Create()`: Create new user
- `FindByID()`: Find by user ID
- `FindByEmail()`: Find by email address
- `FindByProvider()`: Find by OAuth provider and provider ID
- `Update()`: Update user information
- `Delete()`: Soft delete user

---

## Setup Guide

### 1. Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Go to **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
5. Configure OAuth consent screen
6. Set application type to "Web application"
7. Add authorized redirect URIs:
   - Development: `http://localhost:8080/api/v1/auth/google/callback`
   - Production: `https://yourdomain.com/api/v1/auth/google/callback`
8. Copy **Client ID** and **Client Secret** to `.env`

### 2. Facebook OAuth Setup

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app (Consumer type)
3. Add **Facebook Login** product
4. Configure Facebook Login settings:
   - Valid OAuth Redirect URIs:
     - Development: `http://localhost:8080/api/v1/auth/facebook/callback`
     - Production: `https://yourdomain.com/api/v1/auth/facebook/callback`
5. Go to **Settings** → **Basic**
6. Copy **App ID** and **App Secret** to `.env`
7. Make app public (after testing)

### 3. Environment Configuration

Copy `.env.example` to `.env` and fill in credentials:

```bash
cp .env.example .env
# Edit .env with your OAuth credentials
```

### 4. Database Migration

Run the migration to create users table:

```bash
make migrate-up
```

### 5. Start the Application

```bash
# Development mode with live reload
make dev

# Or build and run
make build
./bin/notinoteapp
```

---

## Testing

### Manual Testing with cURL

**Register**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }'
```

**Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Google OAuth**:
```bash
# Get auth URL
curl http://localhost:8080/api/v1/auth/google

# Visit the auth_url in browser, authorize, then:
# You'll be redirected to callback with code and state
```

**Refresh Token**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

### Testing OAuth Locally

For local OAuth testing, you can use tools like:
- **ngrok**: Expose localhost to internet for OAuth callbacks
- **Postman**: OAuth 2.0 authorization helper

```bash
# Using ngrok
ngrok http 8080

# Update redirect URLs in Google/Facebook console to:
# https://your-ngrok-url.ngrok.io/api/v1/auth/google/callback
```

---

## Error Handling

### Common Errors

| Error | Status | Description |
|-------|--------|-------------|
| `ErrUserAlreadyExists` | 409 Conflict | Email already registered |
| `ErrInvalidCredentials` | 401 Unauthorized | Wrong email/password |
| `ErrUserInactive` | 403 Forbidden | Account deactivated |
| `ErrOAuthStateMismatch` | 400 Bad Request | CSRF attack detected |
| `ErrPasswordTooWeak` | 400 Bad Request | Password doesn't meet requirements |
| `ErrInvalidEmail` | 400 Bad Request | Invalid email format |

### Error Response Format

```json
{
  "success": false,
  "error": "User with this email already exists"
}
```

---

## Best Practices

### 1. Production Deployment
- Use HTTPS for all endpoints
- Set strong JWT secret (64+ characters)
- Enable CORS for specific origins only
- Implement rate limiting on auth endpoints
- Monitor failed login attempts
- Use environment-specific OAuth redirect URLs

### 2. Token Management
- Store tokens securely (HttpOnly cookies for web)
- Implement token refresh before expiration
- Consider token blacklisting for logout
- Rotate refresh tokens on use

### 3. OAuth Security
- Always validate state parameter
- Use HTTPS redirect URLs in production
- Request minimal OAuth scopes needed
- Handle provider errors gracefully

### 4. Password Security
- Enforce strong password policies
- Consider implementing password reset flow
- Log password change events
- Never log or expose passwords

---

## Troubleshooting

### OAuth Redirect Mismatch
**Error**: `redirect_uri_mismatch`

**Solution**: Ensure redirect URL in code matches exactly with OAuth provider configuration (including protocol, domain, port, path)

### State Mismatch Error
**Error**: `oauth state mismatch`

**Causes**:
- Redis not running or not accessible
- State expired (>10 minutes)
- Browser cookies disabled

**Solution**: Check Redis connection, ensure state TTL is sufficient

### Email Already Exists
**Error**: When trying to register with OAuth but email exists with different provider

**Solution**: Guide user to login with original provider or implement account linking

---

## Next Steps

1. **Implement password reset flow**:
   - Generate reset tokens
   - Send reset emails
   - Validate reset tokens

2. **Add email verification**:
   - Send verification emails on registration
   - Verify email before activation

3. **Implement account linking**:
   - Allow users to link multiple OAuth providers
   - Merge accounts with same email

4. **Add two-factor authentication (2FA)**:
   - TOTP-based 2FA
   - SMS-based 2FA

5. **Implement session management**:
   - Track active sessions in Redis
   - Allow users to view/revoke sessions
   - Implement device fingerprinting

---

## References

- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Facebook Login Documentation](https://developers.facebook.com/docs/facebook-login)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

---

**Document Version**: 1.0
**Last Updated**: 2025-12-30
**Status**: Implementation Complete
