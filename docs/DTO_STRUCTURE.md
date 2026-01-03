# DTO Structure Documentation

## Overview

This document explains the Data Transfer Object (DTO) structure in NotiNoteApp, following hexagonal architecture principles with clear separation of concerns.

## Directory Structure

```
internal/
├── application/
│   └── dto/                    # Application layer DTOs
│       └── auth.go             # Authentication DTOs (service layer)
│
└── adapters/
    └── primary/
        └── http/
            └── dto/            # HTTP layer DTOs
                ├── request.go  # HTTP request DTOs
                └── response.go # HTTP response DTOs
```

## Layer Separation

### 1. Application Layer DTOs (`internal/application/dto/`)

**Purpose:** Used for data transfer between service layer and handlers/external systems.

**File: `auth.go`**

#### AuthResponse
- Used by all authentication service methods
- Contains user data and tokens
- Returned by: Register, Login, OAuth callbacks, RefreshToken

```go
type AuthResponse struct {
    User         *UserDTO
    AccessToken  string
    RefreshToken string
    ExpiresAt    int64
}
```

#### UserDTO
- Represents user data in application layer
- Excludes sensitive fields (password hash)
- Used in all user-related responses

```go
type UserDTO struct {
    ID        int64
    Email     string
    Name      string
    Provider  domain.AuthProvider
    AvatarURL string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### Input DTOs
Simple structs for service method parameters:
- `LoginInput`
- `RegisterInput`
- `RefreshTokenInput`
- `OAuthCallbackInput`
- `VerifyTokenInput`

#### Helper Functions
- `ToUserDTO(user *domain.User) *UserDTO` - Converts domain User to UserDTO
- `NewAuthResponse(...)` - Creates AuthResponse from domain user and tokens

---

### 2. HTTP Layer DTOs (`internal/adapters/primary/http/dto/`)

**Purpose:** HTTP-specific request/response structures with JSON tags and validation.

**File: `request.go`**

#### Request DTOs
All include Gin binding validation tags:

```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Name     string `json:"name" binding:"required,min=1,max=255"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleTokenRequest struct {
    IDToken string `json:"id_token" binding:"required"`
}

type FacebookTokenRequest struct {
    AccessToken string `json:"access_token" binding:"required"`
}
```

**File: `response.go`**

#### Response DTOs

**AuthResponse** - Full authentication response with nested structures:
```go
type AuthResponse struct {
    Success bool
    Message string
    Data    *struct {
        User struct {
            ID        int64
            Email     string
            Name      string
            Provider  domain.AuthProvider
            AvatarURL string
            CreatedAt time.Time
        }
        AccessToken  string
        RefreshToken string
        TokenType    string
        ExpiresIn    int
    }
}
```

**ErrorResponse** - Standard error format:
```go
type ErrorResponse struct {
    Success bool
    Error   string
}
```

**SuccessResponse** - Generic success format:
```go
type SuccessResponse struct {
    Success bool
    Message string
    Data    interface{}
}
```

**UserResponse** - User profile response:
```go
type UserResponse struct {
    ID        int64
    Email     string
    Name      string
    Provider  domain.AuthProvider
    AvatarURL string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### Helper Functions
- `NewAuthResponse(appResp *appdto.AuthResponse, expiresIn int) AuthResponse`
- `NewUserResponse(user *domain.User) UserResponse`

---

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      CLIENT REQUEST                          │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│           HTTP Handler (auth_handler.go)                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. Bind JSON to HTTP DTO (dto.LoginRequest)         │   │
│  │ 2. Extract fields and call service                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│         Auth Service (auth_service.go)                       │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. Business logic processing                         │   │
│  │ 2. Return appdto.AuthResponse                        │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│           HTTP Handler (auth_handler.go)                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. Convert appdto.AuthResponse → dto.AuthResponse   │   │
│  │ 2. Return JSON to client                             │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      CLIENT RESPONSE                         │
└─────────────────────────────────────────────────────────────┘
```

## Example: Login Flow

### 1. Client sends request
```json
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "Password123!"
}
```

### 2. Handler binds to HTTP DTO
```go
var req dto.LoginRequest
c.ShouldBindJSON(&req)
```

### 3. Handler calls service
```go
authResp, err := h.authService.Login(ctx, req.Email, req.Password)
// authResp is *appdto.AuthResponse
```

### 4. Service returns application DTO
```go
func (s *AuthService) Login(...) (*dto.AuthResponse, error) {
    // ... business logic
    return dto.NewAuthResponse(user, accessToken, refreshToken, 0), nil
}
```

### 5. Handler converts to HTTP DTO
```go
resp := h.buildAuthResponse(authResp)
// resp is dto.AuthResponse (HTTP layer)
c.JSON(http.StatusOK, resp)
```

### 6. Client receives response
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "provider": "email",
      "created_at": "2025-01-03T10:00:00Z"
    },
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

## Benefits

### 1. **Clear Separation of Concerns**
- Application layer doesn't know about HTTP
- HTTP layer doesn't leak into business logic

### 2. **Type Safety**
- Compile-time validation of data structures
- No magic strings or map[string]interface{}

### 3. **Validation**
- HTTP DTOs have Gin validation tags
- Application DTOs are validated by business logic

### 4. **Flexibility**
- Can change HTTP response format without affecting services
- Can add new transport layers (gRPC, WebSocket) easily

### 5. **Testability**
- Services can be tested with simple structs
- No HTTP mocking required for service tests

## Adding New DTOs

### For a new feature (e.g., Notes):

1. **Application Layer DTO** (`internal/application/dto/note.go`)
```go
package dto

type NoteResponse struct {
    Note *NoteDTO
}

type NoteDTO struct {
    ID      int64
    Title   string
    Content string
    Tags    []string
}

type CreateNoteInput struct {
    Title   string
    Content string
    Tags    []string
}
```

2. **HTTP Layer DTOs** (`internal/adapters/primary/http/dto/`)

**In request.go:**
```go
type CreateNoteRequest struct {
    Title   string   `json:"title" binding:"required,min=1,max=255"`
    Content string   `json:"content"`
    Tags    []string `json:"tags"`
}
```

**In response.go:**
```go
type NoteResponse struct {
    Success bool
    Data    *NoteData
}

type NoteData struct {
    ID      int64    `json:"id"`
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Tags    []string `json:"tags"`
}

func NewNoteResponse(note *domain.Note) NoteResponse {
    // conversion logic
}
```

## Best Practices

1. ✅ **Always use DTOs for data transfer between layers**
2. ✅ **Never expose domain entities directly in HTTP responses**
3. ✅ **Keep HTTP concerns (JSON tags, validation) in HTTP layer**
4. ✅ **Keep business logic out of DTOs**
5. ✅ **Use helper functions for conversions**
6. ✅ **Document all DTOs with comments**
7. ✅ **Use consistent naming: Request/Response for HTTP, Input/Output for services**

## Files Modified

- ✅ Created: `internal/application/dto/auth.go`
- ✅ Created: `internal/adapters/primary/http/dto/request.go`
- ✅ Created: `internal/adapters/primary/http/dto/response.go`
- ✅ Updated: `internal/application/services/auth_service.go`
- ✅ Updated: `internal/adapters/primary/http/handlers/auth_handler.go`

---

**Last Updated:** 2026-01-03
**Version:** 1.0
