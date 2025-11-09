// services/video-service/internal/repository/interfaces.go
package repository

import (
	"context"
	"jcloud-project/video-service/internal/domain"
)

type VideoRepository interface {
	Create(ctx context.Context, video *domain.Video) error
}
