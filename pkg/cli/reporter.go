package cli

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"tgdown/pkg/storage"
)

// CLIProgressReporter reporta el progreso de descargas en la terminal
type CLIProgressReporter struct {
	mu         sync.Mutex
	lastUpdate map[string]time.Time
	activeIDs  map[string]bool
}

// NewCLIProgressReporter crea un nuevo reporter de progreso para CLI
func NewCLIProgressReporter() *CLIProgressReporter {
	return &CLIProgressReporter{
		lastUpdate: make(map[string]time.Time),
		activeIDs:  make(map[string]bool),
	}
}

// OnStateChange es llamado cuando cambia el estado de una descarga
func (r *CLIProgressReporter) OnStateChange(item storage.DownloadItem) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// No spamear: solo actualizar si pasó al menos 500ms
	if last, ok := r.lastUpdate[item.ID]; ok {
		if time.Since(last) < 500*time.Millisecond {
			return
		}
	}
	r.lastUpdate[item.ID] = time.Now()

	switch item.Status {
	case "queued":
		fmt.Printf("⏳ En cola: %s\n", truncate(item.FileName, 50))
	case "downloading":
		r.printProgress(item)
	case "completed":
		fmt.Printf("\n✅ Completado: %s (%s)\n", truncate(item.FileName, 50), item.TotalStr)
		delete(r.activeIDs, item.ID)
	case "failed":
		fmt.Printf("\n❌ Fallido: %s\n", truncate(item.FileName, 50))
		if item.Error != "" {
			fmt.Printf   (  "   Error: %s\n", item.Error)
		}
		delete(r.activeIDs, item.ID)
	case "cancelled":
		fmt.Printf("\n⏹️ Cancelado: %s\n", truncate(item.FileName, 50))
		delete(r.activeIDs, item.ID)
	case "skipped":
		fmt.Printf("⏭️  Saltado: %s (ya existe)\n", truncate(item.FileName, 50))
		delete(r.activeIDs, item.ID)
	case "duplicate":
		fmt.Printf("🔄 Duplicado: %s (ya existe)\n", truncate(item.FileName, 50))
		delete(r.activeIDs, item.ID)
	}

	r.activeIDs[item.ID] = true
}

// printProgress muestra la barra de progreso
func (r *CLIProgressReporter) printProgress(item storage.DownloadItem) {
	progress := item.Progress
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	// Barra de progreso simple
	barWidth := 30
	filled := int(progress * float64(barWidth) / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r⬇️ %s [%s] %.1f%% %s/s",
		truncate(item.FileName, 30),
		bar,
		progress,
		item.Speed,
	)

	if item.ETA != "" {
		fmt.Printf(" ETA: %s", item.ETA)
	}
}

// truncate acorta un string a un máximo de caracteres
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
