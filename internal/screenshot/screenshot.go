package screenshot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var imageExts = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".heic": {},
	".tiff": {},
	".tif":  {},
}

var namePrefixes = []string{
	"screenshot",
	"screen shot",
	"bildschirmfoto",
	"capture d'écran",
	"capture d’écran",
	"captura de pantalla",
	"captura de ecrã",
	"captura de tela",
	"schermata",
	"スクリーンショット",
	"스크린샷",
	"截屏",
	"屏幕快照",
	"螢幕快照",
	"schermafbeelding",
	"снимок экрана",
	"skärmavbild",
	"skjermbilde",
	"skærmbillede",
	"kuvakaappaus",
	"zrzut ekranu",
}

// IsImage reports whether path has a screenshot-like image extension.
func IsImage(path string) bool {
	_, ok := imageExts[strings.ToLower(filepath.Ext(path))]
	return ok
}

// MatchesName reports whether the file name looks like a macOS screenshot.
func MatchesName(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	for _, prefix := range namePrefixes {
		if strings.HasPrefix(base, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// IsDotfile reports whether the base name starts with a dot.
func IsDotfile(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}

// IsEmpty reports whether path is missing or has size 0.
func IsEmpty(path string) bool {
	info, err := os.Stat(path)
	return err != nil || info.Size() == 0
}

// HasScreenCaptureMetadata reports kMDItemIsScreenCapture via mdls.
func HasScreenCaptureMetadata(path string) bool {
	out, err := exec.Command("mdls", "-raw", "-name", "kMDItemIsScreenCapture", path).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

// IsScreenshot reports whether path looks like a finished screenshot image.
func IsScreenshot(path string) bool {
	if IsDotfile(path) || !IsImage(path) || IsEmpty(path) {
		return false
	}
	return HasScreenCaptureMetadata(path) || MatchesName(path)
}

// Snapshot lists non-directory entries in dir, keyed by full path.
func Snapshot(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out[filepath.Join(dir, e.Name())] = struct{}{}
	}
	return out, nil
}

// Settler calls a function after path has been quiet for Delay.
type Settler struct {
	Delay time.Duration

	mu      sync.Mutex
	pending map[string]*time.Timer
}

// NewSettler returns a settler. A non-positive delay becomes 400ms.
func NewSettler(delay time.Duration) *Settler {
	if delay <= 0 {
		delay = 400 * time.Millisecond
	}
	return &Settler{
		Delay:   delay,
		pending: make(map[string]*time.Timer),
	}
}

// Touch (re)starts the settle timer for path.
func (s *Settler) Touch(path string, fn func(string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.pending[path]; ok {
		t.Stop()
	}
	s.pending[path] = time.AfterFunc(s.Delay, func() {
		s.mu.Lock()
		delete(s.pending, path)
		s.mu.Unlock()
		fn(path)
	})
}

// Stop cancels all pending timers.
func (s *Settler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.pending {
		t.Stop()
	}
	s.pending = make(map[string]*time.Timer)
}
