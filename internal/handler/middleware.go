package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// TokenParser valida un token y devuelve el id del usuario
type TokenParser interface {
	Parse(token string) (string, error)
}

//userIDKey es la clave para guardar el id del usuario en el contexto
// Es un tipo propio privado para que ningún otro paquete pueda pisarla
type userIDKey struct {}

// UserIDFromContext devuelve el id del usuario autenticado, si lo hay
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok && userID != ""
}

// RequiereAuth devuelve un middleware que exige un token Bearer válido.
//Si es válido, guarda el id del usuario en el contexto y llama a next
func RequireAuth(tokens TokenParser, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			raw, found := strings.CutPrefix(header, "Bearer ")
			if !found || raw == "" {
				unauthorized(w, logger)
				return
			}

			userID, err := tokens.Parse(raw)
			if err != nil {
				unauthorized(w, logger)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//unauthorized()
func unauthorized(w http.ResponseWriter, logger *slog.Logger) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="golazo-api"`)
	writeError(w, logger, http.StatusUnauthorized, "unauthorized", "token ausente, inválido o vencido")
}