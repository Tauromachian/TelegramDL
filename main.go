package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"tgdown/pkg/config"
	"tgdown/pkg/i18n"
	"tgdown/pkg/updater"
)

//go:embed dashboard/dist
var assets embed.FS

func main() {
	// Antes que nada, el registro y la consola hablan el idioma del sistema:
	// el elegido en Ajustes se cargará de la base de datos en NewApp y lo pisará.
	i18n.SetLanguage(i18n.DetectSystemLanguage())

	if len(os.Args) > 1 {
		setupConsole()
	}

	if hasArgument("--help") || hasArgument("-h") {
		printUsage()
		return
	}
	if hasArgument("--server") {
		os.Exit(runServerMode())
	}
	if hasArgument("--update") {
		os.Exit(runUpdateMode())
	}

	runDesktopMode()
}

func hasArgument(target string) bool {
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, target) {
			return true
		}
		if strings.HasPrefix(target, "--") {
			if strings.EqualFold(arg, target[1:]) || strings.EqualFold(arg, "/"+target[2:]) {
				return true
			}
		}
	}
	return false
}

func printUsage() {
	fmt.Println("TelegramDL")
	fmt.Println(i18n.T("cli.usageHeader"))
	fmt.Println(i18n.T("cli.usageDesktop"))
	fmt.Println(i18n.T("cli.usageServer"))
	fmt.Println(i18n.T("cli.usageUpdate"))
}

// waitInstanceLock reintenta unos segundos antes de rendirse: tras una
// actualización, el proceso anterior lanza el nuevo ejecutable y tarda un
// instante en liberar el bloqueo al salir.
func waitInstanceLock(timeout time.Duration) (func(), bool) {
	deadline := time.Now().Add(timeout)
	for {
		release, ok := tryAcquireInstanceLock()
		if ok || time.Now().After(deadline) {
			return release, ok
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func runDesktopMode() {
	release, ok := waitInstanceLock(8 * time.Second)
	if !ok {
		notifyAlreadyRunning()
		return
	}
	defer release()

	app := NewApp(assets)
	if app == nil {
		// NewApp solo devuelve nil cuando no consigue abrir ninguna base de
		// datos. Abierta con doble clic no hay consola donde leer el motivo, así
		// que el aviso tiene que ser una ventana del sistema.
		mostrarAviso(i18n.T("cli.dbOpenAviso", config.DataDir))
		return
	}

	err := wails.Run(&options.App{
		Title:     "Telegram DL",
		Width:     1280,
		Height:    720,
		MinWidth:  950,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.Handler(),
		},
		BackgroundColour: &options.RGBA{R: 7, G: 17, B: 31, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.None,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runServerMode() int {
	release, ok := waitInstanceLock(8 * time.Second)
	if !ok {
		fmt.Fprintln(os.Stderr, i18n.T("cli.alreadyRunning"))
		return 1
	}
	defer release()

	app := NewApp(assets)
	if app == nil {
		fmt.Fprintf(os.Stderr, i18n.T("cli.dbOpenError"), config.DataDir)
		return 1
	}
	if app.server == nil {
		fmt.Fprintln(os.Stderr, i18n.T("cli.serverInitError"))
		return 1
	}

	fmt.Printf(i18n.T("cli.serverStarted"), config.GetServerHost(), config.GetServerPort())
	fmt.Println(i18n.T("cli.serverModeNote"))

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)

	// Manejo de consola multiplataforma
	registerConsoleCtrlHandler(sigCh)

	// Esperar la señal
	sig := <-sigCh
	fmt.Printf("\n%s\n", i18n.T("cli.closing", sig))

	// Watchdog de seguridad: no más de 5 segundos para cerrar
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println(i18n.T("cli.forceClose"))
		os.Exit(1)
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	app.shutdown(shutdownCtx)

	fmt.Println(i18n.T("cli.stopped"))
	_ = os.Stdout.Sync()
	os.Exit(0)
	return 0
}

func runUpdateMode() int {
	appUpdater := updater.NewAppUpdater()
	release, asset, err := appUpdater.CheckForUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("cli.checkError"), err)
		return 1
	}
	if release == nil || asset == nil {
		fmt.Printf(i18n.T("cli.alreadyUpdated"), config.AppVersion)
		return 0
	}

	fmt.Printf(i18n.T("cli.updating"), config.AppVersion, release.TagName, asset.Name)
	if err := appUpdater.InstallUpdate(release); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("cli.installError"), err)
		return 1
	}

	lastStatus := ""
	for {
		progress := appUpdater.GetProgress()
		if progress.Status != lastStatus {
			if strings.HasPrefix(progress.Status, "error:") {
				fmt.Fprintln(os.Stderr, progress.Status)
				return 1
			}
			if progress.Status != "" {
				fmt.Printf(i18n.T("cli.progressState"), progress.Status, progress.Percentage)
			}
			lastStatus = progress.Status
		}
		time.Sleep(500 * time.Millisecond)
	}
}
