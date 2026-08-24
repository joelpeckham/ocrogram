//go:build darwin

package daemon

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/joelpeckham/ocrogram/internal/screenshot"
)

func TestProcessorScreenshotOCROnce(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		seen:     make(map[string]struct{}),
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
		seen:     make(map[string]struct{}),
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

func TestProcessorForgetAllowsReprocess(t *testing.T) {
	var ocr atomic.Int32
	p := newProcessor(processorConfig{
		seen:     make(map[string]struct{}),
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
