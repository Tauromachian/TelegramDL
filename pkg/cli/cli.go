package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tgdown/pkg/config"
	"tgdown/pkg/downloader"
	"tgdown/pkg/i18n"
	"tgdown/pkg/storage"
	"tgdown/pkg/telegram"
)

// RunCLIMode ejecuta el modo CLI sin servidor HTTP
func RunCLIMode(url string) int {
	// 1. Inicializar rutas y configuración
	config.InitPaths()
	i18n.SetLanguage(i18n.DetectSystemLanguage())

	fmt.Println("🚀 TelegramDL - Modo CLI")
	fmt.Println("======================")

	// 2. Abrir base de datos
	dbPath := filepath.Join(config.DataDir, "tgdown.sqlite3")
	st, err := storage.NewStorage(dbPath)
	if err != nil {
		dbPath = filepath.Join(config.DataDir, "tgdown.db")
		st, err = storage.NewStorage(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error abriendo base de datos: %v\n", err)
			return 1
		}
	}
	defer st.Close()

	// 3. Migrar .env legacy si existe
	st.MigrateLegacyEnv()
	bindHost, bindPort := st.ServerBinding()
	config.SetServerBinding(bindHost, bindPort)

	// 4. Cargar configuración desde SQLite
	cfg, err := st.LoadConfig(config.DefaultConfig(), filepath.Join(config.BaseDir, "config.json"))
	if err != nil {
		cfg = config.DefaultConfig()
	}
	cfg = config.NormalizeConfig(cfg)

	// 5. Verificar credenciales de Telegram
	apiID, apiHash, _ := st.GetCredentials()
	if apiID == "" || apiHash == "" {
		fmt.Fprintf(os.Stderr, "Error: No hay credenciales de Telegram configuradas.\n")
		fmt.Fprintf(os.Stderr, "   Abre la aplicación una vez para configurarlas.\n")
		return 1
	}

	// 6. Inicializar cliente Telegram
	fmt.Println("📡 Conectando a Telegram...")
	cm := telegram.NewClientManager()
	if err := cm.InitClient(apiID, apiHash); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error iniciando cliente Telegram: %v\n", err)
		return 1
	}
	defer cm.Stop()

	// Esperar a que el cliente esté listo
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := cm.WaitReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error esperando conexión a Telegram: %v\n", err)
		return 1
	}
	fmt.Println("✅ Conectado a Telegram")

	// 7. Inicializar motor de descargas (SIN servidor)
	eng := downloader.NewEngine(cm, st, cfg)

	// 8. Registrar callback de progreso para CLI
	reporter := NewCLIProgressReporter()
	eng.OnStateChange(reporter.OnStateChange)

	// 9. Procesar la URL
	fmt.Printf("\n📥 Procesando URL: %s\n", url)
	parsed, err := downloader.ParseURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing URL: %v\n", err)
		return 1
	}

	// 10. Resolver username si es necesario
	chatID := parsed.ChatID
	if chatID == 0 && parsed.ChatUsername != "" {
		fmt.Printf("🔍 Resolviendo username: @%s\n", parsed.ChatUsername)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		resolvedID, err := cm.ResolveUsername(ctx, parsed.ChatUsername)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ No se pudo encontrar el canal o usuario: %v\n", err)
			return 1
		}
		chatID = resolvedID
		fmt.Printf("✅ Username resuelto a ID: %d\n", chatID)
	}

	// 11. Crear items para el rango de mensajes
	jobID := config.NewID()
	msgCount := parsed.EndMsgID - parsed.StartMsgID + 1
	fmt.Printf("📋 Creando %d tarea(s) de descarga...\n", msgCount)

	itemIDs := make([]string, 0, msgCount)
	for msgID := parsed.StartMsgID; msgID <= parsed.EndMsgID; msgID++ {
		item := storage.DownloadItem{
			ID:        config.NewID(),
			JobID:     jobID,
			MessageID: int64(msgID),
			ChatID:    chatID,
			Status:    "queued",
			Source:    "manual",
			FileName:  fmt.Sprintf("mensaje_%d", msgID),
		}
		eng.QueueItem(item)
		itemIDs = append(itemIDs, item.ID)
	}

	fmt.Printf("⏳ Descargas encoladas. Esperando finalización...\n\n")

	// 12. Esperar a que termine
	return waitForDownloads(eng, itemIDs)
}

// waitForDownloads espera a que todas las descargas terminen
func waitForDownloads(eng *downloader.Engine, ids []string) int {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	checkInterval := time.NewTicker(100 * time.Millisecond)
	defer checkInterval.Stop()

	for {
		select {
		case <-checkInterval.C:
			// Chequeo rápido para detectar cuando todo termina
		case <-ticker.C:
			// Reporte periódico
		}

		allDone := true
		anyFailed := false
		completedCount := 0
		failedCount := 0
		skippedCount := 0

		for _, id := range ids {
			item, ok := eng.GetItem(id)
			if !ok {
				continue
			}

			if item.Status == "downloading" || item.Status == "queued" || item.Status == "pending" {
				allDone = false
			}
			if item.Status == "failed" {
				anyFailed = true
				failedCount++
			}
			if item.Status == "completed" {
				completedCount++
			}
			if item.Status == "skipped" || item.Status == "duplicate" {
				skippedCount++
			}
		}

		if allDone {
			fmt.Println("\n")
			fmt.Println("======================")
			fmt.Println("📊 Resumen:")
			fmt.Printf("   ✅ Completadas: %d\n", completedCount)
			fmt.Printf("   ⏭️  Saltadas:   %d\n", skippedCount)
			if failedCount > 0 {
				fmt.Printf("   ❌ Fallidas:    %d\n", failedCount)
			}
			fmt.Println("======================")

			if anyFailed {
				fmt.Println("\n⚠️  Algunas descargas fallaron. Revisa el historial en la aplicación para más detalles.")
				return 1
			}
			fmt.Println("\n✨ Todas las descargas finalizaron correctamente.")
			return 0
		}
	}
}
