package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrMismatch = errors.New("la contraseña no coincide")

// BcryptHasher hashea contraseñas con bcrypt.
type BcryptHasher struct {
	cost int 
}

// NewBcryptHasher crea el hasher . cost controla qué tan lento es el hash.
func NewBcryptHasher(cost int) * BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash genera el hash de la contraseña.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(hash), nil
}

func (h *BcryptHasher) Compare(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([] byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrMismatch
	}
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}
	return nil
}
