package models

import "time"

// User represents an account within the VidFriends platform.
type User struct {
	ID        string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FriendRequest represents the invitation workflow between two users.
type FriendRequest struct {
	ID          string
	Requester   string
	Receiver    string
	Status      string
	CreatedAt   time.Time
	RespondedAt *time.Time
}

// VideoShare stores references to a shared video along with cached metadata.
type VideoShare struct {
	ID          string
	OwnerID     string
	URL         string
	Title       string
	Description string
	Thumbnail   string
	CreatedAt   time.Time
	AssetURL    string
	AssetStatus string
	AssetSize   int64
}

const (
	AssetStatusPending = "pending"
	AssetStatusReady   = "ready"
	AssetStatusFailed  = "failed"
)

// SessionTokens groups the bearer credentials issued to authenticated users.
type SessionTokens struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}
