package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Perseo-2025/golazo-api/internal/handler"
	"github.com/joho/godotenv"
)

func main(){

	// carga .env si existe. En producción no habrá archivado ye stá bien.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("error leyendo .env: ", err)
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // asigna un valor si el puerto está vacío
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthCheck)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err) 
	}

	fmt.Println("Servidor escuchando en", listener.Addr())

	if err := http.Serve(listener, mux); err != nil {
		log.Fatal(err)
	}

}