package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

const (
	AppVersion = "2.3.8"
	GithubRepo = "infinityxgame/tgdown"

	// DefaultBindHost es la dirección en la que escucha el panel cuando el
	// usuario no ha elegido ninguna. 0.0.0.0 atiende todas las interfaces, de
	// forma que el panel se puede abrir desde el móvil u otro equipo de la red
	// local. Para limitarlo a este ordenador basta con poner
	// TGDL_BIND_HOST=127.0.0.1 en el .env.
	DefaultBindHost = "0.0.0.0"
)

var SpeedMultipliers = map[string]float64{
	"B":  1.0,
	"KB": 1024.0,
	"MB": 1024.0 * 1024.0,
	"GB": 1024.0 * 1024.0 * 1024.0,
}

type SpeedLimit struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type ListenerChat struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	AutoDownload bool   `json:"auto_download"`
	FPhotos      bool   `json:"f_photos"`
	FVideos      bool   `json:"f_videos"`
	FAudios      bool   `json:"f_audios"`
	FDocs        bool   `json:"f_docs"`
	FStickers    bool   `json:"f_stickers"`
}

type Config struct {
	MaxConcurrentDownloads int            `json:"max_concurrent_downloads"`
	ParallelChunks         bool           `json:"parallel_chunks"`
	ChunkWorkers           int            `json:"chunk_workers"`
	DownloadFolder         string         `json:"download_folder"`
	ColorID                *int           `json:"color_id"`
	LoaderColorID          *int           `json:"loader_color_id"`
	SpeedLimit             SpeedLimit     `json:"speed_limit"`
	ListenerEnabled        bool           `json:"listener_enabled"`
	ListenerChats          []ListenerChat `json:"listener_chats"`
	ListenerChatIDs        []int64        `json:"listener_chat_ids"`
}

var (
	DataDir     string
	UserEnvPath string
	BaseDir     string
	initDirOnce sync.Once
)

func InitPaths() {
	initDirOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		DataDir = filepath.Join(home, ".tgdown")
		_ = os.MkdirAll(DataDir, 0755)

		UserEnvPath = filepath.Join(DataDir, ".env")

		// Base directory
		execPath, err := os.Executable()
		if err == nil {
			BaseDir = filepath.Dir(execPath)
		} else {
			BaseDir = "."
		}

		// Migración de .env legacy si existe
		legacyEnv := filepath.Join(BaseDir, ".env")
		if _, err := os.Stat(UserEnvPath); os.IsNotExist(err) {
			if _, lerr := os.Stat(legacyEnv); lerr == nil {
				if content, rerr := os.ReadFile(legacyEnv); rerr == nil {
					_ = os.WriteFile(UserEnvPath, content, 0600)
				}
			} else if _, cerr := os.Stat(".env"); cerr == nil {
				if content, rerr := os.ReadFile(".env"); rerr == nil {
					_ = os.WriteFile(UserEnvPath, content, 0600)
				}
			}
		}

		// Dejar escrito el valor por defecto de TGDL_BIND_HOST antes de cargar
		// nada, para que el archivo recién creado ya sirva en este mismo
		// arranque.
		ensureEnvDefaults()

		// Cargar variables de entorno prioritariamente desde UserEnvPath (.tgdown/.env)
		if _, err := os.Stat(UserEnvPath); err == nil {
			_ = godotenv.Overload(UserEnvPath)
		} else if _, err := os.Stat(legacyEnv); err == nil {
			_ = godotenv.Overload(legacyEnv)
		} else if _, err := os.Stat(".env"); err == nil {
			_ = godotenv.Overload(".env")
		}
	})
}

// ---------------------------------------------------------------------------
// Edición del archivo .env
//
// El .env del usuario puede tener comentarios y variables que la aplicación no
// conoce, así que nunca se reescribe entero: se cambia solo la línea que toca y
// se respeta el resto tal cual estaba.
// ---------------------------------------------------------------------------

func splitEnvLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	// El salto final del archivo deja un elemento vacío que no es una línea.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func joinEnvLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// envKeyLine devuelve el índice de la línea que define una clave, o -1 si esa
// clave no está definida. Ignora comentarios y líneas en blanco.
func envKeyLine(lines []string, key string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "export ")
		name, _, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), key) {
			return i
		}
	}
	return -1
}

// setEnvValue actualiza una clave conservando su posición, o la añade al final
// si todavía no existía.
func setEnvValue(content, key, value string) string {
	lines := splitEnvLines(content)
	if i := envKeyLine(lines, key); i >= 0 {
		lines[i] = key + "=" + value
		return joinEnvLines(lines)
	}
	return joinEnvLines(append(lines, key+"="+value))
}

// ensureEnvDefaults garantiza que el .env del usuario declare en qué dirección
// escucha el panel. Solo escribe si no hay ninguna elección previa: si ya hay
// un TGDL_BIND_HOST (o BIND_HOST), se respeta aunque sea 127.0.0.1.
func ensureEnvDefaults() {
	content := ""
	if data, err := os.ReadFile(UserEnvPath); err == nil {
		content = string(data)
	}

	lines := splitEnvLines(content)
	if envKeyLine(lines, "TGDL_BIND_HOST") >= 0 || envKeyLine(lines, "BIND_HOST") >= 0 {
		return
	}

	if len(lines) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines,
		"# Dirección en la que escucha el panel.",
		"# 0.0.0.0 permite abrirlo desde el móvil u otro equipo de la red local;",
		"# 127.0.0.1 lo limita únicamente a este ordenador.",
		"TGDL_BIND_HOST="+DefaultBindHost,
	)

	if err := os.WriteFile(UserEnvPath, []byte(joinEnvLines(lines)), 0600); err != nil {
		return
	}
	_ = os.Setenv("TGDL_BIND_HOST", DefaultBindHost)
}

func GetDefaultDownloadFolder() string {
	InitPaths()
	home, err := os.UserHomeDir()
	if err == nil {
		downloadsDir := filepath.Join(home, "Downloads")
		if _, err := os.Stat(downloadsDir); err == nil {
			return filepath.Join(downloadsDir, "TelegramDL")
		}
	}
	return filepath.Join(BaseDir, "descargas")
}

func DefaultConfig() Config {
	return Config{
		MaxConcurrentDownloads: 6,
		ParallelChunks:         true,
		ChunkWorkers:           4,
		DownloadFolder:         GetDefaultDownloadFolder(),
		ColorID:                nil,
		LoaderColorID:          nil,
		SpeedLimit: SpeedLimit{
			Value: 0,
			Unit:  "MB",
		},
		ListenerEnabled: true,
		ListenerChats:   []ListenerChat{},
		ListenerChatIDs: []int64{},
	}
}

func LoadEnvCredentials() (string, string) {
	InitPaths()
	// Cargar primero desde UserEnvPath
	_ = godotenv.Load(UserEnvPath)

	apiID := os.Getenv("TGDL_API_ID")
	if apiID == "" {
		apiID = os.Getenv("API_ID")
	}
	apiHash := os.Getenv("TGDL_API_HASH")
	if apiHash == "" {
		apiHash = os.Getenv("API_HASH")
	}

	if apiID == "" || apiHash == "" {
		// Intentar leer desde BaseDir/.env
		_ = godotenv.Load(filepath.Join(BaseDir, ".env"))
		if apiID == "" {
			apiID = os.Getenv("TGDL_API_ID")
			if apiID == "" {
				apiID = os.Getenv("API_ID")
			}
		}
		if apiHash == "" {
			apiHash = os.Getenv("TGDL_API_HASH")
			if apiHash == "" {
				apiHash = os.Getenv("API_HASH")
			}
		}
	}

	return strings.TrimSpace(apiID), strings.TrimSpace(apiHash)
}

