package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Label = "com.joelpeckham.ocrogram"

const plistName = "com.joelpeckham.ocrogram.plist"

// Start installs the LaunchAgent and bootstraps it so ocrogram runs at login.
func Start() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("service: executable: %w", err)
	}
	exe = stableExecutable(exe)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("service: home: %w", err)
	}
	plist := plistPath(home)
	logFile := logPath(home)

	if err := os.MkdirAll(filepath.Dir(plist), 0o755); err != nil {
		return fmt.Errorf("service: mkdir agents: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return fmt.Errorf("service: mkdir logs: %w", err)
	}

	body := Plist(exe, logFile)
	if err := os.WriteFile(plist, []byte(body), 0o644); err != nil {
		return fmt.Errorf("service: write plist: %w", err)
	}

	_ = bootout()
	if err := bootstrap(plist); err != nil {
		return fmt.Errorf("service: bootstrap: %w", err)
	}
	return nil
}

// Stop unloads and removes the LaunchAgent.
func Stop() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("service: home: %w", err)
	}
	_ = bootout()
	if err := os.Remove(plistPath(home)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("service: remove plist: %w", err)
	}
	return nil
}

// Running reports whether the LaunchAgent is loaded.
func Running() bool {
	return exec.Command("launchctl", "print", target()).Run() == nil
}

// Plist renders a LaunchAgent property list for exe, logging to logFile.
func Plist(exe, logFile string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>%s</string>
	<key>StandardErrorPath</key>
	<string>%s</string>
</dict>
</plist>
`, Label, xmlEscape(exe), xmlEscape(logFile), xmlEscape(logFile))
}

func plistPath(home string) string {
	return filepath.Join(home, "Library", "LaunchAgents", plistName)
}

func logPath(home string) string {
	return filepath.Join(home, "Library", "Logs", "ocrogram.log")
}

func domain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

func target() string {
	return domain() + "/" + Label
}

func bootstrap(plist string) error {
	out, err := exec.Command("launchctl", "bootstrap", domain(), plist).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func bootout() error {
	out, err := exec.Command("launchctl", "bootout", target()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// stableExecutable prefers Homebrew's prefix shim over a Cellar path so
// upgrades do not leave the LaunchAgent pointing at a removed version.
func stableExecutable(exe string) string {
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	const marker = string(filepath.Separator) + "Cellar" + string(filepath.Separator)
	i := strings.Index(resolved, marker)
	if i < 0 {
		return resolved
	}
	candidate := filepath.Join(resolved[:i], "bin", filepath.Base(resolved))
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return resolved
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
