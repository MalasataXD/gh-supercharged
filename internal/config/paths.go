package config

import (
	"os"
	"path/filepath"

	"github.com/cli/go-gh/v2/pkg/config"
)

// ConfigDir returns the directory where config.json and projects.json live.
// Override with GH_SC_CONFIG_DIR.
func ConfigDir() (string, error) {
	if override := os.Getenv("GH_SC_CONFIG_DIR"); override != "" {
		return override, nil
	}
	return filepath.Join(config.ConfigDir(), "extensions", "supercharged"), nil
}