func SaveEnvCredentials(apiID, apiHash string) error {
	InitPaths()
	apiID = strings.TrimSpace(apiID)
	apiHash = strings.TrimSpace(apiHash)

	// Se conserva lo que ya hubiera en el archivo (TGDL_BIND_HOST, TGDL_PORT,
	// comentarios...): antes se reescribía entero y cualquier ajuste del
	// usuario se perdía al volver a guardar las credenciales.
	existing := ""
	if data, rerr := os.ReadFile(UserEnvPath); rerr == nil {
		existing = string(data)
	}

	content := setEnvValue(existing, "API_ID", apiID)
	content = setEnvValue(content, "API_HASH", apiHash)
	content = setEnvValue(content, "TGDL_API_ID", apiID)
	content = setEnvValue(content, "TGDL_API_HASH", apiHash)

	err := os.WriteFile(UserEnvPath, []byte(content), 0600)
	if err != nil {
		return err
	}

	_ = os.Setenv("API_ID", apiID)
	_ = os.Setenv("API_HASH", apiHash)
	_ = os.Setenv("TGDL_API_ID", apiID)
	_ = os.Setenv("TGDL_API_HASH", apiHash)
	return nil
}

func GetServerPort() int {
	InitPaths()
	portStr := os.Getenv("TGDL_PORT")
	if portStr == "" {
		portStr = os.Getenv("PORT")
	}
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		return p
	}
	return 8000
}

func GetServerHost() string {
	InitPaths()
	host := strings.TrimSpace(os.Getenv("TGDL_BIND_HOST"))
	if host == "" {
		host = strings.TrimSpace(os.Getenv("BIND_HOST"))
	}
	if host == "" {
		// Mismo valor que se escribe en el .env al crearlo, para que el código y
		// el archivo nunca digan cosas distintas.
		return DefaultBindHost
	}
	return host
}

func NormalizeConfig(raw Config) Config {
	if raw.MaxConcurrentDownloads < 1 {
		raw.MaxConcurrentDownloads = 1
	} else if raw.MaxConcurrentDownloads > 32 {
		raw.MaxConcurrentDownloads = 32
	}

	if raw.ChunkWorkers < 1 {
		raw.ChunkWorkers = 1
	} else if raw.ChunkWorkers > 8 {
		raw.ChunkWorkers = 8
	}

	if strings.TrimSpace(raw.DownloadFolder) == "" {
		raw.DownloadFolder = GetDefaultDownloadFolder()
	}

	raw.SpeedLimit.Unit = strings.ToUpper(strings.TrimSpace(raw.SpeedLimit.Unit))
	if _, ok := SpeedMultipliers[raw.SpeedLimit.Unit]; !ok {
		raw.SpeedLimit.Unit = "MB"
	}
	if raw.SpeedLimit.Value < 0 {
		raw.SpeedLimit.Value = 0
	}

	chatIDs := make([]int64, 0, len(raw.ListenerChats))
	for i := range raw.ListenerChats {
		chatIDs = append(chatIDs, raw.ListenerChats[i].ID)
	}
	raw.ListenerChatIDs = chatIDs

	return raw
}

func FormatBytes(size float64) string {
	if size <= 0 {
		return "0 B"
	}
	// macOS utiliza base 1000 para todo (Finder, Disk Utility)
	base := 1024.0
	if runtime.GOOS == "darwin" {
		base = 1000.0
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for size >= base && i < len(units)-1 {
		size /= base
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%.0f B", size)
	}
	return fmt.Sprintf("%.2f %s", size, units[i])
}

// GenerateToken crea un secreto aleatorio criptográficamente seguro
// (32 bytes, codificado en hexadecimal) usado como token de acceso a la
// API HTTP local. Se genera una sola vez y se persiste; ver
// storage.GetAPIToken / storage.SaveAPIToken.
func GenerateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func ParseInt64(val any) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		res, _ := strconv.ParseInt(v, 10, 64)
		return res
	case json.Number:
		res, _ := v.Int64()
		return res
	default:
		return 0
	}
}
