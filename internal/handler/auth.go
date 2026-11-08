package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Perseo-2025/golazo-api/internal/service"
)

// @Controlador

// Autenticator es lo que el handler necesita del servicio de autenticación
type Authenticator interface {
	Login(ctx context.Context, email string, password string) (service.LoginResult, error)
}

// AuthHablder atiende las rutas de autenticación
type AuthHandler struct {
	svc    Authenticator
	logger *slog.Logger
}

// AuthHandler crea el handler con sus dependencias.
func NewAuthHanlder(svc Authenticator, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Login atiende POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.logger.Warn("JSON inválido en login", "error", err)
		writeError(w, h.logger, http.StatusBadRequest, "invalid_json", "el cuerpo debe ser un JSON válido")
		return 
	}

	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, h.logger, http.StatusUnauthorized, "invalid_credentials", err.Error())
		return
	case errors.Is(err, service.ErrAccountInactive):
		writeError(w, h.logger, http.StatusForbidden, "account_inactive", err.Error())
		return
	case err != nil:
		h.logger.Error("login falló", "error", err)
		writeError(w, h.logger, http.StatusInternalServerError, "internal_error", "ocurrió un error interno, intenta más tarde")
		return
	}

	writeJSON(w, h.logger, http.StatusOK, loginResponse{
		AccessToken: res.AccessToken,
		TokenType: "Bearer",
		ExpiresIn: int(time.Until(res.ExpiresAt).Seconds()),
	})

}