package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeTokens acepta solo el token "valido" y lo asocia al usuario "ana-id"
type fakeTokens struct {}

func (fakeTokens) Parse(token string) (string, error) {
	if token == "valido" {
		return "ana-id", nil
	}
	return "", errors.New("token inválido")
}

func TestRequireAuth(t *testing.T) {
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name  string
		header  string
		wantStatus  int
		wantCalled  bool
	} {
		{"token válido", "Bearer valido", http.StatusOK, true},
		{"sin header", "", http.StatusUnauthorized, false},
		{"token inválido", "Bearer falso", http.StatusUnauthorized, false},
		{"sin la palabra Bearer", "valido", http.StatusUnauthorized, false},
		{"Bearer vacío", "Bearer ", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false 
			var gotUserID string

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
				called = true
				gotUserID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/wallets/w-1/deposits", nil) //<- estblemos la ruta
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			RequireAuth(fakeTokens{}, discard)(next).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tt.wantStatus)
			}
			if called != tt.wantCalled {
				t.Errorf("¿se llamó al hanlder? = %v, se esperaba %v", called, tt.wantCalled)
			}
			if tt.wantCalled && gotUserID != "ana-id" {
				t.Errorf("userID en el contexto= %q, se esperaba %q", gotUserID, "ana-id")
			}

		})
	}

}
