package telegram

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestCodeEntitiesUsaDesplazamientosUTF16(t *testing.T) {
	// El emoji ocupa dos unidades UTF-16 pero cuatro bytes y una sola runa: si
	// se midiera en bytes o en runas, Telegram resaltaría el tramo equivocado.
	const token = "abc123"
	text := "🔐 TelegramDL · token de acceso remoto\n\n" + token + "\n\nfin"

	entities := codeEntities(text, []string{token})
	if len(entities) != 1 {
		t.Fatalf("se esperaba una entidad, hay %d", len(entities))
	}

	code, ok := entities[0].(*tg.MessageEntityCode)
	if !ok {
		t.Fatalf("tipo de entidad inesperado: %T", entities[0])
	}
	if code.Offset != 40 {
		t.Fatalf("desplazamiento incorrecto: got %d, want 40", code.Offset)
	}
	if code.Length != len(token) {
		t.Fatalf("longitud incorrecta: got %d, want %d", code.Length, len(token))
	}
}

func TestCodeEntitiesIgnoraFragmentosAusentes(t *testing.T) {
	entities := codeEntities("hola mundo", []string{"", "no está", "mundo"})
	if len(entities) != 1 {
		t.Fatalf("solo el fragmento presente debe generar entidad: %+v", entities)
	}

	code := entities[0].(*tg.MessageEntityCode)
	if code.Offset != 5 || code.Length != 5 {
		t.Fatalf("entidad incorrecta: offset=%d length=%d", code.Offset, code.Length)
	}
}

func TestMessageRandomIDNoSeRepite(t *testing.T) {
	vistos := make(map[int64]bool, 64)
	for i := 0; i < 64; i++ {
		id, err := messageRandomID()
		if err != nil {
			t.Fatalf("error generando el identificador: %v", err)
		}
		if vistos[id] {
			t.Fatalf("identificador repetido: %d", id)
		}
		vistos[id] = true
	}
}

func TestSendToSavedMessagesSinClienteFalla(t *testing.T) {
	// Se construye a mano en vez de con NewClientManager para no tocar
	// ~/.tgdown ni la sesión real del usuario durante los tests.
	cm := &ClientManager{}

	if err := cm.SendToSavedMessages(t.Context(), "hola"); err == nil {
		t.Fatal("sin cliente configurado debe devolver error, no enviar nada")
	}
	if err := cm.SendToSavedMessages(t.Context(), "   "); err == nil {
		t.Fatal("un mensaje vacío debe rechazarse")
	}
}
