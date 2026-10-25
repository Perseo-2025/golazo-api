package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err!= nil {
		return err
	}

	defer pool.Close()

	fmt.Println("Conectado a PostgreSQL")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthCheck)

	srv := &http.Server{
		Handler: mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("escuchar en el puerto %s: %w", port, err)
	}

	fmt.Println("Servidor escuchando en", listener.Addr())

	// El servidor correindo en su propio goroutine.
	serverErr := make(chan error, 1)
	go func(){
		serverErr <- srv.Serve(listener)
	}()

	// Esperamos lo que ocurra primero
	select {
	case  err := <- serverErr:
		return fmt.Errorf("servidor HTTP: %w", err)
	case <- ctx.Done():
		fmt.Println("Señal recibida, apagando...")
	}

	//Damos hasta 15 segundos ara que terminen as peticiones en curso.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("apagar servidor: %w", err)
	}

	fmt.Println("Servidor detenido correctamente")
	return nil

}
