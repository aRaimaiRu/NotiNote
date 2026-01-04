# NotiNoteApp API - Bruno Collection

This is a comprehensive Bruno API collection for the NotiNoteApp backend. It includes all currently implemented endpoints with complete documentation, test suites, and environment configurations.

## Overview

**Collection Status**: Phase 0 (Authentication) + Phase 1 (Notes System)
**Last Updated**: 2026-01-04
**Total Endpoints**: 26 (Health Check + 6 Auth Endpoints + 19 Note Endpoints)

### Implemented Endpoints

#### Health Check
- `GET /health` - Server health status check

#### Authentication (Public Endpoints)
1. `POST /api/v1/auth/register` - User registration with email/password
2. `POST /api/v1/auth/login` - User login with credentials
3. `POST /api/v1/auth/refresh` - Refresh expired access token
4. `POST /api/v1/auth/google/verify` - Google OAuth token verification
5. `POST /api/v1/auth/facebook/verify` - Facebook OAuth token verification

#### User Profile (Protected Endpoints)
6. `GET /api/v1/me` - Get current authenticated user profile

#### Notes - Basic CRUD (6 Protected Endpoints)
7. `GET /api/v1/notes` - List notes with pagination & filtering
8. `GET /api/v1/notes/:id` - Get specific note
9. `POST /api/v1/notes` - Create new note
10. `PUT /api/v1/notes/:id` - Update note metadata
11. `DELETE /api/v1/notes/:id` - Soft delete note
12. `GET /api/v1/notes/search` - Search notes by query

#### Notes - Lifecycle (4 Protected Endpoints)
13. `POST /api/v1/notes/:id/archive` - Archive note
14. `POST /api/v1/notes/:id/unarchive` - Unarchive note
15. `POST /api/v1/notes/:id/restore` - Restore soft-deleted note
16. `POST /api/v1/notes/:id/move` - Move note to different parent

#### Notes - Hierarchy (2 Protected Endpoints)
17. `GET /api/v1/notes/:id/children` - Get direct child notes
18. `GET /api/v1/notes/:id/ancestors` - Get ancestor breadcrumb path

#### Notes - Blocks (5 Protected Endpoints)
19. `PUT /api/v1/notes/:id/blocks` - Replace all blocks
20. `POST /api/v1/notes/:id/blocks` - Add new block
21. `PATCH /api/v1/notes/:id/blocks/:block_id` - Update block content
22. `DELETE /api/v1/notes/:id/blocks/:block_id` - Delete block
23. `POST /api/v1/notes/:id/blocks/reorder` - Reorder blocks

#### Notes - Views & Properties (2 Protected Endpoints)
24. `PUT /api/v1/notes/:id/view` - Update view metadata (database views)
25. `PUT /api/v1/notes/:id/properties` - Update custom properties

## Quick Start

### 1. Import Collection into Bruno

- Open Bruno application
- Click "Open Collection"
- Navigate to `bruno/` directory in NotiNoteApp project
- Select and open

### 2. Configure Environment

1. Click on environment selector (top-left)
2. Choose "Local Development" for local testing
3. Leave authToken empty for now

### 3. Run Health Check

1. Open health-check.bru
2. Click "Send" button
3. Verify you get 200 OK response

### 4. Register New User

1. Open auth/register.bru
2. Update email in body (use unique email each time)
3. Click "Send"
4. Copy access_token from response

### 5. Test Protected Endpoint

1. Open auth/get-current-user.bru
2. Paste access token into authToken environment variable
3. Click "Send"
4. Verify you get user profile data

## Directory Structure

