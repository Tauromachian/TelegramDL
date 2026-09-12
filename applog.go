package main

// Traduce los cambios de estado de los motores de descarga y de escucha a
// entradas legibles del registro (la vista "Logs" del panel). Va aparte de
// app.go para no mezclar el cableado de la aplicación con el del registro.

import (
	"fmt"
	"strings"
	"sync"

	"tgdown/pkg/listener"
	"tgdown/pkg/logbus"
	"tgdown/pkg/storage"
)

// attachLogWatchers engancha los observadores que convierten transiciones de
// estado en líneas de registro. Solo se emite una línea cuando el estado
// cambia de verdad, no en cada actualización de progreso.
func (a *App) attachLogWatchers() {
	if a.downloader != nil {
		var mu sync.Mutex
		lastStatus := make(map[string]string)

		a.downloader.OnStateChange(func(item storage.DownloadItem) {
			mu.Lock()
			previous, seen := lastStatus[item.ID]
			if seen && previous == item.Status {
				mu.Unlock()
				return
			}
			// Evita que el mapa crezca sin límite en sesiones muy largas.
			if len(lastStatus) > 5000 {
				lastStatus = make(map[string]string)
			}
			lastStatus[item.ID] = item.Status
			mu.Unlock()

			logDownloadStatus(item, previous)
		})
	}

	if a.listener != nil {
		var mu sync.Mutex
		announced := make(map[string]bool)

		a.listener.OnStateChange(func(item listener.ListenerItem) {
			if item.ID == "" || item.Status != "available" {
				return
			}
			mu.Lock()
			if announced[item.ID] {
				mu.Unlock()
				return
			}
			if len(announced) > 5000 {
				announced = make(map[string]bool)
			}
			announced[item.ID] = true
			mu.Unlock()

			logbus.Info(logbus.CatListener,
				fmt.Sprintf("Nuevo archivo detectado: %s", describeListenerItem(item)),
				fmt.Sprintf("Chat: %s (%d) · Mensaje: %d · Pendiente de descargar",
					item.ChatName, item.ChatID, item.MessageID),
			)
		})
	}
}

// logDownloadStatus emite la línea correspondiente a una transición de estado.
func logDownloadStatus(item storage.DownloadItem, previous string) {
	name := describeDownload(item)
	origin := "manual"
	if item.Source == "listener" {
		origin = "escucha"
	}
	context := fmt.Sprintf("Chat: %d · Mensaje: %d · Origen: %s", item.ChatID, item.MessageID, origin)

	switch item.Status {
	case "queued":
		if previous == "" {
			logbus.Info(logbus.CatDownloads, "En cola: "+name, context)
		} else {
			logbus.Info(logbus.CatDownloads, "De vuelta en cola: "+name, context)
		}
	case "downloading":
		detail := context
		if item.TotalStr != "" {
			detail = fmt.Sprintf("Tamaño: %s · %s", item.TotalStr, context)
		}
		logbus.Info(logbus.CatDownloads, "Descargando: "+name, detail)
	case "completed":
		detail := context
		if item.FilePath != "" {
			detail = fmt.Sprintf("Guardado en: %s · %s", item.FilePath, context)
		}
		logbus.Success(logbus.CatDownloads, "Descarga completada: "+name, detail)
	case "failed":
		reason := strings.TrimSpace(item.Error)
		if reason == "" {
			reason = "motivo desconocido"
		}
		logbus.Error(logbus.CatDownloads,
			fmt.Sprintf("Falló la descarga de %s: %s", name, explainError(reason)),
			fmt.Sprintf("Detalle técnico: %s · %s", reason, context),
		)
	case "paused":
		logbus.Warn(logbus.CatDownloads, "Descarga pausada: "+name, context)
	case "cancelled":
		logbus.Warn(logbus.CatDownloads, "Descarga cancelada: "+name, context)
	case "duplicate":
		logbus.Warn(logbus.CatDownloads,
			"Omitido por duplicado: "+name,
			"Ya existe un archivo idéntico en la carpeta de descargas · "+context)
	}
}

// explainError traduce los fallos más habituales a una explicación corta y
// entendible, dejando el texto original en el detalle de la entrada.
func explainError(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "no space") || strings.Contains(lower, "espacio"):
		return "no queda espacio suficiente en el disco"
	case strings.Contains(lower, "mensaje no encontrado"):
		return "el mensaje ya no existe en Telegram (borrado o sin acceso)"
	case strings.Contains(lower, "flood") || strings.Contains(lower, "too many requests"):
		return "Telegram está limitando las peticiones (FLOOD_WAIT); hay que esperar"
	case strings.Contains(lower, "auth") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "sesión"):
		return "la sesión de Telegram no es válida; vuelve a iniciar sesión"
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return "se agotó el tiempo de espera de la conexión"
	case strings.Contains(lower, "connection") || strings.Contains(lower, "conexión") || strings.Contains(lower, "eof"):
		return "se perdió la conexión con Telegram"
	case strings.Contains(lower, "permission") || strings.Contains(lower, "denied") || strings.Contains(lower, "permiso"):
		return "sin permisos para escribir en la carpeta de descargas"
	case strings.Contains(lower, "file_reference"):
		return "la referencia del archivo caducó; hay que reintentar"
	case strings.Contains(lower, "interrumpida"):
		return "la descarga se interrumpió de forma inesperada"
	default:
		return raw
	}
}

func describeDownload(item storage.DownloadItem) string {
	if name := strings.TrimSpace(item.FileName); name != "" {
		return name
	}
	if item.MessageID != 0 {
		return fmt.Sprintf("mensaje %d", item.MessageID)
	}
	return item.ID
}

func describeListenerItem(item listener.ListenerItem) string {
	name := strings.TrimSpace(item.FileName)
	if name == "" {
		name = fmt.Sprintf("mensaje %d", item.MessageID)
	}
	if item.TotalStr != "" {
		return fmt.Sprintf("%s (%s)", name, item.TotalStr)
	}
	return name
}
