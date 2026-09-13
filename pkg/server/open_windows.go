//go:build windows

package server

// Abrir un archivo con la aplicación que le corresponda, en Windows.
//
// Este archivo existe por un problema de seguridad concreto. La primera versión
// usaba exec.Command("cmd", "/c", "start", "", ruta), y eso era una inyección de
// comandos: cmd.exe interpreta «&», «^» y «%», Go solo entrecomilla los
// argumentos que llevan espacios, y el nombre del archivo lo elige quien lo sube
// a Telegram. Un archivo llamado «video&calc.mp4» ejecutaba calc al pulsar
// «abrir».
//
// El siguiente intento fue rundll32.exe url.dll,FileProtocolHandler. Resolvía la
// inyección, pero rundll32 es una de las vías que más usa el malware para
// ejecutar cosas, así que los antivirus lo bloquean con frecuencia y el archivo
// simplemente no se abría.
//
// Lo que queda es la llamada de verdad: ShellExecuteW, que es exactamente lo que
// hace el Explorador al hacer doble clic. La ruta viaja como parámetro de la
// función en UTF-16, no como línea de comandos, así que no hay ningún intérprete
// de por medio y no hay nada que un «&» pueda romper.

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32.NewProc("ShellExecuteW")
)

// swShowNormal es SW_SHOWNORMAL. Se define aquí en vez de importarlo para no
// depender de ningún paquete externo: esto es solo la stdlib.
const swShowNormal = 1

func abrirEnElSistema(target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}

	if _, err := os.Stat(target); err != nil {
		// El archivo ya no está: se abre la carpeta que lo contenía, que es más
		// útil que no hacer nada.
		abrirConExplorer(filepath.Dir(target))
		return
	}

	if err := shellExecute("open", target, filepath.Dir(target)); err == nil {
		return
	} else {
		log.Printf("[SERVER] ShellExecuteW no pudo abrir %s: %v", target, err)
	}

	// Segundo intento. explorer.exe abre un archivo con su aplicación
	// predeterminada cuando se le pasa la ruta, y tampoco es un intérprete de
	// comandos, así que sigue sin haber inyección posible.
	abrirConExplorer(target)
}

func abrirConExplorer(ruta string) {
	if strings.TrimSpace(ruta) == "" {
		return
	}
	// explorer.exe devuelve código 1 incluso cuando funciona, así que su error
	// no dice nada y no se comprueba.
	_ = exec.Command("explorer.exe", ruta).Start()
}

// shellExecute llama a ShellExecuteW de shell32.dll.
//
// La función devuelve un valor mayor que 32 cuando ha conseguido lanzar algo;
// cualquier valor igual o menor es un código de error.
func shellExecute(verbo, archivo, directorio string) error {
	verboPtr, err := syscall.UTF16PtrFromString(verbo)
	if err != nil {
		return err
	}
	archivoPtr, err := syscall.UTF16PtrFromString(archivo)
	if err != nil {
		return err
	}
	directorioPtr, err := syscall.UTF16PtrFromString(directorio)
	if err != nil {
		return err
	}

	ret, _, errLlamada := procShellExecute.Call(
		0, // sin ventana padre
		uintptr(unsafe.Pointer(verboPtr)),
		uintptr(unsafe.Pointer(archivoPtr)),
		0, // sin argumentos: se abre un archivo, no se ejecuta un programa
		uintptr(unsafe.Pointer(directorioPtr)),
		uintptr(swShowNormal),
	)

	if ret <= 32 {
		return fmt.Errorf("código %d (%v)", ret, errLlamada)
	}
	return nil
}