```
bruno/
├── bruno.json                 # Collection metadata
├── README.md                  # This file
├── environments/
│   ├── local.bru             # Local development environment
│   ├── development.bru       # Development environment
│   └── production.bru        # Production environment
├── health-check.bru          # Server health status endpoint
├── auth/
│   ├── register.bru          # User registration endpoint
│   ├── login.bru             # User login endpoint
│   ├── refresh-token.bru     # Token refresh endpoint
│   ├── google-verify.bru     # Google OAuth verification
│   ├── facebook-verify.bru   # Facebook OAuth verification
│   └── get-current-user.bru  # Get user profile endpoint
└── notes/
    ├── list-notes.bru        # List all notes with pagination
    ├── get-note.bru          # Get single note by ID
    ├── create-note.bru       # Create new note
    ├── update-note.bru       # Update note metadata
    ├── delete-note.bru       # Soft delete note
    ├── search-notes.bru      # Search notes by query
    ├── archive-note.bru      # Archive note
    ├── unarchive-note.bru    # Unarchive note
    ├── restore-note.bru      # Restore soft-deleted note
    ├── move-note.bru         # Move note to different parent
    ├── get-children.bru      # Get direct child notes
    ├── get-ancestors.bru     # Get ancestor breadcrumb path
    ├── replace-blocks.bru    # Replace all note blocks
    ├── add-block.bru         # Add new block to note
    ├── update-block.bru      # Update block content
    ├── delete-block.bru      # Delete block from note
    ├── reorder-blocks.bru    # Reorder blocks in note
    ├── update-view-metadata.bru  # Configure database view
    └── update-properties.bru     # Update custom properties
```

## Environment Variables

### Local Development (environments/local.bru)
```
baseUrl: http://localhost:8080
authToken: [leave empty initially, populate after login]
expiresIn: 86400
```

### Development (environments/development.bru)
```
baseUrl: https://api-dev.notinoteapp.com
authToken: [populate after login]
expiresIn: 86400
```

### Production (environments/production.bru)
```
baseUrl: https://api.notinoteapp.com
authToken: [populate after login]
expiresIn: 86400
```

**Note**: The authToken variable is automatically used in the Authorization header for protected endpoints.

## Authentication Flow

### Email/Password Authentication

1. POST /api/v1/auth/register
   - Send: email, password, name
   - Receive: user data + access_token + refresh_token

2. Use access_token in requests:
   Authorization: Bearer {access_token}

3. When access_token expires (24 hours):
   POST /api/v1/auth/refresh
   - Send: refresh_token
   - Receive: new access_token + new refresh_token

4. If refresh_token also expires (7 days):
   - Redirect user to login page

### Google OAuth Authentication

1. Frontend: User clicks "Sign in with Google"
2. Frontend: Get ID token from Google Sign-In SDK
3. Frontend: Send ID token to POST /api/v1/auth/google/verify
4. Backend: Verify token, create/link user, return tokens
5. Frontend: Use returned access_token for API requests

### Facebook OAuth Authentication

1. Frontend: User clicks "Login with Facebook"
2. Frontend: Get access token from Facebook Login SDK
3. Frontend: Send access token to POST /api/v1/auth/facebook/verify
4. Backend: Verify token, create/link user, return tokens
5. Frontend: Use returned access_token for API requests

## Common Tasks

### Task 1: Register and Login

1. Open register.bru, update email, send
2. Copy access_token from response
3. In environment settings, set authToken = copied_token
4. Open get-current-user.bru, send
5. Verify user profile displays correctly

### Task 2: Test Token Refresh

1. Get refresh_token from login response
2. Open refresh-token.bru
3. Update refresh_token in body
4. Send request
5. Use new access_token for subsequent requests

### Task 3: Test Google OAuth

1. Get ID token from Google Sign-In (frontend)
2. Open google-verify.bru
3. Update id_token in body
4. Send request
5. Verify new user created or existing user linked

### Task 4: Test Facebook OAuth

1. Get access token from Facebook Login (frontend)
2. Open facebook-verify.bru
3. Update access_token in body
4. Send request
5. Verify new user created or existing user linked

## Password Requirements

