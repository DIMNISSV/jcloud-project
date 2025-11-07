// internal/service/video_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"jcloud-project/video-service/internal/domain"
	"jcloud-project/video-service/internal/repository"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type VideoService interface {
	UploadVideo(ctx context.Context, userID int64, title, description string, fileHeader *multipart.FileHeader) (*domain.Video, error)
}

type videoService struct {
	repo repository.VideoRepository
}

func NewVideoService(repo repository.VideoRepository) VideoService {
	return &videoService{repo: repo}
}

func (s *videoService) UploadVideo(ctx context.Context, userID int64, title, description string, fileHeader *multipart.FileHeader) (*domain.Video, error) {
	// Open the file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// --- File Saving Logic ---
	// NOTE: This is a temporary local storage implementation.
	// This will be replaced with Nextcloud API calls later.

	// Create a unique filename to avoid collisions
	uniqueFileName := fmt.Sprintf("%d-%d-%s", userID, time.Now().UnixNano(), fileHeader.Filename)
	uploadPath := filepath.Join("/uploads", uniqueFileName)

	// Create the destination file
	dst, err := os.Create(uploadPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}
	// --- End of File Saving Logic ---

	video := &domain.Video{
		Title:       title,
		Description: description,
		UserID:      userID,
		FilePath:    uploadPath, // Store the path for future processing
		Status:      "PENDING",
	}

	if err := s.repo.Create(ctx, video); err != nil {
		// Here you might want to add logic to delete the saved file if DB insert fails
		return nil, err
	}

	return video, nil
}
