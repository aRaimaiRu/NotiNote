# NotiNoteApp Architecture Documentation

## Hexagonal Architecture Overview

This project follows **Hexagonal Architecture** (also known as Ports and Adapters pattern), which provides:

- **Clear separation of concerns**: Business logic is independent of external frameworks
- **Testability**: Easy to test business logic in isolation
- **Flexibility**: Easy to swap implementations (e.g., change database or notification provider)
- **Maintainability**: Changes in one layer don't affect others

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    DRIVING ADAPTERS                          │
│         (Primary/Inbound - HTTP, WebSocket, CLI)            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ HTTP Handler │  │  WebSocket   │  │     CLI      │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │              │
└─────────┼──────────────────┼──────────────────┼──────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    APPLICATION LAYER                         │
│                (Use Cases, Services)                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  AuthUseCase, NoteUseCase, NotificationUseCase       │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                        │
└─────────────────────┼────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      CORE DOMAIN                             │
│                 (Business Logic)                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Domain Entities: User, Note, Notification, Device   │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  Domain Services: Password, JWT, Validation          │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  Ports (Interfaces):                                 │  │
│  │  - UserRepository, NoteRepository                    │  │
│  │  - NotificationQueue, Cache, MessageSender           │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                        │
└─────────────────────┼────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   DRIVEN ADAPTERS                            │
│        (Secondary/Outbound - Database, Cache, etc)          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  PostgreSQL  │  │    Redis     │  │     FCM      │     │
│  │  Repository  │  │    Cache     │  │   Sender     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure Explained

### `/cmd` - Application Entry Points

```
cmd/
└── server/
    └── main.go          # Application bootstrap and dependency injection
```

**Purpose**: Contains the main application entry point. Responsible for:
- Loading configuration
- Setting up dependencies
- Wiring up all layers (dependency injection)
- Starting the HTTP server

**Example**:
```go
func main() {
    // Load config
    cfg := config.Load()

    // Setup infrastructure
    db := setupDatabase(cfg)
    redis := setupRedis(cfg)

    // Setup repositories (adapters)
    userRepo := postgres.NewUserRepository(db)
    noteRepo := postgres.NewNoteRepository(db)

    // Setup services (application layer)
    authService := services.NewAuthService(userRepo, cfg.JWT)
    noteService := services.NewNoteService(noteRepo)

    // Setup handlers (adapters)
    authHandler := handlers.NewAuthHandler(authService)
    noteHandler := handlers.NewNoteHandler(noteService)

    // Start server
    router := setupRouter(authHandler, noteHandler)
    router.Run(":8080")
}
```

---

### `/internal/core` - Core Business Logic

#### `/internal/core/domain` - Domain Entities

```
internal/core/domain/
├── user.go              # User entity
├── note.go              # Note entity
├── notification.go      # Notification entity
├── device.go            # Device entity
└── errors.go            # Domain-specific errors
```

**Purpose**: Contains pure business entities with no external dependencies.

**Example**:
```go
// user.go
type User struct {
    ID           int64
    Email        string
    PasswordHash string
    Name         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (u *User) ValidateEmail() error {
    // Business validation logic
}

func (u *User) SetPassword(password string) error {
    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hash)
    return nil
}
```

#### `/internal/core/ports` - Interface Definitions

```
internal/core/ports/
├── repositories.go      # Repository interfaces
├── services.go          # External service interfaces
├── cache.go            # Cache interface
├── queue.go            # Queue interface
└── messaging.go        # Messaging interface
```

**Purpose**: Defines interfaces (contracts) that adapters must implement. This is the key to dependency inversion.

**Example**:
```go
// repositories.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id int64) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id int64) error
}

type NoteRepository interface {
    Create(ctx context.Context, note *domain.Note) error
    FindByID(ctx context.Context, id int64) (*domain.Note, error)
    FindByUserID(ctx context.Context, userID int64, filters Filters) ([]*domain.Note, error)
    Update(ctx context.Context, note *domain.Note) error
    Delete(ctx context.Context, id int64) error
}

// messaging.go
type MessageSender interface {
    SendPushNotification(ctx context.Context, deviceToken string, message Message) error
    SendToTopic(ctx context.Context, topic string, message Message) error
}

// queue.go
type NotificationQueue interface {
    Enqueue(ctx context.Context, notification *domain.Notification) error
    Dequeue(ctx context.Context) (*domain.Notification, error)
    MarkProcessing(ctx context.Context, id int64) error
    MarkComplete(ctx context.Context, id int64) error
    MarkFailed(ctx context.Context, id int64, err error) error
}
```

---

### `/internal/adapters` - Implementation of Ports

