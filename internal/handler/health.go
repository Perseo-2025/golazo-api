package handler

import (
	"log"
	"net/http"
)

// si el servidor está vacío
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		log.Println("error escribiendo respuesta: ", err)
	}
}