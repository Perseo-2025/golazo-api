package token

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const testSecrets = "claveeeeeeeeeeeeeeeeepepepepepep"

func newTestManager(t *testing.T, now time.Time) *JWTManager {
	t.Helper()
	m, err := NewJWTManager(testSecrets, 15*time.Minute)
	if err != nil {
		t.Fatalf("crear manager: %v", err)
	}
	m.now = func() time.Time { return now }
	return m
}

func TestJWT_EmitirYValidar(t *testing.T) {
	m := newTestManager(t, time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC))

	tok, _, err := m.Issue("user-123")
	if err != nil {
		t.Fatalf("emitir: %v", err)
	}

	userID, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("validar: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("userID = %q, se esperaba %q", userID, "user-123")
	}
}

func TestJWT_Rechazos(t *testing.T) {
	issuedAt := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	tok, _, err := newTestManager(t, issuedAt).Issue("user-123")
	if err != nil {
		t.Fatalf("emitir: %v", err)
	}

	otherSecret, err := NewJWTManager("otra-clave-distinta-de-al-menos-32-caracteres", 15*time.Minute)
	if err != nil {
		t.Fatalf("crear manager: %v", err)
	}
	otherSecret.now = func() time.Time { return issuedAt }

	parts := strings.Split(tok, ".")
	tampered := parts[0] + "." + parts[1] + "x." + parts[2]

	tests := []struct {
		name  string
		m     *JWTManager
		token string
	}{
		{"vencido (16 miunutos después)", newTestManager(t, issuedAt.Add(16*time.Minute)), tok},
		{"firmado con otra clave", otherSecret, tok},
		{"payload modificado", newTestManager(t, issuedAt), tampered},
		{"texto basura", newTestManager(t, issuedAt), "no-es-un-jwt"},
		{"vacío", newTestManager(t, issuedAt), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.m.Parse(tt.token)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("error = %v, se esperaba ErrInvalidToken", err)
			}
		})
	}

}

func TestJWT_ClaveCorta(t *testing.T) {
	if _, err := NewJWTManager("corta", 15*time.Minute); err == nil {
		t.Error("una clave corta debería ser rechazada")
	}
}
