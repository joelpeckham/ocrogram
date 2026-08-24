//go:build darwin

package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joelpeckham/ocrogram/internal/config"
	"github.com/joelpeckham/ocrogram/internal/screenshot"
)

const (
	helperTimeout     = 30 * time.Second
	unknownRetryWait  = 400 * time.Millisecond
	unknownRetryTries = 5
)

// Run is the long-lived process launchd starts.
func Run() error {
	lock, err := acquireInstanceLock()
	if err != nil {
		return err
	}
	defer lock.Close()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	dir := cfg.ScreenshotDir
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("daemon: screenshot dir %s: %w", dir, err)
	}

	helper, err := helperPath()
	if err != nil {
		return fmt.Errorf("daemon: helper: %w", err)
	}

	seen, err := screenshot.Snapshot(dir)
	if err != nil {
		return fmt.Errorf("daemon: snapshot: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("daemon: watcher: %w", err)
	}
	defer watcher.Close()
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("daemon: watch: %w", err)
	}

	log.Printf("watching %s", dir)

	settler := screenshot.NewSettler(400 * time.Millisecond)
	defer settler.Stop()

	var helperMu sync.Mutex
	proc := newProcessor(processorConfig{
		seen:     seen,
		maxTries: unknownRetryTries,
		classify: screenshot.Classify,
		ocr: func(path string) {
			helperMu.Lock()
			defer helperMu.Unlock()
			runHelper(helper, path)
		},
		sleep: time.Sleep,
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	watchDir := filepath.Clean(dir)
	for {
		select {
		case <-sigs:
			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("daemon: watcher errors closed")
			}
			log.Printf("watch error: %v", err)
		case ev, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("daemon: watcher closed")
			}
			path := ev.Name
			if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				if filepath.Clean(path) == watchDir {
					return fmt.Errorf("daemon: screenshot dir gone: %s", dir)
				}
				proc.forget(path)
			}
			if !ev.Has(fsnotify.Create) && !ev.Has(fsnotify.Write) {
				continue
			}
			if proc.known(path) {
				continue
			}
			settler.Touch(path, proc.settled)
		}
	}
}

type processorConfig struct {
	seen     map[string]struct{}
	maxTries int
	classify func(string) screenshot.Kind
	ocr      func(string)
	sleep    func(time.Duration)
}

type processor struct {
	mu       sync.Mutex
	seen     map[string]struct{}
	busy     map[string]struct{}
	maxTries int
	classify func(string) screenshot.Kind
	ocr      func(string)
	sleep    func(time.Duration)
}

func newProcessor(cfg processorConfig) *processor {
	if cfg.seen == nil {
		cfg.seen = make(map[string]struct{})
	}
	if cfg.maxTries <= 0 {
		cfg.maxTries = unknownRetryTries
	}
	if cfg.sleep == nil {
		cfg.sleep = time.Sleep
	}
	return &processor{
		seen:     cfg.seen,
		busy:     make(map[string]struct{}),
		maxTries: cfg.maxTries,
		classify: cfg.classify,
		ocr:      cfg.ocr,
		sleep:    cfg.sleep,
	}
}

func (p *processor) known(path string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.seen[path]
	return ok
}

func (p *processor) mark(path string) {
	p.mu.Lock()
	p.seen[path] = struct{}{}
	p.mu.Unlock()
}

func (p *processor) forget(path string) {
	p.mu.Lock()
	delete(p.seen, path)
	p.mu.Unlock()
}

func (p *processor) settled(path string) {
	p.mu.Lock()
	if _, ok := p.seen[path]; ok {
		p.mu.Unlock()
		return
	}
	if _, ok := p.busy[path]; ok {
		p.mu.Unlock()
		return
	}
	p.busy[path] = struct{}{}
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.busy, path)
		p.mu.Unlock()
	}()

	switch p.classify(path) {
	case screenshot.KindIncomplete:
		return
	case screenshot.KindNotScreenshot:
		p.mark(path)
		return
	case screenshot.KindScreenshot:
		p.mark(path)
		p.ocr(path)
		return
	case screenshot.KindUnknown:
		p.retryUnknown(path)
	}
}

func (p *processor) retryUnknown(path string) {
	for i := 1; i < p.maxTries; i++ {
		p.sleep(unknownRetryWait)
		switch p.classify(path) {
		case screenshot.KindScreenshot:
			p.mark(path)
			p.ocr(path)
			return
		case screenshot.KindNotScreenshot:
			p.mark(path)
			return
		case screenshot.KindIncomplete:
			return
		}
	}
	p.mark(path)
}

func helperPath() (string, error) {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "ocrogram-helper")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	path, err := exec.LookPath("ocrogram-helper")
	if err != nil {
		return "", err
	}
	return path, nil
}

func runHelper(helper, image string) {
	ctx, cancel := context.WithTimeout(context.Background(), helperTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, helper, image)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("helper timeout: %s", filepath.Base(image))
			return
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			log.Printf("no text: %s", filepath.Base(image))
			return
		}
		log.Printf("helper %s: %v", filepath.Base(image), err)
		return
	}
	n := len(strings.TrimSpace(string(out)))
	log.Printf("copied %d chars from %s", n, filepath.Base(image))
}

func acquireInstanceLock() (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("daemon: lock: %w", err)
	}
	dir := filepath.Join(home, "Library", "Application Support", "ocrogram")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("daemon: lock dir: %w", err)
	}
	path := filepath.Join(dir, "daemon.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("daemon: lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, flockError(err)
	}
	return f, nil
}

func flockError(err error) error {
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return fmt.Errorf("daemon: already running")
	}
	return fmt.Errorf("daemon: lock: %w", err)
}
