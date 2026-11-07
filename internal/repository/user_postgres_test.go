//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Perseo-2025/golazo-api/internal/domain"
	"github.com/Perseo-2025/golazo-api/internal/repository"
)

func TestUserPostgres_CreaUsuarioConBilletera(t *testing.T) {
	pool := newTestPool(t)
	repo := repository.NewUserPostgres(pool)
	ctx := context.Background()

	user, err := repo.Create(ctx, newTestUser("ana@golazo.pe", "12345678"))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if user.ID == "" {
		t.Error("el usuario debería tener el id que generó Postgres")
	}
	if user.Status != "active" {
		t.Errorf("status = %q, se esperaba %q", user.Status, "active")
	}

	var wallets int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM wallets WHERE user_id = $1", user.ID).Scan(&wallets)
	if err != nil {
		t.Fatalf("contar billeteras: %v", err)
	}
	if wallets != 1 {
		t.Errorf("billeteras = %d, se esperaba 1", wallets)
	}
}

func TestUserPostgres_Duplicados(t *testing.T) {
	pool := newTestPool(t)
	repo := repository.NewUserPostgres(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, newTestUser("ana@golazo.pe", "12345678")); err != nil {
		t.Fatalf("crear el primer usuario: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		document string
		wantErr  error
	}{
		{"email repetido", "ana@golazo.pe", "87654321", domain.ErrEmailTaken},
		{"documento repetido", "otra@golazo.pe", "12345678", domain.ErrDocumentTaken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.Create(ctx, newTestUser(tt.email, tt.document))
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, se esperaba %v", err, tt.wantErr)
			}
		})
	}
}