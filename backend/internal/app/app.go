package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/vidfriends/backend/internal/config"
	"github.com/vidfriends/backend/internal/db"
	"github.com/vidfriends/backend/internal/handlers"
	"github.com/vidfriends/backend/internal/httpserver"
	"github.com/vidfriends/backend/internal/middleware"
)

// Run bootstraps the VidFriends backend application.
func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("expected command: serve, migrate, or seed")
	}

	switch args[0] {
	case "serve":
		return serve(ctx)
	case "migrate":
		return runMigrations(ctx, args[1:])
	case "seed":
		return runSeed(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func serve(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	deps := buildDependencies(pool, cfg)

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, deps)

	handler := middleware.RequestLogger(logger)(mux)

	srv := httpserver.New(cfg.AppPort, handler)

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
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	command := "up"
	if len(args) > 0 {
		command = args[0]
	}

	migrationDir := cfg.MigrationDir
	if !filepath.IsAbs(migrationDir) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		migrationDir = filepath.Join(wd, migrationDir)
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var migrations []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		migrations = append(migrations, entry.Name())
	}

	sort.Strings(migrations)

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
                version TEXT PRIMARY KEY,
                applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )`); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("fetch applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate applied migrations: %w", err)
	}

	switch command {
	case "status":
		for _, name := range migrations {
			if _, ok := applied[name]; ok {
				fmt.Printf("[x] %s\n", name)
			} else {
				fmt.Printf("[ ] %s\n", name)
			}
		}
		return nil
	case "up", "":
		if len(migrations) == 0 {
			fmt.Println("no migrations to apply")
			return nil
		}

		for _, name := range migrations {
			if _, ok := applied[name]; ok {
				continue
			}

			path := filepath.Join(migrationDir, name)
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read migration %s: %w", name, err)
			}

			tx, err := conn.Begin(ctx)
			if err != nil {
				return fmt.Errorf("begin migration transaction for %s: %w", name, err)
			}

			if _, err := tx.Exec(ctx, string(contents)); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("apply migration %s: %w", name, err)
			}

			if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("record migration %s: %w", name, err)
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("commit migration %s: %w", name, err)
			}

			fmt.Printf("applied migration %s\n", name)
		}
		return nil
	case "down":
		return errors.New("down migrations are not supported yet")
	default:
		return fmt.Errorf("unknown migrate command %q", command)
	}
}

func runSeed(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("expected seed name (e.g. dev)")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	seedDir := cfg.SeedDir
	if !filepath.IsAbs(seedDir) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		seedDir = filepath.Join(wd, seedDir)
	}

	seedName := args[0]
	if !strings.HasSuffix(seedName, ".sql") {
		seedName = fmt.Sprintf("%s_seed.sql", seedName)
	}

	seedPath := filepath.Join(seedDir, seedName)
	contents, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("read seed %s: %w", seedName, err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, string(contents)); err != nil {
		return fmt.Errorf("apply seed %s: %w", seedName, err)
	}

	fmt.Printf("applied seed %s\n", seedName)
	return nil
}
