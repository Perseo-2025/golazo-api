package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

// fakeRepo cumple UserRepository sin tocar una base de datos.
type fakeRepo struct {
	saved domain.User // lo que el servicio intentó guardar
	err error  // e error que queremos simular 
}

func (f *fakeRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if f.err != nil {
		return domain.User{}, f.err
	}
	f.saved = u
	u.ID = "id-false"
	return u, nil
}

//fakeHasher cumple PasswordHasher de forma instantánea y predecible.
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) {
	return "hash:" + plain, nil
}

// creamos el test
func newTestService(repo *fakeRepo) *UserService {
	s := NewUserService(repo, fakeHasher{})
	s.today = func() time.Time {
		return time.Date(2026, time.October, 28,0,0,0,0, time.UTC)
	}
	return s
}

func validInput() RegisterInput {
	return RegisterInput{
		NewUserParams: domain.NewUserParams{
			DocumentType: domain.DocumentDNI,
			DocumentNumber: "12345678",
			FirstName: "Ana",
			LastName: "Lopez",
			BirthDate: time.Date(2000, time.January, 10,0,0,0,0, time.UTC),
			Email: "ana@golazo.pe",
			Phone: "987879653",
		},
		Password: "MiClave123",
	}
}


func TestRegister_GuardarConHash(t *testing.T) {
	repo := &fakeRepo{}
	s := newTestService(repo)

	user, err := s.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if user.ID != "id-false" {
		t.Errorf("ID = %q, se esperaba %q", user.ID, "id-false")
	}
	if repo.saved.PasswordHash != "hash:MiClave123" {
		t.Errorf("se guardó el hash %q", repo.saved.PasswordHash)
	}
}

func TestRegister_ContrasenaDebil( t*testing.T) {
	repo := &fakeRepo{}
	s := newTestService(repo)

	in := validInput()
	in.Password = "123"
	_, err := s.Register(context.Background(), in)
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("error = %v, se esperaba ErrWeakPassword", err)
	}
	if repo.saved.Email != "" {
		t.Error("no debería haberse guardado nada")
	}
}

func TestRegister_EmailDuplicado(t *testing.T) {
	repo := &fakeRepo{err: domain.ErrEmailTaken}
	s := newTestService(repo)

	_, err := s.Register(context.Background(), validInput())
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("error = %v, se esperaba ErrEmailTaken", err)
	}
}