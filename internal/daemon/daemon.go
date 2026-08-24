package daemon

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joelpeckham/ocrogram/internal/config"
	"github.com/joelpeckham/ocrogram/internal/screenshot"
)

// Run is the long-lived process launchd starts.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	dir := cfg.ScreenshotDir
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("daemon: screenshot dir: %w", err)
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

	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)
	log.Printf("watching %s", dir)

	settler := screenshot.NewSettler(400 * time.Millisecond)
	defer settler.Stop()

	var mu sync.Mutex
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigs:
			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("watch error: %v", err)
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !ev.Has(fsnotify.Create) && !ev.Has(fsnotify.Write) && !ev.Has(fsnotify.Rename) {
				continue
			}
			path := ev.Name
			mu.Lock()
			_, known := seen[path]
			mu.Unlock()
			if known {
				continue
			}
			settler.Touch(path, func(p string) {
				if !screenshot.IsScreenshot(p) {
					if !screenshot.IsEmpty(p) && !screenshot.IsDotfile(p) {
						mu.Lock()
						seen[p] = struct{}{}
						mu.Unlock()
					}
					return
				}
				mu.Lock()
				seen[p] = struct{}{}
				mu.Unlock()
				runHelper(helper, p)
			})
		}
	}
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
	cmd := exec.Command(helper, image)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			log.Printf("no text: %s", filepath.Base(image))
			return
		}
		log.Printf("helper %s: %v", filepath.Base(image), err)
		return
	}
	if n := len(string(out)); n > 0 {
		log.Printf("copied %d chars from %s", n, filepath.Base(image))
	}
}
