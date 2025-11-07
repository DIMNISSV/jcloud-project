// internal/domain/video.go
package domain

import "time"

//
// Video Domain Model
//

type Video struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      int64     `json:"user_id"` // Owner of the video
	FilePath    string    `json:"-"`       // Path in the file storage, hidden from public JSON
	Status      string    `json:"status"`  // e.g., "PENDING", "PROCESSING", "PUBLISHED"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
