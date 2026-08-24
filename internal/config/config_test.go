package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	want := Config{ScreenshotDir: "/tmp/shots"}
	if err := SaveTo(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ScreenshotDir != want.ScreenshotDir {
		t.Fatalf("ScreenshotDir = %q, want %q", got.ScreenshotDir, want.ScreenshotDir)
	}
}

func TestLoadFromMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ScreenshotDir == "" {
		t.Fatal("expected discovered screenshot dir")
	}
}

func TestLoadFromEmptyDirUsesDiscovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("screenshot_dir = \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ScreenshotDir == "" {
		t.Fatal("expected discovered screenshot dir")
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		in, home, want string
	}{
		{"~", "/Users/ada", "/Users/ada"},
		{"~/Desktop", "/Users/ada", "/Users/ada/Desktop"},
		{"/tmp/shots", "/Users/ada", "/tmp/shots"},
	}
	for _, tt := range tests {
		if got := expandHome(tt.in, tt.home); got != tt.want {
			t.Errorf("expandHome(%q, %q) = %q, want %q", tt.in, tt.home, got, tt.want)
		}
	}
}
