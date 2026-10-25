package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("crear pool: %w", err) // crea la conexion
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // cerramos
	

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping a postgres: %w", err)
	}

	return pool, nil
}