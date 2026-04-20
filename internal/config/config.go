package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed default.json
var defaultJSON []byte

type Config struct {
	GithubHandle     string `json:"github_handle"`
	DigestWindowDays int    `json:"digest_window_days"`
	StandupFormat    string `json:"standup_format"`
}

// ErrFirstRun is returned when config.json doesn't exist or github_handle is unset.
var ErrFirstRun = errors.New("first run")

// Load reads config.json. On first run, creates it from the embedded default
// and returns ErrFirstRun so callers can print a setup message and exit.
func Load() (*Config, string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, "", fmt.Errorf("config dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, "", fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if writeErr := os.WriteFile(path, defaultJSON, 0o600); writeErr != nil {
			return nil, path, fmt.Errorf("write default config: %w", writeErr)
		}
		return nil, path, ErrFirstRun
	}
	if err != nil {
		return nil, path, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, path, fmt.Errorf("parse config: %w", err)
	}
	if cfg.GithubHandle == "" || cfg.GithubHandle == "your-github-username" {
		return nil, path, fmt.Errorf("%w: set github_handle in %s", ErrFirstRun, path)
	}
	return &cfg, path, nil
}
