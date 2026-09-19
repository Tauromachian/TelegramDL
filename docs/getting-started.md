# Primeros Pasos e Instalación

Esta guía te guiará en la configuración inicial, los requisitos del sistema y los métodos para ejecutar y compilar **TelegramDL**.

---

## :material-clipboard-list: Requisitos del Sistema

TelegramDL está optimizado para funcionar en múltiples plataformas:

- **Sistemas Operativos**: Windows 10/11, macOS 11+, Linux (Ubuntu, Debian, Fedora, Arch).
- **Go**: Versión 1.22 o superior (para compilar desde código fuente).
- **Node.js**: Versión 18 o superior y `npm` (para compilar el panel web).
- **Wails v2 CLI**: Necesario para el empaquetado de escritorio.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

## :material-key: Obtención de Credenciales de Telegram

Para conectarte a la red MTProto de Telegram necesitas un **API ID** y un **API HASH** personales:

1. Inicia sesión en [my.telegram.org](https://my.telegram.org/) con tu número de teléfono.
2. Ve a la sección **API development tools**.
3. Completa el formulario de registro de aplicación (puedes poner cualquier nombre y título).
4. Copia los valores generados: `App api_id` y `App api_hash`.

> [!IMPORTANT]
> Nunca compartas tu `API ID`, `API HASH`, ni los archivos de sesión generados (`.session`) con terceros.

---

## :material-cog: Configuración Inicial

Existen dos formas de introducir tus credenciales:

### Opción 1: Asistente Gráfico en el Primer Inicio
Al abrir TelegramDL por primera vez, la interfaz mostrará un asistente interactivo que te solicitará:
1. `API ID` y `API HASH`.
2. Número de teléfono internacional (ejemplo: `+34600000000`).
3. Código de verificación (enviado por la app oficial de Telegram o SMS).
4. Contraseña de verificación en dos pasos (2FA), en caso de tenerla activada.

### Opción 2: Archivo de Configuración `.env`
Puedes crear un archivo `.env` en la raíz del proyecto para preconfigurar variables de entorno:

```env
TGDL_API_ID=12345678
TGDL_API_HASH=abcdef0123456789abcdef0123456789
TGDL_BIND_HOST=127.0.0.1
TGDL_PORT=8000
```

---

## :material-folder: Ubicación de Datos y Almacenamiento

TelegramDL guarda la configuración, sesiones y base de datos en una ruta dedicada en tu directorio de usuario:

- **Windows**: `%USERPROFILE%\.tgdown\tgdown.sqlite3`
- **Linux / macOS**: `~/.tgdown/tgdown.sqlite3`

Las descargas se guardan por defecto en la carpeta del sistema `Descargas/TelegramDL` o en la ruta que configures desde la pestaña de **Ajustes**.

---

## :material-laptop: Desarrollo y Compilación

### 1. Instalar dependencias del panel (Vue 3)

```bash
npm --prefix dashboard install
```

### 2. Ejecución en Modo Desarrollo (con Live Reload)

```bash
wails dev
```

Este comando inicia el backend en Go y el servidor de desarrollo de Vite en paralelo, abriendo la ventana de la aplicación y actualizando los cambios al guardar.

### 3. Compilar el Panel Web

```bash
npm --prefix dashboard run build
```

> [!NOTE]
> La carpeta `dashboard/dist/` es generada por Vite y debe existir antes de compilar Go o ejecutar `go vet`, ya que `main.go` utiliza la directiva `//go:embed all:dashboard/dist`.

### 4. Compilar el Ejecutable de Escritorio

```bash
wails build
```

El binario ejecutable final se genera dentro de `build/bin/` (`TelegramDL.exe` en Windows, o binario ejecutable en macOS/Linux).
