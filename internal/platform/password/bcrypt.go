package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
