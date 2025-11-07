// internal/repository/video_postgres.go
package repository

import (
	"context"
	"jcloud-project/video-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoRepository interface {
	Create(ctx context.Context, video *domain.Video) error
}

type videoPostgresRepository struct {
	db *pgxpool.Pool
}

func NewVideoPostgresRepository(db *pgxpool.Pool) VideoRepository {
	return &videoPostgresRepository{db: db}
}

func (r *videoPostgresRepository) Create(ctx context.Context, video *domain.Video) error {
	query := `INSERT INTO videos (title, description, user_id, file_path, status)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query,
		video.Title, video.Description, video.UserID, video.FilePath, video.Status,
	).Scan(&video.ID, &video.CreatedAt, &video.UpdatedAt)
	return err
}
