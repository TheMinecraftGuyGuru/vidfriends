package videos

import "errors"

var (
	// ErrProviderUnavailable indicates the metadata provider is not configured.
	ErrProviderUnavailable = errors.New("video metadata provider unavailable")
	// ErrAssetStorageUnavailable indicates persistence of downloaded media is not configured.
	ErrAssetStorageUnavailable = errors.New("video asset storage unavailable")
)
