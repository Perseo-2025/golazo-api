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

func main() {
	if err := run(); err != nil {
		log.Fatal(err) // el ÚNICO lugar donde el programa termina con error
	}
}

func run() error {

	// carga .env si existe. En producción no habrá archivado ye stá bien.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("leer .env: %w", err)
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // asigna un valor si el puerto está vacío
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL es obligatorio")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, databaseURL)

	if err != nil {
		return err
	}
	defer func() {
		pool.Close()
		fmt.Println("Pool de PostgreSQL cerrado")
	}()

	fmt.Println("Conectado a PostgreSQL")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthCheck)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("escuchar en el puerto %s: %w", port, err)
	}

	fmt.Println("Servidor escuchando en", listener.Addr())

	return http.Serve(listener, mux)

}
