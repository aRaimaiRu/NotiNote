# NotiNoteApp - Notes API Bruno Collection

## Overview

This document describes the comprehensive Bruno collection for all 19 note-related endpoints in the NotiNoteApp backend. All endpoints are fully authenticated, include realistic request/response examples, comprehensive test suites, and complete documentation.

**Collection Version**: 1.1.0
**Total Note Endpoints**: 19
**All Endpoints Status**: Production-Ready for Testing

---

## Table of Contents

1. [Collection Structure](#collection-structure)
2. [Endpoint Groups](#endpoint-groups)
3. [Authentication Requirements](#authentication-requirements)
4. [Request/Response Examples](#requestresponse-examples)
5. [Testing Notes](#testing-notes)
6. [Advanced Features](#advanced-features)

---

## Collection Structure

All note endpoints are located in `bruno/notes/` directory:

```
bruno/notes/
├── BASIC CRUD (6 endpoints)
│   ├── list-notes.bru           - GET /api/v1/notes
│   ├── get-note.bru             - GET /api/v1/notes/:id
│   ├── create-note.bru          - POST /api/v1/notes
│   ├── update-note.bru          - PUT /api/v1/notes/:id
│   ├── delete-note.bru          - DELETE /api/v1/notes/:id
│   └── search-notes.bru         - GET /api/v1/notes/search
│
├── LIFECYCLE (4 endpoints)
│   ├── archive-note.bru         - POST /api/v1/notes/:id/archive
│   ├── unarchive-note.bru       - POST /api/v1/notes/:id/unarchive
│   ├── restore-note.bru         - POST /api/v1/notes/:id/restore
│   └── move-note.bru            - POST /api/v1/notes/:id/move
│
├── HIERARCHY (2 endpoints)
│   ├── get-children.bru         - GET /api/v1/notes/:id/children
│   └── get-ancestors.bru        - GET /api/v1/notes/:id/ancestors
│
├── BLOCKS (5 endpoints)
│   ├── replace-blocks.bru       - PUT /api/v1/notes/:id/blocks
│   ├── add-block.bru            - POST /api/v1/notes/:id/blocks
│   ├── update-block.bru         - PATCH /api/v1/notes/:id/blocks/:block_id
│   ├── delete-block.bru         - DELETE /api/v1/notes/:id/blocks/:block_id
│   └── reorder-blocks.bru       - POST /api/v1/notes/:id/blocks/reorder
│
└── VIEWS & PROPERTIES (2 endpoints)
    ├── update-view-metadata.bru - PUT /api/v1/notes/:id/view
    └── update-properties.bru    - PUT /api/v1/notes/:id/properties
```

---

## Endpoint Groups

### 1. Basic CRUD Operations (6 endpoints)

The foundation for note management - create, read, update, and delete operations.

#### Sequence: 1-6

| # | Method | Endpoint | Purpose |
|---|--------|----------|---------|
| 1 | GET | `/api/v1/notes` | List all notes with pagination, filtering, and sorting |
| 2 | GET | `/api/v1/notes/:id` | Retrieve a specific note with all details |
| 3 | POST | `/api/v1/notes` | Create a new note (root or with parent) |
| 4 | PUT | `/api/v1/notes/:id` | Update note metadata (title, icon, cover) |
| 5 | DELETE | `/api/v1/notes/:id` | Soft delete a note |
| 6 | GET | `/api/v1/notes/search` | Full-text search notes |

**Key Features**:
- Pagination support (page, limit parameters)
- Filtering by parent_id, archived status, search query
- Sorting by created_at, updated_at, title
- Parent-child relationships (max 10 levels deep)
- Soft delete (data preserved, flagged as deleted)

---

### 2. Lifecycle Management (4 endpoints)

Advanced note state management - archiving, restoration, and moving.

#### Sequence: 7-10

| # | Method | Endpoint | Purpose |
|---|--------|----------|---------|
| 7 | POST | `/api/v1/notes/:id/archive` | Archive note (hide from default view) |
| 8 | POST | `/api/v1/notes/:id/unarchive` | Restore archived note to active view |
| 9 | POST | `/api/v1/notes/:id/restore` | Restore soft-deleted note |
| 10 | POST | `/api/v1/notes/:id/move` | Move note to different parent/position |

**State Transitions**:
```
Created Note
    ↓
[Active] ←→ [Archived]  (toggle with archive/unarchive)
    ↓
[Soft Deleted] → [Active] (restore)

Move: Change parent and position while maintaining state
```

---

### 3. Hierarchy Operations (2 endpoints)

Navigate parent-child relationships for breadcrumb and tree views.

#### Sequence: 11-12

| # | Method | Endpoint | Purpose |
|---|--------|----------|---------|
| 11 | GET | `/api/v1/notes/:id/children` | Get direct child notes (immediate children only) |
| 12 | GET | `/api/v1/notes/:id/ancestors` | Get breadcrumb path (root to parent) |

**Use Cases**:
- Building breadcrumb navigation
- Rendering note trees in UI
- Context understanding for a note
- Hierarchical organization visualization

---

### 4. Block Management (5 endpoints)

Rich content editing through block-based system (Notion-like).

#### Sequence: 13-17

| # | Method | Endpoint | Purpose |
|---|--------|----------|---------|
| 13 | PUT | `/api/v1/notes/:id/blocks` | Replace all blocks (full content update) |
| 14 | POST | `/api/v1/notes/:id/blocks` | Add single block to end |
| 15 | PATCH | `/api/v1/notes/:id/blocks/:block_id` | Update block content |
| 16 | DELETE | `/api/v1/notes/:id/blocks/:block_id` | Delete specific block |
| 17 | POST | `/api/v1/notes/:id/blocks/reorder` | Reorder blocks by ID array |

**Block Types Supported**:
- **Text**: paragraph, heading_1-6, quote
- **Lists**: bullet_list, numbered_list, checkbox
- **Code**: code block with language syntax highlighting
- **Divider**: horizontal divider line

**Rich Text Features**:
- Bold, italic, underline, strikethrough
- Inline code, hyperlinks
- Text colors and background colors
- Nested children (for lists)

---

### 5. Database Views & Properties (2 endpoints)

Notion-like database views with custom properties and metadata.

#### Sequence: 18-19

| # | Method | Endpoint | Purpose |
|---|--------|----------|---------|
| 18 | PUT | `/api/v1/notes/:id/view` | Configure database view (table, board, gallery, list) |
| 19 | PUT | `/api/v1/notes/:id/properties` | Set custom properties (key-value pairs) |

**View Types**:
- **table**: Traditional spreadsheet-like view
- **board**: Kanban board (group by select property)
- **gallery**: Grid/gallery layout
- **list**: Linear list view

**Property Types**:
- text, number, select, multi_select
- date, checkbox
- url, email, person reference

---

## Authentication Requirements

All note endpoints require JWT Bearer token authentication.

### Setup Steps

1. **Register/Login** to get access token:
   ```bash
   POST /api/v1/auth/register
   or
   POST /api/v1/auth/login
   ```

2. **Store Token** in environment variable:
   - In Bruno: Set `authToken` environment variable
   - Format: `Authorization: Bearer {{authToken}}`

3. **Token Expiration**:
   - Access token: 24 hours
   - Refresh token: 7 days
   - Use `POST /api/v1/auth/refresh` when token expires

### Header Format

All requests include:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

---

## Request/Response Examples

### Example 1: Create a Note

**Request**:
```bash
POST {{baseUrl}}/api/v1/notes
Authorization: Bearer {{authToken}}
Content-Type: application/json

{
  "title": "Q1 Planning Meeting",
  "parent_id": null,
  "icon": "📊",
  "cover_image": ""
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 42,
    "parent_id": null,
    "title": "Q1 Planning Meeting",
    "icon": "📊",
    "cover_image": "",
    "blocks": [],
    "view_metadata": null,
    "properties": {},
    "path": "/1",
    "depth": 0,
    "position": 0,
    "is_archived": false,
    "is_deleted": false,
    "created_at": "2025-12-30T10:00:00Z",
    "updated_at": "2025-12-30T10:00:00Z"
  }
}
```

---

### Example 2: Add Block with Rich Text

**Request**:
```bash
POST {{baseUrl}}/api/v1/notes/1/blocks
Authorization: Bearer {{authToken}}
Content-Type: application/json

{
  "type": "paragraph",
  "content": {
    "rich_text": [
      {
        "text": "Key objectives for Q1:",
        "style": {
          "bold": true,
          "italic": false,
          "color": "#000000"
        }
      }
    ]
  }
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "blocks": [
      {
        "id": "auto-generated-uuid",
        "type": "paragraph",
        "content": {
          "rich_text": [
            {
              "text": "Key objectives for Q1:",
              "style": {
                "bold": true,
                "italic": false,
                "color": "#000000"
              }
            }
          ]
        },
        "order": 0
      }
    ],
    "updated_at": "2025-12-30T10:30:00Z"
  }
}
```

---

### Example 3: Configure Database View

**Request**:
```bash
PUT {{baseUrl}}/api/v1/notes/1/view
Authorization: Bearer {{authToken}}
Content-Type: application/json

{
  "view_type": "table",
  "properties": [
    {
      "id": "status",
      "name": "Status",
      "type": "select",
      "options": ["Not Started", "In Progress", "Completed"],
      "visible": true,
      "width": 150,
      "position": 0
    },
    {
      "id": "priority",
      "name": "Priority",
      "type": "select",
      "options": ["Low", "Medium", "High"],
      "visible": true,
      "width": 120,
      "position": 1
    }
  ],
  "filters": [
    {
      "property_id": "status",
      "operator": "equals",
      "value": "In Progress"
    }
  ],
  "sorts": [
    {
      "property_id": "priority",
      "direction": "desc"
    }
  ]
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "view_metadata": {
      "view_type": "table",
      "properties": [...],
      "filters": [...],
      "sorts": [...]
    },
    "updated_at": "2025-12-30T11:00:00Z"
  }
}
```

---

### Example 4: Update Custom Properties

**Request**:
```bash
PUT {{baseUrl}}/api/v1/notes/1/properties
Authorization: Bearer {{authToken}}
Content-Type: application/json

{
  "properties": {
    "status": "In Progress",
    "priority": "High",
    "assigned_to": "john@example.com",
    "due_date": "2025-12-31",
    "budget": 50000,
    "completion_percentage": 75,
    "team_members": ["alice@example.com", "bob@example.com"]
  }
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "properties": {
      "status": "In Progress",
      "priority": "High",
      "assigned_to": "john@example.com",
      "due_date": "2025-12-31",
      "budget": 50000,
      "completion_percentage": 75,
      "team_members": ["alice@example.com", "bob@example.com"]
    },
    "updated_at": "2025-12-30T11:30:00Z"
  }
}
```

---

## Testing Notes

### Recommended Test Sequence

1. **Authentication**: Login/Register to get token
2. **Create Root Note**: `POST /api/v1/notes` (no parent)
3. **List Notes**: `GET /api/v1/notes?page=1&limit=20`
4. **Get Note**: `GET /api/v1/notes/{id}`
5. **Create Child Note**: `POST /api/v1/notes` (with parent_id)
6. **Get Children**: `GET /api/v1/notes/{parent_id}/children`
7. **Get Ancestors**: `GET /api/v1/notes/{child_id}/ancestors`
8. **Add Block**: `POST /api/v1/notes/{id}/blocks`
9. **Update Block**: `PATCH /api/v1/notes/{id}/blocks/{block_id}`
10. **Reorder Blocks**: `POST /api/v1/notes/{id}/blocks/reorder`
11. **Delete Block**: `DELETE /api/v1/notes/{id}/blocks/{block_id}`
12. **Update View**: `PUT /api/v1/notes/{id}/view`
13. **Update Properties**: `PUT /api/v1/notes/{id}/properties`
14. **Archive Note**: `POST /api/v1/notes/{id}/archive`
15. **Unarchive Note**: `POST /api/v1/notes/{id}/unarchive`
16. **Delete Note**: `DELETE /api/v1/notes/{id}`
17. **Restore Note**: `POST /api/v1/notes/{id}/restore`
18. **Move Note**: `POST /api/v1/notes/{id}/move`
19. **Search Notes**: `GET /api/v1/notes/search?q=query`

### Each Endpoint Includes

- **meta block**: Endpoint name and sequence number
- **HTTP block**: Method, URL with variables, auth type
- **auth block**: Bearer token configuration
- **headers block**: Content-Type and custom headers
- **body**: Realistic sample data (if applicable)
- **query parameters**: Filter and pagination options (if applicable)
- **tests block**: Comprehensive test assertions
  - Status code validation
  - Response structure validation
  - Field type validation
  - Data consistency checks
- **docs block**: Complete documentation
  - Endpoint description
  - Authentication requirements
  - Parameter specifications
  - Response format
  - Error responses
  - Status codes
  - Use cases

---

## Advanced Features

### 1. Hierarchical Notes

Create nested note structures up to 10 levels deep:

```
Project Root (depth 0)
├── Phase 1 (depth 1)
│   ├── Task 1 (depth 2)
│   ├── Task 2 (depth 2)
│   └── Task 3 (depth 2)
├── Phase 2 (depth 1)
│   └── Review (depth 2)
└── Phase 3 (depth 1)
    ├── Planning (depth 2)
    └── Execution (depth 2)
```

**Constraints**:
- Max depth: 10 levels
- No circular references (note cannot be own ancestor)
- Moving enforces depth limits
- Path field tracks materialized path (e.g., "/1/5/12")

### 2. Rich Block Content

Blocks support complex content structures:

```
Paragraph Block:
  - Rich text with inline formatting
  - Multiple text segments
  - Hyperlinks and colors

List Block:
  - Text content
  - Nested children (sub-items)
  - Checkbox tracking

Code Block:
  - Raw code content
  - Language syntax highlighting
  - Support for 100+ languages
```

### 3. Database Views

Transform notes into databases with custom columns:

```
Table View:
┌─────────────┬──────────┬──────────┐
│ Title       │ Status   │ Priority │
├─────────────┼──────────┼──────────┤
│ Task 1      │ Done     │ High     │
│ Task 2      │ In Prog  │ Medium   │
│ Task 3      │ Todo     │ Low      │
└─────────────┴──────────┴──────────┘

Board View:
┌────────────┬────────────┬──────────┐
│   Todo     │  In Prog   │   Done   │
├────────────┼────────────┼──────────┤
│ Task 1     │ Task 2     │ Task 3   │
│ Task 4     │            │ Task 5   │
└────────────┴────────────┴──────────┘
```

### 4. Custom Properties

Attach arbitrary metadata to notes:

```
properties: {
  status: "In Progress",
  priority: "High",
  assigned_to: "user@example.com",
  due_date: "2025-12-31",
  budget: 50000,
  tags: ["important", "urgent"],
  team: ["alice@example.com", "bob@example.com"],
  progress: 75,
  custom_field: "any value"
}
```

Supports:
- Strings, numbers, booleans
- Arrays and nested objects
- Null values
- No schema validation - store any data

### 5. Soft Delete & Restore

Safe deletion with recovery option:

```
Active Note
    ↓
DELETE /api/v1/notes/:id (soft delete)
    ↓
Soft Deleted (is_deleted: true)
    ↓
POST /api/v1/notes/:id/restore (recovery)
    ↓
Active Note (restored)
```

---

## File Locations

All 19 note endpoint files:

- **g:\NotiNoteApp\bruno\notes\list-notes.bru**
- **g:\NotiNoteApp\bruno\notes\get-note.bru**
- **g:\NotiNoteApp\bruno\notes\create-note.bru**
- **g:\NotiNoteApp\bruno\notes\update-note.bru**
- **g:\NotiNoteApp\bruno\notes\delete-note.bru**
- **g:\NotiNoteApp\bruno\notes\search-notes.bru**
- **g:\NotiNoteApp\bruno\notes\archive-note.bru**
- **g:\NotiNoteApp\bruno\notes\unarchive-note.bru**
- **g:\NotiNoteApp\bruno\notes\restore-note.bru**
- **g:\NotiNoteApp\bruno\notes\move-note.bru**
- **g:\NotiNoteApp\bruno\notes\get-children.bru**
- **g:\NotiNoteApp\bruno\notes\get-ancestors.bru**
- **g:\NotiNoteApp\bruno\notes\replace-blocks.bru**
- **g:\NotiNoteApp\bruno\notes\add-block.bru**
- **g:\NotiNoteApp\bruno\notes\update-block.bru**
- **g:\NotiNoteApp\bruno\notes\delete-block.bru**
- **g:\NotiNoteApp\bruno\notes\reorder-blocks.bru**
- **g:\NotiNoteApp\bruno\notes\update-view-metadata.bru**
- **g:\NotiNoteApp\bruno\notes\update-properties.bru**

Related files:
- **g:\NotiNoteApp\bruno\bruno.json** - Collection metadata
- **g:\NotiNoteApp\bruno\README.md** - Main documentation
- **g:\NotiNoteApp\bruno\NOTES_COLLECTION.md** - This file

---

## Summary

This comprehensive Bruno collection provides:

✅ **19 Production-Ready Endpoints** - All note operations covered
✅ **Complete Test Suites** - Status, structure, and data validation
✅ **Realistic Examples** - Request/response samples with actual data
✅ **Full Documentation** - Specs, validation rules, error responses
✅ **Environment Variables** - Configurable for local/dev/prod
✅ **Authentication** - Bearer token for protected endpoints
✅ **Advanced Features** - Hierarchy, blocks, views, custom properties

Use these endpoints immediately for API testing, integration development, and documentation reference.

---

**Generated**: 2026-01-04
**Framework**: Gin + Go
**Database**: PostgreSQL
**Collection Tool**: Bruno
**Status**: Ready for Production Testing
