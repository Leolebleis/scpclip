# scpclip

Cross-platform Go CLI. Reads clipboard image, uploads to SSH host, copies remote path to clipboard.

## Architecture

- **Zero CGO** — shells out to native clipboard tools (pngpaste, xclip/wl-paste, PowerShell) instead of `golang-design/clipboard`. Enables trivial cross-compilation with `CGO_ENABLED=0`.
- **Interfaces for testability** — `Clipboard` and `Uploader` interfaces in clipboard.go/upload.go. Tests use mock implementations; production uses `OSClipboard` and `SSHUploader`.
- **Subcommands without a framework** — `os.Args[1]` check before `flag.Parse()` handles `default` and `update` subcommands. No Cobra/urfave.
- **Self-update via minio/selfupdate** — DIY GitHub Releases API query + minio for atomic binary replacement. Chosen over creativeprojects/go-selfupdate (19 deps vs 3).

## Build & Release

- `go build .` — local build, version is `dev`
- GoReleaser on tag push (`v*`) — injects version via `-ldflags "-s -w -X main.version={{.Version}}"`
- **GoReleaser strips `v` prefix** — `{{.Version}}` = `0.1.0`, GitHub tag = `v0.1.0`. Use `strings.TrimPrefix` when comparing.
- 6 targets: linux/darwin/windows x amd64/arm64, all `CGO_ENABLED=0`
- Archives: `.tar.gz` (Unix), `.zip` (Windows)

## Testing

- `go test -v ./...` — runs all tests
- **exec subprocess mock pattern** in upload_test.go — `TestHelperProcess` + `fakeExecCommand` for mocking `os/exec.Command`
- **Interface mocks** in main_test.go — `mockClipboard`, `mockUploader` for testing `run()` orchestration
- **Archive extraction tests** in update_test.go — create tar.gz/zip in memory, verify extraction
- Config tests use `t.Setenv` to redirect `os.UserConfigDir()`

## Config

- Persistent config at `os.UserConfigDir()/scpclip/config.json` (Linux: `~/.config/`, macOS: `~/Library/Application Support/`, Windows: `%AppData%`)
- Priority: `--host` flag > `SCPCLIP_HOST` env > saved config > error

## Clipboard Tools (runtime deps)

| OS | Read image | Write text |
|----|-----------|------------|
| macOS | `pngpaste -` (stdout) | `pbcopy` (stdin) |
| Linux X11 | `xclip -selection clipboard -target image/png -o` | `xclip -selection clipboard` |
| Linux Wayland | `wl-paste --type image/png` | `wl-copy` |
| Windows | PowerShell `System.Windows.Forms.Clipboard::GetImage()` | PowerShell `Set-Clipboard` |
