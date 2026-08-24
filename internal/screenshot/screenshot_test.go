//go:build darwin

package screenshot

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/synctest"
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
		if got := matchesName(tt.name); got != tt.want {
			t.Errorf("matchesName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsImage(t *testing.T) {
	if !isImage("a.PNG") || !isImage("a.jpg") || isImage("a.txt") {
		t.Fatal("isImage mismatch")
	}
}

func TestIsDotfile(t *testing.T) {
	if !isDotfile("/tmp/.hidden.png") || isDotfile("/tmp/Screenshot.png") {
		t.Fatal("isDotfile mismatch")
	}
}

func TestClassifyNameAndEmpty(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "Screenshot empty.png")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if Classify(empty) != KindIncomplete {
		t.Fatal("empty file should not be a screenshot")
	}

	ok := filepath.Join(dir, "Screenshot ok.png")
	if err := os.WriteFile(ok, []byte("not-a-real-png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Classify(ok) != KindScreenshot {
		t.Fatal("named screenshot with size should match")
	}

	other := filepath.Join(dir, "notes.png")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Classify(other) == KindScreenshot {
		t.Fatal("unrelated png without metadata should not match")
	}
}

func TestParseScreenCaptureMetadata(t *testing.T) {
	tests := []struct {
		out  string
		err  error
		want Kind
	}{
		{"1", nil, KindScreenshot},
		{"1\n", nil, KindScreenshot},
		{"0", nil, KindNotScreenshot},
		{"(null)", nil, KindUnknown},
		{"", nil, KindUnknown},
		{"1", os.ErrPermission, KindUnknown},
	}
	for _, tt := range tests {
		if got := parseScreenCaptureMetadata(tt.out, tt.err); got != tt.want {
			t.Errorf("parseScreenCaptureMetadata(%q, %v) = %s, want %s", tt.out, tt.err, got, tt.want)
		}
	}
}

func TestClassify(t *testing.T) {
	dir := t.TempDir()
	named := filepath.Join(dir, "Screenshot ok.png")
	if err := os.WriteFile(named, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "notes.png")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "Screenshot empty.png")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	text := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(text, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(dir, ".hidden.png")
	if err := os.WriteFile(hidden, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	meta := func(kind Kind) func(string) Kind {
		return func(string) Kind { return kind }
	}

	if got := classify(named, meta(KindNotScreenshot)); got != KindScreenshot {
		t.Fatalf("named screenshot = %s, want screenshot", got)
	}
	if got := classify(empty, meta(KindScreenshot)); got != KindIncomplete {
		t.Fatalf("empty = %s, want incomplete", got)
	}
	if got := classify(hidden, meta(KindScreenshot)); got != KindIncomplete {
		t.Fatalf("dotfile = %s, want incomplete", got)
	}
	if got := classify(text, meta(KindScreenshot)); got != KindNotScreenshot {
		t.Fatalf("txt = %s, want not-screenshot", got)
	}
	if got := classify(other, meta(KindScreenshot)); got != KindScreenshot {
		t.Fatalf("mdls 1 = %s, want screenshot", got)
	}
	if got := classify(other, meta(KindNotScreenshot)); got != KindNotScreenshot {
		t.Fatalf("mdls 0 = %s, want not-screenshot", got)
	}
	if got := classify(other, meta(KindUnknown)); got != KindUnknown {
		t.Fatalf("mdls null = %s, want unknown", got)
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
	synctest.Test(t, func(t *testing.T) {
		s := NewSettler(40 * time.Millisecond)
		defer s.Stop()

		got := make(chan string, 2)
		s.Touch("/tmp/a.png", func(p string) { got <- p })
		s.Touch("/tmp/a.png", func(p string) { got <- p })
		time.Sleep(40 * time.Millisecond)
		synctest.Wait()

		select {
		case p := <-got:
			if p != "/tmp/a.png" {
				t.Fatalf("got %q", p)
			}
		default:
			t.Fatal("did not settle")
		}
		select {
		case <-got:
			t.Fatal("settled twice")
		default:
		}
	})
}

func TestSettlerStaleTimerIgnored(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := NewSettler(10 * time.Millisecond)
		defer s.Stop()

		var n atomic.Int32
		s.Touch("/tmp/a.png", func(string) { n.Add(1) })
		s.Touch("/tmp/a.png", func(string) { n.Add(1) })
		s.Touch("/tmp/a.png", func(string) { n.Add(1) })
		time.Sleep(10 * time.Millisecond)
		synctest.Wait()
		if got := n.Load(); got != 1 {
			t.Fatalf("calls = %d, want 1", got)
		}
	})
}
