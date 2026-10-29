package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// MinAge es la edad mínima para apostar en Perú.
const MinAge = 18

// DocumentType es el tipo de documento de identidad
type DocumentType string

const (
	DocumentDNI      DocumentType = "DNI"
	DocumentCE       DocumentType = "CE"
	DocumentPassport DocumentType = "PASSPORT"
)

// UserStatus es el estado de la cuenta
type UserStatus string

const (
	UserActive       UserStatus = "active"        // usuario activo
	UserSuspended    UserStatus = "suspended"     // usuario suspendido
	UserSelfExcluded UserStatus = "self_excluded" // usuario autoexcluido
)

// Errores del dominio. Quien llame a NewUser puede compararlos con errors.Is.
// aqui sse visualizan los mensajes
var (
	ErrInvalidDocument = errors.New("documento inválido")
	ErrInvalidName     = errors.New("nombre y apellido son obligatorios")
	ErrInvalidEmail    = errors.New("email inválido")
	ErrEmailTaken      = errors.New("el email ya está registrado")
	ErrDocumentTaken   = errors.New("el documento ya tiene una cuenta")
	ErrInvalidPhone    = errors.New("teléfono inválido")
	ErrUnderage        = errors.New("debes ser mayor de 18 años")
)

var (
	dniRegex      = regexp.MustCompile(`^\d{8}$`)
	foreignDocRgx = regexp.MustCompile(`^[A-Za-z0-9]{6,12}$`)
	phoneRegex    = regexp.MustCompile(`^9\d{8}$`)
)

// User representa a un usuario de Golazo.
type User struct {
	ID             string
	DocumentType   DocumentType
	DocumentNumber string
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Email          string
	Phone          string
	PasswordHash   string
	Status         UserStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewUserParams son los datos que llegan para registrar un usuario.
type NewUserParams struct {
	DocumentType   DocumentType
	DocumentNumber string
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Email          string
	Phone          string
}

// NewUser valida los datos y crea un usuario. Es la única forma correcta de construir un Usuario nuevo.
func NewUser(p NewUserParams, now time.Time) (User, error) {
	docNumber := strings.TrimSpace(string(p.DocumentNumber))
	if err := validateDocument(p.DocumentType, docNumber); err != nil {
		return User{}, err
	}

	firstName := strings.TrimSpace(p.FirstName)
	lastName := strings.TrimSpace(p.LastName)
	if firstName == "" || lastName == "" {
		return User{}, ErrInvalidName
	}

	email := strings.ToLower(strings.TrimSpace(p.Email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return User{}, ErrInvalidEmail
	}

	phone := strings.TrimSpace(p.Phone)
	if !phoneRegex.MatchString(phone) {
		return User{}, ErrInvalidPhone
	}

	if !isAdult(p.BirthDate, now) {
		return User{}, ErrUnderage
	}

	return User{
		DocumentType:   p.DocumentType,
		DocumentNumber: docNumber,
		FirstName:      firstName,
		LastName:       lastName,
		BirthDate:      p.BirthDate,
		Email:          email,
		Phone:          phone,
		Status:         UserActive,
	}, nil

}

// Fnc: funcion para validar Documento
func validateDocument(docType DocumentType, number string) error {
	switch docType {
	case DocumentDNI:
		if !dniRegex.MatchString(number) {
			return ErrInvalidDocument
		}
	case DocumentCE, DocumentPassport:
		if !foreignDocRgx.MatchString(number) {
			return ErrInvalidDocument
		}
	default:
		return ErrInvalidDocument
	}
	return nil
}

func isAdult(birthDate, now time.Time) bool {
	return !birthDate.AddDate(MinAge, 0, 0).After(now)
}
