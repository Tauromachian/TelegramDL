package main

// Traduce los cambios de estado de los motores de descarga y de escucha a
// entradas legibles del registro (la vista "Logs" del panel). Va aparte de
// app.go para no mezclar el cableado de la aplicación con el del registro.

import (
	"fmt"
	"strconv"
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

			a.logDownloadStatus(item, previous)
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
				fmt.Sprintf("Chat: %s · Mensaje: %d · Pendiente de descargar",
					a.chatLabel(item.ChatID, item.ChatName), item.MessageID),
			)
		})
	}
}

// chatLabel devuelve el nombre legible de un chat. Prueba primero el nombre que
// venga con el propio elemento, luego la caché de títulos de Telegram, y solo
// como último recurso el ID numérico.
func (a *App) chatLabel(chatID int64, hints ...string) string {
	for _, hint := range hints {
		hint = strings.TrimSpace(hint)
		if hint != "" && hint != strconv.FormatInt(chatID, 10) {
			return hint
		}
	}
	if a.clientMgr != nil {
		if name := strings.TrimSpace(a.clientMgr.GetChatName(chatID)); name != "" {
			return name
		}
	}
	if a.listener != nil {
		if name := strings.TrimSpace(a.listener.ChatName(chatID)); name != "" {
			return name
		}
	}
	return strconv.FormatInt(chatID, 10)
}

// logDownloadStatus emite la línea correspondiente a una transición de estado.
//
// Solo dos estados son visibles por defecto en cada descarga —iniciada y
// terminada—; el resto de transiciones internas se quedan en nivel detalle para
// no llenar el registro.
func (a *App) logDownloadStatus(item storage.DownloadItem, previous string) {
	name := describeDownload(item)
	origin := "manual"
	if item.Source == "listener" {
		origin = "escucha"
	}
	context := fmt.Sprintf("Chat: %s · Mensaje: %d · Origen: %s",
		a.chatLabel(item.ChatID), item.MessageID, origin)

	switch item.Status {
	case "queued":
		if previous == "" {
			logbus.Debug(logbus.CatDownloads, "En cola: "+name, context)
		} else {
			logbus.Debug(logbus.CatDownloads, "De vuelta en cola: "+name, context)
		}
	case "downloading":
		detail := context
		if item.TotalStr != "" {
			detail = fmt.Sprintf("%s · Tamaño: %s", context, item.TotalStr)
		}
		logbus.Info(logbus.CatDownloads, "Descarga iniciada: "+name, detail)
	case "completed":
		detail := context
		if item.FilePath != "" {
			detail = fmt.Sprintf("%s · Guardado en: %s", context, item.FilePath)
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