#### `/internal/adapters/primary` - Driving Adapters (Inbound)

These adapters receive requests from the outside world and call application use cases.

##### HTTP Handlers

```
internal/adapters/primary/http/
├── handlers/
│   ├── auth_handler.go          # Authentication endpoints
│   ├── note_handler.go          # Note CRUD endpoints
│   ├── notification_handler.go  # Notification endpoints
│   ├── device_handler.go        # Device registration
│   └── health_handler.go        # Health check endpoint
├── middleware/
│   ├── auth.go                  # JWT authentication middleware
│   ├── logging.go               # Request logging
│   ├── cors.go                  # CORS configuration
│   ├── rate_limit.go            # Rate limiting
│   └── recovery.go              # Panic recovery
└── router.go                    # Route definitions
```

**Example**:
```go
// handlers/note_handler.go
type NoteHandler struct {
    noteUseCase usecases.NoteUseCase
}

func (h *NoteHandler) Create(c *gin.Context) {
    var req CreateNoteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    userID := c.GetInt64("user_id") // From auth middleware

    note, err := h.noteUseCase.CreateNote(c, userID, req.Title, req.Content, req.Tags)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"success": true, "data": note})
}
```

##### WebSocket

```
internal/adapters/primary/websocket/
├── hub.go              # WebSocket hub (manages connections)
├── client.go           # WebSocket client
├── message.go          # Message types
└── handler.go          # WebSocket handler
```

**Example**:
```go
// hub.go
type Hub struct {
    clients    map[int64][]*Client // userID -> clients
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
    mu         sync.RWMutex
}

func (h *Hub) BroadcastToUser(userID int64, message *Message) {
    h.mu.RLock()
    clients := h.clients[userID]
    h.mu.RUnlock()

    for _, client := range clients {
        select {
        case client.send <- message:
        default:
            // Client buffer full, close connection
            close(client.send)
        }
    }
}
```

#### `/internal/adapters/secondary` - Driven Adapters (Outbound)

These adapters implement the port interfaces and interact with external systems.

##### Database (PostgreSQL)

```
internal/adapters/secondary/database/postgres/
├── repositories/
│   ├── user_repository.go           # UserRepository implementation
│   ├── note_repository.go           # NoteRepository implementation
│   ├── notification_repository.go   # NotificationRepository implementation
│   └── device_repository.go         # DeviceRepository implementation
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_notes_table.up.sql
│   └── ...
├── models/                          # Database models (GORM)
│   ├── user.go
│   ├── note.go
│   └── notification.go
└── postgres.go                      # Database connection setup
```

**Example**:
```go
// repositories/user_repository.go
type PostgresUserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) ports.UserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
    dbUser := toDBUser(user) // Convert domain to DB model
    if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
        return err
    }
    user.ID = dbUser.ID
    return nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var dbUser models.User
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&dbUser).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return toDomainUser(&dbUser), nil
}
```

##### Cache (Redis)

```
internal/adapters/secondary/cache/redis/
├── cache.go            # Cache interface implementation
├── session.go          # Session management
└── client.go           # Redis client setup
```

**Example**:
```go
// cache.go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return err
    }
    return json.Unmarshal(data, dest)
}
```

##### Messaging (FCM)

```
internal/adapters/secondary/messaging/fcm/
├── client.go           # FCM client setup
└── sender.go           # MessageSender implementation
```

**Example**:
```go
// sender.go
type FCMSender struct {
    client *messaging.Client
}

func (s *FCMSender) SendPushNotification(ctx context.Context, deviceToken string, msg ports.Message) error {
    message := &messaging.Message{
        Token: deviceToken,
        Notification: &messaging.Notification{
            Title: msg.Title,
            Body:  msg.Body,
        },
        Data: msg.Data,
    }

    _, err := s.client.Send(ctx, message)
    return err
}
```

##### Queue

```
internal/adapters/secondary/queue/
├── redis_queue.go      # Redis-based queue implementation
├── worker_pool.go      # Worker pool
└── scheduler.go        # Notification scheduler
```

**Example**:
```go
// redis_queue.go
type RedisQueue struct {
    client    *redis.Client
    queueName string
}

func (q *RedisQueue) Enqueue(ctx context.Context, notification *domain.Notification) error {
    data, err := json.Marshal(notification)
    if err != nil {
        return err
    }
    return q.client.LPush(ctx, q.queueName, data).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (*domain.Notification, error) {
    result, err := q.client.BRPop(ctx, 5*time.Second, q.queueName).Result()
    if err != nil {
        return nil, err
    }

    var notification domain.Notification
    if err := json.Unmarshal([]byte(result[1]), &notification); err != nil {
        return nil, err
    }

    return &notification, nil
}
```

