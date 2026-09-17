package server

// Endpoints del registro de actividad (vista "Logs" del panel) y traducción de
// los cambios de configuración a entradas legibles del mismo registro.

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"tgdown/pkg/config"
	"tgdown/pkg/logbus"
)

// handleLogs devuelve las entradas del registro. Con ?since=<id> solo devuelve
// lo ocurrido después de ese ID, que es como el panel se mantiene al día sin
// recargar todo el historial en cada refresco.
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var since int64 = -1
	if raw := strings.TrimSpace(r.URL.Query().Get("since")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			since = parsed
		}
	}

	limit := 500
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	entries, lastID := logbus.Since(since, limit)
	s.jsonResponse(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"entries": entries,
		"last_id": lastID,
		"file":    logbus.FilePath(),
	})
}

// handleLogsClear vacía el registro en memoria (el archivo en disco se
// conserva: es el histórico de la sesión).
func (s *Server) handleLogsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	logbus.Clear()
	s.jsonResponse(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"last_id": logbus.LastID(),
	})
}

// handleLogsExport entrega el registro completo en memoria como texto plano.
func (s *Server) handleLogsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	name := fmt.Sprintf("telegramdl-%s.log", time.Now().Format("2006-01-02_15-04-05"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(logbus.ExportText()))
}

// ---------------------------------------------------------------------------
// Registro de cambios en los ajustes
// ---------------------------------------------------------------------------

// logConfigChanges compara la configuración anterior con la nueva y deja una
// línea por cada cosa que el usuario haya cambiado realmente.
func logConfigChanges(before, after config.Config) {
	changes := make([]string, 0, 8)

	if before.MaxConcurrentDownloads != after.MaxConcurrentDownloads {
		changes = append(changes, fmt.Sprintf("Descargas simultáneas: %d → %d",
			before.MaxConcurrentDownloads, after.MaxConcurrentDownloads))
	}
	if before.ParallelChunks != after.ParallelChunks {
		changes = append(changes, fmt.Sprintf("Descarga por fragmentos: %s → %s",
			onOff(before.ParallelChunks), onOff(after.ParallelChunks)))
	}
	if before.ChunkWorkers != after.ChunkWorkers {
		changes = append(changes, fmt.Sprintf("Hilos por archivo: %d → %d",
			before.ChunkWorkers, after.ChunkWorkers))
	}
	if before.DownloadFolder != after.DownloadFolder {
		changes = append(changes, fmt.Sprintf("Carpeta de descargas: %s → %s",
			before.DownloadFolder, after.DownloadFolder))
	}
	if before.OrganizeByChat != after.OrganizeByChat {
		changes = append(changes, fmt.Sprintf("Carpeta por chat en la escucha: %s → %s",
			onOff(before.OrganizeByChat), onOff(after.OrganizeByChat)))
	}
	if before.SpeedLimit != after.SpeedLimit {
		changes = append(changes, fmt.Sprintf("Límite de velocidad: %s → %s",
			describeSpeedLimit(before.SpeedLimit), describeSpeedLimit(after.SpeedLimit)))
	}
	if !sameIntPtr(before.ColorID, after.ColorID) {
		changes = append(changes, "Tema de color cambiado")
	}
	if !sameIntPtr(before.LoaderColorID, after.LoaderColorID) {
		changes = append(changes, "Color del indicador de carga cambiado")
	}

	if len(changes) > 0 {
		logbus.Info(logbus.CatSettings, "Ajustes guardados", strings.Join(changes, " · "))
	}

	logListenerConfigChanges(before, after)
}

