//go:build !darwin

package updater

// En Windows y Linux la aplicación se distribuye como un ejecutable suelto, así
// que reemplazar el binario con selfupdate es lo correcto y no hay nada
// específico de plataforma que hacer. El caso especial es macOS, donde la
// aplicación es un bundle .app y hay que sustituirlo entero: eso vive en
// install_darwin.go.

// installPlatform devuelve siempre handled=false para que InstallUpdate siga
// por el camino genérico.
func (u *AppUpdater) installPlatform(_, _ string) (bool, error) {
	return false, nil
}

// cleanPlatform no tiene nada que limpiar fuera de macOS: de los archivos .old
// del ejecutable ya se encarga CleanOldVersion.
func (u *AppUpdater) cleanPlatform() {}
