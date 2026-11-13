package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kingperseocs/golazo-api/internal/domain"
)

type fakeUserFinder struct {
	user domain.User
	err  error
}

func (f fakeUserFinder) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	if f.err != nil {
		return domain.User{}, f.err
	}
	return f.user, nil
}

// fakeComparer acepta solo la contraseña "correcta" contra el hash "hash-real" y cuenta cuántas veces lo llamaron
type fakeComparer struct{ calls *int }

func (f fakeComparer) Compare(hash, plain string) error {
	*f.calls++
	if hash == "hash-real" && plain == "correcta" {
		return nil
	}
	return errors.New("no coincide :(")
}

type fakeIssuer struct{}

func (fakeIssuer) Issue(userID string) (string, time.Time, error) {
	return "token-de-" + userID, time.Date(2026, 11, 7, 12, 15, 0, 0, time.UTC), nil
}

func TestLogin(t *testing.T) {
	ana := domain.User{ID: "ana-id", PasswordHash: "hash-real", Status: domain.UserActive}
	suspended := ana
	suspended.Status = "suspended"

	tests := []struct {
		name      string
		finder    fakeUserFinder
		password  string
		wantErr   error
		wantToken string
	}{
		{"credenciales correctas", fakeUserFinder{user: ana}, "correcta", nil, "token-de-ana-id"},
		{"contraseña incorrecta", fakeUserFinder{user: ana}, "otra", ErrInvalidCredentials, ""},
		{"email inexistente", fakeUserFinder{err: domain.ErrUserNotFound}, "correcta", ErrInvalidCredentials, ""},
		{"cuenta suspendida", fakeUserFinder{user: suspended}, "correcta", ErrAccountInactive, ""},
		{"suspendida y contraseña incorrecta", fakeUserFinder{user: suspended}, "otra", ErrInvalidCredentials, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			s := NewAuthService(tt.finder, fakeComparer{calls: &calls}, fakeIssuer{}, "hash-falso")

			res, err := s.Login(context.Background(), "  Ana@Golazo.pe ", tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, se esperaba %v", err, tt.wantErr)
			}
			if res.AccessToken != tt.wantToken {
				t.Errorf("token = %q, se esperaba %q", res.AccessToken, tt.wantToken)
			}
			if calls != 1 {
				t.Errorf("Compare se llamó %d veces, se esperaba 1 (también con email inexistente)", calls)
			}
		})
	}

}