// logListenerConfigChanges detalla lo que cambió en la escucha: si se activó o
// se pausó, qué chats se añadieron o quitaron y qué filtros se tocaron.
func logListenerConfigChanges(before, after config.Config) {
	if before.ListenerEnabled != after.ListenerEnabled {
		if after.ListenerEnabled {
			logbus.Success(logbus.CatListener, "Escucha activada", "")
		} else {
			logbus.Warn(logbus.CatListener, "Escucha pausada", "")
		}
	}

	// Las entradas se indexan por (chat, tema): un mismo grupo puede aparecer
	// varias veces, una por cada tema vigilado, y cada una se registra aparte.
	oldChats := make(map[string]config.ListenerChat, len(before.ListenerChats))
	for _, c := range before.ListenerChats {
		oldChats[c.Key()] = c
	}
	newChats := make(map[string]config.ListenerChat, len(after.ListenerChats))
	for _, c := range after.ListenerChats {
		newChats[c.Key()] = c
	}

	// Chats añadidos y modificados, en el orden en que están configurados.
	for _, chat := range after.ListenerChats {
		previous, existed := oldChats[chat.Key()]
		if !existed {
			logbus.Success(logbus.CatListener,
				fmt.Sprintf("Chat añadido a la escucha: %s", describeChat(chat)),
				fmt.Sprintf("Descarga automática: %s · Filtros: %s",
					onOff(chat.AutoDownload), describeFilters(chat)))
			continue
		}

		details := make([]string, 0, 3)
		if previous.AutoDownload != chat.AutoDownload {
			details = append(details, fmt.Sprintf("Descarga automática: %s → %s",
				onOff(previous.AutoDownload), onOff(chat.AutoDownload)))
		}
		if describeFilters(previous) != describeFilters(chat) {
			details = append(details, fmt.Sprintf("Filtros: %s → %s",
				describeFilters(previous), describeFilters(chat)))
		}
		if previous.Name != chat.Name && strings.TrimSpace(chat.Name) != "" {
			details = append(details, fmt.Sprintf("Nombre: %s → %s", previous.Name, chat.Name))
		}
		if previous.TopicName != chat.TopicName && strings.TrimSpace(chat.TopicName) != "" {
			details = append(details, fmt.Sprintf("Tema: %s", chat.TopicName))
		}
		if len(details) > 0 {
			logbus.Info(logbus.CatListener,
				fmt.Sprintf("Chat actualizado: %s", describeChat(chat)),
				strings.Join(details, " · "))
		}
	}

	// Chats eliminados, en orden estable por ID y tema para que el log sea
	// reproducible.
	removed := make([]config.ListenerChat, 0)
	for key, chat := range oldChats {
		if _, still := newChats[key]; !still {
			removed = append(removed, chat)
		}
	}
	sort.Slice(removed, func(i, j int) bool {
		if removed[i].ID != removed[j].ID {
			return removed[i].ID < removed[j].ID
		}
		return removed[i].Topic() < removed[j].Topic()
	})
	for _, chat := range removed {
		logbus.Warn(logbus.CatListener,
			fmt.Sprintf("Chat eliminado de la escucha: %s", describeChat(chat)), "")
	}
}

// describeChat identifica una entrada de escucha en el registro. Cuando está
// acotada a un tema, se nombra primero el tema y después el grupo, igual que en
// el panel, para saber de dónde sale cada archivo.
func describeChat(chat config.ListenerChat) string {
	name := strings.TrimSpace(chat.Name)
	if name == "" || name == strconv.FormatInt(chat.ID, 10) {
		if topic := chat.TopicLabel(); topic != "" {
			return fmt.Sprintf("%s · %d", topic, chat.ID)
		}
		return strconv.FormatInt(chat.ID, 10)
	}
	if topic := chat.TopicLabel(); topic != "" {
		return fmt.Sprintf("%s · %s (%d, tema %d)", topic, name, chat.ID, chat.Topic())
	}
	return fmt.Sprintf("%s (%d)", name, chat.ID)
}

func describeFilters(chat config.ListenerChat) string {
	active := make([]string, 0, 5)
	if chat.FPhotos {
		active = append(active, "fotos")
	}
	if chat.FVideos {
		active = append(active, "vídeos")
	}
	if chat.FAudios {
		active = append(active, "audios")
	}
	if chat.FDocs {
		active = append(active, "documentos")
	}
	if chat.FStickers {
		active = append(active, "stickers")
	}
	if len(active) == 0 {
		return "ninguno"
	}
	return strings.Join(active, ", ")
}

func describeSpeedLimit(limit config.SpeedLimit) string {
	if limit.Value <= 0 {
		return "sin límite"
	}
	return fmt.Sprintf("%g %s/s", limit.Value, limit.Unit)
}

func onOff(v bool) string {
	if v {
		return "activada"
	}
	return "desactivada"
}

func sameIntPtr(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
