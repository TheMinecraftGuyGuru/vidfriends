package videos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
		Args:    []string{"--dump-single-json", "--no-warnings", "--no-playlist", "--skip-download"},
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
	args = append(args, url)

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

func defaultCommandRunner(ctx context.Context, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	return cmd.Output()
}