For email/password registration, passwords must:
- Be at least 8 characters long
- Contain at least one uppercase letter (A-Z)
- Contain at least one lowercase letter (a-z)
- Contain at least one number (0-9)
- Contain at least one special character (!@#$%^&*(),.?":{}|<>)

Example valid password: SecurePass123!

## Response Format

### Success Response (2xx)

```json
{
  "success": true,
  "data": {
    // Endpoint-specific data
  }
}
```

### Error Response (4xx/5xx)

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

## Error Codes

### 400 Bad Request
- Invalid email format
- Password too weak
- Missing required fields
- Invalid request body

### 401 Unauthorized
- Invalid email/password combination
- Invalid or expired token
- Token verification failed

### 403 Forbidden
- Account is inactive/deactivated
- User account disabled

### 409 Conflict
- Email already registered

### 500 Internal Server Error
- Database connection error
- Token generation error
- OAuth provider communication error
- Unexpected server error

## Troubleshooting

### 401 Unauthorized on Protected Endpoint

Problem: Getting 401 when accessing /api/v1/me

Solutions:
1. Verify authToken is set in environment variables
2. Check token hasn't expired (24 hour expiration)
3. Use /api/v1/auth/refresh to get new access token
4. Verify Authorization header format: Bearer {token}

### Email Already Exists Error

Problem: Registration fails with "email already exists"

Solutions:
1. Use a different email address
2. Login with existing email instead of registering
3. Use OAuth authentication (google/facebook)

### Token Refresh Fails

Problem: Getting 401 on /api/v1/auth/refresh

Solutions:
1. Verify refresh_token is still valid (7 day expiration)
2. Check refresh_token hasn't been blacklisted
3. Login again to get new tokens

### Google/Facebook OAuth Fails

Problem: Verification fails with "Failed to verify token"

Solutions:
1. Verify token from Google/Facebook SDK (not manually created)
2. Check token hasn't expired
3. Confirm frontend is using correct Client ID
4. Verify scopes include email and public_profile

## Testing Notes Endpoints

All notes endpoints require authentication via Bearer token. Follow the authentication flow above to get an access token, then set it in the environment variables.

### Quick Test Sequence

1. Register/Login to get authToken
2. Create a note: `POST /api/v1/notes` with title "Test Note"
3. Get the note ID from response
4. List all notes: `GET /api/v1/notes`
5. Get specific note: `GET /api/v1/notes/{id}`
6. Add a block: `POST /api/v1/notes/{id}/blocks`
7. Update note: `PUT /api/v1/notes/{id}`
8. Archive note: `POST /api/v1/notes/{id}/archive`
9. Restore note: `POST /api/v1/notes/{id}/restore`
10. Delete note: `DELETE /api/v1/notes/{id}`

### Important Note Features

#### Block Types
The notes system supports rich content through blocks:
- **Text blocks**: paragraph, heading_1 through heading_6, quote
- **List blocks**: bullet_list, numbered_list, checkbox
- **Code blocks**: code (with language syntax highlighting)
- **Divider blocks**: divider (horizontal rule)

#### Rich Text Formatting
Text blocks support inline formatting:
- Bold, italic, underline, strikethrough
- Inline code, links, colors, background highlighting

#### Hierarchy
- Notes can have parent-child relationships
- Maximum nesting depth: 10 levels
- Materialized path tracking for efficient queries

#### Database Views
Notes support Notion-like database views:
- Table, Board (Kanban), Gallery, List views
- Custom properties (text, number, select, date, checkbox, etc.)
- Filters and sorts on properties
- Property types: text, number, select, multi_select, date, checkbox, url, email, person

## Version History

- **v1.1.0** (2026-01-04): Added Phase 1 - Notes System
  - Basic CRUD operations (6 endpoints)
  - Note lifecycle management (4 endpoints)
  - Hierarchy support with breadcrumbs (2 endpoints)
  - Block management with rich text (5 endpoints)
  - Database views and custom properties (2 endpoints)
  - Total: 19 new note endpoints

- **v1.0.0** (2026-01-03): Initial release with Phase 0 endpoints
  - Health check endpoint
  - User registration with email/password
  - User login
  - Token refresh mechanism
  - Google OAuth verification
  - Facebook OAuth verification
  - Get current user profile

---

**Created for**: NotiNoteApp Development
**Framework**: Gin + Go
**Database**: PostgreSQL
**Authentication**: JWT tokens
**Collection Tool**: Bruno

**Status**: Ready for Development Testing
