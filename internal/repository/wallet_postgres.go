package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

// Código de error de Postgres para  "texto con formato" (por ejemplo, un UUID mal escrito)
const pgInvalidTextRepresentation = "22P02"

// WalletPostgres guarda billeteras y movimientos en PostgreSQL.
type WalletPostgres struct {
	pool *pgxpool.Pool
}

// NewWalletPostgres crea el repositorio con el pool de conexios.
func NewWalletPostgres(pool *pgxpool.Pool) *WalletPostgres {
	return &WalletPostgres{pool: pool}
}

// Consultas SQL
const lockWalletSQL = `
SELECT id, user_id, type, currency, balance_cents, created_at, updated_at
FROM wallets
WHERE id = $1
FOR UPDATE`

const findTransactionByKeySQL = `
SELECT id, wallet_id, type, amount_cents, balance_after_cents, idempotency_key, created_at
FROM wallet_transactions
WHERE wallet_id = $1 AND idempotency_key = $2`

const updateBalanceSQL = `
UPDATE wallets
SET balance_cents = $2, updated_at = now()
WHERE id = $1`

const insertTransactionSQL = `
INSERT INTO wallet_transactions (wallet_id, type, amount_cents, balance_after_cents, idempotency_key)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`

func (r *WalletPostgres) Deposit(ctx context.Context, walletID string, amountCents int64, key string) (tx domain.WalletTrasaction, created bool, err error) {
	err = pgx.BeginFunc(ctx, r.pool, func(dbTx pgx.Tx) error {
		// 1. Bloquear la billetera. Otro depósito a la misma billetera espera aquí.
		var w domain.Wallet
		err := dbTx.QueryRow(ctx, lockWalletSQL, walletID).Scan(
			&w.ID, &w.UserID, &w.Type, &w.Currency, &w.BalanceCents, &w.CreatedAt, &w.UpdatedAt, 
		)
		if err != nil {
			return mapWalletError(err)
		}

		// 2. ¿Ya procesamos esta clave? (se revisa DESPUÉS del bloqueo)
		existing, found, err := findTransactionByKey(ctx, dbTx, walletID, key)
		if err != nil {
			return err
		}
		if found {
			if existing.Type != domain.TxDeposit || existing.AmountCents != amountCents {
				return domain.ErrIdempotencyConflict
			}
			tx = existing
			return nil
		}

		// 3. Reglas del negocio en el dominio
		newTx, err := w.Deposit(amountCents, key)
		if err != nil {
			return err
		}

		// 4. Guardar el nuevo saldo.
		if _, err := dbTx.Exec(ctx, updateBalanceSQL, w.ID, w.BalanceCents); err != nil {
			return fmt.Errorf("actualizar saldo: %w", err)
		}

		// 5. Registrar el movimiento en el ledger.
		err = dbTx.QueryRow(ctx, insertTransactionSQL,
			newTx.WalletID, string(newTx.Type), newTx.AmountCents, newTx.BalanceAfterCents, newTx.IdempotencyKey,
		).Scan(&newTx.ID, &newTx.CreateAt)
		if err != nil {
			return fmt.Errorf("insertar movimiento: %w", err)
		}

		tx = newTx
		created = true
		return nil
	})
	if err != nil {
		return domain.WalletTrasaction{}, false, err
	}
	return tx, created, nil
}

// findTransactionByKey busca un movimiento por su calve de idempotencia.
func findTransactionByKey(ctx context.Context, dbTx pgx.Tx, walletID, key string) (domain.WalletTrasaction, bool, error) {
	var t domain.WalletTrasaction
	var txType string
	err :=dbTx.QueryRow(ctx, findTransactionByKeySQL,  walletID, key).Scan(
		&t.ID, &t.WalletID, &txType, &t.AmountCents, &t.BalanceAfterCents, &t.IdempotencyKey, &t.CreateAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WalletTrasaction{}, false, nil
	}
	if err != nil {
		return domain.WalletTrasaction{}, false, fmt.Errorf("buscar movimiento por claves: %w", err)
	}
	t.Type = domain.TransactionType(txType)
	return t, true, nil
}

func mapWalletError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrWalletNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgInvalidTextRepresentation {
		return domain.ErrWalletNotFound
	}
	return fmt.Errorf("bloquear billetera: %w", err)
}