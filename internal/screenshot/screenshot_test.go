package screenshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMatchesName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"Screenshot 2026-08-24 at 10.10.00 AM.png", true},
		{"Screen Shot 2020-01-01 at 1.00.00 PM.png", true},
		{"Bildschirmfoto 2026-01-01.png", true},
		{"Capture d’écran 2026-01-01.png", true},
		{"スクリーンショット 2026-01-01.png", true},
		{"random.png", false},
		{"notes.txt", false},
	}
	for _, tt := range tests {
		if got := MatchesName(tt.name); got != tt.want {
			t.Errorf("MatchesName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsImage(t *testing.T) {
	if !IsImage("a.PNG") || !IsImage("a.jpg") || IsImage("a.txt") {
		t.Fatal("IsImage mismatch")
	}
}

func TestIsDotfile(t *testing.T) {
	if !IsDotfile("/tmp/.hidden.png") || IsDotfile("/tmp/Screenshot.png") {
		t.Fatal("IsDotfile mismatch")
	}
}

func TestIsScreenshotNameAndEmpty(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "Screenshot empty.png")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsScreenshot(empty) {
		t.Fatal("empty file should not be a screenshot")
	}

	ok := filepath.Join(dir, "Screenshot ok.png")
	if err := os.WriteFile(ok, []byte("not-a-real-png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsScreenshot(ok) {
		t.Fatal("named screenshot with size should match")
	}

	other := filepath.Join(dir, "notes.png")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsScreenshot(other) {
		t.Fatal("unrelated png without metadata should not match")
	}
}

func TestSnapshot(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.png")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Snapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got[file]; !ok {
		t.Fatalf("missing %s: %#v", file, got)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestSettlerDebounce(t *testing.T) {
	s := NewSettler(40 * time.Millisecond)
	defer s.Stop()

	got := make(chan string, 2)
	s.Touch("/tmp/a.png", func(p string) { got <- p })
	s.Touch("/tmp/a.png", func(p string) { got <- p })

	select {
	case p := <-got:
		if p != "/tmp/a.png" {
			t.Fatalf("got %q", p)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for settle")
	}

	select {
	case <-got:
		t.Fatal("settled twice")
	case <-time.After(80 * time.Millisecond):
	}
}
