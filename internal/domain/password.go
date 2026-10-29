package domain

import (
	"errors"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var ErrWeakPassword = errors.New("la contraseña debe tener entre 8 y 72 caracteres")


// ValidatePassword revisa la política de contraseñas sobre el texto plano.

func ValidatePassword(plain string) error {
	if len(plain) < MinPasswordLength || len(plain) > MaxPasswordLength {
		return ErrWeakPassword
	}
	return nil
}
