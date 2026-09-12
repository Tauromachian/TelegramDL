package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	goupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/minio/selfupdate"
	"tgdown/pkg/config"
	"tgdown/pkg/logbus"
)

type Progress struct {
	Status     string `json:"status"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Percentage int    `json:"percentage"`
}

type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
	release *goupdate.Release
}

type AppUpdater struct {
	currentVersion  string
	repoURL         string
	mu              sync.RWMutex
	progress        Progress
	postponedUpdate string

	// Estado para no repetir en el registro lo mismo cada 5 minutos: la
	// comprobación periódica solo se narra la primera vez (al abrir la app) y
	// cuando el resultado cambia de verdad.
	checkCount       int
	announcedVersion string
	lastCheckError   string

	// rateLimitUntil marca hasta cuándo GitHub nos tiene cortados por exceso de
	// peticiones. Mientras no pase esa hora no tiene sentido volver a preguntar.
	rateLimitUntil time.Time
}

func NewAppUpdater() *AppUpdater {
	config.InitPaths()
	u := &AppUpdater{
		currentVersion: config.AppVersion,
		repoURL:        config.GithubRepo,
		progress: Progress{
			Status: "idle",
		},
	}
	u.CleanOldVersion()
	return u
}

// CleanOldVersion busca y elimina archivos temporales de actualizaciones previas (.old)
func (u *AppUpdater) CleanOldVersion() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	dir := filepath.Dir(exePath)
	base := filepath.Base(exePath)

	// Intentar detectar archivos .old (estándar de selfupdate y variantes con punto inicial)
	targets := []string{
		exePath + ".old",
		filepath.Join(dir, "."+base+".old"),
	}

	for _, target := range targets {
		if _, err := os.Stat(target); err == nil {
			log.Printf("[UPDATER] Detectado archivo de versión antigua: %s. Eliminando...", target)
			// En Windows, intentamos eliminarlo. os.Remove funciona incluso con atributo oculto.
			if err := os.Remove(target); err != nil {
				log.Printf("[UPDATER] No se pudo eliminar la versión antigua %s: %v", target, err)
			} else {
				log.Printf("[UPDATER] Versión antigua eliminada: %s", target)
			}
		}
	}
}

func (u *AppUpdater) GetProgress() Progress {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.progress
}

func (u *AppUpdater) setProgress(status string, downloaded, total int64, pct int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.progress = Progress{
		Status:     status,
		Downloaded: downloaded,
		Total:      total,
		Percentage: pct,
	}
}

// Postpone marca una versión específica para no volver a notificar en esta sesión
func (u *AppUpdater) Postpone(version string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.postponedUpdate = version
}

// IsPostponed verifica si una versión ya fue pospuesta
func (u *AppUpdater) IsPostponed(version string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.postponedUpdate == version
}

// routineLog registra un paso rutinario de la comprobación de actualizaciones:
// visible la primera vez (al abrir la aplicación) y en nivel detalle en las
// comprobaciones periódicas posteriores, que se repiten cada pocos minutos.
func (u *AppUpdater) routineLog(first bool, message, detail string) {
	if first {
		logbus.Info(logbus.CatUpdater, message, detail)
	} else {
		logbus.Debug(logbus.CatUpdater, message, detail)
	}
}

var (
	// GitHub añade al error una cuenta atrás ("[rate reset in 23m20s]") que
	// cambia a cada segundo. Hay que quitarla para comparar errores, o el mismo
	// fallo parecería distinto cada vez y se avisaría una y otra vez.
	cuotaCuentaAtras = regexp.MustCompile(`\[rate reset in [^\]]*\]`)
	cuotaEspera      = regexp.MustCompile(`rate reset in ([0-9hms.]+)`)
)

// checkErrorLog evita repetir el mismo fallo: avisa la primera vez y cada vez
// que el error cambia de verdad, y el resto queda en nivel detalle.
func (u *AppUpdater) checkErrorLog(first bool, message string, err error) {
	text := ""
	if err != nil {
		text = err.Error()
	}
	clave := cuotaCuentaAtras.ReplaceAllString(text, "[rate reset]")

	u.mu.Lock()
	changed := u.lastCheckError != clave
	u.lastCheckError = clave
	u.mu.Unlock()

	if first || changed {
		logbus.Warn(logbus.CatUpdater, message, text)
		return
	}
	logbus.Debug(logbus.CatUpdater, message, text)
}

// anotarLimiteDeCuota detecta si GitHub rechazó la petición por haber agotado
// las peticiones por hora y, en ese caso, apunta hasta cuándo no merece la pena
// volver a preguntar. GitHub dice en el propio error cuánto falta; si no se
// puede leer, se espera un cuarto de hora.
func (u *AppUpdater) anotarLimiteDeCuota(err error) bool {
	if err == nil {
		return false
	}
	texto := err.Error()
	if !strings.Contains(strings.ToLower(texto), "rate limit") {
		return false
	}

	espera := 15 * time.Minute
	if m := cuotaEspera.FindStringSubmatch(texto); m != nil {
		if d, perr := time.ParseDuration(m[1]); perr == nil && d > 0 {
			espera = d + 30*time.Second // un margen por si el reloj va justo
		}
	}

	u.mu.Lock()
	u.rateLimitUntil = time.Now().Add(espera)
	u.mu.Unlock()

	logbus.Warn(logbus.CatUpdater,
		"GitHub limitó las comprobaciones de actualización",
		fmt.Sprintf("Se alcanzó el máximo de peticiones por hora; no se volverá a preguntar hasta dentro de %s",
			espera.Round(time.Second)))
	return true
}

func (u *AppUpdater) CheckForUpdate() (*ReleaseInfo, *ReleaseAsset, error) {
	u.mu.Lock()
	u.checkCount++
	first := u.checkCount == 1
	pausaHasta := u.rateLimitUntil
	u.mu.Unlock()

	// Si GitHub ya nos dijo que agotamos la cuota, esperamos a que pase el
	// tiempo que él mismo indicó en vez de seguir insistiendo (y llenando el
	// registro) cada pocos minutos.
	if !pausaHasta.IsZero() && time.Now().Before(pausaHasta) {
		logbus.Debug(logbus.CatUpdater, "Comprobación de actualizaciones en pausa",
			fmt.Sprintf("GitHub limitó las peticiones; se reintentará en %s",
				time.Until(pausaHasta).Round(time.Second)))
		return nil, nil, nil
	}

	u.routineLog(first, fmt.Sprintf("Comprobando actualizaciones de %s", u.repoURL),
		fmt.Sprintf("Versión instalada: v%s", u.currentVersion))

	// 1. Intentar detección automática primero
	updater, err := goupdate.NewUpdater(goupdate.Config{})
	if err != nil {
		u.checkErrorLog(first, "No se pudo preparar el buscador de actualizaciones", err)
		return nil, nil, err
	}

	latest, found, err := updater.DetectLatest(context.Background(), goupdate.ParseSlug(u.repoURL))
	if err != nil {
		if !u.anotarLimiteDeCuota(err) {
			u.checkErrorLog(first, "No se pudo consultar la última versión publicada", err)
		}
		return nil, nil, err
	}

	// 2. Validar si el asset detectado es para nuestra plataforma
	if found {
		assetName := strings.ToLower(latest.AssetName)
		isWin := strings.Contains(assetName, "win") || strings.Contains(assetName, "exe")
		isLin := strings.Contains(assetName, "linux")
		isMac := strings.Contains(assetName, "darwin") || strings.Contains(assetName, "mac")

		currentOS := runtime.GOOS
		if (currentOS == "windows" && !isWin) || (currentOS == "linux" && !isLin) || (currentOS == "darwin" && !isMac) {
			logbus.Debug(logbus.CatUpdater,
				fmt.Sprintf("El archivo detectado (%s) no parece de %s; se reintenta con filtros", latest.AssetName, currentOS), "")
			found = false
		}
	}

	// 3. Si no se encontró o no era válido, forzar con filtros específicos
	if !found {
		var filters []string
		switch runtime.GOOS {
		case "windows":
			filters = []string{"win"}
		case "darwin":
			filters = []string{"mac"}
		case "linux":
			filters = []string{"linux"}
		}

		updater, _ = goupdate.NewUpdater(goupdate.Config{
			Filters: filters,
		})
		latest, found, err = updater.DetectLatest(context.Background(), goupdate.ParseSlug(u.repoURL))
		if err != nil {
			// Este segundo intento también gasta cuota, así que su fallo se
			// trata igual que el primero en vez de pasar por "no hay versión
			// compatible", que es lo que parecía antes.
			if !u.anotarLimiteDeCuota(err) {
				u.checkErrorLog(first, "No se pudo consultar la última versión publicada", err)
			}
			return nil, nil, err
		}
	}

	if !found || latest == nil {
		u.routineLog(first, "No se encontró ninguna versión compatible con este sistema", "")
		return nil, nil, nil
	}

	u.routineLog(first, fmt.Sprintf("Última versión publicada: %s", latest.Version()),
		"Archivo: "+latest.AssetName)

	// Comparar versiones (normalizando)
	currV := strings.TrimPrefix(u.currentVersion, "v")
	if latest.LessOrEqual(currV) {
		u.routineLog(first, fmt.Sprintf("TelegramDL está al día (v%s)", u.currentVersion), "")
		return nil, nil, nil
	}

	// La novedad se anuncia una sola vez por versión: la comprobación se repite
	// cada pocos minutos y no tiene sentido repetir el aviso en cada ronda.
	u.mu.Lock()
	alreadyAnnounced := u.announcedVersion == latest.Version()
	u.announcedVersion = latest.Version()
	u.mu.Unlock()

	if alreadyAnnounced {
		logbus.Debug(logbus.CatUpdater,
			fmt.Sprintf("Sigue disponible la versión %s", latest.Version()), "")
	} else {
		logbus.Success(logbus.CatUpdater,
			fmt.Sprintf("¡Nueva versión disponible: %s!", latest.Version()),
			fmt.Sprintf("Tienes la v%s · Puedes actualizar desde Ajustes", u.currentVersion))
	}

	rel := &ReleaseInfo{
		TagName: latest.Version(),
		Body:    latest.ReleaseNotes,
		release: latest,
	}
	asset := &ReleaseAsset{
		Name:        latest.AssetName,
		DownloadURL: latest.AssetURL,
		Size:        int64(latest.AssetByteSize),
	}
	rel.Assets = []ReleaseAsset{*asset}

	return rel, asset, nil
}

func (u *AppUpdater) InstallUpdate(rel *ReleaseInfo) error {
	if rel.release == nil {
		return fmt.Errorf("información de actualización no válida")
	}

	go func() {
		u.setProgress("starting", 0, 0, 0)

		tempDir, err := os.MkdirTemp("", "tgdown_update")
		if err != nil {
			u.setProgress("error: no se pudo crear directorio temporal", 0, 0, 0)
			return
		}
		defer os.RemoveAll(tempDir)

		// 1. Descarga con progreso manual
		archivePath := filepath.Join(tempDir, rel.release.AssetName)
		u.setProgress("downloading", 0, int64(rel.release.AssetByteSize), 0)

		if err := u.downloadWithProgress(rel.release.AssetURL, archivePath); err != nil {
			u.setProgress("error: descarga fallida: "+err.Error(), 0, 0, 0)
			return
		}

		// 2. Extracción
		u.setProgress("extracting", 0, 0, 100)
		extractPath := filepath.Join(tempDir, "extracted")
		_ = os.MkdirAll(extractPath, 0755)

		lowerArch := strings.ToLower(archivePath)
		if strings.HasSuffix(lowerArch, ".zip") {
			if err := unzip(archivePath, extractPath); err != nil {
				u.setProgress("error: fallo al descomprimir zip: "+err.Error(), 0, 0, 0)
				return
			}
		} else if strings.HasSuffix(lowerArch, ".tar.gz") || strings.HasSuffix(lowerArch, ".tgz") {
			if err := untarGz(archivePath, extractPath); err != nil {
				u.setProgress("error: fallo al descomprimir tar.gz: "+err.Error(), 0, 0, 0)
				return
			}
		} else {
			// Si no es un archivo comprimido, lo tratamos como el binario directo
			extractPath = archivePath
		}

		// 3. Buscar el binario ejecutable real dentro de lo extraído
		newBinaryPath := u.findExecutable(extractPath)
		if newBinaryPath == "" {
			u.setProgress("error: no se encontró el ejecutable en la descarga", 0, 0, 0)
			return
		}

		// 4. Aplicar actualización atómica (reemplazo seguro)
		u.setProgress("finishing", 0, 0, 100)
		exePath, err := os.Executable()
		if err != nil {
			u.setProgress("error: no se pudo obtener ruta del ejecutable", 0, 0, 0)
			return
		}

		newBinaryFile, err := os.Open(newBinaryPath)
		if err != nil {
			u.setProgress("error: no se pudo abrir nuevo binario", 0, 0, 0)
			return
		}
		defer newBinaryFile.Close()

		log.Printf("[UPDATER] Aplicando actualización sobre: %s", exePath)
		err = selfupdate.Apply(newBinaryFile, selfupdate.Options{
			TargetPath: exePath,
		})
		if err != nil {
			u.setProgress("error: fallo al aplicar actualización: "+err.Error(), 0, 0, 0)
			return
		}

		u.setProgress("finishing", 0, 0, 100)
		time.Sleep(2 * time.Second)

		u.restartApp(exePath)
	}()

	return nil
}

func (u *AppUpdater) downloadWithProgress(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub devolvió código %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64
	buffer := make([]byte, 64*1024)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, werr := out.Write(buffer[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			pct := 0
			if total > 0 {
				pct = int(float64(downloaded) / float64(total) * 100) // 0% a 100%
			}
			u.setProgress("downloading", downloaded, total, pct)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *AppUpdater) findExecutable(path string) string {
	fi, err := os.Stat(path)
	if err == nil && !fi.IsDir() {
		return path
	}

	var found string
	var priority int = -1

	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := strings.ToLower(info.Name())
		ext := strings.ToLower(filepath.Ext(name))

		// En Windows solo aceptamos .exe
		// En otros sistemas aceptamos sin extensión o cualquier cosa que parezca el binario
		isWindows := runtime.GOOS == "windows"
		if isWindows && ext != ".exe" {
			return nil
		}

		currPriority := -1
		// Nombres exactos tienen prioridad máxima
		if name == "telegramdl.exe" || name == "tgdown.exe" || name == "telegramdl" || name == "tgdown" {
			currPriority = 100
		} else if strings.Contains(name, "telegramdl") || strings.Contains(name, "tgdown") {
			currPriority = 50
		} else if !isWindows && ext == "" {
			// En Linux/Mac, un archivo sin extensión podría ser el binario
			currPriority = 10
		}

		if currPriority > priority {
			priority = currPriority
			found = p
		}

		if priority == 100 {
			return io.EOF // Encontrado el mejor match posible
		}
		return nil
	})
	return found
}

func (u *AppUpdater) restartApp(exePath string) {
	// Usamos la ruta original donde instalamos el nuevo binario
	// En lugar de os.Executable() que podría devolver la ruta del archivo .old en Windows

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// En Windows, ejecutar el binario directamente suele funcionar si no heredamos nada
		cmd = exec.Command(exePath, os.Args[1:]...)
	} else {
		cmd = exec.Command(exePath, os.Args[1:]...)
	}

	log.Printf("[UPDATER] Lanzando nueva versión: %s", exePath)
	err := cmd.Start()
	if err != nil {
		log.Printf("[UPDATER] Error crítico al reiniciar aplicación: %v", err)
	}

	os.Exit(0)
}

// Funciones auxiliares de extracción
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func untarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(outFile, tr)
			outFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
