package repositories

import (
	"context"

	"github.com/vidfriends/backend/internal/models"
)

// VideoRepository exposes data access for shared videos.
type VideoRepository interface {
	Create(ctx context.Context, share models.VideoShare) error
	ListFeed(ctx context.Context, userID string) ([]models.VideoShare, error)
}
