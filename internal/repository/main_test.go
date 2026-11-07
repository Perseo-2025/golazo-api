//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Perseo-2025/golazo-api/internal/domain"
	"github.com/Perseo-2025/golazo-api/internal/repository"
)

// newTestPool conecta a la base de tests y la deja vacía.
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definido; ejecuta con: task test:integration")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("conectar a la base de tests: %v", err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(context.Background(),
		"TRUNCATE wallet_transactions, wallets, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("limpiar tablas: %v", err)
	}
	return pool
}

// newTestUser devuelve un usuario válido. Cambia lo que necesites en cada test.
func newTestUser(email, document string) domain.User {
	return domain.User{
		DocumentType:   domain.DocumentDNI,
		DocumentNumber: document,
		FirstName:      "Ana",
		LastName:       "Lopez",
		BirthDate:      time.Date(2000, time.January, 10, 0, 0, 0, 0, time.UTC),
		Email:          email,
		Phone:          "987654321",
		PasswordHash:   "hash-de-prueba",
	}
}

// createUserWithWallet registra un usuario y devuelve el id de su billetera real.
func createUserWithWallet(t *testing.T, pool *pgxpool.Pool, email, document string) string {
	t.Helper()
	ctx := context.Background()

	user, err := repository.NewUserPostgres(pool).Create(ctx, newTestUser(email, document))
	if err != nil {
		t.Fatalf("crear usuario: %v", err)
	}

	var walletID string
	err = pool.QueryRow(ctx, "SELECT id FROM wallets WHERE user_id = $1", user.ID).Scan(&walletID)
	if err != nil {
		t.Fatalf("buscar billetera: %v", err)
	}
	return walletID
}

// assertLedgerMatches comprueba que el saldo es el esperado y que coincide
// con la suma de los movimientos del ledger.
func assertLedgerMatches(t *testing.T, pool *pgxpool.Pool, walletID string, want int64) {
	t.Helper()
	ctx := context.Background()

	var balance, ledgerSum int64
	if err := pool.QueryRow(ctx, "SELECT balance_cents FROM wallets WHERE id = $1", walletID).Scan(&balance); err != nil {
		t.Fatalf("leer saldo: %v", err)
	}
	if err := pool.QueryRow(ctx,
		"SELECT COALESCE(SUM(amount_cents), 0) FROM wallet_transactions WHERE wallet_id = $1", walletID,
	).Scan(&ledgerSum); err != nil {
		t.Fatalf("sumar ledger: %v", err)
	}

	if balance != want {
		t.Errorf("saldo = %d, se esperaba %d", balance, want)
	}
	if balance != ledgerSum {
		t.Errorf("el saldo (%d) no coincide con la suma del ledger (%d)", balance, ledgerSum)
	}
}