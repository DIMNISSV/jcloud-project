// services/video-service/internal/service/video_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"jcloud-project/libs/go-common/ierr"
	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/video-service/internal/domain"
	"jcloud-project/video-service/internal/repository"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type VideoService interface {
	ProcessNewVideoUpload(ctx context.Context, claims *commontypes.JwtCustomClaims, title, description string, fileHeader *multipart.FileHeader) (*domain.Video, error)
}

type videoService struct {
	repo repository.VideoRepository
}

func NewVideoService(repo repository.VideoRepository) VideoService {
	return &videoService{repo: repo}
}

// ProcessNewVideoUpload handles the business logic of saving a video file and creating a DB record.
func (s *videoService) ProcessNewVideoUpload(ctx context.Context, claims *commontypes.JwtCustomClaims, title, description string, fileHeader *multipart.FileHeader) (*domain.Video, error) {
	// Шаг 1: Проверка прав доступа из JWT
	maxSizeMb, ok := claims.Permissions["max_upload_size_mb"].(float64)
	if !ok {
		return nil, fmt.Errorf("permission 'max_upload_size_mb' is missing: %w", ierr.ErrForbidden)
	}

	if fileHeader.Size > int64(maxSizeMb*1024*1024) {
		return nil, fmt.Errorf("file size exceeds the allowed limit of %.0f MB: %w", maxSizeMb, ierr.ErrForbidden)
	}

	// Шаг 2: Сохранение файла (временная логика)
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	uniqueFileName := fmt.Sprintf("%d-%d-%s", claims.UserID, time.Now().UnixNano(), fileHeader.Filename)
	uploadPath := filepath.Join("/uploads", uniqueFileName)

	dst, err := os.Create(uploadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to save file content: %w", err)
	}

	// Шаг 3: Создание записи в БД
	video := &domain.Video{
		Title:       title,
		Description: description,
		UserID:      claims.UserID,
		FilePath:    uploadPath,
		Status:      "PENDING",
	}

	if err := s.repo.Create(ctx, video); err != nil {
		_ = os.Remove(uploadPath)
		return nil, fmt.Errorf("failed to create video record in database: %w", err)
	}

	return video, nil
}
