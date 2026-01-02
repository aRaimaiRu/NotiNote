# NotiNoteApp - Context & Architecture Documentation

## Project Overview

**NotiNoteApp** is a full-featured note-taking application with intelligent notification capabilities, supporting both mobile and web platforms. The backend is built with Golang, providing a RESTful API with real-time notification delivery.

### Key Features
- Multi-user support with email/password authentication
- CRUD operations for notes
- Scheduled notifications for notes
- Real-time notifications via WebSocket (web) and FCM (mobile)
- Cross-platform support (iOS, Android, Web)

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT APPLICATIONS                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Web App     │  │  iOS App     │  │ Android App  │      │
│  │  (WebSocket) │  │    (FCM)     │  │    (FCM)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   API GATEWAY / LOAD BALANCER                │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│               GOLANG BACKEND (GIN FRAMEWORK)                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  HTTP API Server (Port 8080)                         │   │
│  │  - Auth endpoints (/api/v1/auth/*)                   │   │
│  │  - Note endpoints (/api/v1/notes/*)                  │   │
│  │  - Notification endpoints (/api/v1/notifications/*)  │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  WebSocket Server (Port 8080/ws)                     │   │
│  │  - Real-time notifications for web clients           │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Background Workers (Notification Scheduler)         │   │
│  │  - Poll Redis for due notifications                  │   │
│  │  - Send via FCM/WebSocket                            │   │
│  │  - Update delivery status                            │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
┌──────────────────┐ ┌──────────┐ ┌─────────────────┐
│  PostgreSQL RDS  │ │  Redis   │ │  Firebase FCM   │
│  (Primary DB)    │ │ (Queue)  │ │  (Push Notify)  │
└──────────────────┘ └──────────┘ └─────────────────┘
```

### Component Responsibilities

#### **API Server**
- Handle HTTP requests (REST API)
- JWT authentication & authorization
- Request validation
- Business logic orchestration
- Response formatting

#### **WebSocket Server**
- Maintain persistent connections with web clients
- Broadcast real-time notifications
- Handle connection lifecycle (connect, disconnect, reconnect)

#### **Background Workers**
- Poll database for scheduled notifications
- Push notification jobs to Redis queue
- Process Redis queue messages
- Send notifications via FCM or WebSocket
- Update notification delivery status
- Retry failed deliveries

#### **PostgreSQL Database**
- Store users, notes, notifications
- Track notification delivery status
- Manage sessions (optional, can use Redis)

#### **Redis**
- Queue for notification jobs
- Cache for frequently accessed data
- WebSocket connection state
- Rate limiting data

#### **Firebase Cloud Messaging**
- Deliver push notifications to mobile devices
- Handle device token management

---

## Database Schema

### Entity Relationship Diagram

```
┌─────────────┐         ┌──────────────┐         ┌──────────────────┐
│   users     │ 1     * │    notes     │ 1     * │  notifications   │
│─────────────│◄────────│──────────────│◄────────│──────────────────│
│ id          │         │ id           │         │ id               │
│ email       │         │ user_id (FK) │         │ note_id (FK)     │
│ password    │         │ title        │         │ scheduled_time   │
│ name        │         │ content      │         │ status           │
│ created_at  │         │ tags         │         │ type             │
│ updated_at  │         │ created_at   │         │ created_at       │
└─────────────┘         │ updated_at   │         │ updated_at       │
                        └──────────────┘         └──────────────────┘
                                                           │
                                                           │ 1
                                                           │
                                                           │ *
                                                  ┌────────────────────┐
                                                  │ notification_logs  │
                                                  │────────────────────│
                                                  │ id                 │
                                                  │ notification_id(FK)│
                                                  │ delivery_status    │
                                                  │ delivery_channel   │
                                                  │ error_message      │
                                                  │ attempted_at       │
                                                  └────────────────────┘

┌─────────────┐
│   devices   │
│─────────────│
│ id          │
│ user_id(FK) │◄─── Links to users table
│ device_token│
│ platform    │
│ created_at  │
│ updated_at  │
└─────────────┘
```

### Table Definitions

#### **users**
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**Fields:**
- `id`: Auto-incrementing primary key
- `email`: Unique email address for login
- `password_hash`: Bcrypt hashed password
- `name`: User's display name
- `created_at`: Account creation timestamp
- `updated_at`: Last update timestamp

---

#### **notes**
```sql
CREATE TABLE notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    tags TEXT[], -- PostgreSQL array for tags
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_created_at ON notes(created_at DESC);
CREATE INDEX idx_notes_tags ON notes USING GIN(tags); -- For tag search
```

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users table
- `title`: Note title (required)
- `content`: Note body (optional, can be long text)
- `tags`: Array of tags for categorization
- `created_at`: Note creation timestamp
- `updated_at`: Last modification timestamp

---

#### **notifications**
```sql
CREATE TYPE notification_status AS ENUM ('pending', 'processing', 'sent', 'failed', 'cancelled');
CREATE TYPE notification_type AS ENUM ('one_time', 'recurring_daily', 'recurring_weekly');

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    note_id BIGINT NOT NULL,
    scheduled_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status notification_status DEFAULT 'pending',
    type notification_type DEFAULT 'one_time',
    message TEXT, -- Custom notification message (optional)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_note_id ON notifications(note_id);
CREATE INDEX idx_notifications_scheduled_time ON notifications(scheduled_time);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_pending_due ON notifications(status, scheduled_time)
    WHERE status = 'pending' AND scheduled_time <= CURRENT_TIMESTAMP;
```

**Fields:**
- `id`: Auto-incrementing primary key
- `note_id`: Foreign key to notes table
- `scheduled_time`: When to send the notification
- `status`: Current status (pending, processing, sent, failed, cancelled)
- `type`: Notification type (one-time, recurring)
- `message`: Optional custom notification message (defaults to note title)
- `created_at`: Notification creation timestamp
- `updated_at`: Last status update timestamp

---

#### **notification_logs**
```sql
CREATE TYPE delivery_channel AS ENUM ('fcm_android', 'fcm_ios', 'websocket', 'email');
CREATE TYPE delivery_status AS ENUM ('success', 'failed', 'retrying');

CREATE TABLE notification_logs (
    id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL,
    delivery_status delivery_status NOT NULL,
    delivery_channel delivery_channel NOT NULL,
    error_message TEXT,
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

CREATE INDEX idx_notification_logs_notification_id ON notification_logs(notification_id);
CREATE INDEX idx_notification_logs_attempted_at ON notification_logs(attempted_at DESC);
```

**Fields:**
- `id`: Auto-incrementing primary key
- `notification_id`: Foreign key to notifications table
- `delivery_status`: Success, failed, or retrying
- `delivery_channel`: Which channel was used (FCM, WebSocket, email)
- `error_message`: Error details if delivery failed
- `attempted_at`: When delivery was attempted

---

#### **devices**
```sql
CREATE TYPE device_platform AS ENUM ('ios', 'android', 'web');

CREATE TABLE devices (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_token VARCHAR(500) UNIQUE NOT NULL,
    platform device_platform NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_token ON devices(device_token);
CREATE INDEX idx_devices_active ON devices(user_id, is_active) WHERE is_active = true;
```

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users table
- `device_token`: FCM token or WebSocket connection identifier
- `platform`: iOS, Android, or Web
- `is_active`: Whether device is currently active
- `created_at`: Device registration timestamp
- `updated_at`: Last token update timestamp

---

## API Endpoints

### Base URL
```
Production: https://api.notinoteapp.com/api/v1
Development: http://localhost:8080/api/v1
```

### Authentication Endpoints

#### **POST /auth/register**
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2025-12-30T10:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Validation:**
- Email: Valid email format, unique
- Password: Min 8 characters, must contain uppercase, lowercase, number, special char
- Name: Required, 1-255 characters

---

#### **POST /auth/login**
Authenticate and receive JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-12-31T10:00:00Z"
  }
}
```

---

#### **POST /auth/refresh**
Refresh JWT token.

**Headers:**
```
Authorization: Bearer <current_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-12-31T10:00:00Z"
  }
}
```

---

#### **POST /auth/logout**
Invalidate current token (optional implementation).

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

### Note Endpoints

All note endpoints require authentication via JWT token.

**Headers:**
```
Authorization: Bearer <token>
```

#### **GET /notes**
Retrieve all notes for authenticated user.

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 20, max 100
- `tags` (optional): Filter by tags (comma-separated)
- `search` (optional): Search in title and content
- `sort` (optional): Sort by `created_at`, `updated_at`, default `-created_at` (desc)

**Example:**
```
GET /notes?page=1&limit=20&tags=work,important&sort=-created_at
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "notes": [
      {
        "id": 1,
        "user_id": 1,
        "title": "Meeting Notes",
        "content": "Discuss Q1 goals...",
        "tags": ["work", "important"],
        "created_at": "2025-12-30T10:00:00Z",
        "updated_at": "2025-12-30T11:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

---

#### **GET /notes/:id**
Retrieve a specific note by ID.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Meeting Notes",
    "content": "Discuss Q1 goals...",
    "tags": ["work", "important"],
    "created_at": "2025-12-30T10:00:00Z",
    "updated_at": "2025-12-30T11:00:00Z"
  }
}
```

---

#### **POST /notes**
Create a new note.

**Request:**
```json
{
  "title": "Shopping List",
  "content": "Milk, Eggs, Bread",
  "tags": ["personal", "shopping"]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "user_id": 1,
    "title": "Shopping List",
    "content": "Milk, Eggs, Bread",
    "tags": ["personal", "shopping"],
    "created_at": "2025-12-30T12:00:00Z",
    "updated_at": "2025-12-30T12:00:00Z"
  }
}
```

---

#### **PUT /notes/:id**
Update an existing note.

**Request:**
```json
{
  "title": "Updated Shopping List",
  "content": "Milk, Eggs, Bread, Butter",
  "tags": ["personal", "shopping", "groceries"]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "user_id": 1,
    "title": "Updated Shopping List",
    "content": "Milk, Eggs, Bread, Butter",
    "tags": ["personal", "shopping", "groceries"],
    "created_at": "2025-12-30T12:00:00Z",
    "updated_at": "2025-12-30T12:30:00Z"
  }
}
```

---

#### **DELETE /notes/:id**
Delete a note.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Note deleted successfully"
}
```

---

### Notification Endpoints

#### **GET /notifications**
Retrieve all notifications for user's notes.

**Query Parameters:**
- `page`, `limit`: Pagination
- `status`: Filter by status (pending, sent, failed, cancelled)
- `note_id`: Filter by specific note

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": 1,
        "note_id": 1,
        "scheduled_time": "2025-12-31T09:00:00Z",
        "status": "pending",
        "type": "one_time",
        "message": "Reminder: Meeting Notes",
        "created_at": "2025-12-30T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 10,
      "total_pages": 1
    }
  }
}
```

---

#### **POST /notifications**
Create a new notification for a note.

**Request:**
```json
{
  "note_id": 1,
  "scheduled_time": "2025-12-31T09:00:00Z",
  "type": "one_time",
  "message": "Don't forget the meeting!"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "note_id": 1,
    "scheduled_time": "2025-12-31T09:00:00Z",
    "status": "pending",
    "type": "one_time",
    "message": "Don't forget the meeting!",
    "created_at": "2025-12-30T10:00:00Z"
  }
}
```

---

#### **PUT /notifications/:id**
Update notification (reschedule or change message).

**Request:**
```json
{
  "scheduled_time": "2025-12-31T10:00:00Z",
  "message": "Updated reminder message"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "note_id": 1,
    "scheduled_time": "2025-12-31T10:00:00Z",
    "status": "pending",
    "type": "one_time",
    "message": "Updated reminder message",
    "updated_at": "2025-12-30T11:00:00Z"
  }
}
```

---

#### **DELETE /notifications/:id**
Cancel a notification.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notification cancelled successfully"
}
```

---

### Device Management Endpoints

#### **POST /devices**
Register a device for push notifications.

**Request:**
```json
{
  "device_token": "fcm_token_string_here",
  "platform": "android"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "device_token": "fcm_token_string_here",
    "platform": "android",
    "is_active": true,
    "created_at": "2025-12-30T10:00:00Z"
  }
}
```

---

#### **DELETE /devices/:id**
Unregister a device.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Device unregistered successfully"
}
```

---

### WebSocket Endpoint

#### **WS /ws**
WebSocket connection for real-time notifications.

**Connection:**
```javascript
const ws = new WebSocket('wss://api.notinoteapp.com/ws?token=<jwt_token>');
```

**Message Format (Server → Client):**
```json
{
  "type": "notification",
  "data": {
    "id": 1,
    "note_id": 1,
    "title": "Meeting Notes",
    "message": "Don't forget the meeting!",
    "scheduled_time": "2025-12-31T09:00:00Z",
    "timestamp": "2025-12-31T09:00:01Z"
  }
}
```

**Heartbeat (Client → Server):**
```json
{
  "type": "ping"
}
```

**Heartbeat Response (Server → Client):**
```json
{
  "type": "pong"
}
```

---

## Notification System Design

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│              NOTIFICATION SCHEDULING FLOW                    │
└─────────────────────────────────────────────────────────────┘

1. User creates notification via API
         ↓
2. Save to PostgreSQL (status: pending)
         ↓
3. Scheduler Worker (runs every 30 seconds)
   - Query: SELECT * FROM notifications
            WHERE status='pending'
            AND scheduled_time <= NOW()
         ↓
4. Push to Redis Queue
   - LPUSH notification_queue '{"notification_id": 1, ...}'
   - Update status to 'processing'
         ↓
5. Worker Pool (3-5 workers) process queue
   - BRPOP notification_queue
         ↓
6. Send Notification
   ├── Mobile: FCM API call
   └── Web: WebSocket broadcast
         ↓
7. Update Status & Log
   - notifications.status = 'sent' or 'failed'
   - INSERT INTO notification_logs
         ↓
8. Retry Logic (if failed)
   - Max 3 retries with exponential backoff
   - After 3 failures: status = 'failed'
```

### Redis Queue Structure

#### **Queue Keys**
```
notification_queue          - Main queue (LIST)
notification_processing     - Currently processing (HASH)
notification_failed         - Failed notifications (LIST)
notification_retry_{id}     - Retry counter (STRING with TTL)
```

#### **Queue Operations**

**Push to Queue:**
```go
type NotificationJob struct {
    NotificationID int64     `json:"notification_id"`
    NoteID         int64     `json:"note_id"`
    UserID         int64     `json:"user_id"`
    Message        string    `json:"message"`
    ScheduledTime  time.Time `json:"scheduled_time"`
    RetryCount     int       `json:"retry_count"`
}

// Push to queue
job := NotificationJob{...}
jobJSON, _ := json.Marshal(job)
redis.LPush(ctx, "notification_queue", jobJSON)
```

**Pop from Queue (Blocking):**
```go
// BRPOP blocks until item available or timeout
result, err := redis.BRPop(ctx, 5*time.Second, "notification_queue").Result()
if err == redis.Nil {
    // No items, continue waiting
}
```

### Scheduler Worker (Cron)

**Runs every 30 seconds:**
```go
func NotificationScheduler(db *gorm.DB, redis *redis.Client) {
    ticker := time.NewTicker(30 * time.Second)

    for range ticker.C {
        var notifications []Notification

        // Find due notifications
        db.Where("status = ? AND scheduled_time <= ?",
                 "pending", time.Now()).
           Find(&notifications)

        for _, notif := range notifications {
            // Create job
            job := NotificationJob{
                NotificationID: notif.ID,
                NoteID:         notif.NoteID,
                UserID:         notif.UserID,
                Message:        notif.Message,
                ScheduledTime:  notif.ScheduledTime,
                RetryCount:     0,
            }

            // Push to Redis queue
            jobJSON, _ := json.Marshal(job)
            redis.LPush(ctx, "notification_queue", jobJSON)

            // Update status
            db.Model(&notif).Update("status", "processing")
        }
    }
}
```

### Notification Worker Pool

**Worker Implementation:**
```go
func NotificationWorker(
    id int,
    redis *redis.Client,
    db *gorm.DB,
    fcmClient *fcm.Client,
    wsHub *WebSocketHub,
) {
    log.Printf("Worker %d started", id)

    for {
        // Blocking pop from queue (5 second timeout)
        result, err := redis.BRPop(ctx, 5*time.Second, "notification_queue").Result()
        if err == redis.Nil {
            continue // No jobs, keep waiting
        }

        // Parse job
        var job NotificationJob
        json.Unmarshal([]byte(result[1]), &job)

        // Get user devices
        var devices []Device
        db.Where("user_id = ? AND is_active = ?", job.UserID, true).Find(&devices)

        // Send to each device
        success := true
        for _, device := range devices {
            err := sendNotification(device, job, fcmClient, wsHub)
            if err != nil {
                success = false
                logDeliveryFailure(db, job, device, err)
            } else {
                logDeliverySuccess(db, job, device)
            }
        }

        // Update notification status
        if success {
            db.Model(&Notification{}).
               Where("id = ?", job.NotificationID).
               Update("status", "sent")
        } else {
            handleRetry(redis, db, job)
        }
    }
}

func sendNotification(
    device Device,
    job NotificationJob,
    fcmClient *fcm.Client,
    wsHub *WebSocketHub,
) error {
    switch device.Platform {
    case "android", "ios":
        return sendFCM(fcmClient, device.DeviceToken, job)
    case "web":
        return sendWebSocket(wsHub, device.UserID, job)
    }
    return nil
}
```

### FCM Integration

**Send Push Notification:**
```go
func sendFCM(client *fcm.Client, token string, job NotificationJob) error {
    message := &fcm.Message{
        Token: token,
        Notification: &fcm.Notification{
            Title: "NotiNote Reminder",
            Body:  job.Message,
        },
        Data: map[string]string{
            "note_id":         strconv.FormatInt(job.NoteID, 10),
            "notification_id": strconv.FormatInt(job.NotificationID, 10),
            "type":            "note_reminder",
        },
        Android: &fcm.AndroidConfig{
            Priority: "high",
            Notification: &fcm.AndroidNotification{
                Sound: "default",
                ChannelID: "note_reminders",
            },
        },
        APNS: &fcm.APNSConfig{
            Payload: &fcm.APNSPayload{
                Aps: &fcm.Aps{
                    Sound: "default",
                    Badge: 1,
                },
            },
        },
    }

    response, err := client.Send(context.Background(), message)
    if err != nil {
        return fmt.Errorf("FCM send failed: %w", err)
    }

    log.Printf("FCM sent successfully: %s", response)
    return nil
}
```

### WebSocket Integration

**Broadcast to User:**
```go
type WebSocketHub struct {
    clients    map[int64][]*WebSocketClient // user_id -> clients
    broadcast  chan *NotificationMessage
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    mu         sync.RWMutex
}

func sendWebSocket(hub *WebSocketHub, userID int64, job NotificationJob) error {
    message := &NotificationMessage{
        Type: "notification",
        Data: map[string]interface{}{
            "id":             job.NotificationID,
            "note_id":        job.NoteID,
            "message":        job.Message,
            "scheduled_time": job.ScheduledTime,
            "timestamp":      time.Now(),
        },
    }

    hub.mu.RLock()
    clients := hub.clients[userID]
    hub.mu.RUnlock()

    if len(clients) == 0 {
        return fmt.Errorf("no active WebSocket clients for user %d", userID)
    }

    for _, client := range clients {
        select {
        case client.send <- message:
            // Sent successfully
        default:
            // Client buffer full, close connection
            close(client.send)
        }
    }

    return nil
}
```

### Retry Logic

**Exponential Backoff:**
```go
const (
    MaxRetries = 3
    InitialBackoff = 1 * time.Minute
)

func handleRetry(redis *redis.Client, db *gorm.DB, job NotificationJob) {
    job.RetryCount++

    if job.RetryCount >= MaxRetries {
        // Max retries reached, mark as failed
        db.Model(&Notification{}).
           Where("id = ?", job.NotificationID).
           Update("status", "failed")

        log.Printf("Notification %d failed after %d retries",
                   job.NotificationID, MaxRetries)
        return
    }

    // Calculate backoff: 1min, 2min, 4min, ...
    backoff := InitialBackoff * time.Duration(math.Pow(2, float64(job.RetryCount-1)))

    // Push back to queue with delay
    jobJSON, _ := json.Marshal(job)
    redis.ZAdd(ctx, "notification_delayed", &redis.Z{
        Score:  float64(time.Now().Add(backoff).Unix()),
        Member: jobJSON,
    })

    log.Printf("Notification %d scheduled for retry %d in %v",
               job.NotificationID, job.RetryCount, backoff)
}

// Separate worker to move delayed jobs back to main queue
func DelayedQueueWorker(redis *redis.Client) {
    ticker := time.NewTicker(10 * time.Second)

    for range ticker.C {
        now := float64(time.Now().Unix())

        // Get jobs due for retry
        jobs, _ := redis.ZRangeByScore(ctx, "notification_delayed", &redis.ZRangeBy{
            Min: "0",
            Max: fmt.Sprintf("%f", now),
        }).Result()

        for _, jobJSON := range jobs {
            // Move to main queue
            redis.LPush(ctx, "notification_queue", jobJSON)
            redis.ZRem(ctx, "notification_delayed", jobJSON)
        }
    }
}
```

---

## Authentication System

### JWT Token Structure

**Claims:**
```go
type JWTClaims struct {
    UserID int64  `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

// Token expiration: 24 hours
// Refresh token: 7 days
```

**Token Generation:**
```go
func GenerateToken(user *User) (string, error) {
    claims := JWTClaims{
        UserID: user.ID,
        Email:  user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "notinoteapp",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

**Token Validation Middleware:**
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        claims := &JWTClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)

        c.Next()
    }
}
```

### Password Hashing

**Using bcrypt:**
```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Password Validation

**Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

```go
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }

    var (
        hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
        hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
        hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
        hasSpecial = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
    )

    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number, and special character")
    }

    return nil
}
```

---

## Project Structure

```
NotiNoteApp/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth.go            # Auth endpoints
│   │   │   ├── notes.go           # Note CRUD endpoints
│   │   │   ├── notifications.go   # Notification endpoints
│   │   │   └── devices.go         # Device registration
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT middleware
│   │   │   ├── logging.go         # Request logging
│   │   │   ├── cors.go            # CORS configuration
│   │   │   └── rate_limit.go      # Rate limiting
│   │   └── router.go              # Route registration
│   │
│   ├── models/
│   │   ├── user.go                # User model
│   │   ├── note.go                # Note model
│   │   ├── notification.go        # Notification model
│   │   ├── device.go              # Device model
│   │   └── notification_log.go    # Log model
│   │
│   ├── repository/
│   │   ├── user_repo.go           # User database operations
│   │   ├── note_repo.go           # Note database operations
│   │   ├── notification_repo.go   # Notification DB ops
│   │   └── device_repo.go         # Device DB operations
│   │
│   ├── services/
│   │   ├── auth_service.go        # Auth business logic
│   │   ├── note_service.go        # Note business logic
│   │   ├── notification_service.go # Notification logic
│   │   └── device_service.go      # Device management
│   │
│   ├── notification/
│   │   ├── scheduler.go           # Notification scheduler
│   │   ├── worker.go              # Worker pool
│   │   ├── fcm.go                 # FCM integration
│   │   ├── websocket.go           # WebSocket hub
│   │   └── retry.go               # Retry logic
│   │
│   └── websocket/
│       ├── hub.go                 # WebSocket hub
│       ├── client.go              # WebSocket client
│       └── message.go             # Message types
│
├── pkg/
│   ├── config/
│   │   └── config.go              # Configuration loader
│   ├── database/
│   │   └── postgres.go            # PostgreSQL connection
│   ├── redis/
│   │   └── client.go              # Redis connection
│   ├── validator/
│   │   └── validator.go           # Input validation
│   └── utils/
│       ├── response.go            # API response helpers
│       ├── errors.go              # Error handling
│       └── jwt.go                 # JWT utilities
│
├── migrations/
│   ├── 001_create_users_table.up.sql
│   ├── 001_create_users_table.down.sql
│   ├── 002_create_notes_table.up.sql
│   ├── 002_create_notes_table.down.sql
│   ├── 003_create_notifications_table.up.sql
│   ├── 003_create_notifications_table.down.sql
│   ├── 004_create_notification_logs_table.up.sql
│   ├── 004_create_notification_logs_table.down.sql
│   ├── 005_create_devices_table.up.sql
│   └── 005_create_devices_table.down.sql
│
├── config/
│   ├── config.yaml                # Configuration file
│   └── config.example.yaml        # Example config
│
├── docs/
│   └── api.md                     # API documentation
│
├── scripts/
│   ├── migrate.sh                 # Database migration script
│   └── seed.sh                    # Database seeding
│
├── .env.example                   # Environment variables example
├── .gitignore
├── go.mod                         # Go module dependencies
├── go.sum
├── Dockerfile                     # Docker container
├── docker-compose.yml             # Local development stack
├── Makefile                       # Build and run commands
└── README.md                      # Project documentation
```

---

## Technology Stack

### Core Dependencies

**Web Framework:**
```go
github.com/gin-gonic/gin v1.9.1
```

**Database:**
```go
gorm.io/gorm v1.25.5
gorm.io/driver/postgres v1.5.4
```

**Redis:**
```go
github.com/redis/go-redis/v9 v9.3.0
```

**JWT Authentication:**
```go
github.com/golang-jwt/jwt/v5 v5.2.0
```

**Password Hashing:**
```go
golang.org/x/crypto v0.17.0
```

**Firebase Cloud Messaging:**
```go
firebase.google.com/go/v4 v4.13.0
```

**WebSocket:**
```go
github.com/gorilla/websocket v1.5.1
```

**Configuration:**
```go
github.com/spf13/viper v1.18.2
```

**Validation:**
```go
github.com/go-playground/validator/v10 v10.16.0
```

**Database Migrations:**
```go
github.com/golang-migrate/migrate/v4 v4.17.0
```

**Logging:**
```go
github.com/sirupsen/logrus v1.9.3
```

**CORS:**
```go
github.com/gin-contrib/cors v1.5.0
```

**Rate Limiting:**
```go
github.com/ulule/limiter/v3 v3.11.2
```

**Testing:**
```go
github.com/stretchr/testify v1.8.4
```

### Development Tools

- **Air**: Live reload for Go apps
- **golangci-lint**: Linting
- **swag**: Swagger API documentation generation
- **mockgen**: Mock generation for testing

---

## AWS Deployment Architecture

### Infrastructure Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        AWS CLOUD                             │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │             Route 53 (DNS)                          │    │
│  │  api.notinoteapp.com                                │    │
│  └─────────────────────┬──────────────────────────────┘    │
│                        ▼                                     │
│  ┌────────────────────────────────────────────────────┐    │
│  │    CloudFront (CDN) + SSL Certificate               │    │
│  └─────────────────────┬──────────────────────────────┘    │
│                        ▼                                     │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Application Load Balancer (ALB)                    │    │
│  │  - Health checks                                     │    │
│  │  - SSL termination                                   │    │
│  └─────────────────────┬──────────────────────────────┘    │
│                        ▼                                     │
│  ┌────────────────────────────────────────────────────┐    │
│  │         ECS Fargate Cluster (Auto-scaling)          │    │
│  │  ┌──────────────┐  ┌──────────────┐                │    │
│  │  │ API Server   │  │ API Server   │                │    │
│  │  │ Container 1  │  │ Container 2  │                │    │
│  │  └──────────────┘  └──────────────┘                │    │
│  │  ┌──────────────┐  ┌──────────────┐                │    │
│  │  │   Worker     │  │   Worker     │                │    │
│  │  │ Container 1  │  │ Container 2  │                │    │
│  │  └──────────────┘  └──────────────┘                │    │
│  └────────────┬────────────────┬──────────────────────┘    │
│               │                │                            │
│               ▼                ▼                            │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  RDS PostgreSQL  │  │ ElastiCache Redis│               │
│  │  (Multi-AZ)      │  │  (Cluster Mode)  │               │
│  │  - Primary       │  │  - Replication   │               │
│  │  - Standby       │  │  - Auto-failover │               │
│  └──────────────────┘  └──────────────────┘               │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │         CloudWatch (Monitoring & Logs)              │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │            Secrets Manager                          │    │
│  │  - DB credentials                                   │    │
│  │  - JWT secret                                       │    │
│  │  - FCM credentials                                  │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Service Specifications

#### **ECS Fargate Tasks**

**API Server Task:**
- CPU: 512 (.5 vCPU)
- Memory: 1GB
- Auto-scaling: 2-10 tasks based on CPU/memory
- Health check endpoint: `/health`

**Worker Task:**
- CPU: 256 (.25 vCPU)
- Memory: 512MB
- Auto-scaling: 2-5 tasks based on queue depth
- Runs scheduler + worker pool

#### **RDS PostgreSQL**

**Instance Type:** db.t3.micro (dev) / db.t3.medium (prod)
- Storage: 20GB SSD (auto-scaling enabled)
- Multi-AZ: Enabled (production)
- Automated backups: Daily, 7-day retention
- Encryption: At-rest and in-transit

#### **ElastiCache Redis**

**Node Type:** cache.t3.micro (dev) / cache.t3.medium (prod)
- Cluster mode: Disabled (simple use case)
- Replicas: 1-2
- Automatic failover: Enabled
- Backup retention: 5 days

#### **Application Load Balancer**

- Listeners: HTTP (80) → HTTPS (443)
- SSL Certificate: AWS Certificate Manager
- Health checks: `/health` every 30s
- Target groups: API servers

---

## Implementation Phases

### **Phase 1: Project Setup & Core Infrastructure** (Week 1)

**Tasks:**
1. Initialize Go module and project structure
2. Set up PostgreSQL database locally
3. Create database migrations
4. Implement configuration management (Viper)
5. Set up Redis locally
6. Create Makefile for common tasks
7. Set up Docker Compose for local development

**Deliverables:**
- Running local development environment
- Database schema implemented
- Basic project structure

---

### **Phase 2: Authentication System** (Week 1-2)

**Tasks:**
1. Implement user model and repository
2. Create auth handlers (register, login, refresh)
3. Implement JWT token generation/validation
4. Create auth middleware
5. Add password hashing and validation
6. Write unit tests for auth service

**Deliverables:**
- Working authentication endpoints
- JWT-based authentication
- Protected routes

---

### **Phase 3: Notes CRUD** (Week 2)

**Tasks:**
1. Implement note model and repository
2. Create note handlers (CRUD)
3. Add pagination and filtering
4. Implement tag search
5. Add validation middleware
6. Write unit tests for note service

**Deliverables:**
- Complete note management API
- Search and filter functionality
- Unit tests passing

---

### **Phase 4: Notification System - Core** (Week 3)

**Tasks:**
1. Implement notification model and repository
2. Create notification handlers
3. Build notification scheduler (cron)
4. Implement Redis queue operations
5. Create worker pool
6. Add notification logging

**Deliverables:**
- Notification scheduling working
- Redis queue processing
- Worker pool running

---

### **Phase 5: Push Notification Integration** (Week 3-4)

**Tasks:**
1. Set up Firebase project
2. Integrate FCM SDK
3. Implement device registration
4. Create FCM sender
5. Add retry logic
6. Test on mobile devices (simulator)

**Deliverables:**
- FCM integration working
- Push notifications sent to mobile
- Retry mechanism implemented

---

### **Phase 6: WebSocket Implementation** (Week 4)

**Tasks:**
1. Create WebSocket hub
2. Implement client connection handling
3. Add authentication for WebSocket
4. Build broadcast mechanism
5. Add heartbeat/ping-pong
6. Test real-time delivery

**Deliverables:**
- WebSocket server running
- Real-time notifications on web
- Connection management

---

### **Phase 7: AWS Deployment Setup** (Week 5)

**Tasks:**
1. Create Dockerfile
2. Set up ECR repository
3. Configure ECS cluster and task definitions
4. Set up RDS PostgreSQL
5. Configure ElastiCache Redis
6. Set up ALB and target groups
7. Configure Route 53 and SSL

**Deliverables:**
- Application running on AWS
- Production database and cache
- HTTPS endpoint accessible

---

### **Phase 8: Monitoring & Optimization** (Week 6)

**Tasks:**
1. Set up CloudWatch logs
2. Create CloudWatch dashboards
3. Add performance metrics
4. Implement rate limiting
5. Add CORS configuration
6. Performance testing and optimization

**Deliverables:**
- Production monitoring
- Performance optimizations
- Rate limiting active

---

### **Phase 9: Testing & Documentation** (Week 6-7)

**Tasks:**
1. Write integration tests
2. Load testing (notification system)
3. Security audit
4. API documentation (Swagger)
5. Deployment documentation
6. User guide

**Deliverables:**
- Comprehensive test coverage
- Complete documentation
- Production-ready system

---

## Security Considerations

### **Authentication & Authorization**
- ✅ JWT tokens with expiration
- ✅ HTTPS only in production
- ✅ Bcrypt password hashing (cost 12+)
- ✅ Rate limiting on auth endpoints
- ✅ Token refresh mechanism

### **Data Protection**
- ✅ PostgreSQL encryption at rest
- ✅ Redis encryption in transit (TLS)
- ✅ Secrets in AWS Secrets Manager
- ✅ Environment variables never committed
- ✅ Input validation on all endpoints

### **API Security**
- ✅ CORS configuration (whitelist origins)
- ✅ SQL injection prevention (GORM parameterized queries)
- ✅ XSS prevention (sanitize input)
- ✅ Request size limits
- ✅ Rate limiting per user/IP

### **Infrastructure Security**
- ✅ VPC with private subnets for DB/Redis
- ✅ Security groups (least privilege)
- ✅ ALB with WAF (optional)
- ✅ CloudTrail audit logging
- ✅ Regular dependency updates

---

## Monitoring & Logging

### **CloudWatch Metrics**
- API request latency
- Error rates (4xx, 5xx)
- Database connection pool usage
- Redis queue depth
- Worker processing time
- FCM delivery success rate
- WebSocket connection count

### **CloudWatch Alarms**
- High error rate (> 5%)
- Database CPU > 80%
- Redis memory > 75%
- Queue depth > 1000
- Worker failure rate > 10%

### **Logging Strategy**
- Structured JSON logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID tracing
- Sensitive data redaction
- CloudWatch Logs retention: 30 days

---

## Configuration Files

### **config.yaml**
```yaml
server:
  port: 8080
  mode: production # debug, release, test

database:
  host: ${DB_HOST}
  port: 5432
  name: notinoteapp
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: ${REDIS_HOST}
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0
  pool_size: 10

jwt:
  secret: ${JWT_SECRET}
  expiration: 24h
  refresh_expiration: 168h # 7 days

fcm:
  credentials_file: ${FCM_CREDENTIALS_FILE}

notification:
  scheduler_interval: 30s
  worker_count: 5
  max_retries: 3
  retry_backoff: 1m

cors:
  allowed_origins:
    - https://notinoteapp.com
    - https://app.notinoteapp.com
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Authorization
    - Content-Type

rate_limit:
  requests_per_second: 10
  burst: 20
```

### **.env.example**
```bash
# Server
SERVER_PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=notinoteapp
DB_USER=postgres
DB_PASSWORD=your_password_here

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your_super_secret_jwt_key_here

# Firebase
FCM_CREDENTIALS_FILE=/path/to/firebase-credentials.json

# AWS (for production)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=

# Logging
LOG_LEVEL=info
```

---

## Next Steps

### **Immediate Actions**
1. Review this architecture document
2. Set up development environment (PostgreSQL, Redis)
3. Initialize Go project with dependencies
4. Create database migrations
5. Start with Phase 1 implementation

### **Questions to Answer**
- [ ] Do you have AWS account and access?
- [ ] Do you have Firebase project set up?
- [ ] Do you need help with any specific phase?
- [ ] Any additional features or requirements?

### **Useful Resources**
- [Gin Framework Docs](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging)
- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)

---

**Document Version:** 1.0
**Last Updated:** 2025-12-30
**Author:** Claude Code
**Status:** Ready for Implementation
