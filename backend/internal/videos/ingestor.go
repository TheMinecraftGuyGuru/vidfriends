package videos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/vidfriends/backend/internal/models"
)

// ShareAssetUpdater persists ingestion status updates for video shares.
type ShareAssetUpdater interface {
	MarkAssetReady(ctx context.Context, shareID, location string, size int64) error
	MarkAssetFailed(ctx context.Context, shareID string) error
}

// AssetIngestorConfig controls the concurrency characteristics of the ingestor.
type AssetIngestorConfig struct {
	QueueSize int
	Workers   int
}

// AssetIngestor asynchronously persists downloaded video assets using yt-dlp.
type AssetIngestor struct {
	provider *YTDLPProvider
	storage  AssetStorage
	updater  ShareAssetUpdater
	logger   *slog.Logger

	jobs   chan ingestJob
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

type ingestJob struct {
	share models.VideoShare
}

var errIngestorClosed = errors.New("asset ingestor closed")

// NewAssetIngestor constructs a background worker that persists assets.
func NewAssetIngestor(provider *YTDLPProvider, storage AssetStorage, updater ShareAssetUpdater, cfg AssetIngestorConfig, logger *slog.Logger) *AssetIngestor {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 16
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ing := &AssetIngestor{
		provider: provider,
		storage:  storage,
		updater:  updater,
		logger:   logger,
		jobs:     make(chan ingestJob, cfg.QueueSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	ing.wg.Add(cfg.Workers)
	for i := 0; i < cfg.Workers; i++ {
		go ing.worker()
	}

	return ing
}

// Enqueue schedules asset persistence for the supplied share.
func (i *AssetIngestor) Enqueue(ctx context.Context, share models.VideoShare) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-i.ctx.Done():
		return errIngestorClosed
	default:
	}

	job := ingestJob{share: share}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-i.ctx.Done():
		return errIngestorClosed
	case i.jobs <- job:
		return nil
	}
}

// Shutdown waits for the worker pool to drain outstanding jobs.
func (i *AssetIngestor) Shutdown(ctx context.Context) error {
	i.once.Do(func() {
		i.cancel()
		close(i.jobs)
	})

	done := make(chan struct{})
	go func() {
		i.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (i *AssetIngestor) worker() {
	defer i.wg.Done()

	for {
		select {
		case <-i.ctx.Done():
			return
		case job, ok := <-i.jobs:
			if !ok {
				return
			}
			i.handleJob(job)
		}
	}
}

func (i *AssetIngestor) handleJob(job ingestJob) {
	if i.provider == nil || i.storage == nil || i.updater == nil {
		i.logger.Error("asset ingestor missing dependencies", "hasProvider", i.provider != nil, "hasStorage", i.storage != nil, "hasUpdater", i.updater != nil)
		return
	}

	fetchCtx, cancel := context.WithTimeout(context.Background(), maxDuration(2*i.provider.Timeout, 2*time.Minute))
	defer cancel()

	prefixed := &prefixedStorage{prefix: job.share.ID, base: i.storage}
	_, assets, err := i.provider.Fetch(fetchCtx, job.share.URL, FetchOptions{DownloadVideo: true, Storage: prefixed})
	if err != nil {
		i.logger.Error("asset ingestion failed", "shareId", job.share.ID, "url", job.share.URL, "error", err)
		i.recordFailure(job.share.ID)
		return
	}

	var videoAsset *DownloadedAsset
	for idx := range assets {
		if assets[idx].Type == AssetTypeVideo {
			videoAsset = &assets[idx]
			break
		}
	}

	if videoAsset == nil {
		i.logger.Error("yt-dlp did not produce a video asset", "shareId", job.share.ID)
		i.recordFailure(job.share.ID)
		return
	}

	if err := i.recordSuccess(job.share.ID, videoAsset.Location, videoAsset.Size); err != nil {
		i.logger.Error("mark asset ready", "shareId", job.share.ID, "error", err)
		i.recordFailure(job.share.ID)
	}
}

func (i *AssetIngestor) recordFailure(shareID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := i.updater.MarkAssetFailed(ctx, shareID); err != nil {
		i.logger.Error("record asset failure", "shareId", shareID, "error", err)
	}
}

func (i *AssetIngestor) recordSuccess(shareID, location string, size int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return i.updater.MarkAssetReady(ctx, shareID, location, size)
}

type prefixedStorage struct {
	prefix string
	base   AssetStorage
}

func (p *prefixedStorage) Save(ctx context.Context, name string, r io.Reader) (string, error) {
	if p.base == nil {
		return "", fmt.Errorf("prefix storage: %w", ErrAssetStorageUnavailable)
	}
	key := path.Join(p.prefix, name)
	if strings.TrimSpace(key) == "" {
		return "", errors.New("prefix storage: empty key")
	}
	return p.base.Save(ctx, key, r)
}

func maxDuration(a, b time.Duration) time.Duration {
	if a >= b {
		return a
	}
	return b
}
