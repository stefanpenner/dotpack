#!/bin/bash
# Dotpack installer — downloads and extracts the hermetic tool bundle.
#
# Usage:
#   curl -fsSL https://github.com/stefanpenner/dotpack/releases/latest/download/install.sh | bash
#
# Options (env vars):
#   DOTPACK_VERSION=v0.1.0   Pin to a specific release (default: latest)
#   DOTPACK_DIR=$HOME/.local  Install directory (default: ~/.local)
#   DOTPACK_REPO=user/repo   Override repo (default: stefanpenner/dotpack)
#
set -euo pipefail

REPO="${DOTPACK_REPO:-stefanpenner/dotpack}"
VERSION="${DOTPACK_VERSION:-latest}"
INSTALL_DIR="${DOTPACK_DIR:-$HOME/.local}"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Normalize arch
case "$ARCH" in
  x86_64|amd64)  ARCH=x86_64 ;;
  aarch64|arm64)
    case "$OS" in
      linux)  ARCH=aarch64 ;;
      darwin) ARCH=arm64 ;;
    esac
    ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="dotpack-${OS}-${ARCH}.tar.gz"
CHECKSUM_ASSET="dotpack-${OS}-${ARCH}.tar.gz.sha256"

if [ "$VERSION" = "latest" ]; then
  BASE_URL="https://github.com/${REPO}/releases/latest/download"
else
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
fi

echo "dotpack: installing ${VERSION} for ${OS}/${ARCH}"
echo "  target: ${INSTALL_DIR}"

# Download
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "  downloading ${ASSET}..."
curl -fsSL "${BASE_URL}/${ASSET}" -o "${TMP}/${ASSET}"

# Verify checksum if available
if curl -fsSL "${BASE_URL}/${CHECKSUM_ASSET}" -o "${TMP}/${CHECKSUM_ASSET}" 2>/dev/null; then
  echo "  verifying checksum..."
  cd "$TMP"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "${CHECKSUM_ASSET}"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "${CHECKSUM_ASSET}"
  else
    echo "  warning: no sha256sum available, skipping verification"
  fi
  cd - >/dev/null
fi

# Extract
echo "  extracting..."
mkdir -p "${INSTALL_DIR}"
tar xzf "${TMP}/${ASSET}" -C "${INSTALL_DIR}"

echo ""
echo "dotpack installed to ${INSTALL_DIR}"
echo ""
echo "Add to your shell profile:"
echo '  export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$PATH"'
echo ""
echo "Then install your dotfiles:"
echo "  git clone https://github.com/stefanpenner/dotfiles.git ~/src/stefanpenner/dotfiles"
echo "  ~/src/stefanpenner/dotfiles/sync.sh"
