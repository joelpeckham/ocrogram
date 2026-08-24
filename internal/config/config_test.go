//go:build darwin

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

func TestNormalizeDir(t *testing.T) {
	dir := t.TempDir()
	got, err := NormalizeDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(want) {
		t.Fatalf("NormalizeDir = %q, want %q", got, want)
	}

	file := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeDir(file); err == nil {
		t.Fatal("expected error for file")
	}
	if _, err := NormalizeDir(""); err == nil {
		t.Fatal("expected error for empty")
	}
	if _, err := NormalizeDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected error for missing")
	}
}

func TestResolveDirExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveDir("~/Desktop")
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(filepath.Join(home, "Desktop"))
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(want) {
		t.Fatalf("resolveDir = %q, want %q", got, want)
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
