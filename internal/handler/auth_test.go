package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Perseo-2025/golazo-api/internal/service"
)

// fakeAuthenticator devuelve siempre el resultado y el error que le indiquemos.
type fakeAuthenticator struct {
	res service.LoginResult
	err error
}

func (f fakeAuthenticator) Login(ctx context.Context, email, password string) (service.LoginResult, error) {
	return f.res, f.err
}

func TestAuthHandler_Login(t *testing.T) {
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	okResult := service.LoginResult{AccessToken: "token-ok", ExpiresAt: time.Now().Add(15 * time.Minute)}

	tests := []struct {
		name       string
		auth       fakeAuthenticator
		body       string
		wantStatus int
		wantCode   string // code del error; vacío si es exitoso
	}{
		{"login correcto", fakeAuthenticator{res: okResult}, `{"email":"a@b.pe","password":"x"}`, http.StatusOK, ""},
		{"credenciales inválidas", fakeAuthenticator{err: service.ErrInvalidCredentials}, `{"email":"a@b.pe","password":"x"}`, http.StatusUnauthorized, "invalid_credentials"},
		{"cuenta inactiva", fakeAuthenticator{err: service.ErrAccountInactive}, `{"email":"a@b.pe","password":"x"}`, http.StatusForbidden, "account_inactive"},
		{"error desconocido", fakeAuthenticator{err: errors.New("postgres caído")}, `{"email":"a@b.pe","password":"x"}`, http.StatusInternalServerError, "internal_error"},
		{"JSON roto", fakeAuthenticator{}, `{"email":`, http.StatusBadRequest, "invalid_json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			h := NewAuthHanlder(tt.auth, discard)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tt.wantStatus)
			}

			// el cuerpo debe se UN solo JSO, si hay algo despues, falta un return
			decoder := json.NewDecoder(rec.Body)
			var body map[string]any
			if err := decoder.Decode(&body); err != nil {
				t.Fatalf("el cuerpo no es un JSON: %v", err)
			}
			if decoder.More(){ 
				t.Fatal("la respuesta tiene más de un JSON: ¿falta un return después de writeError?")
			}

			if tt.wantCode == "" {
				if body["access_token"] != "token-ok" {
					t.Errorf("access_token = %v", body["access_token"])
				}
				return 
			}
			errObj, ok := body["error"].(map[string]any)
			if !ok {
				t.Fatalf("se esperaba un objeto error, llegó %v", body)
			}
			if errObj["code"] != tt.wantCode {
				t.Errorf("code = %v, se esperaba %q", errObj["code"], tt.wantCode)
			}
 
		})
	}

}
