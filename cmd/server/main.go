package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

	looger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(looger); err != nil {
		looger.Error("la aplicación terminó con error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("leer .env :%w", err)
	}

	// validamos que entre las variables de entorno
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DARABASE_URL es Obligatorio")
	}

	// creamos el contexto
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// creamos el pool de conexiones
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("conectado a POSTGRESQL")

	mux := http.NewServeMux()
	mux.HandleFunc("GET  /health", handler.HealthCheck)

	srv := &http.Server {
		Handler:  mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("escuchar en el puerto %s: %w", port, err)
	}

	logger.Info("servidor HTTP escuchando", "addr", listener.Addr().String(), "env", env)

	serverErr := make(chan error, 1)
	go func(){
		serverErr <- srv.Serve(listener)
	}()

	select {
	case err := <- serverErr:
		return fmt.Errorf("servidor HTTP: %w", err)
	case <- ctx.Done():
		logger.Info("señal recibida, apagando")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return  fmt.Errorf("apagar servidor: %w", err)
	}

	logger.Info("servidor detenido correctamente")

	return nil
}
