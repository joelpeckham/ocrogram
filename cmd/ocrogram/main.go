//go:build darwin

package main

import (
	"fmt"
	"os"

	"github.com/joelpeckham/ocrogram/internal/daemon"
	"github.com/joelpeckham/ocrogram/internal/service"
)

// version is set via -ldflags at build time.
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "daemon":
		return daemon.Run()
	case "start":
		return service.Start()
	case "stop":
		return service.Stop()
	case "version", "--version", "-v":
		fmt.Println("ocrogram", version)
		return nil
	case "-h", "--help", "help":
		usage()
		return nil
	default:
		usage()
		os.Exit(2)
		return nil
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `ocrogram extracts text from macOS screenshots into the clipboard.

Usage:
  ocrogram start     install and start the login item
  ocrogram stop      stop and remove the login item
  ocrogram daemon    run the background watcher (used by launchd)
`)
}
