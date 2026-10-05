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

	fmt.Println(" TelegramDL - CLI mode")
	fmt.Println("======================")

	// 2. Abrir base de datos
	dbPath := filepath.Join(config.DataDir, "tgdown.sqlite3")
	st, err := storage.NewStorage(dbPath)
	if err != nil {
		dbPath = filepath.Join(config.DataDir, "tgdown.db")
		st, err = storage.NewStorage(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error opening database: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "Error: No Telegram credentials configured..\n")
		fmt.Fprintf(os.Stderr, "   Open the app once to configure them..\n")
		return 1
	}

	// 6. Inicializar cliente Telegram
	fmt.Println(" Connecting to Telegram...")
	cm := telegram.NewClientManager()
	if err := cm.InitClient(apiID, apiHash); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Telegram client: %v\n", err)
		return 1
	}
	defer cm.Stop()

	// Esperar a que el cliente esté listo
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := cm.WaitReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for connection to Telegram: %v\n", err)
		return 1
	}
	fmt.Println(" Connected to Telegram")

	// 7. Inicializar motor de descargas (SIN servidor)
	eng := downloader.NewEngine(cm, st, cfg)

	// 8. Registrar callback de progreso para CLI
	reporter := NewCLIProgressReporter()
	eng.OnStateChange(reporter.OnStateChange)

	// 9. Procesar la URL
	fmt.Printf("\n Processing URL: %s\n", url)
	parsed, err := downloader.ParseURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
		return 1
	}

	// 10. Resolver username si es necesario
	chatID := parsed.ChatID
	if chatID == 0 && parsed.ChatUsername != "" {
		fmt.Printf(" Resolving username: @%s\n", parsed.ChatUsername)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		resolvedID, err := cm.ResolveUsername(ctx, parsed.ChatUsername)
		if err != nil {
			fmt.Fprintf(os.Stderr, "The channel or user could not be found.: %v\n", err)
			return 1
		}
		chatID = resolvedID
		fmt.Printf("Username resolved to ID: %d\n", chatID)
	}

	// 11. Crear items para el rango de mensajes
	jobID := config.NewID()
	msgCount := parsed.EndMsgID - parsed.StartMsgID + 1
	fmt.Printf("Creating %d download task(s)...\n", msgCount)

	itemIDs := make([]string, 0, msgCount)
	for msgID := parsed.StartMsgID; msgID <= parsed.EndMsgID; msgID++ {
		item := storage.DownloadItem{
			ID:        config.NewID(),
			JobID:     jobID,
			MessageID: int64(msgID),
			ChatID:    chatID,
			Status:    "queued",
			Source:    "manual",
			FileName:  fmt.Sprintf("message_%d", msgID),
		}
		eng.QueueItem(item)
		itemIDs = append(itemIDs, item.ID)
	}

	fmt.Printf("Queued downloads. Waiting for completion....\n\n")

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
			fmt.Println("Summary:")
			fmt.Printf("   Completed: %d\n", completedCount)
			fmt.Printf("   Sautéed:   %d\n", skippedCount)
			if failedCount > 0 {
				fmt.Printf("   Failed:    %d\n", failedCount)
			}
			fmt.Println("======================")

			if anyFailed {
				fmt.Println("\n Some downloads failed. Check the history in the app for more details..")
				return 1
			}
			fmt.Println("\n All downloads completed successfully.")
			return 0
		}
	}
}
