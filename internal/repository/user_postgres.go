package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Perseo-2025/golazo-api/internal/domain"
)

// código de error de Postgres para "valor duplicado en un UNIQUE"
const pgUniqueViolation = "23505"

// UserPostgres guarda usuarios en PostgreSQL
type UserPostgres struct {
	pool *pgxpool.Pool
}

// NewUserPostgres crea el repositorio con el pool de conexio.
func NewUserPostgres(pool *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{pool: pool}
}

// 1. insert
const insertUserSQL = `
	INSERT INTO users (document_type, document_number, first_name, last_name, birth_date, email, phone, password_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, status, created_at, updated_at
`
// 2. insert
const insertWalletSQL = `
	INSERT INTO wallets (user_id, type, currency)
	VALUES  ($1, 'real', 'PEN') 
`

// Create inserta el usuario y lo devuelve con los datos que generó Postgres
func (r *UserPostgres) Create(ctx context.Context, u domain.User) (domain.User, error) {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var status string
		err := tx.QueryRow(ctx, insertUserSQL,
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
			return mapUserError(err)
		}
		u.Status = domain.UserStatus(status)

		if _, err := tx.Exec(ctx, insertWalletSQL, u.ID); err != nil {
			return fmt.Errorf("crear billetera: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.User{}, err
	}
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
