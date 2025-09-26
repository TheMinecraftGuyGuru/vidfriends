package handlers

import (
	"context"

	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/videos"
)

// UserStore captures the persistence operations required by the auth handlers.
type UserStore interface {
	Create(ctx context.Context, user models.User) error
	FindByEmail(ctx context.Context, email string) (models.User, error)
}

// SessionManager issues and refreshes authentication tokens for users.
type SessionManager interface {
	Issue(ctx context.Context, userID string) (models.SessionTokens, error)
	Refresh(ctx context.Context, refreshToken string) (models.SessionTokens, error)
}

// FriendStore captures operations required by the friend handlers.
type FriendStore interface {
	CreateRequest(ctx context.Context, request models.FriendRequest) error
	ListForUser(ctx context.Context, userID string) ([]models.FriendRequest, error)
	UpdateStatus(ctx context.Context, requestID, status string) error
}

// VideoStore captures persistence for video sharing workflows.
type VideoStore interface {
	Create(ctx context.Context, share models.VideoShare) error
	ListFeed(ctx context.Context, userID string) ([]models.VideoShare, error)
}

// VideoMetadataProvider resolves video details for shared URLs.
type VideoMetadataProvider interface {
	Lookup(ctx context.Context, url string) (videos.Metadata, error)
}

// VideoAssetIngestor schedules background persistence of video files.
type VideoAssetIngestor interface {
	Enqueue(ctx context.Context, share models.VideoShare) error
}
