//go:build darwin

package screenshot

import (
	"os"
	"path/filepath"
	"testing"
)

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
		t.Fatalf("named screenshot = %d, want screenshot", got)
	}
	if got := classify(empty, meta(KindScreenshot)); got != KindIncomplete {
		t.Fatalf("empty = %d, want incomplete", got)
	}
	if got := classify(hidden, meta(KindScreenshot)); got != KindIncomplete {
		t.Fatalf("dotfile = %d, want incomplete", got)
	}
	if got := classify(text, meta(KindScreenshot)); got != KindNotScreenshot {
		t.Fatalf("txt = %d, want not-screenshot", got)
	}
	if got := classify(other, meta(KindScreenshot)); got != KindScreenshot {
		t.Fatalf("mdls 1 = %d, want screenshot", got)
	}
	if got := classify(other, meta(KindNotScreenshot)); got != KindNotScreenshot {
		t.Fatalf("mdls 0 = %d, want not-screenshot", got)
	}
	if got := classify(other, meta(KindUnknown)); got != KindUnknown {
		t.Fatalf("mdls null = %d, want unknown", got)
	}
}
