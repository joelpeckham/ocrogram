//go:build darwin

package main

import (
	"fmt"
	"os"

	"github.com/joelpeckham/ocrogram/internal/daemon"
	"github.com/joelpeckham/ocrogram/internal/service"
	"github.com/joelpeckham/ocrogram/internal/tui"
	"github.com/spf13/cobra"
)

// version is set via -ldflags at build time.
var version = "dev"

func main() {
	root := &cobra.Command{
		Use:           "ocrogram",
		Short:         "Extract text from macOS screenshots into the clipboard",
		Long:          "ocrogram watches for macOS screenshots, extracts text with Apple Vision, and copies it to the clipboard.",
		Version:       version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "daemon",
			Short: "Run the background screenshot watcher (used by launchd)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return daemon.Run()
			},
		},
		&cobra.Command{
			Use:   "start",
			Short: "Install and start the login item",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return service.Start()
			},
		},
		&cobra.Command{
			Use:   "stop",
			Short: "Stop and remove the login item",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return service.Stop()
			},
		},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
