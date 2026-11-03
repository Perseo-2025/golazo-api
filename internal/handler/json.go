package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorDetail struct {
	Code  	 string 	`json:"code"`
	Message  string 	`json:"message"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

// writeJSON escribe body como JSON con el código de estado indicado.
func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Error("no se pudo escribir la respuesta JSON", "error", err)
	}
}

// writeError escribe un error con el formato estándar de a API.
func writeError(w http.ResponseWriter, logger *slog.Logger, status int, code, message string) {
	writeJSON(w, logger, status, errorResponse {
		Error: errorDetail{Code: code, Message: message},
	})
}