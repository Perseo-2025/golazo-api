package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestWalletDeposit(t *testing.T) {
	test := []struct {
		name        string
		amount      int64
		key         string
		wantErr     error
		wantBalance int64
	} {
		{"depósito válido", 5_000, "yape-001", nil, 15_000},
		{"monto cero", 0, "yape-001", ErrInvalidAmount, 10_000},
		{"monto negativo", -5_000, "yape-001", ErrInvalidAmount, 10_000},
		{"justo el máximo", MaxDepositCents, "yape-001", nil, 10_000 + MaxDepositCents},
		{"un céntimo sobre el máximo", MaxDepositCents + 1, "yape-001", ErrDepositLimit, 10_000},
		{"clave vacía", 5_000, "  ", ErrInvalidIdempotency, 10_000},
		{"clave muy larga", 5_000, strings.Repeat("a", 101), ErrInvalidIdempotency, 10_000},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			w := Wallet{ID: "w-1", BalanceCents: 10_000} // S/ 100.00

			tx, err := w.Deposit(tt.amount, tt.key)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, se esperaba %v", err, tt.wantErr)
			}
			if w.BalanceCents != tt.wantBalance {
				t.Errorf("saldo = %d, se esperaba %d", w.BalanceCents, tt.wantBalance)
			}
			if err == nil && tx.BalanceAfterCents != w.BalanceCents{
				t.Errorf("balance_after = %d, saldo = %d", tx.BalanceAfterCents, w.BalanceCents)
			}
		})
	}

}
