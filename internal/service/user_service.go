package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

// la interface sirve para realizar una accion
// UserRepository es lo que el servicio necesita para guardar usuarios.
type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
}

// PasswordHasher es lo que el servicio necesita para hashear constraseñas.
type PasswordHasher interface {
	Hash(plain string) (string, error)
}

// peryTime es la zona horaria de Perú (UTP-5, sin horario de verano).
var peruTime = time.FixedZone("America/Lima", -5*60*60)

// RegisterInput son los datos que llegan para registrar un usuario.
// Es la entrada de los datos
type RegisterInput struct {
	domain.NewUserParams
	Password string
}

// User service coordina el registro de usuarios.
type UserService struct {
	userRepo UserRepository
	hasher PasswordHasher
	today func() time.Time
}

// NewUserService crea el servicio con sus dependencias.
func NewUserService(repo UserRepository, hasher PasswordHasher) *UserService {
	return &UserService{
		userRepo: repo,
		hasher: hasher,
		today: todayInPeru,
	}
}

// Register valida, hashea la contraseña y guarda al usuario.
func (s *UserService) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	if err := domain.ValidatePassword(in.Password); err != nil {
		return domain.User{}, err
	}

	user, err := domain.NewUser(in.NewUserParams, s.today())
	if err != nil {
		return domain.User{}, err
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hashear contraseña: %w", err)
	}

	user.PasswordHash = hash

	return s.userRepo.Create(ctx, user)

}

// devuelve la fecha de perú
func todayInPeru() time.Time{
	now := time.Now().In(peruTime)
	return time.Date(now.Year(), now.Month(), now.Day(), 0,0,0,0, time.UTC)
}