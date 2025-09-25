package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/config"
	"github.com/vidfriends/backend/internal/db"
	"github.com/vidfriends/backend/internal/handlers"
	"github.com/vidfriends/backend/internal/httpserver"
	"github.com/vidfriends/backend/internal/repositories"
	"github.com/vidfriends/backend/internal/videos"
)

// Run bootstraps the VidFriends backend application.
func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("expected command: serve or migrate")
	}

	switch args[0] {
	case "serve":
		return serve(ctx)
	case "migrate":
		return runMigrations(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func serve(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	ytDlp := videos.NewYTDLPProvider(cfg.YTDLPPath, cfg.YTDLPTimeout)
	metadataProvider := videos.NewCachingProvider(ytDlp, cfg.MetadataCacheTTL)

	deps := handlers.Dependencies{
		Users:         repositories.NewPostgresUserRepository(pool),
		Sessions:      auth.NewManager(15*time.Minute, 24*time.Hour),
		Friends:       repositories.NewPostgresFriendRepository(pool),
		Videos:        repositories.NewPostgresVideoRepository(pool),
		VideoMetadata: metadataProvider,
	}

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, deps)

	srv := httpserver.New(cfg.AppPort, mux)

	logger.Info("starting http server", "port", cfg.AppPort)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Start()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("context canceled, shutting down server")
	case sig := <-signalCh:
		logger.Info("received signal, shutting down", "signal", sig.String())
	case err := <-srvErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), httpserver.ShutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

func runMigrations(ctx context.Context, args []string) error {
	_ = ctx
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Printf("running migrations in %s with args %v\n", cfg.MigrationDir, args)
	return nil
}
