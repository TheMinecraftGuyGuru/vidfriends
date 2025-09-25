package app

import (
	"time"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/config"
	"github.com/vidfriends/backend/internal/db"
	"github.com/vidfriends/backend/internal/handlers"
	"github.com/vidfriends/backend/internal/repositories"
	"github.com/vidfriends/backend/internal/videos"
)

// buildDependencies wires together concrete implementations used by the HTTP handlers.
func buildDependencies(pool db.Pool, cfg config.Config) handlers.Dependencies {
	ytDlp := videos.NewYTDLPProvider(cfg.YTDLPPath, cfg.YTDLPTimeout)
	metadataProvider := videos.NewCachingProvider(ytDlp, cfg.MetadataCacheTTL)
	sessionStore := repositories.NewPostgresSessionStore(pool)

	return handlers.Dependencies{
		Users:         repositories.NewPostgresUserRepository(pool),
		Sessions:      auth.NewManager(15*time.Minute, 24*time.Hour, sessionStore),
		Friends:       repositories.NewPostgresFriendRepository(pool),
		Videos:        repositories.NewPostgresVideoRepository(pool),
		VideoMetadata: metadataProvider,
	}
}
