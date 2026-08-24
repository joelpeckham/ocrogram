//go:build darwin

package daemon

import (
	"errors"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/joelpeckham/ocrogram/internal/screenshot"
)

func TestProcessorIncompleteNotSeen(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		classify: func(string) screenshot.Kind { return screenshot.KindIncomplete },
		ocr:      func(string) { ocr.Add(1) },
		sleep:    func(time.Duration) {},
	})
	p.settled("/tmp/Screenshot.png")
	if p.known("/tmp/Screenshot.png") {
		t.Fatal("incomplete path should not be marked seen")
	}
	if ocr.Load() != 0 {
		t.Fatal("incomplete path should not be OCR'd")
	}
}

func TestProcessorNotScreenshotMarkedSeen(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		classify: func(string) screenshot.Kind { return screenshot.KindNotScreenshot },
		ocr:      func(string) { ocr.Add(1) },
		sleep:    func(time.Duration) {},
	})
	p.settled("/tmp/notes.png")
	if !p.known("/tmp/notes.png") {
		t.Fatal("non-screenshot should be marked seen")
	}
	if ocr.Load() != 0 {
		t.Fatal("non-screenshot should not be OCR'd")
	}
}

func TestProcessorScreenshotOCROnce(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		classify: func(string) screenshot.Kind { return screenshot.KindScreenshot },
		ocr:      func(string) { ocr.Add(1) },
		sleep:    func(time.Duration) {},
	})
	p.settled("/tmp/Screenshot.png")
	p.settled("/tmp/Screenshot.png")
	if !p.known("/tmp/Screenshot.png") {
		t.Fatal("screenshot should be marked seen")
	}
	if ocr.Load() != 1 {
		t.Fatalf("ocr calls = %d, want 1", ocr.Load())
	}
}

func TestProcessorUnknownRetriesThenScreenshot(t *testing.T) {
	var n atomic.Int32
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		maxTries: 5,
		classify: func(string) screenshot.Kind {
			if n.Add(1) < 3 {
				return screenshot.KindUnknown
			}
			return screenshot.KindScreenshot
		},
		ocr:   func(string) { ocr.Add(1) },
		sleep: func(time.Duration) {},
	})
	p.settled("/tmp/custom.png")
	if ocr.Load() != 1 {
		t.Fatalf("ocr calls = %d, want 1", ocr.Load())
	}
	if !p.known("/tmp/custom.png") {
		t.Fatal("should be marked seen after becoming a screenshot")
	}
}

func TestProcessorUnknownExhaustedMarkedSeen(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		maxTries: 3,
		classify: func(string) screenshot.Kind { return screenshot.KindUnknown },
		ocr:      func(string) { ocr.Add(1) },
		sleep:    func(time.Duration) {},
	})
	p.settled("/tmp/custom.png")
	if ocr.Load() != 0 {
		t.Fatal("exhausted unknown should not be OCR'd")
	}
	if !p.known("/tmp/custom.png") {
		t.Fatal("exhausted unknown should be marked seen")
	}
}

func TestProcessorUnknownBecomesNotScreenshot(t *testing.T) {
	var n atomic.Int32
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		maxTries: 5,
		classify: func(string) screenshot.Kind {
			if n.Add(1) == 1 {
				return screenshot.KindUnknown
			}
			return screenshot.KindNotScreenshot
		},
		ocr:   func(string) { ocr.Add(1) },
		sleep: func(time.Duration) {},
	})
	p.settled("/tmp/photo.png")
	if ocr.Load() != 0 {
		t.Fatal("should not OCR a non-screenshot")
	}
	if !p.known("/tmp/photo.png") {
		t.Fatal("should be marked seen")
	}
}

func TestProcessorForgetAllowsReprocess(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		classify: func(string) screenshot.Kind { return screenshot.KindScreenshot },
		ocr:      func(string) { ocr.Add(1) },
		sleep:    func(time.Duration) {},
	})
	p.settled("/tmp/Screenshot.png")
	if ocr.Load() != 1 {
		t.Fatalf("ocr calls = %d, want 1", ocr.Load())
	}
	p.forget("/tmp/Screenshot.png")
	if p.known("/tmp/Screenshot.png") {
		t.Fatal("forgotten path should not be known")
	}
	p.settled("/tmp/Screenshot.png")
	if ocr.Load() != 2 {
		t.Fatalf("ocr calls = %d, want 2", ocr.Load())
	}
}

func TestProcessorBusySkipsDuplicateSettle(t *testing.T) {
	inClassify := make(chan struct{})
	release := make(chan struct{})
	var classifyCalls atomic.Int32
	p := newProcessor(processorConfig{
		classify: func(string) screenshot.Kind {
			if classifyCalls.Add(1) == 1 {
				close(inClassify)
				<-release
			}
			return screenshot.KindScreenshot
		},
		ocr:   func(string) {},
		sleep: func(time.Duration) {},
	})
	done := make(chan struct{})
	go func() {
		p.settled("/tmp/Screenshot.png")
		close(done)
	}()
	<-inClassify
	p.settled("/tmp/Screenshot.png")
	close(release)
	<-done
	if classifyCalls.Load() != 1 {
		t.Fatalf("classify calls = %d, want 1", classifyCalls.Load())
	}
}

func TestFlockError(t *testing.T) {
	err := flockError(syscall.EWOULDBLOCK)
	if !strings.Contains(err.Error(), "already running") {
		t.Fatalf("EWOULDBLOCK = %v, want already running", err)
	}
	other := flockError(syscall.EIO)
	if strings.Contains(other.Error(), "already running") {
		t.Fatal("EIO should not be treated as already running")
	}
	if !errors.Is(other, syscall.EIO) {
		t.Fatalf("EIO should be wrapped, got %v", other)
	}
}
