package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/config"
	"github.com/vidfriends/backend/internal/db"
	"github.com/vidfriends/backend/internal/handlers"
	"github.com/vidfriends/backend/internal/repositories"
	"github.com/vidfriends/backend/internal/storage"
	"github.com/vidfriends/backend/internal/videos"
)

// buildDependencies wires together concrete implementations used by the HTTP handlers.
func buildDependencies(ctx context.Context, pool db.Pool, cfg config.Config) (handlers.Dependencies, func(context.Context) error, error) {
	ytDlp := videos.NewYTDLPProvider(cfg.YTDLPPath, cfg.YTDLPTimeout)
	metadataProvider := videos.NewCachingProvider(ytDlp, cfg.MetadataCacheTTL)
	sessionStore := repositories.NewPostgresSessionStore(pool)
	videoRepo := repositories.NewPostgresVideoRepository(pool)

	objectStore, err := storage.NewS3Storage(ctx, cfg.ObjectStore)
	if err != nil {
		return handlers.Dependencies{}, nil, fmt.Errorf("configure object storage: %w", err)
	}

	assetIngestor := videos.NewAssetIngestor(ytDlp, objectStore, videoRepo, videos.AssetIngestorConfig{
		QueueSize: 32,
		Workers:   2,
	}, slog.Default())

	deps := handlers.Dependencies{
		Users:         repositories.NewPostgresUserRepository(pool),
		Sessions:      auth.NewManager(15*time.Minute, 24*time.Hour, sessionStore),
		Friends:       repositories.NewPostgresFriendRepository(pool),
		Videos:        videoRepo,
		VideoMetadata: metadataProvider,
		VideoAssets:   assetIngestor,
	}

	cleanup := func(shutdownCtx context.Context) error {
		if assetIngestor == nil {
			return nil
		}
		return assetIngestor.Shutdown(shutdownCtx)
	}

	return deps, cleanup, nil
}
