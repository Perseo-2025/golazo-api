package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kingperseocs/golazo-api/internal/domain"
	"github.com/kingperseocs/golazo-api/internal/service"
)

// UserRegisterer es lo que el handler necesita del servicio.
type UserRegisterer interface{
	Register(ctx context.Context, in service.RegisterInput) (domain.User, error)
}

// UserHanlder atiende las rutas HTTP de usuarios
type UserHandler struct {
	svc UserRegisterer
	logger *slog.Logger
}

// NewUserHanlder crea el hanlder con sus dependencias.
func NewUserHandler(svc UserRegisterer, logger *slog.Logger) *UserHandler{
	return &UserHandler{svc: svc, logger: logger}
}

// registreRequest es el JSON que llega en POST /users.
type registerRequest struct {
	DocumentType 	string `json:"document_type"`
	DocumentNumber	string `json:"document_number"`
	FirstName 		string `json:"first_name"`
	LastName		string `json:"last_name"`
	BirthDate		string `json:"birth_date"`
	Email			string `json:"email"`
	Phone			string `json:"phone"`
	Password 		string `json:"password"`
}

// userResponse es lo que la API devuelve de usuario. Sin password_hash.
type userResponse struct {
	ID 				string 		`json:"id"`
	DocumentType	string 		`json:"document_type"`
	FirstName		string 		`json:"first_name"`
	LastName		string 		`json:"last_name"`
	Email 			string 		`json:"email"`
	Status			string 		`json:"status"`
	CreatedAt 		time.Time	`json:"created_at"`
}

// Register atiende POST /users.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req registerRequest 
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.logger.Warn("JSON inválido en registro", "error", err)
		writeError(w, h.logger, http.StatusBadRequest, "invalid_json", "el cuerpo debe ser un JSON válido")
		return
	}

	birthDate, err := time.Parse(time.DateOnly, req.BirthDate)
	if err != nil {
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_birth_date", "birth_date debe tener el formato AAAA-MM-DD")

	}

	user, err := h.svc.Register(r.Context(), service.RegisterInput{
		NewUserParams:  domain.NewUserParams{
			DocumentType: domain.DocumentType(req.DocumentType),
			DocumentNumber: req.DocumentNumber,
			FirstName: req.FirstName,
			LastName: req.LastName,
			BirthDate: birthDate,
			Email: req.Email,
			Phone: req.Phone,
		},
		Password: req.Password,
	})
	if err != nil {
		h.writeRegisterError(w, err)
		return
	}

	writeJSON(w, h.logger, http.StatusCreated, userResponse {
		ID:  user.ID,
		DocumentType: string(user.DocumentType),
		LastName: user.LastName,
		Email: user.Email,
		Status: string(user.Status),
		CreatedAt: user.CreatedAt,
	})

}

// imprimiendo o escribiendo los errores, writeRegisterError traduce los errores del registro a respuesta HTTP.
func (h *UserHandler) writeRegisterError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		writeError(w, h.logger, http.StatusConflict, "email_taken", err.Error())
	case errors.Is(err, domain.ErrDocumentTaken):
		writeError(w, h.logger, http.StatusConflict, "document_taken", err.Error())
	case errors.Is(err, domain.ErrInvalidDocument):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_document", err.Error())
	case errors.Is(err, domain.ErrInvalidName):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_name", err.Error())
	case errors.Is(err, domain.ErrInvalidEmail):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_email", err.Error())
	case errors.Is(err, domain.ErrInvalidPhone):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "invalid_phone", err.Error())
	case errors.Is(err, domain.ErrUnderage):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "underage", err.Error())
	case errors.Is(err, domain.ErrWeakPassword):
		writeError(w, h.logger, http.StatusUnprocessableEntity, "weak_password", err.Error())
	default:
		h.logger.Error("registro de usuario falló", "error", err)
		writeError(w, h.logger, http.StatusInternalServerError, "internal_error", "ocurrió un error interno, intenta más tarde")
	}
}