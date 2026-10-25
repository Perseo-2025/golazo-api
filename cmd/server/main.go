package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/Perseo-2025/golazo-api/internal/handler"
	"github.com/Perseo-2025/golazo-api/internal/platform/postgres"
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

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL es obligatorio")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, databaseURL)

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("Conectado a PostgreSQL")

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