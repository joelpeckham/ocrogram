//go:build darwin

package daemon

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelperRecognizesHello(t *testing.T) {
	helper, ok := lookupHelper()
	if !ok {
		t.Skip("ocrogram-helper not built; run make helper")
	}

	image := filepath.Join(repoRoot(t), "internal", "daemon", "testdata", "hello.png")
	cmd := exec.Command(helper, image)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper: %v", err)
	}
	text := strings.TrimSpace(string(out))
	if !strings.Contains(strings.ToUpper(text), "HELLO") {
		t.Fatalf("recognized %q, want HELLO", text)
	}
}

func TestHelperUsage(t *testing.T) {
	helper, ok := lookupHelper()
	if !ok {
		t.Skip("ocrogram-helper not built; run make helper")
	}
	cmd := exec.Command(helper)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit 2")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 2 {
		t.Fatalf("exit %v, want 2; output %q", err, out)
	}
	if !strings.Contains(string(out), "usage: ocrogram-helper") {
		t.Fatalf("got %q", out)
	}
}

func lookupHelper() (string, bool) {
	root := findGoMod()
	if root == "" {
		return "", false
	}
	for _, rel := range []string{
		filepath.Join("bin", "ocrogram-helper"),
		filepath.Join("helper", ".build", "release", "ocrogram-helper"),
	} {
		path := filepath.Join(root, rel)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	if path, err := exec.LookPath("ocrogram-helper"); err == nil {
		return path, true
	}
	return "", false
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root := findGoMod()
	if root == "" {
		t.Fatal("go.mod not found")
	}
	return root
}

func findGoMod() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
