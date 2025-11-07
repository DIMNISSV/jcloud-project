// internal/repository/user_postgres.go
package repository

import (
	"context"
	"jcloud-project/user-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

//
// User Repository Interface
//

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
}

//
// Postgres Implementation
//

type userPostgresRepository struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepository(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// pgx.ErrNoRows is a common error we should handle gracefully
		return nil, err
	}

	return user, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, created_at, updated_at FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