---

### `/internal/application` - Application Layer

```
internal/application/
├── services/
│   ├── auth_service.go          # Authentication business logic
│   ├── note_service.go          # Note business logic
│   ├── notification_service.go  # Notification business logic
│   └── device_service.go        # Device management
└── usecases/
    ├── auth_usecase.go          # Auth use cases
    ├── note_usecase.go          # Note use cases
    └── notification_usecase.go  # Notification use cases
```

**Purpose**: Orchestrates domain objects and coordinates application flow. Contains use cases and application services.

**Example**:
```go
// usecases/auth_usecase.go
type AuthUseCase struct {
    userRepo ports.UserRepository
    jwtService *services.JWTService
    cache    ports.Cache
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
    // Check if user exists
    existing, _ := uc.userRepo.FindByEmail(ctx, email)
    if existing != nil {
        return nil, "", domain.ErrUserAlreadyExists
    }

    // Create user
    user := &domain.User{
        Email: email,
        Name:  name,
    }

    if err := user.SetPassword(password); err != nil {
        return nil, "", err
    }

    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, "", err
    }

    // Generate JWT
    token, err := uc.jwtService.GenerateToken(user.ID, user.Email)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}
```

---

### `/pkg` - Shared Packages

```
pkg/
├── config/
│   └── config.go           # Configuration loader
├── logger/
│   └── logger.go           # Logging utilities
├── validator/
│   └── validator.go        # Input validation
└── utils/
    ├── response.go         # HTTP response helpers
    ├── errors.go           # Error utilities
    └── jwt.go              # JWT utilities
```

**Purpose**: Reusable utilities that can be used across the application or even in other projects.

---

## Data Flow Example: Create Note

```
1. HTTP Request
   POST /api/v1/notes
   Headers: Authorization: Bearer <token>
   Body: {"title": "Meeting Notes", "content": "...", "tags": ["work"]}

   ↓

2. HTTP Handler (Primary Adapter)
   - Parse request
   - Validate input
   - Extract user ID from JWT (middleware)
   - Call use case

   ↓

3. Note Use Case (Application Layer)
   - Validate business rules
   - Create domain entity
   - Call repository

   ↓

4. Note Repository (Secondary Adapter)
   - Convert domain entity to DB model
   - Save to PostgreSQL
   - Return result

   ↓

5. Response flows back up
   - Repository → Use Case → Handler → HTTP Response
```

## Dependency Injection

All dependencies are injected at the application startup in `cmd/server/main.go`:

```go
func main() {
    // 1. Infrastructure
    db := setupDatabase()
    redis := setupRedis()
    fcm := setupFCM()

    // 2. Adapters (Secondary - Outbound)
    userRepo := postgres.NewUserRepository(db)
    noteRepo := postgres.NewNoteRepository(db)
    cache := rediscache.NewCache(redis)
    queue := queue.NewRedisQueue(redis)
    fcmSender := fcm.NewSender(fcm)

    // 3. Application Layer
    authUseCase := usecases.NewAuthUseCase(userRepo, cache)
    noteUseCase := usecases.NewNoteUseCase(noteRepo)
    notificationUseCase := usecases.NewNotificationUseCase(queue, fcmSender)

    // 4. Adapters (Primary - Inbound)
    authHandler := handlers.NewAuthHandler(authUseCase)
    noteHandler := handlers.NewNoteHandler(noteUseCase)
    notificationHandler := handlers.NewNotificationHandler(notificationUseCase)

    // 5. Start server
    router := setupRouter(authHandler, noteHandler, notificationHandler)
    router.Run(":8080")
}
```

## Benefits of This Architecture

1. **Testability**: Easy to mock interfaces for unit testing
2. **Flexibility**: Easy to swap implementations (e.g., PostgreSQL → MySQL)
3. **Maintainability**: Clear separation of concerns
4. **Independence**: Business logic doesn't depend on frameworks
5. **Scalability**: Can split into microservices easily
6. **Clean**: No circular dependencies

## Testing Strategy

```
tests/
├── unit/
│   ├── domain/          # Test domain logic
│   ├── services/        # Test application services
│   └── handlers/        # Test HTTP handlers (with mocks)
└── integration/
    ├── api/             # Test full API flow
    ├── database/        # Test repository implementations
    └── queue/           # Test queue operations
```

**Unit Tests**: Test each layer in isolation using mocks
**Integration Tests**: Test with real dependencies (Docker containers)

## Next Steps

1. Implement domain entities
2. Define all port interfaces
3. Implement adapters
4. Wire everything together in main.go
5. Add tests
6. Deploy

For detailed implementation guide, see [claude.md](../claude.md).
