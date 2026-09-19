# Línea de Comandos (CLI) y Modos de Ejecución

Además de la interfaz gráfica de escritorio habitual, el ejecutable de **TelegramDL** puede iniciarse desde la terminal con diversos parámetros y modos de operación.

---

## :material-laptop: Modos de Ejecución

```bash
# Windows
TelegramDL.exe [opciones]

# Linux / macOS
./TelegramDL [opciones]
```

### Parámetros Disponibles:

| Parámetro | Alias | Descripción |
| :--- | :--- | :--- |
| *(sin parámetros)* | — | Inicia la aplicación completa en modo escritorio con interfaz Wails. |
| `--server` | `-s`, `/server` | **Modo Servidor Headless**: Inicia el servidor REST y WebSocket sin abrir ninguna ventana gráfica. |
| `--update` | `-u`, `/update` | Comprueba si existe una versión más reciente en GitHub y la instala automáticamente. |
| `--help` | `-h`, `/?` | Muestra el mensaje de ayuda con la sintaxis de uso en consola. |

---

## :material-server: Modo Servidor Headless (`--server`)

El modo servidor es ideal para:
- Servidores domésticos o NAS (Synology, TrueNAS, etc.).
- Máquinas virtuales Linux o servidores sin entorno gráfico (headless).
- Despliegues en segundo plano que se controlan exclusivamente desde el teléfono móvil o navegador web mediante [Acceso Remoto](remote-access.md).

```bash
TelegramDL.exe --server
```

Al iniciarse en este modo:
1. Adquiere el bloqueo de instancia única para evitar colisiones.
2. Inicia la conexión MTProto con Telegram.
3. Arranca el servidor HTTP y WebSocket en la IP y puerto configurados (por defecto `0.0.0.0:8000`).
4. Permanece activo escuchando descargas y eventos del Listener hasta recibir una señal de terminación (`Ctrl+C` / `SIGINT`).

---

## :material-sync: Actualización Automática (`--update`)

Puedes forzar una búsqueda y actualización automática de versión directamente desde la línea de comandos:

```bash
TelegramDL.exe --update
```

El actualizador integrado (`pkg/updater/updater.go`):
1. Consulta la API pública de releases en GitHub de TelegramDL.
2. Compara la versión actual con la última versión publicada.
3. Descarga el binario optimizado correspondiente a tu sistema operativo y arquitectura.
4. Sustituye el ejecutable actual de forma atómica y segura.

---

## :material-lock: Bloqueo de Instancia Única

TelegramDL implementa un mecanismo de bloqueo inteligente para evitar que dos procesos utilicen la misma base de datos SQLite o el mismo puerto de red simultáneamente:

- Si intentas abrir una segunda ventana mientras la app ya está corriendo, el nuevo proceso detecta la instancia activa, muestra un aviso y se cierra ordenadamente.
- Al actualizarse, el nuevo proceso espera hasta 8 segundos para que el proceso anterior libere el bloqueo antes de tomar el relevo.
