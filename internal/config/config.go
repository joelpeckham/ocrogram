//go:build darwin

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
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: home: %w", err)
	}
	return filepath.Join(home, ".config", "ocrogram", "config.toml"), nil
}

// Load reads the default config path. A missing file is not an error;
// discovered screenshot-folder defaults are returned instead.
func Load() (Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return Config{}, err
	}
	return LoadFrom(path)
}

// LoadFrom reads config from path.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		dir, derr := DiscoverScreenshotDir()
		if derr != nil {
			return Config{}, derr
		}
		dir, derr = resolveDir(dir)
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

	dir, err := resolveDir(cfg.ScreenshotDir)
	if err != nil {
		return Config{}, err
	}
	cfg.ScreenshotDir = dir
	return cfg, nil
}

// Save writes the default config path.
func Save(cfg Config) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return SaveTo(path, cfg)
}

// NormalizeDir expands ~, makes path absolute, and checks that it is a directory.
func NormalizeDir(p string) (string, error) {
	abs, err := resolveDir(p)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("config: screenshot dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("config: screenshot dir: not a directory: %s", abs)
	}
	return abs, nil
}

func resolveDir(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("config: screenshot dir is empty")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: home: %w", err)
	}
	p = expandHome(p, home)
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("config: abs: %w", err)
	}
	return filepath.Clean(abs), nil
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
