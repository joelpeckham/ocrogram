//go:build darwin

package screenshot

import (
	"context"
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

// Kind is the classification of a path in the screenshot folder.
type Kind int

const (
	// KindIncomplete is missing, empty, or a dotfile — do not mark seen.
	KindIncomplete Kind = iota
	// KindScreenshot is a finished screenshot — OCR it.
	KindScreenshot
	// KindNotScreenshot is finished but not a screenshot — ignore it.
	KindNotScreenshot
	// KindUnknown is an image whose Spotlight metadata is not ready yet.
	KindUnknown
)

func isImage(path string) bool {
	_, ok := imageExts[strings.ToLower(filepath.Ext(path))]
	return ok
}

func matchesName(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	for _, prefix := range namePrefixes {
		if strings.HasPrefix(base, prefix) {
			return true
		}
	}
	return false
}

func isDotfile(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}

func isEmpty(path string) bool {
	info, err := os.Stat(path)
	return err != nil || info.Size() == 0
}

const metaTimeout = 5 * time.Second

func lookupScreenCaptureMetadata(path string) Kind {
	ctx, cancel := context.WithTimeout(context.Background(), metaTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "mdls", "-raw", "-name", "kMDItemIsScreenCapture", path).Output()
	return parseScreenCaptureMetadata(string(out), err)
}

func parseScreenCaptureMetadata(out string, err error) Kind {
	if err != nil {
		return KindUnknown
	}
	switch strings.TrimSpace(out) {
	case "1":
		return KindScreenshot
	case "0":
		return KindNotScreenshot
	default:
		return KindUnknown
	}
}

func classify(path string, meta func(string) Kind) Kind {
	if isDotfile(path) || isEmpty(path) {
		return KindIncomplete
	}
	if !isImage(path) {
		return KindNotScreenshot
	}
	if matchesName(path) {
		return KindScreenshot
	}
	return meta(path)
}

// Classify reports whether path is a screenshot, not one, still writing, or unknown.
func Classify(path string) Kind {
	return classify(path, lookupScreenCaptureMetadata)
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

// NewSettler returns a settler that fires after delay of quiet.
func NewSettler(delay time.Duration) *Settler {
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
	var timer *time.Timer
	timer = time.AfterFunc(s.Delay, func() {
		s.mu.Lock()
		current, ok := s.pending[path]
		if !ok || current != timer {
			s.mu.Unlock()
			return
		}
		delete(s.pending, path)
		s.mu.Unlock()
		fn(path)
	})
	s.pending[path] = timer
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
