# scpclip

Clipboard image to SSH host in one command. Grab a screenshot, run `scpclip`, get the remote path on your clipboard.

## Install

**macOS / Linux:**

```sh
curl -sSfL https://raw.githubusercontent.com/leolebleis/scpclip/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/leolebleis/scpclip/main/install.ps1 | iex
```

**From source** (requires Go):

```sh
go install github.com/leolebleis/scpclip@latest
```

Re-run the install command to update to the latest version.

## Usage

```
scpclip [--host user@host] [--dir /remote/dir] [--version]
```

1. Copy an image to your clipboard (screenshot, snip, etc.)
2. Run `scpclip`
3. The image is uploaded to `/tmp/scpclip_<timestamp>.png` on the remote host
4. The remote path is copied to your clipboard and printed to stdout

## Configuration

Set a default host so you don't need `--host` every time:

```sh
export SCPCLIP_HOST=user@myserver    # add to your shell profile
```

| Flag / Env Var | Description | Default |
|----------------|-------------|---------|
| `--host` / `SCPCLIP_HOST` | SSH host (required) | — |
| `--dir` / `SCPCLIP_DIR` | Remote directory | `/tmp` |

Flags override env vars. The tool uses your system `ssh`, so all `~/.ssh/config` settings (ProxyJump, keys, aliases) work automatically.

Or save a default host so you never need `--host`:

```sh
scpclip default pi           # save default (uses ~/.ssh/config aliases)
scpclip default              # show current default
scpclip                      # just works
```

Priority: `--host` flag > `SCPCLIP_HOST` env var > saved default.

## Requirements

`ssh` must be on your PATH (ships with all modern OSes). For clipboard access:

| OS | Tool | Install |
|----|------|---------|
| macOS | `pngpaste` | `brew install pngpaste` |
| Linux (X11) | `xclip` | `apt install xclip` |
| Linux (Wayland) | `wl-clipboard` | `apt install wl-clipboard` |
| Windows | PowerShell | Built-in |

## How it works

1. Reads PNG image from clipboard via native OS tools
2. Writes to a local temp file
3. Pipes the file to `ssh host "umask 077 && cat > /tmp/scpclip_<ts>.png"` (0600 permissions)
4. Copies the remote path to your clipboard
5. Cleans up the local temp file

## License

MIT
