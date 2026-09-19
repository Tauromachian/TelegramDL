# Getting Started & Installation

This guide walks you through the system requirements, initial setup, and methods to run and compile **TelegramDL**.

---

## :material-clipboard-list: System Requirements

TelegramDL is optimized for cross-platform desktop environments:

- **Operating Systems**: Windows 10/11, macOS 11+, Linux (Ubuntu, Debian, Fedora, Arch).
- **Go**: Version 1.22 or higher (to build from source).
- **Node.js**: Version 18 or higher with `npm` (to build the web dashboard).
- **Wails v2 CLI**: Required for desktop compilation and packaging.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

## :material-key: Obtaining Telegram Credentials

To connect to the Telegram MTProto network, you need personal **API ID** and **API HASH** keys:

1. Sign in to [my.telegram.org](https://my.telegram.org/) with your phone number.
2. Navigate to **API development tools**.
3. Fill out the application registration form (any app name or short title is valid).
4. Copy the generated values: `App api_id` and `App api_hash`.

> [!IMPORTANT]
> Never share your `API ID`, `API HASH`, or session files (`.session`) with third parties.

---

## :material-cog: Initial Configuration

Credentials can be supplied in two ways:

### Method 1: Graphical Setup Wizard
When launching TelegramDL for the first time, an intuitive setup wizard will ask you for:
1. `API ID` and `API HASH`.
2. International phone number format (e.g. `+12025550123`).
3. Verification code (sent via the official Telegram app or SMS).
4. Two-Factor Authentication (2FA) password, if enabled on your account.

### Method 2: Local `.env` File
Create a `.env` file in the root directory to preset configuration:

```env
TGDL_API_ID=12345678
TGDL_API_HASH=abcdef0123456789abcdef0123456789
TGDL_BIND_HOST=127.0.0.1
TGDL_PORT=8000
```

---

## :material-folder: Data Directory & Storage

TelegramDL stores all application state, sessions, and SQLite databases in your user directory:

- **Windows**: `%USERPROFILE%\.tgdown\tgdown.sqlite3`
- **Linux / macOS**: `~/.tgdown/tgdown.sqlite3`

Downloads default to your system's `Downloads/TelegramDL` folder, which can be modified at any time in **Settings**.

---

## :material-laptop: Development & Building

### 1. Install Dashboard Dependencies (Vue 3)

```bash
npm --prefix dashboard install
```

### 2. Run in Development Mode (Live Reload)

```bash
wails dev
```

This runs the Go backend and Vite dev server concurrently, opening the desktop window with hot module replacement.

### 3. Build Web Dashboard

```bash
npm --prefix dashboard run build
```

> [!NOTE]
> The `dashboard/dist/` folder is produced by Vite and must exist prior to compiling Go or running `go vet`, as `main.go` embeds it via `//go:embed all:dashboard/dist`.

### 4. Build Desktop Binary

```bash
wails build
```

The compiled binary will be placed inside `build/bin/` (`TelegramDL.exe` on Windows, or the desktop executable on macOS/Linux).
