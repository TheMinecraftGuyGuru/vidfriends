package videos

import "context"

// Metadata captures the subset of video details used by VidFriends.
type Metadata struct {
	Title       string
	Description string
	Thumbnail   string
}

// Provider returns metadata for the supplied video URL.
type Provider interface {
	Lookup(ctx context.Context, url string) (Metadata, error)
}
