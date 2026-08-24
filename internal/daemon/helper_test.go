//go:build darwin

package daemon

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHelperRecognizesHello(t *testing.T) {
	helper, ok := lookupHelper()
	if !ok {
		t.Skip("ocrogram-helper not built; run make helper")
	}

	cmd := exec.Command(helper, "testdata/hello.png")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper: %v", err)
	}
	text := strings.TrimSpace(string(out))
	if !strings.Contains(strings.ToUpper(text), "HELLO") {
		t.Fatalf("recognized %q, want HELLO", text)
	}
}

func lookupHelper() (string, bool) {
	if info, err := os.Stat("../../bin/ocrogram-helper"); err == nil && !info.IsDir() {
		return "../../bin/ocrogram-helper", true
	}
	if path, err := exec.LookPath("ocrogram-helper"); err == nil {
		return path, true
	}
	return "", false
}
