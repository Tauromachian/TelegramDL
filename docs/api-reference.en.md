# API & WebSocket Reference

TelegramDL exposes a comprehensive HTTP REST API and a WebSocket feed on its configured port (default `http://127.0.0.1:8000`).

---

## 🔒 Authentication

All endpoints under `/api/*` (except theme queries and the send-token-to-telegram endpoint) require authentication via HTTP header or query parameter:

```http
Authorization: Bearer <YOUR_API_TOKEN>
```
Or alternatively:
```
http://127.0.0.1:8000/api/downloads?token=<YOUR_API_TOKEN>
```

---

## 📥 Download Endpoints

### 1. List Downloads
- **Method**: `GET`
- **Route**: `/api/downloads`
- **Description**: Returns all items (queued, downloading, paused, completed, skipped, and failed).

### 2. Add Download (or Range)
- **Method**: `POST`
- **Route**: `/api/download`
- **JSON Body**:
```json
{
  "url": "https://t.me/c/2121902112/31449"
}
```
*Also accepts range links such as `https://t.me/c/2121902112/31449-31455`.*

### 3. Individual Task Controls
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

### 4. Batch Operations
- `POST /api/downloads/pause-all`: Pause all active transfers.
- `POST /api/downloads/resume-all`: Resume all paused transfers.
- `POST /api/downloads/cancel-all`: Cancel all active and queued transfers.
- `POST /api/clear-history`: Clear completed, failed, and cancelled records from history.

---

## 🎧 Listener Endpoints

- **`GET /api/listener/items`**: List pending detected media items.
- **`GET /api/listener/settings`**: Fetch monitored chat rules and active media filters.
- **`POST /api/listener/settings`**: Update monitored chats and filter flags (`f_photos`, `f_videos`, `f_audios`, `f_docs`, `f_stickers`, `auto_download`).
- **`POST /api/listener/download`**: Enqueue detected item `{ "id": "<id>" }`.
- **`POST /api/listener/clear`**: Clear all detected media records.
- **`POST /api/listener/resolve-chat`**: Resolve channel or group by username or link `{ "query": "channelname" }`.
- **`GET /api/listener/topics`**: Fetch forum topics list for a specific supergroup.

---

## ⚙️ Configuration & System

- **`GET /api/settings`**: Fetch current application settings.
- **`POST /api/settings`**: Update download directory, max concurrency, etc.
- **`POST /api/settings/speed`**: Update global speed limit dynamically `{ "limit_kb": 1024 }`.
- **`GET /api/system/info`**: OS platform, memory usage, and application version.
- **`GET /api/system/disk`**: Available and total storage on the download volume.
- **`POST /api/fs/browse`**: Browse directories on the host file system.
- **`GET /api/logs`**: Fetch current runtime log buffer.
- **`POST /api/app/exit`**: Cleanly terminate the TelegramDL process.

---

## 🔄 Real-time WebSocket

To maintain UI reactivity without polling:

- **Route**: `ws://127.0.0.1:8000/api/ws?token=<YOUR_API_TOKEN>`
- **Behavior**: Upon connection, emits a complete system state snapshot. Any subsequent state change (chunk progress, speed, listener event) broadcasts an incremental JSON update immediately.
