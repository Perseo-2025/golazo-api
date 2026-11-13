package service

import (
	"context"
	"github.com/Perseo-2025/golazo-api/internal/domain"
)

//  WalletRepository es lo que el servicio necesita para mover dinero.
type WalletRepository interface {
	Deposit(ctx context.Context, userID, walletID string, amountCents int64, key string) (domain.WalletTrasaction, bool, error)
}

//WalletService
type WalletService struct {
	repo WalletRepository
}

// NewWalletService crea el servicio con su repositorio.
func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// Deposit acredita dinero en la billetera. created indica si el movimiento es nuevo (true) o si es un rendimiento de uno ya procesado (false)
func (s *WalletService) Deposit(ctx context.Context, userID,walletID string, amountCents int64, key string) (domain.WalletTrasaction, bool, error) {
	return s.repo.Deposit(ctx, userID, walletID, amountCents, key)
}