package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config is the on-disk settings persisted after TUI setup.
type Config struct {
	ScreenshotDir string `toml:"screenshot_dir"`
}

// DefaultPath is ~/.config/ocrogram/config.toml.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "ocrogram", "config.toml")
}

// Load reads the default config path. A missing file is not an error;
// discovered screenshot-folder defaults are returned instead.
func Load() (Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom reads config from path.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		dir, derr := DiscoverScreenshotDir()
		if derr != nil {
			return Config{}, derr
		}
		return Config{ScreenshotDir: dir}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("config: read: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config: parse: %w", err)
	}
	if cfg.ScreenshotDir == "" {
		dir, derr := DiscoverScreenshotDir()
		if derr != nil {
			return Config{}, derr
		}
		cfg.ScreenshotDir = dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("config: home: %w", err)
	}
	cfg.ScreenshotDir = expandHome(cfg.ScreenshotDir, home)
	return cfg, nil
}

// Save writes the default config path.
func Save(cfg Config) error {
	return SaveTo(DefaultPath(), cfg)
}

// SaveTo writes config to path, creating parent directories as needed.
func SaveTo(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: mkdir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("config: write: %w", err)
	}
	return nil
}

// DiscoverScreenshotDir returns the macOS screenshot folder, or ~/Desktop.
func DiscoverScreenshotDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: home: %w", err)
	}
	out, err := exec.Command("defaults", "read", "com.apple.screencapture", "location").Output()
	if err == nil {
		loc := strings.TrimSpace(string(out))
		if loc != "" {
			return expandHome(loc, home), nil
		}
	}
	return filepath.Join(home, "Desktop"), nil
}

func expandHome(p, home string) string {
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, strings.TrimPrefix(p, "~/"))
	}
	return p
}
