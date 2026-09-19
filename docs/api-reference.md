# Referencia de API y WebSockets

TelegramDL expone una API HTTP REST completa y un canal WebSocket en el puerto configurado (por defecto `http://127.0.0.1:8000`).

---

## :material-lock: Autenticación

Todas las rutas bajo `/api/*` (excepto la petición para enviar el token a Telegram y el tema público) requieren autenticación mediante encabezado HTTP o parámetro de consulta:

```http
Authorization: Bearer <TU_API_TOKEN>
```
O bien:
```
http://127.0.0.1:8000/api/downloads?token=<TU_API_TOKEN>
```

---

## :material-download: Endpoints de Descargas

### 1. Listar Descargas
- **Método**: `GET`
- **Ruta**: `/api/downloads`
- **Descripción**: Devuelve la lista completa de descargas (en cola, descargando, pausadas, completadas, omitidas y fallidas).

### 2. Añadir Descarga (o Rango)
- **Método**: `POST`
- **Ruta**: `/api/download`
- **Cuerpo JSON**:
```json
{
  "url": "https://t.me/c/2121902112/31449"
}
```
*Soporta también rangos como `https://t.me/c/2121902112/31449-31455`.*

### 3. Control Individual de Tareas
- **Pausar**: `POST /api/pause` con `{"id": "<uuid>"}`
- **Reanudar**: `POST /api/resume` con `{"id": "<uuid>"}`
- **Cancelar**: `POST /api/cancel` con `{"id": "<uuid>"}`
- **Reintentar**: `POST /api/retry` con `{"id": "<uuid>"}`
- **Eliminar**: `POST /api/delete` con:
```json
{
  "id": "<uuid>",
  "delete_file": false
}
```
- **Abrir en Explorador**: `POST /api/open` con `{"id": "<uuid>"}`

### 4. Control Masivo
- `POST /api/downloads/pause-all`: Pausa todas las descargas activas.
- `POST /api/downloads/resume-all`: Reanuda todas las descargas pausadas.
- `POST /api/downloads/cancel-all`: Cancela todas las descargas activas y en cola.
- `POST /api/clear-history`: Elimina del historial las descargas completadas, fallidas o canceladas.

---

## :material-headphones: Endpoints del Modo Escucha (Listener)

- **`GET /api/listener/items`**: Obtiene la lista de archivos detectados pendientes de descarga.
- **`GET /api/listener/settings`**: Devuelve la configuración de chats vigilados y filtros activos.
- **`POST /api/listener/settings`**: Actualiza los chats vigilados y sus filtros multimedia (`f_photos`, `f_videos`, `f_audios`, `f_docs`, `f_stickers`, `auto_download`).
- **`POST /api/listener/download`**: Encola un elemento detectado para descargarlo `{ "id": "<id>" }`.
- **`POST /api/listener/clear`**: Limpia la lista de multimedia detectada.
- **`POST /api/listener/resolve-chat`**: Resuelve un canal o grupo por su nombre de usuario o enlace `{ "query": "nombrecanal" }`.
- **`GET /api/listener/topics`**: Obtiene la lista de temas/hilos de un supergrupo con foros.

---

## :material-cog: Configuración y Sistema

- **`GET /api/settings`**: Obtiene la configuración general actual.
- **`POST /api/settings`**: Actualiza rutas de guardado, workers de descarga, etc.
- **`POST /api/settings/speed`**: Modifica el límite global de velocidad en tiempo real `{ "limit_kb": 1024 }`.
- **`GET /api/system/info`**: Información del sistema operativo, versión de la app y uso de memoria.
- **`GET /api/system/disk`**: Espacio disponible y total en el disco de descargas.
- **`POST /api/fs/browse`**: Explora directorios en el sistema de archivos del servidor.
- **`GET /api/logs`**: Obtiene el buffer de logs en tiempo real.
- **`POST /api/app/exit`**: Cierra limpiamente la aplicación.

---

## :material-sync: WebSocket en Tiempo Real

Para sincronizar la interfaz sin necesidad de peticiones HTTP constantes:

- **Ruta**: `ws://127.0.0.1:8000/api/ws?token=<TU_API_TOKEN>`
- **Comportamiento**: Al conectarse, el servidor envía un snapshot completo del estado. Posteriormente, cada cambio (progreso de descarga, velocidad, nuevo elemento en escucha) emite inmediatamente una actualización estructurada en JSON.
