// services/user-service/internal/repository/postgres_user_repository.go
package repository

import (
	"context"
	"errors"
	"jcloud-project/libs/go-common/ierr"
	"jcloud-project/user-service/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userPostgresRepository struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepository(db *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	// По умолчанию роль 'USER'
	if user.Role == "" {
		user.Role = "USER"
	}
	err := r.db.QueryRow(ctx, query, user.Email, user.Password, user.Role).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Проверяем на ошибку дублирования email
		if strings.Contains(err.Error(), "unique constraint") {
			return ierr.ErrConflict
		}
		return err
	}
	return nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET email = $1, role = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, user.Email, user.Role, time.Now(), user.ID)
	return err
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	var u domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ierr.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userPostgresRepository) FindAll(ctx context.Context) ([]domain.UserPublic, error) {
	query := `SELECT id, email, role, created_at, updated_at FROM users ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.UserPublic
	for rows.Next() {
		var u domain.UserPublic
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
