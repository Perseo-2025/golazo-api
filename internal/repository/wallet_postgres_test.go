//go:build integration

package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Perseo-2025/golazo-api/internal/domain"
	"github.com/Perseo-2025/golazo-api/internal/repository"
)

func TestWalletPostgres_DepositoEIdempotencia(t *testing.T) {
	pool := newTestPool(t)
	userID, walletID := createUserWithWallet(t, pool, "ana@golazo.pe", "12345678")
	repo := repository.NewWalletPostgres(pool)
	ctx := context.Background()

	first, created, err := repo.Deposit(ctx, userID, walletID, 5_000, "yape-001")
	if err != nil {
		t.Fatalf("primer depósito: %v", err)
	}
	if !created || first.BalanceAfterCents != 5_000 {
		t.Fatalf("created = %v, balance_after = %d", created, first.BalanceAfterCents)
	}

	// Reintento con la misma clave y el mismo monto.
	again, created, err := repo.Deposit(ctx, userID, walletID, 5_000, "yape-001")
	if err != nil {
		t.Fatalf("reintento: %v", err)
	}
	if created {
		t.Error("el reintento no debería crear un movimiento nuevo")
	}
	if again.ID != first.ID {
		t.Errorf("id del reintento = %s, se esperaba %s", again.ID, first.ID)
	}

	// Misma clave, otro monto.
	_, _, err = repo.Deposit(ctx, userID, walletID, 9_000, "yape-001")
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Errorf("error = %v, se esperaba ErrIdempotencyConflict", err)
	}

	assertLedgerMatches(t, pool, walletID, 5_000)
}

func TestWalletPostgres_BilleteraInexistente(t *testing.T) {
	pool := newTestPool(t)
	repo := repository.NewWalletPostgres(pool)
	anyUser := "00000000-0000-0000-0000-000000000000"

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid"} {
		_, _, err := repo.Deposit(context.Background(), anyUser, id, 5_000, "yape-001")
		if !errors.Is(err, domain.ErrWalletNotFound) {
			t.Errorf("id %q: error = %v, se esperaba ErrWalletNotFound", id, err)
		}
	}
}

func TestWalletPostgres_BilleteraAjena(t *testing.T) {
	pool := newTestPool(t)
	_, anaWallet := createUserWithWallet(t, pool, "ana@golazo.pe", "12345678")
	luisID, _ := createUserWithWallet(t, pool, "luis@golazo.pe", "87654321")
	repo := repository.NewWalletPostgres(pool)

	// Luis intenta depositar en la billetera de Ana.
	_, _, err := repo.Deposit(context.Background(), luisID, anaWallet, 5_000, "yape-001")
	if !errors.Is(err, domain.ErrWalletNotFound) {
		t.Fatalf("error = %v, se esperaba ErrWalletNotFound", err)
	}

	assertLedgerMatches(t, pool, anaWallet, 0)
}

func TestWalletPostgres_DepositosConcurrentes(t *testing.T) {
	pool := newTestPool(t)
	userID, walletID := createUserWithWallet(t, pool, "ana@golazo.pe", "12345678")
	repo := repository.NewWalletPostgres(pool)

	const workers = 50
	const amount int64 = 100 // S/ 1.00 cada uno

	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrente-%d", i)
			if _, _, err := repo.Deposit(context.Background(), userID, walletID, amount, key); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("depósito concurrente falló: %v", err)
	}

	assertLedgerMatches(t, pool, walletID, workers*amount)
}