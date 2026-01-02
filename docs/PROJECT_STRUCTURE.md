# NotiNoteApp Project Structure

## Overview

This document provides a complete overview of the project structure following **Hexagonal Architecture** principles.

## Directory Tree

```
NotiNoteApp/
├── .air.toml                        # Live reload configuration
├── .env.example                     # Environment variables template
├── .gitignore                       # Git ignore rules
├── claude.md                        # Complete architecture documentation
├── docker-compose.yml               # Local development stack
├── Dockerfile                       # Production container image
├── go.mod                          # Go module dependencies
├── Makefile                        # Build and development commands
├── README.md                       # Project overview and quick start
│
├── cmd/                            # Application entry points
│   └── server/                     # Main HTTP server
│       └── main.go                 # Bootstrap and dependency injection
│
├── config/                         # Configuration files
│   └── config.example.yaml         # Configuration template
│
├── docs/                           # Documentation
│   ├── ARCHITECTURE.md             # Detailed architecture guide
│   ├── API.md                      # API documentation (to be created)
│   └── PROJECT_STRUCTURE.md        # This file
│
├── internal/                       # Private application code
│   ├── core/                       # Core business logic (Domain Layer)
│   │   ├── domain/                 # Domain entities and business rules
│   │   │   ├── user.go            # User entity
│   │   │   ├── note.go            # Note entity
│   │   │   ├── notification.go    # Notification entity
│   │   │   ├── device.go          # Device entity
│   │   │   └── errors.go          # Domain-specific errors
│   │   │
│   │   └── ports/                  # Interface definitions (Dependency Inversion)
│   │       ├── repositories.go    # Repository interfaces
│   │       ├── services.go        # External service interfaces
│   │       ├── cache.go           # Cache interface
│   │       ├── queue.go           # Queue interface
│   │       └── messaging.go       # Messaging/notification interface
│   │
│   ├── application/                # Application Layer (Use Cases)
│   │   ├── services/              # Application services
│   │   │   ├── auth_service.go    # Authentication logic
│   │   │   ├── note_service.go    # Note management logic
│   │   │   ├── notification_service.go  # Notification logic
│   │   │   ├── device_service.go  # Device management
│   │   │   └── jwt_service.go     # JWT utilities
│   │   │
│   │   └── usecases/              # Business use cases
│   │       ├── auth_usecase.go    # Auth operations
│   │       ├── note_usecase.go    # Note operations
│   │       └── notification_usecase.go  # Notification operations
│   │
│   └── adapters/                   # Adapters (Infrastructure Layer)
│       ├── primary/               # Driving Adapters (Inbound)
│       │   ├── http/              # HTTP REST API
│       │   │   ├── handlers/      # Request handlers
│       │   │   │   ├── auth_handler.go
│       │   │   │   ├── note_handler.go
│       │   │   │   ├── notification_handler.go
│       │   │   │   ├── device_handler.go
│       │   │   │   └── health_handler.go
│       │   │   │
│       │   │   ├── middleware/    # HTTP middleware
│       │   │   │   ├── auth.go    # JWT authentication
│       │   │   │   ├── logging.go # Request logging
│       │   │   │   ├── cors.go    # CORS configuration
│       │   │   │   ├── rate_limit.go  # Rate limiting
│       │   │   │   └── recovery.go    # Panic recovery
│       │   │   │
│       │   │   └── router.go      # Route definitions
│       │   │
│       │   └── websocket/         # WebSocket server
│       │       ├── hub.go         # Connection hub
│       │       ├── client.go      # Client connection
│       │       ├── message.go     # Message types
│       │       └── handler.go     # WebSocket handler
│       │
│       └── secondary/             # Driven Adapters (Outbound)
│           ├── database/
│           │   └── postgres/
│           │       ├── repositories/  # Repository implementations
│           │       │   ├── user_repository.go
│           │       │   ├── note_repository.go
│           │       │   ├── notification_repository.go
│           │       │   └── device_repository.go
│           │       │
│           │       ├── migrations/    # SQL migrations
│           │       │   ├── 000001_create_users_table.up.sql
│           │       │   ├── 000001_create_users_table.down.sql
│           │       │   ├── 000002_create_notes_table.up.sql
│           │       │   └── ...
│           │       │
│           │       ├── models/        # GORM database models
│           │       │   ├── user.go
│           │       │   ├── note.go
│           │       │   ├── notification.go
│           │       │   └── device.go
│           │       │
│           │       └── postgres.go    # Database connection setup
│           │
│           ├── cache/
│           │   └── redis/
│           │       ├── cache.go       # Cache implementation
│           │       ├── session.go     # Session management
│           │       └── client.go      # Redis client setup
│           │
│           ├── messaging/
│           │   └── fcm/
│           │       ├── client.go      # FCM client setup
│           │       └── sender.go      # Push notification sender
│           │
│           └── queue/
│               ├── redis_queue.go     # Redis-based queue
│               ├── worker_pool.go     # Worker pool for processing
│               └── scheduler.go       # Notification scheduler
│
├── pkg/                            # Shared/reusable packages
│   ├── config/
│   │   └── config.go              # Configuration loader (Viper)
│   │
│   ├── logger/
│   │   └── logger.go              # Structured logging (Logrus)
│   │
│   ├── validator/
│   │   └── validator.go           # Input validation utilities
│   │
│   └── utils/
│       ├── response.go            # HTTP response helpers
│       ├── errors.go              # Error handling utilities
│       └── jwt.go                 # JWT generation/validation
│
├── scripts/                        # Utility scripts
│   ├── migrate.sh                 # Database migration script
│   ├── seed.sh                    # Database seeding
│   └── deploy.sh                  # Deployment script
│
└── tests/                         # Test files
    ├── unit/                      # Unit tests
    │   ├── domain/               # Domain logic tests
    │   ├── services/             # Service tests (with mocks)
    │   └── handlers/             # Handler tests (with mocks)
    │
    └── integration/               # Integration tests
        ├── api/                  # Full API flow tests
        ├── database/             # Repository tests (real DB)
        └── queue/                # Queue operation tests
```

