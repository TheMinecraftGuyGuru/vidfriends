package handlers

import (
	"context"

	"github.com/vidfriends/backend/internal/models"
)

// UserStore captures the persistence operations required by the auth handlers.
type UserStore interface {
	Create(ctx context.Context, user models.User) error
	FindByEmail(ctx context.Context, email string) (models.User, error)
}

// FriendStore captures operations required by the friend handlers.
type FriendStore interface {
	CreateRequest(ctx context.Context, request models.FriendRequest) error
	ListForUser(ctx context.Context, userID string) ([]models.FriendRequest, error)
}

// VideoStore captures persistence for video sharing workflows.
type VideoStore interface {
	Create(ctx context.Context, share models.VideoShare) error
	ListFeed(ctx context.Context, userID string) ([]models.VideoShare, error)
}
