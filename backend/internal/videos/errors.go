package videos

import "errors"

var (
	// ErrProviderUnavailable indicates the metadata provider is not configured.
	ErrProviderUnavailable = errors.New("video metadata provider unavailable")
)
