---
title: API & WebSocket Reference
description: Detailed catalog of REST endpoints and real-time WebSocket communication.
---

TelegramDL exposes a comprehensive HTTP REST API and a WebSocket feed on its configured port (default `http://127.0.0.1:8000`).

---

## Authentication

All endpoints under `/api/*` (except the public endpoints listed below) require authentication via HTTP header or query parameter:

```http
Authorization: Bearer <YOUR_API_TOKEN>
```

Or alternatively:

```text
http://127.0.0.1:8000/api/downloads?token=<YOUR_API_TOKEN>
```

### Public Endpoints (no token required)

The following endpoints can be accessed without authentication:

- **`GET /api/theme`**: Gets the panel theme (color).
- **`POST /api/auth/token/send`**: Sends the access token to Telegram Saved Messages (for remote access).
- **`GET /api/auth/has-credentials`**: Indicates if Telegram credentials are configured and session is active.
- **`GET /api/auth/status`**: Telegram authentication status.
- **`POST /api/auth/credentials`**: Configures Telegram API credentials.
- **`POST /api/auth/send-code`**: Requests verification code to phone.
- **`POST /api/auth/verify-code`**: Verifies the received code.
- **`POST /api/auth/verify-2fa`**: Verifies two-factor password.
- **`POST /api/auth/logout`**: Logs out of Telegram.

### Token Endpoints (require authentication)

- **`GET /api/auth/token`**: Gets the current access token.
- **`POST /api/auth/token/regenerate`**: Regenerates a new access token.

---

## Download Endpoints

### 1. List Downloads

- **Method**: `GET`
- **Route**: `/api/downloads`
- **Description**: Returns all items (queued, downloading, paused, completed, skipped, and failed).

### 2. Additional Download Routes

- **`DELETE /api/downloads/history`**: Clear completed, failed, and cancelled records from history (alias for `/api/clear-history`).
- **`POST /api/downloads/open`**: Open a download in file explorer (alias for `/api/open`).
- **`DELETE /api/downloads/{id}`**: Delete a specific download by ID. Supports `?delete_file=true` parameter to also delete the physical file.

### 3. Add Download (or Range)

- **Method**: `POST`
- **Route**: `/api/download`
- **JSON Body**:

```json

{
  "url": "https://t.me/c/2121902112/31449"
}
```

*Also accepts range links such as `https://t.me/c/2121902112/31449-31455`.*

### 4. Individual Task Controls

- **Pause**: `POST /api/pause` with `{"id": "<uuid>"}`
- **Resume**: `POST /api/resume` with `{"id": "<uuid>"}`
- **Cancel**: `POST /api/cancel` with `{"id": "<uuid>"}`
- **Retry**: `POST /api/retry` with `{"id": "<uuid>"}`
- **Delete**: `POST /api/delete` with:

```json

{
  "id": "<uuid>",
  "delete_file": false
}

```

- **Open in File Explorer**: `POST /api/open` with `{"id": "<uuid>"}`

### 5. Batch Operations

- `POST /api/downloads/pause-all`: Pause all active transfers.
- `POST /api/downloads/resume-all`: Resume all paused transfers.
- `POST /api/downloads/cancel-all`: Cancel all active and queued transfers.
- `POST /api/clear-history`: Clear completed, failed, and cancelled records from history.

---

## Listener Endpoints

- **`GET /api/listener/items`**: List pending detected media items.
- **`GET /api/listener/settings`**: Fetch monitored chat rules and active media filters.
- **`POST /api/listener/settings`**: Update monitored chats and filter flags (`f_photos`, `f_videos`, `f_audios`, `f_docs`, `f_stickers`, `auto_download`).
- **`POST /api/listener/download`**: Enqueue detected item `{ "id": "<id>" }`.
- **`POST /api/listener/clear`**: Clear all detected media records.
- **`POST /api/listener/delete`**: Delete a detected item `{ "id": "<id>" }`.
- **`DELETE /api/listener/item/{id}`**: Delete a detected item by ID (alternative route).
- **`POST /api/listener/resolve-chat`**: Resolve channel or group by username or link `{ "query": "channelname" }`.
- **`GET /api/listener/chat/{query}`**: Resolve channel or group by name (alternative route).
- **`GET /api/listener/topics`**: Fetch forum topics list for a specific supergroup.
- **`GET /api/listener/topics/{chat_id}`**: Fetch topics for a specific chat.
- **`POST /api/listener/update-filename`**: Update the filename of a detected item.

---

## Configuration & System

- **`GET /api/settings`**: Fetch current application settings.
- **`POST /api/settings`**: Update download directory, max concurrency, etc.
- **`POST /api/settings/speed`**: Update global speed limit dynamically `{ "limit_kb": 1024 }`.
- **`POST /api/settings/speed-limit`**: Alias for `/api/settings/speed`.

---

## Log Endpoints

- **`GET /api/logs`**: Fetch current runtime log buffer. Supports parameters:
  - `since=<id>`: Only returns entries after that ID.
  - `limit=<n>`: Maximum number of entries (default: 500).
- **`POST /api/logs/clear`**: Clears the in-memory log buffer (disk file is preserved).
- **`GET /api/logs/export`**: Exports the complete in-memory log as plain text.

---

## Filesystem Endpoints

- **`GET /api/filesystem`**: Browse directories on the host file system (alias).
- **`GET /api/fs/browse`**: Browse directories on the host file system.
- **`POST /api/fs/pick`**: Select a folder (used by native folder picker).

---

## System Endpoints

- **`GET /api/system/info`**: OS platform, memory usage, and application version.
- **`GET /api/server/info`**: Alias for `/api/system/info`.
- **`GET /api/system/disk`**: Available and total storage on the download volume.

---

## Update Endpoints

- **`GET /api/update/check`**: Check for available updates.
- **`GET /api/updates/check`**: Alias for `/api/update/check`.
- **`GET /api/update/progress`**: Get the progress of the current update.
- **`GET /api/updates/progress`**: Alias for `/api/update/progress`.
- **`POST /api/update/install`**: Install the available update.
- **`POST /api/updates/install`**: Alias for `/api/update/install`.
- **`POST /api/update/postpone`**: Postpone the update.

---

## App Control Endpoints

- **`POST /api/app/exit`**: Cleanly terminate the TelegramDL process.
- **`POST /api/exit`**: Alias for `/api/app/exit`.

---

## Real-time WebSocket

To maintain UI reactivity without polling:

- **Route**: `ws://127.0.0.1:8000/api/ws?token=<YOUR_API_TOKEN>`
- **Behavior**: Upon connection, emits a complete system state snapshot. Any subsequent state change (chunk progress, speed, listener event) broadcasts an incremental JSON update immediately.
