package domain

import (
	"errors"
	"testing"
	"time"
)

// date es un atajp para crear fechas en los test.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// validParams retorna datos válidos. Cada caso cambia solo lo que quiere probar.
func validParams() NewUserParams {
	return NewUserParams{
		DocumentType:   DocumentDNI,
		DocumentNumber: "12345678",
		FirstName:      "Rey",
		LastName:       "Cardenas",
		BirthDate:      date(2003, time.August, 15),
		Email:          "rey@gmail.com",
		Phone:          "959878813",
	}
}

func TestNewUser(t *testing.T) {
	now := date(2026, time.October, 27) // "hoy" queda fijo para siepre

	tests := []struct {
		name    string
		modify  func(p *NewUserParams)
		wantErr error
	}{
		{"datos válidos", func(p *NewUserParams) {}, nil},
		{"DNI con 7 dígitos", func(p *NewUserParams) { p.DocumentNumber = "1234567" }, ErrInvalidDocument},
		{"DNI con letras", func(p *NewUserParams) { p.DocumentNumber = "ABC12345" }, ErrInvalidDocument},
		{"tipo de documento desconocido", func(p *NewUserParams) { p.DocumentType = "RUC" }, ErrInvalidDocument},
		{"nombre vacío", func(p *NewUserParams) { p.FirstName = "  " }, ErrInvalidName},
		{"email sin dominio", func(p *NewUserParams) { p.Email = "rey@" }, ErrInvalidEmail},
		{"celular que no empieza con 9", func(p *NewUserParams) { p.Phone = "812345678" }, ErrInvalidPhone},
		{"cumple 18 mañana", func(p *NewUserParams) { p.BirthDate = date(2008, time.October, 28) }, ErrUnderage},
		{"cumple 18 hoy", func(p *NewUserParams) { p.BirthDate = date(2008, time.October, 27) }, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validParams()
			tt.modify(&p)

			_, err := NewUser(p, now)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, se esperaba %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUser_NormalizarEmail(t *testing.T) {
	p := validParams()
	p.Email = "  Rey@Golazo.PE  "

	user, err := NewUser(p, date(2026, time.October, 27))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if user.Email != "rey@golazo.pe" {
		t.Errorf("email = %q, se esperaba %q", user.Email, "rey@golazo.pe")
	}
}
