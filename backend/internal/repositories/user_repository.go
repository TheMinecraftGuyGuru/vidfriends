package repositories

import (
	"context"

	"github.com/vidfriends/backend/internal/models"
)

// UserRepository defines the data access contract for users.
type UserRepository interface {
	Create(ctx context.Context, user models.User) error
	FindByEmail(ctx context.Context, email string) (models.User, error)
	Update(ctx context.Context, user models.User) error
}
