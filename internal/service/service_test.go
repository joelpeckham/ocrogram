//go:build darwin

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlistContainsLabelAndPaths(t *testing.T) {
	body := Plist("/opt/homebrew/bin/ocrogram", "/Users/ada/Library/Logs/ocrogram.log")
	for _, want := range []string{
		Label,
		"/opt/homebrew/bin/ocrogram",
		"daemon",
		"/Users/ada/Library/Logs/ocrogram.log",
		"<key>KeepAlive</key>",
		"<key>SuccessfulExit</key>",
		"<false/>",
		"<key>RunAtLoad</key>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("plist missing %q", want)
		}
	}
	if strings.Contains(body, "<key>KeepAlive</key>\n	<true/>") {
		t.Fatal("KeepAlive should not be a bare true")
	}
}

func TestIsNotLoaded(t *testing.T) {
	if !isNotLoaded("Boot-out failed: 3: No such process", fmt.Errorf("exit status 5")) {
		t.Fatal("expected no such process to be ignored")
	}
	if !isNotLoaded("Could not find service", nil) {
		t.Fatal("expected could not find to be ignored")
	}
	if isNotLoaded("permission denied", fmt.Errorf("exit status 1")) {
		t.Fatal("permission error should not be ignored")
	}
}

func TestStableExecutablePrefersHomebrewPrefix(t *testing.T) {
	dir := t.TempDir()
	cellar := filepath.Join(dir, "Cellar", "ocrogram", "0.1.0", "bin")
	prefixBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(cellar, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prefixBin, 0o755); err != nil {
		t.Fatal(err)
	}
	cellarExe := filepath.Join(cellar, "ocrogram")
	prefixExe := filepath.Join(prefixBin, "ocrogram")
	if err := os.WriteFile(cellarExe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prefixExe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := stableExecutable(cellarExe)
	want, err := filepath.EvalSymlinks(prefixExe)
	if err != nil {
		want = prefixExe
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPlistEscapesXML(t *testing.T) {
	body := Plist("/tmp/a&b/ocrogram", "/tmp/x<y>")
	if strings.Contains(body, "/tmp/a&b/") {
		t.Fatal("unescaped & in exe path")
	}
	if !strings.Contains(body, "/tmp/a&amp;b/ocrogram") {
		t.Fatal("expected escaped exe path")
	}
	if !strings.Contains(body, "/tmp/x&lt;y&gt;") {
		t.Fatal("expected escaped log path")
	}
}
