package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

//código de error de Postgres para "valor duplicado en un UNIQUE"
const pgUniqueViolation = "23505"

//UserPostgres guarda usuarios en PostgreSQL
type UserPostgres struct {
	pool *pgxpool.Pool
}

// NewUserPostgres crea el repositorio con el pool de conexio.
func NewUserPostgres(pool *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{pool: pool}
}

const insertUserSQL = `
	INSERT INTO users (document_type, document_number, firts_name, last_name, birth_date, email, phone, passhord_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, status, created_at, updated_at
`

// Create inserta el usuario y lo devuelve con los datos que generó Postgres
func (r *UserPostgres) Create(ctx context.Context, u domain.User) (domain.User, error) {
	var status string

	err := r.pool.QueryRow(ctx, insertUserSQL,
		string(u.DocumentType),
		u.DocumentNumber,
		u.FirstName,
		u.LastName,
		u.BirthDate,
		u.Email,
		u.Phone,
		u.PasswordHash,
	).Scan(&u.ID, &status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	u.Status = domain.UserStatus(status)
	return u, nil
}

// mapUserError traduceerrores de Postgres a errores del dominio.
func mapUserError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "users_email_unique":
			return domain.ErrEmailTaken
		case "users_document_unique": 
			return domain.ErrEmailTaken
		}
	}

	return fmt.Errorf("insertar usuario: %w", err)
} 


