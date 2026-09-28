//go:build windows

package i18n

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// DetectSystemLanguage averigua el idioma de Windows leyendo del registro el
// locale del usuario ("es-ES", "en-US"...). Los procesos de escritorio no
// heredan LANG, así que la variable de entorno solo es un fallback (sirve para
// Cygwin/MSYS o un arranque forzado). Cualquier idioma que no sea español se
// traduce a inglés.
func DetectSystemLanguage() string {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\International`, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		if name, _, err := key.GetStringValue("LocaleName"); err == nil {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "es") {
				return LangES
			}
			return LangEN
		}
	}

	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := strings.ToLower(strings.TrimSpace(os.Getenv(name))); v != "" {
			if strings.HasPrefix(v, "es") {
				return LangES
			}
			return LangEN
		}
	}
	return LangEN
}
