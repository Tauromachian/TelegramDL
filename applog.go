package main

// Traduce los cambios de estado de los motores de descarga y de escucha a
// entradas legibles del registro (la vista "Logs" del panel). Va aparte de
// app.go para no mezclar el cableado de la aplicación con el del registro.

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"tgdown/pkg/i18n"
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
				i18n.T("listener.newFile", describeListenerItem(item)),
				i18n.T("listener.newFileDetail",
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
	origin := i18n.T("downloads.originManual")
	if item.Source == "listener" {
		origin = i18n.T("downloads.originListener")
	}
	context := i18n.T("downloads.context",
		a.chatLabel(item.ChatID), item.MessageID, origin)

	switch item.Status {
	case "queued":
		if previous == "" {
			logbus.Debug(logbus.CatDownloads, i18n.T("downloads.queued", name), context)
		} else {
			logbus.Debug(logbus.CatDownloads, i18n.T("downloads.requeued", name), context)
		}
	case "downloading":
		detail := context
		if item.TotalStr != "" {
			detail = i18n.T("downloads.contextSize", context, item.TotalStr)
		}
		logbus.Info(logbus.CatDownloads, i18n.T("downloads.started", name), detail)
	case "completed":
		detail := context
		if item.FilePath != "" {
			detail = i18n.T("downloads.contextSaved", context, item.FilePath)
		}
		logbus.Success(logbus.CatDownloads, i18n.T("downloads.completed", name), detail)
	case "failed":
		reason := strings.TrimSpace(item.Error)
		if reason == "" {
			reason = i18n.T("downloads.unknownReason")
		}
		logbus.Error(logbus.CatDownloads,
			i18n.T("downloads.failed", name, explainError(reason)),
			i18n.T("downloads.failedDetail", reason, context),
		)
	case "paused":
		logbus.Warn(logbus.CatDownloads, i18n.T("downloads.paused", name), context)
	case "cancelled":
		logbus.Warn(logbus.CatDownloads, i18n.T("downloads.cancelled", name), context)
	case "duplicate":
		logbus.Warn(logbus.CatDownloads,
			i18n.T("downloads.duplicate", name),
			i18n.T("downloads.duplicateDetail", context))
	}
}

// explainError traduce los fallos más habituales a una explicación corta y
// entendible, dejando el texto original en el detalle de la entrada.
//
// La detección mira el texto crudo del motor, que llega en inglés desde
// Telegram (flood, timeout, file_reference) o en español de errores propios
// antiguos (espacio, conexión, interrumpida); la salida va en el idioma del
// registro.
func explainError(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "no space") || strings.Contains(lower, "espacio"):
		return i18n.T("downloads.errNoSpace")
	case strings.Contains(lower, "mensaje no encontrado") || strings.Contains(lower, "message not found"):
		return i18n.T("downloads.errMsgGone")
	case strings.Contains(lower, "flood") || strings.Contains(lower, "too many requests"):
		return i18n.T("downloads.errFlood")
	case strings.Contains(lower, "auth") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "sesión"):
		return i18n.T("downloads.errAuth")
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return i18n.T("downloads.errTimeout")
	case strings.Contains(lower, "connection") || strings.Contains(lower, "conexión") || strings.Contains(lower, "eof"):
		return i18n.T("downloads.errConnLost")
	case strings.Contains(lower, "permission") || strings.Contains(lower, "denied") || strings.Contains(lower, "permiso"):
		return i18n.T("downloads.errPerm")
	case strings.Contains(lower, "file_reference"):
		return i18n.T("downloads.errFileRef")
	case strings.Contains(lower, "interrumpida"):
		return i18n.T("downloads.errInterrupted")
	default:
		return raw
	}
}

func describeDownload(item storage.DownloadItem) string {
	if name := strings.TrimSpace(item.FileName); name != "" {
		return name
	}
	if item.MessageID != 0 {
		return i18n.T("downloads.messageN", item.MessageID)
	}
	return item.ID
}

func describeListenerItem(item listener.ListenerItem) string {
	name := strings.TrimSpace(item.FileName)
	if name == "" {
		name = i18n.T("downloads.messageN", item.MessageID)
	}
	if item.TotalStr != "" {
		return fmt.Sprintf("%s (%s)", name, item.TotalStr)
	}
	return name
}