## Layer Responsibilities

### Core Layer (`/internal/core`)

**No external dependencies** - Pure business logic

- **Domain**: Business entities and rules
- **Ports**: Interface definitions (contracts)

### Application Layer (`/internal/application`)

**Depends only on Core Layer**

- **Services**: Business logic orchestration
- **Use Cases**: Application-specific operations

### Adapters Layer (`/internal/adapters`)

**Implements interfaces from Core Layer**

#### Primary Adapters (Inbound)
- Receive external requests
- Convert to domain operations
- Examples: HTTP handlers, WebSocket, CLI

#### Secondary Adapters (Outbound)
- Implement port interfaces
- Interact with external systems
- Examples: Database, Cache, Message Queue, FCM

### Shared Layer (`/pkg`)

**Reusable utilities** - Can be extracted to separate packages

## Dependency Flow

```
Primary Adapters → Application Layer → Core Layer ← Secondary Adapters
     (HTTP)            (Use Cases)      (Domain)      (Database, etc.)
```

**Key Principle**: Dependencies point inward toward the Core Layer

## File Naming Conventions

- **Entities**: Singular noun (e.g., `user.go`, `note.go`)
- **Interfaces**: Descriptive name ending with purpose (e.g., `user_repository.go`, `cache.go`)
- **Implementations**: Prefix with technology (e.g., `postgres_user_repository.go`, `redis_cache.go`)
- **Handlers**: Suffix with `_handler` (e.g., `auth_handler.go`)
- **Services**: Suffix with `_service` (e.g., `auth_service.go`)
- **Use Cases**: Suffix with `_usecase` (e.g., `auth_usecase.go`)
- **Tests**: Suffix with `_test` (e.g., `user_test.go`)

## Configuration Files

| File | Purpose |
|------|---------|
| `.env.example` | Environment variables template |
| `config/config.example.yaml` | YAML configuration template |
| `.air.toml` | Live reload configuration for development |
| `docker-compose.yml` | Local development infrastructure |
| `Dockerfile` | Production container image |
| `Makefile` | Development and build commands |
| `.gitignore` | Files to exclude from version control |
| `go.mod` | Go module dependencies |

## Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview and quick start |
| `claude.md` | Complete architecture and implementation guide |
| `docs/ARCHITECTURE.md` | Detailed hexagonal architecture explanation |
| `docs/PROJECT_STRUCTURE.md` | This file - project structure overview |
| `docs/API.md` | API endpoint documentation (to be created) |

## Build Artifacts (Not in Git)

These directories are created during build/runtime:

- `bin/` - Compiled binaries
- `tmp/` - Temporary files (Air live reload)
- `logs/` - Application logs
- `vendor/` - Vendored dependencies (optional)
- `coverage.out` / `coverage.html` - Test coverage reports

## Getting Started

1. **Read documentation**:
   - Start with `README.md`
   - Understand architecture: `docs/ARCHITECTURE.md`
   - Review complete design: `claude.md`

2. **Set up environment**:
   ```bash
   cp .env.example .env
   cp config/config.example.yaml config/config.yaml
   # Edit .env and config.yaml
   ```

3. **Install dependencies**:
   ```bash
   make deps
   make install-tools
   ```

4. **Start development**:
   ```bash
   make docker-up    # Start PostgreSQL and Redis
   make migrate-up   # Run database migrations
   make dev          # Start with live reload
   ```

## Next Implementation Steps

1. **Phase 1**: Core Domain Models
   - Implement entities in `/internal/core/domain`
   - Define all ports in `/internal/core/ports`

2. **Phase 2**: Database Layer
   - Create migrations in `/internal/adapters/secondary/database/postgres/migrations`
   - Implement repositories

3. **Phase 3**: Application Layer
   - Implement use cases
   - Implement services

4. **Phase 4**: HTTP Layer
   - Implement handlers
   - Set up routing and middleware

5. **Phase 5**: WebSocket & Queue
   - Implement WebSocket hub
   - Implement notification queue and workers

6. **Phase 6**: Main Application
   - Wire everything together in `cmd/server/main.go`

For detailed implementation guide, see [claude.md](../claude.md).

## Key Design Decisions

1. **Hexagonal Architecture**: Clean separation, easy testing
2. **PostgreSQL**: Relational data, ACID guarantees
3. **Redis**: Queue and cache for performance
4. **FCM**: Cross-platform push notifications
5. **WebSocket**: Real-time web notifications
6. **JWT**: Stateless authentication
7. **GORM**: Type-safe database operations
8. **Gin**: Fast HTTP framework

## Testing Strategy

- **Unit Tests**: Test each layer in isolation with mocks
- **Integration Tests**: Test with real infrastructure (Docker)
- **API Tests**: End-to-end API testing

See test structure in `/tests` directory.
