# ocrogram

A set-and-forget Mac background tool: install via Homebrew, run the TUI once, add it to login items, then ignore it. Whenever you take a normal macOS screenshot, ocrogram extracts high-quality text and puts it on the clipboard so Cmd+V pastes that text anywhere.

The name is a portmanteau of OCR and program — and *gram*, something written down.

## Goal

Take a default Mac screenshot. Paste the text. That is the whole workflow.

You should not open an app, crop a region into a dedicated OCR window, or remember a hotkey beyond the screenshot shortcut you already use. ocrogram runs in the background, notices the screenshot, does the extraction, and leaves the text waiting on the clipboard.

## Intended design

Nothing below is implemented yet. This repo is a stack scaffold.

- **Watch** the user’s screenshot folder — Desktop, `~/Pictures/Screenshots`, or whatever `defaults read com.apple.screencapture location` points at
- **OCR** with Apple Vision (best quality available on Mac)
- **Write** recognized text to the clipboard
- **Persist** as a per-user LaunchAgent so it starts at login

```
brew install ocrogram
ocrogram          # TUI: confirm folder, enable login item, start daemon
# then forget about it
```

Later, launchd runs `ocrogram daemon`, which calls `ocrogram-helper` for Vision and clipboard work.

## Stack

- **Go** — CLI (`cobra`), setup TUI (`bubbletea` + `lipgloss`), daemon and LaunchAgent install
- **Swift** — `ocrogram-helper` for Vision OCR, clipboard, and screenshot-file APIs
- **Homebrew** — install path; formula in `Formula/ocrogram.rb` is a stub until there is a release
- **launchd** — login item; template in `contrib/com.joelpeckham.ocrogram.plist`

## Commands (wired, not implemented)

| Command | Role |
| --- | --- |
| `ocrogram` | Setup TUI |
| `ocrogram daemon` | Background watcher (what launchd will run) |
| `ocrogram start` | Install and start the login item |
| `ocrogram stop` | Stop and remove the login item |

## Develop

Requires Go 1.27+ and Swift 6 (Xcode / Command Line Tools).

```bash
make
./bin/ocrogram --help
./bin/ocrogram-helper
```

`make go` builds only the CLI. `make helper` builds only the Swift helper. `make clean` removes `bin/` and Swift build products.

## Status

Scaffold only. No screenshot watching, no Vision OCR, no clipboard writes, no real plist install. No Homebrew tap and no GitHub release yet.
