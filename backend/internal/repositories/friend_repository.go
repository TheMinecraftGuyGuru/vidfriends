package repositories

import (
	"context"

	"github.com/vidfriends/backend/internal/models"
)

// FriendRepository defines data access for friend requests and relationships.
type FriendRepository interface {
	CreateRequest(ctx context.Context, request models.FriendRequest) error
	ListForUser(ctx context.Context, userID string) ([]models.FriendRequest, error)
	UpdateStatus(ctx context.Context, requestID, status string) error
}
