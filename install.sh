#!/bin/sh
set -e

REPO="leolebleis/scpclip"
INSTALL_DIR="${SCPCLIP_INSTALL_DIR:-$HOME/.local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64)    ARCH="amd64" ;;
    aarch64|arm64)   ARCH="arm64" ;;
    *)               echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

VERSION=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version" >&2
    exit 1
fi
VERSION_TRIMMED="${VERSION#v}"

ARCHIVE="scpclip_${VERSION_TRIMMED}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
CHECKSUMS_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Installing scpclip $VERSION ($OS/$ARCH)..."
curl -sSfL "$URL" -o "$TMPDIR/$ARCHIVE"
curl -sSfL "$CHECKSUMS_URL" -o "$TMPDIR/checksums.txt"

cd "$TMPDIR"
if command -v sha256sum >/dev/null 2>&1; then
    grep "$ARCHIVE" checksums.txt | sha256sum -c --quiet
elif command -v shasum >/dev/null 2>&1; then
    grep "$ARCHIVE" checksums.txt | shasum -a 256 -c --quiet
else
    echo "Warning: no checksum tool found, skipping verification" >&2
fi

tar xzf "$ARCHIVE" scpclip
mkdir -p "$INSTALL_DIR"
mv scpclip "$INSTALL_DIR/scpclip"
chmod +x "$INSTALL_DIR/scpclip"

echo ""
echo "✓ Installed scpclip $VERSION to $INSTALL_DIR/scpclip"
echo ""
echo "  Get started:"
echo "    scpclip default <host>   set your default SSH host"
echo "    scpclip --help           see all options"

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) echo "Add $INSTALL_DIR to your PATH if it isn't already." >&2 ;;
esac
