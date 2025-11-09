// internal/repository/user_postgres.go
package repository

import (
	"context"
	"errors"
	"jcloud-project/user-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//
// User Repository Interface
//

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindAll(ctx context.Context) ([]domain.UserPublic, error)
	Update(ctx context.Context, user *domain.User) error
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
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
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
	query := `SELECT id, email, role, created_at, updated_at FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userPostgresRepository) FindAll(ctx context.Context) ([]domain.UserPublic, error) {
	query := `SELECT id, email, role, created_at, updated_at FROM users ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Теперь мы сопоставляем результат с нашей новой, безопасной структурой
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.UserPublic])
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, role = $2, updated_at = NOW()
		WHERE id = $3
	`
	tag, err := r.db.Exec(ctx, query, user.Email, user.Role, user.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
