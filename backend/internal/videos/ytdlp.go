package videos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandRunner executes external commands and returns stdout bytes.
type CommandRunner func(ctx context.Context, binary string, args ...string) ([]byte, error)

// YTDLPProvider fetches metadata using the yt-dlp CLI tool.
type YTDLPProvider struct {
	Binary  string
	Args    []string
	Run     CommandRunner
	Timeout time.Duration
}

// AssetType identifies the type of media that was downloaded by yt-dlp.
type AssetType string

const (
	// AssetTypeVideo represents the primary video file for a share.
	AssetTypeVideo AssetType = "video"
)

// AssetStorage persists downloaded media to durable storage (filesystem, S3, etc).
type AssetStorage interface {
	Save(ctx context.Context, name string, r io.Reader) (string, error)
}

// DownloadedAsset captures information about a media file that was persisted after
// being downloaded by yt-dlp.
type DownloadedAsset struct {
	Type     AssetType
	Location string
	Name     string
	Size     int64
}

// FetchOptions configure how metadata lookup should behave.
type FetchOptions struct {
	// DownloadVideo toggles whether the full video file should be downloaded.
	DownloadVideo bool
	// Storage specifies where downloaded assets should be persisted. It is
	// required when DownloadVideo is true.
	Storage AssetStorage
}

// NewYTDLPProvider constructs a Provider that shells out to yt-dlp.
func NewYTDLPProvider(binary string, timeout time.Duration) *YTDLPProvider {
	if strings.TrimSpace(binary) == "" {
		binary = "yt-dlp"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &YTDLPProvider{
		Binary:  binary,
		Args:    []string{"--dump-single-json", "--no-warnings", "--no-playlist"},
		Run:     defaultCommandRunner,
		Timeout: timeout,
	}
}

// Lookup executes yt-dlp for the provided URL and parses the JSON response.
func (p *YTDLPProvider) Lookup(ctx context.Context, url string) (Metadata, error) {
	if p == nil {
		return Metadata{}, ErrProviderUnavailable
	}
	if p.Run == nil {
		p.Run = defaultCommandRunner
	}

	execCtx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	args := append([]string{}, p.Args...)
	args = append(args, "--skip-download", url)

	out, err := p.Run(execCtx, p.Binary, args...)
	if err != nil {
		return Metadata{}, fmt.Errorf("yt-dlp fetch: %w", err)
	}

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   string `json:"thumbnail"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return Metadata{}, fmt.Errorf("parse yt-dlp response: %w", err)
	}

	if payload.Title == "" && payload.Description == "" && payload.Thumbnail == "" {
		return Metadata{}, errors.New("yt-dlp returned empty metadata")
	}

	return Metadata{
		Title:       payload.Title,
		Description: payload.Description,
		Thumbnail:   payload.Thumbnail,
	}, nil
}

// Fetch resolves metadata for the supplied URL and, when configured, downloads
// the primary video asset and persists it using the provided storage backend.
func (p *YTDLPProvider) Fetch(ctx context.Context, url string, opts FetchOptions) (Metadata, []DownloadedAsset, error) {
	if p == nil {
		return Metadata{}, nil, ErrProviderUnavailable
	}
	if p.Run == nil {
		p.Run = defaultCommandRunner
	}
	if opts.DownloadVideo && opts.Storage == nil {
		return Metadata{}, nil, fmt.Errorf("yt-dlp fetch: %w", ErrAssetStorageUnavailable)
	}

	execCtx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	args := append([]string{}, p.Args...)
	if !opts.DownloadVideo {
		args = append(args, "--skip-download")
	}
	args = append(args, url)

	out, err := p.Run(execCtx, p.Binary, args...)
	if err != nil {
		return Metadata{}, nil, fmt.Errorf("yt-dlp fetch: %w", err)
	}

	var payload struct {
		Title              string `json:"title"`
		Description        string `json:"description"`
		Thumbnail          string `json:"thumbnail"`
		RequestedDownloads []struct {
			Filepath string `json:"filepath"`
			Filename string `json:"filename"`
			Filesize int64  `json:"filesize"`
		} `json:"requested_downloads"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return Metadata{}, nil, fmt.Errorf("parse yt-dlp response: %w", err)
	}

	if payload.Title == "" && payload.Description == "" && payload.Thumbnail == "" {
		return Metadata{}, nil, errors.New("yt-dlp returned empty metadata")
	}

	metadata := Metadata{
		Title:       payload.Title,
		Description: payload.Description,
		Thumbnail:   payload.Thumbnail,
	}

	if !opts.DownloadVideo {
		return metadata, nil, nil
	}

	if len(payload.RequestedDownloads) == 0 {
		return metadata, nil, errors.New("yt-dlp did not return download metadata")
	}

	var assets []DownloadedAsset
	for _, item := range payload.RequestedDownloads {
		localPath := item.Filepath
		if localPath == "" {
			localPath = item.Filename
		}
		if strings.TrimSpace(localPath) == "" {
			return metadata, nil, errors.New("yt-dlp provided empty download path")
		}

		// Normalize to absolute paths to avoid surprises when opening files.
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(".", localPath)
		}

		f, err := os.Open(localPath)
		if err != nil {
			return metadata, nil, fmt.Errorf("open downloaded asset: %w", err)
		}

		name := filepath.Base(localPath)
		location, persistErr := opts.Storage.Save(ctx, name, f)
		closeErr := f.Close()
		removeErr := os.Remove(localPath)

		if persistErr != nil {
			return metadata, nil, fmt.Errorf("persist asset %s: %w", name, persistErr)
		}
		if closeErr != nil {
			return metadata, nil, fmt.Errorf("close asset %s: %w", name, closeErr)
		}
		if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return metadata, nil, fmt.Errorf("cleanup asset %s: %w", name, removeErr)
		}

		assets = append(assets, DownloadedAsset{
			Type:     AssetTypeVideo,
			Location: location,
			Name:     name,
			Size:     item.Filesize,
		})
	}

	return metadata, assets, nil
}

func defaultCommandRunner(ctx context.Context, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	return cmd.Output()
}
