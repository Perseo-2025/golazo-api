package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken se devuelve cuando el token no es válido o venció.
var ErrInvalidToken = errors.New("token inválido o venció")

const issuer = "golazo-api"

// JWTManager firma y valida tokens de acceso con HS256
type JWTManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewJWTManager crea el manejador. ttkl es cuánto dura cada token.
func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET debe tener al menos 32 caracteres")
	}
	return &JWTManager{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}	

// Issue firma un token para el usuario y devuelve también cuando vence.
func (m *JWTManager) Issue(userID string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)

	claims := jwt.RegisteredClaims {
		Subject: userID,
		Issuer: issuer,
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("firmar token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse valida la firma y el vencimiento, y devuelve el id del usuario.
func (m *JWTManager) Parse(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return "", ErrInvalidToken
	}
	if claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
