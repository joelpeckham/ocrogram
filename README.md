# ocrogram

Take a screenshot. Paste the text.

A set-and-forget Mac background tool. Install it, run `ocrogram start`, leave it on as a login item. After that, a normal macOS screenshot puts the text on the clipboard. Cmd+V pastes it anywhere. Apple Vision runs on-device. Nothing is uploaded.

Site: [ocrogram.com](https://ocrogram.com) — [install](https://ocrogram.com/#install), [requirements](https://ocrogram.com/#requirements), [troubleshooting](https://ocrogram.com/#troubleshooting), [guides](https://ocrogram.com/guides).

The name is a portmanteau of OCR and program — and *gram*, something written down.

## Goal

Take a default Mac screenshot. Paste the text. That is the whole workflow.

You should not open an app, crop a region into a dedicated OCR window, or remember a hotkey beyond the screenshot shortcut you already use.

## Install

```bash
brew install joelpeckham/ocrogram/ocrogram
ocrogram start
```

That taps [joelpeckham/homebrew-ocrogram](https://github.com/joelpeckham/homebrew-ocrogram) and trusts this formula (Homebrew 6 requires that for third-party taps). `ocrogram start` installs the login item; `ocrogram stop` removes it.

Homebrew compiles from source, so you need Go and a Swift toolchain it accepts. If it says Xcode is outdated, update Xcode (or the Command Line Tools) to the version for your macOS, or remove an older `/Applications/Xcode.app` so Homebrew can use the CLT. Full requirements and a step-by-step walkthrough: [ocrogram.com/#install](https://ocrogram.com/#install).

To build the latest `main` instead of the last release:

```bash
brew install --HEAD joelpeckham/ocrogram/ocrogram
```

## How it works

ocrogram watches the folder from `defaults read com.apple.screencapture location`, or `~/Desktop` if that key is unset. New screenshot images are sent to `ocrogram-helper`, which runs Apple Vision and writes the text to the clipboard.

`ocrogram start` generates a LaunchAgent at `~/Library/LaunchAgents/com.joelpeckham.ocrogram.plist` that runs `ocrogram daemon` at login. Logs go to `~/Library/Logs/ocrogram.log`.

Clipboard-only captures (`Cmd+Ctrl+Shift+3` / `Cmd+Ctrl+Shift+4`) never write a file, so ocrogram ignores them.

Watching Desktop may need Full Disk Access. If the daemon cannot read that folder, point screenshots at a folder you own:

```bash
mkdir -p ~/Pictures/Screenshots
defaults write com.apple.screencapture location ~/Pictures/Screenshots
killall SystemUIServer
```

More cases (build failures, renamed screenshots, the lock file): [ocrogram.com/#troubleshooting](https://ocrogram.com/#troubleshooting).

## Commands

| Command | Role |
| --- | --- |
| `ocrogram start` | Install and start the login item |
| `ocrogram stop` | Stop and remove the login item |
| `ocrogram daemon` | Background watcher (what launchd runs) |
| `ocrogram version` | Print the installed version (`--version` and `-v` work too) |

## Develop

Requires Go 1.27+ and Swift 6 (Xcode / Command Line Tools).

```bash
make
./bin/ocrogram --help
./bin/ocrogram-helper
make test
```

`make go` builds only the CLI. `make helper` builds only the Swift helper. `make clean` removes `bin/` and Swift build products.

The helper contract is `ocrogram-helper <image-path>`: print text, copy it, exit 0 on success, 1 if the image had no text, 2 on usage or IO errors.

Custom-named screenshots are still picked up once Spotlight sets `kMDItemIsScreenCapture` (ocrogram retries for a couple of seconds). Default localized names match immediately.

## Release

1. Tag `vX.Y.Z` and push. GitHub Actions builds, tests, and opens a Release.
2. `sha256` the source tarball at `https://github.com/joelpeckham/ocrogram/archive/refs/tags/vX.Y.Z.tar.gz`.
3. Set `url` and `sha256` in the tap formula.
