package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsPublicAPIPath(t *testing.T) {
	// El envío del token tiene que responder sin token: quien pulsa el botón es
	// justo quien no lo tiene.
	for _, path := range []string{publicAPIPath, tokenSendAPIPath} {
		if !isPublicAPIPath(path) {
			t.Fatalf("%s debería ser público", path)
		}
	}
	// Cualquier otra cosa sigue exigiendo token.
	for _, path := range []string{"/api/settings", "/api/auth/token", "/api/auth/token/regenerate", "/api/downloads"} {
		if isPublicAPIPath(path) {
			t.Fatalf("%s NO debería ser público", path)
		}
	}
}

func TestBuildTokenMessageDejaElTokenEnSuPropiaLinea(t *testing.T) {
	const token = "0123456789abcdef0123456789abcdef"
	msg := buildTokenMessage(token)

	lines := strings.Split(msg, "\n")
	found := false
	for _, line := range lines {
		if line == token {
			found = true
			break
		}
	}
	if !found {
		// Si el token comparte línea con otro texto, el formato de código lo
		// abarcaría de más y copiarlo de un toque dejaría de funcionar.
		t.Fatalf("el token debe ocupar una línea entera:\n%s", msg)
	}
	if strings.Count(msg, token) != 1 {
		t.Fatalf("el token no debe aparecer dos veces:\n%s", msg)
	}
}

func TestRequestOriginIgnoraCabecerasDeProxy(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, tokenSendAPIPath, nil)
	r.RemoteAddr = "192.168.1.40:51234"
	// Falsificable por quien llama: no debe influir en lo que se registra.
	r.Header.Set("X-Forwarded-For", "8.8.8.8")

	if got := requestOrigin(r); got != "192.168.1.40" {
		t.Fatalf("origen incorrecto: got %q, want %q", got, "192.168.1.40")
	}
}

func TestSendTokenRechazaMetodoIncorrecto(t *testing.T) {
	s := &Server{}
	w := httptest.NewRecorder()

	s.handleSendTokenToTelegram(w, httptest.NewRequest(http.MethodGet, tokenSendAPIPath, nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("un GET debe rechazarse: got %d", w.Code)
	}
}

func TestSendTokenAplicaLimiteDeFrecuencia(t *testing.T) {
	s := &Server{apiToken: "abc"}
	// Turno ya consumido: el siguiente intento no debe llegar a Telegram.
	s.tokenSendNextAt = time.Now().Add(30 * time.Second)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, tokenSendAPIPath, nil)
	r.RemoteAddr = "10.0.0.5:1234"
	s.handleSendTokenToTelegram(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("se esperaba 429, got %d", w.Code)
	}

	var body struct {
		Status     string `json:"status"`
		RetryAfter int    `json:"retry_after"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("respuesta ilegible: %v", err)
	}
	if body.Status != "rate_limited" {
		t.Fatalf("estado inesperado: %q", body.Status)
	}
	if body.RetryAfter < 1 || body.RetryAfter > 31 {
		t.Fatalf("la espera devuelta no es razonable: %d", body.RetryAfter)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("falta la cabecera Retry-After")
	}
	// La respuesta nunca debe filtrar el token.
	if strings.Contains(w.Body.String(), "abc") {
		t.Fatalf("la respuesta no debe incluir el token: %s", w.Body.String())
	}
}

func TestSendTokenSinTokenGenerado(t *testing.T) {
	s := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, tokenSendAPIPath, nil)
	r.RemoteAddr = "10.0.0.5:1234"

	s.handleSendTokenToTelegram(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("sin token generado se esperaba 503, got %d", w.Code)
	}
}
