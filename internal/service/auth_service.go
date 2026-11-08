package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

// Errores del login. Se devuelve el mismo error si el email no existe o si la contraseña es incorrecta, para no relevar qué email tiene cuenta.
var (
	ErrInvalidCredentials = errors.New("email o contraseña incorrectos")
	ErrAccountInactive    = errors.New("la cuenta no está activada")
)

// UserFinder busca usuarios pór email
type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

// PassworedComparer compara una contraseña con su hash
type PasswordComparer interface {
	Compare(hash, plain string) error
}

// TokenIssuer emite token de acceso.
type TokenIssuer interface {
	Issue(userID string) (string, time.Time, error)
}

// LoginResult es lo que devuelve un login exitoso
type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
}

// AuthService maneja el inicio de sesión
type AuthService struct {
	users     UserFinder
	passwords PasswordComparer
	tokens    TokenIssuer
	dummyHash string
}

// NewAuthService crea el servicio. dummyHash es un hash bcrypt cualquiera, usado para que un email inexistente tarde lo mismo que uno real
// servicio de autenticacion 
func NewAuthService(users UserFinder, passwords PasswordComparer, tokens TokenIssuer, dummyHash string) *AuthService{
	return &AuthService{users: users, passwords: passwords, tokens: tokens, dummyHash: dummyHash}
}

//Login valida las credenciales y emite un token de acceso
func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		// comparamos igual contra un hash falso para que la respuesta tarde lo mismo que un email real (evita la enumeración por tiempo).
		if cmpErr := s.passwords.Compare(s.dummyHash, password); cmpErr == nil {
			return LoginResult{}, errors.New("el hash falso no debería coincidir nunca")
		}
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.passwords.Compare(user.PasswordHash, password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	// Recién después de verificar la contraseña revelamos el estado de la cuetna
	if user.Status != domain.UserActive {
		return LoginResult{}, ErrAccountInactive
	}

	accessToken, expiresAt, err := s.tokens.Issue(user.ID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("emitir token: %w", err)
	}
	return LoginResult{AccessToken: accessToken, ExpiresAt: expiresAt}, nil
}

