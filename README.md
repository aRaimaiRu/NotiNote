# NotiNoteApp

A full-featured note-taking application with intelligent notification capabilities, supporting both mobile and web platforms. Built with Golang backend following hexagonal architecture principles.

## Features

- 📝 **Note Management**: Complete CRUD operations for notes with tags
- 🔔 **Smart Notifications**: Scheduled notifications via FCM (mobile) and WebSocket (web)
- 👥 **Multi-user Support**: Secure authentication with email/password
- 🔒 **JWT Authentication**: Token-based authentication with refresh tokens
- 🌐 **Cross-platform**: iOS, Android, and Web support
- ⚡ **Real-time**: WebSocket for instant web notifications
- 🏗️ **Hexagonal Architecture**: Clean, maintainable, and testable codebase

## Architecture

This project follows **Hexagonal Architecture** (Ports and Adapters pattern):

```
├── cmd/                          # Application entry points
├── internal/
│   ├── core/                     # Business logic (domain layer)
│   │   ├── domain/              # Domain entities
│   │   └── ports/               # Interfaces (ports)
│   ├── adapters/                # Implementation of ports
│   │   ├── primary/             # Driving adapters (HTTP, WebSocket)
│   │   └── secondary/           # Driven adapters (Database, Cache, FCM)
│   └── application/             # Application services and use cases
└── pkg/                         # Shared utilities
```

## Tech Stack

- **Backend**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 16
- **Cache/Queue**: Redis 7
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Push Notifications**: Firebase Cloud Messaging
- **Real-time**: WebSocket (gorilla/websocket)
- **Configuration**: Viper
- **Logging**: Logrus

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 16
- Redis 7
- Docker & Docker Compose (optional, for local development)

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/notinoteapp.git
cd notinoteapp
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Using Docker Compose (Recommended)

```bash
# Start all services (PostgreSQL, Redis, App)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### 4. Manual Setup

#### Install dependencies

```bash
make deps
```

#### Start PostgreSQL and Redis

```bash
# Using Docker
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

#### Run database migrations

```bash
make migrate-up
```

#### Run the application

```bash
# Development mode with live reload
make dev

# Or build and run
make build
./bin/notinoteapp
```

## Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build the application binary
make run               # Run the application
make dev               # Run with live reload (requires air)
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests only
make coverage          # Run tests with coverage report
make lint              # Run linter
make fmt               # Format code
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make migrate-create NAME=migration_name  # Create new migration
make docker-build      # Build Docker image
make docker-up         # Start Docker stack
make docker-down       # Stop Docker stack
```

### Install Development Tools

```bash
make install-tools
```

This installs:
- `air` - Live reload for Go apps
- `migrate` - Database migration tool
- `golangci-lint` - Go linter

### Project Structure Details

```
NotiNoteApp/
├── cmd/
│   └── server/                           # Main application
├── internal/
│   ├── core/
│   │   ├── domain/                       # Domain models (User, Note, Notification)
│   │   └── ports/                        # Interfaces for repositories and services
│   ├── adapters/
│   │   ├── primary/
│   │   │   ├── http/                     # HTTP handlers and middleware
│   │   │   └── websocket/                # WebSocket hub and clients
│   │   └── secondary/
│   │       ├── database/postgres/        # PostgreSQL implementation
│   │       ├── cache/redis/              # Redis implementation
│   │       ├── messaging/fcm/            # FCM implementation
│   │       └── queue/                    # Queue implementation
│   └── application/
│       ├── services/                     # Application services
│       └── usecases/                     # Business use cases
├── pkg/                                  # Shared packages
│   ├── config/                           # Configuration management
│   ├── logger/                           # Logging utilities
│   ├── validator/                        # Input validation
│   └── utils/                            # Helper functions
├── config/                               # Configuration files
├── migrations/                           # Database migrations
├── tests/                                # Test files
├── docs/                                 # Documentation
└── scripts/                              # Utility scripts
```

## API Documentation

### Authentication

```
POST /api/v1/auth/register   - Register new user
POST /api/v1/auth/login      - Login user
POST /api/v1/auth/refresh    - Refresh JWT token
POST /api/v1/auth/logout     - Logout user
```

### Notes

```
GET    /api/v1/notes         - Get all notes (with pagination, search, filters)
GET    /api/v1/notes/:id     - Get note by ID
POST   /api/v1/notes         - Create new note
PUT    /api/v1/notes/:id     - Update note
DELETE /api/v1/notes/:id     - Delete note
```

### Notifications

```
GET    /api/v1/notifications      - Get all notifications
POST   /api/v1/notifications      - Create notification
PUT    /api/v1/notifications/:id  - Update notification
DELETE /api/v1/notifications/:id  - Cancel notification
```

### Devices

```
POST   /api/v1/devices       - Register device for push notifications
DELETE /api/v1/devices/:id   - Unregister device
```

### WebSocket

```
WS /ws?token=<jwt_token>     - WebSocket connection for real-time notifications
```

See [claude.md](claude.md) for complete API documentation.

## Database Migrations

### Create a new migration

```bash
make migrate-create NAME=add_user_profile
```

### Run migrations

```bash
make migrate-up
```

### Rollback migrations

```bash
make migrate-down
```

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make coverage
```

## Deployment

### AWS Deployment Architecture

- **Compute**: ECS Fargate
- **Database**: RDS PostgreSQL (Multi-AZ)
- **Cache**: ElastiCache Redis
- **Load Balancer**: Application Load Balancer
- **CDN**: CloudFront
- **DNS**: Route 53
- **Monitoring**: CloudWatch

See [claude.md](claude.md) for detailed deployment guide.

## Configuration

Configuration can be set via:
1. Environment variables (`.env`)
2. Configuration file (`config/config.yaml`)
3. Command-line flags

Priority: Command-line flags > Environment variables > Config file

## Security

- ✅ JWT authentication with refresh tokens
- ✅ Bcrypt password hashing
- ✅ HTTPS only in production
- ✅ CORS configuration
- ✅ Rate limiting
- ✅ Input validation
- ✅ SQL injection prevention (parameterized queries)
- ✅ Secrets management (AWS Secrets Manager in production)

## Monitoring & Logging

- Structured JSON logging
- Request ID tracing
- CloudWatch metrics and alarms
- Health check endpoint: `/health`

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Support

For issues and questions, please create an issue in the GitHub repository.

## Roadmap

- [x] Project structure setup
- [ ] Core domain models
- [ ] Authentication system
- [ ] Notes CRUD
- [ ] Notification scheduler
- [ ] FCM integration
- [ ] WebSocket server
- [ ] AWS deployment
- [ ] Mobile app integration
- [ ] Web frontend

See [claude.md](claude.md) for detailed implementation phases.
