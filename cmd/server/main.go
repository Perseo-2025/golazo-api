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

	"github.com/kingperseocs/golazo-api/internal/handler"
	"github.com/kingperseocs/golazo-api/internal/platform/password"
	"github.com/kingperseocs/golazo-api/internal/platform/postgres"
	"github.com/kingperseocs/golazo-api/internal/platform/token"
	"github.com/kingperseocs/golazo-api/internal/repository"
	"github.com/kingperseocs/golazo-api/internal/service"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET es obligatorio")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("conectado a POSTGRESQL")

	hasher := password.NewBcryptHasher(12)

	tokens, err := token.NewJWTManager(jwtSecret, 15*time.Minute)
	if err != nil {
		return err
	}

	dummyHash, err := hasher.Hash("contraseña-que-nadie-usaparaparaigualartiempos")
	if err != nil {
		return fmt.Errorf("generar hash false: %w", err)
	}


	userRepo := repository.NewUserPostgres(pool)
	userService := service.NewUserService(userRepo, hasher)
	userHandler := handler.NewUserHandler(userService, logger)

	authService := service.NewAuthService(userRepo, hasher, tokens, dummyHash)
	authHandler := handler.NewAuthHanlder(authService, logger)

	walletRepo := repository.NewWalletPostgres(pool)
	walletService := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /users", userHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /wallets/{id}/deposits", walletHandler.Deposit)

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("escuchar en el puerto %s: %w", port, err)
	}

	logger.Info("servidor HTTP escuchando", "addr", listener.Addr().String(), "env", env)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Serve(listener)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("servidor HTTP: %w", err)
	case <-ctx.Done():
		logger.Info("señal recibida, apagando")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("apagar servidor: %w", err)
	}

	logger.Info("servidor detenido correctamente")
	return nil
}
