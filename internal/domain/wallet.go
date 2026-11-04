package domain

import (
	"errors"
	"strings"
	"time"

)

// MaxDepositCents es el depósito máximo permitido: S/ 5 000.00
const MaxDepositCents int64 = 500_000

// MMaxIdempotencyKeyLength es el largo máximo de una clave de idempotencia.
const MaxIdempotencyKeyLength = 100

// TransactionType es el tipo de movimiento en el ledger.
type TransactionType string

const (
	TxDeposit    TransactionType = "deposit"
	TxWithdrawal TransactionType = "withdrawal"
	TxBetStake   TransactionType = "bet_stake"
	TxBetWin     TransactionType = "bet_win"
	TxBetRefund  TransactionType = "bet_refund"
	TxBonus      TransactionType = "bonus"
	TxAdjustment TransactionType = "adjustment"
)

var (
	ErrInvalidAmount       = errors.New("el monto debe ser mayor que cero")
	ErrDepositLimit        = errors.New("el depósito máximo es de S/ 5 000.00")
	ErrInvalidIdempotency  = errors.New("la clave de idempotencia es obligatoria (máximo 100 caracteres)")
	ErrWalletNotFound      = errors.New("billetera no enontrada")
	ErrIdempotencyConflict = errors.New("la clave de idempotencia ya se usó con otros datos")
)

// Wallet es la billetera de un usuario. Los montos van en céntimos.
type Wallet struct {
	ID                string
	WalletID          string
	Type              string
	AmountCents       int64
	BalanceAfterCents int64
	IdempotencyKey    int64
	CreatedAt         time.Time
}

// WalletTransaction es un movimiento del ledger. Nunca se modifica
type WalletTrasaction struct {
	ID                string
	WalletID          string
	Type              TransactionType
	AmontCents        int64
	BalanceAfterCents int64
	IdempotencyKey    string
	CreateAt          time.Time
}

// Deposit suma amountCents al saldo y devuelve el movimiento que lo registra.
// si algo no es válido, el saldo no cambia.
func (w *Wallet) Deposit(amountCents int64, idempotencyKey string) (WalletTrasaction, error) {
	if amountCents <= 0 {
		return  WalletTrasaction{}, ErrInvalidAmount
	}
	if amountCents > MaxDepositCents {
		return WalletTrasaction{}, ErrDepositLimit
	}

	key := strings.TrimSpace(idempotencyKey)
	if key == "" || len(key) > MaxIdempotencyKeyLength {
		return WalletTrasaction{}, ErrInvalidIdempotency
	}

	w.BalanceAfterCents += amountCents

	return WalletTrasaction{
		WalletID: w.ID,
		Type: TxDeposit,
		AmontCents: amountCents,
		BalanceAfterCents: w.BalanceAfterCents,
		IdempotencyKey: key,
	}, nil

}
