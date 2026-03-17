#!/bin/bash
# Download pre-built static/musl binaries for dotpack.
# Usage: download-binaries.sh <output-dir> [os] [arch]
#   os defaults to current (linux/darwin)
#   arch defaults to current machine
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=../versions.env
source "$SCRIPT_DIR/../versions.env"

OUT="${1:?usage: download-binaries.sh <output-dir> [os] [arch]}"
OS="${2:-$(uname -s | tr '[:upper:]' '[:lower:]')}"
ARCH="${3:-$(uname -m)}"

mkdir -p "$OUT/bin"

# Normalize architecture
case "$ARCH" in
  x86_64|amd64)  RUST_ARCH=x86_64  GOARCH=amd64  ARCH_GENERIC=x86_64 ;;
  aarch64|arm64)  RUST_ARCH=aarch64 GOARCH=arm64  ARCH_GENERIC=arm64  ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# OS-specific target triples & naming
case "$OS" in
  linux)
    RUST_TARGET="${RUST_ARCH}-unknown-linux-musl"
    LAZYGIT_OS=Linux
    NVIM_OS=linux
    JQ_OS=linux
    ;;
  darwin)
    RUST_TARGET="${RUST_ARCH}-apple-darwin"
    LAZYGIT_OS=Darwin
    NVIM_OS=macos
    JQ_OS=macos
    ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

dl() {
  local url=$1 dest=$2
  echo "  $(basename "$dest")"
  curl -fsSL "$url" -o "$dest"
}

dl_tar() {
  local url=$1 binary=$2
  local tmp; tmp=$(mktemp -d)
  echo "  $binary"
  curl -fsSL "$url" | tar xz -C "$tmp"
  find "$tmp" -name "$binary" -type f -exec cp {} "$OUT/bin/$binary" \;
  rm -rf "$tmp"
}

echo "==> Downloading binaries ($OS/$ARCH)"

# --- Rust tools (musl for Linux, apple-darwin for Mac) ---
dl_tar "https://github.com/sharkdp/fd/releases/download/v${FD_VERSION}/fd-v${FD_VERSION}-${RUST_TARGET}.tar.gz" fd
dl_tar "https://github.com/sharkdp/bat/releases/download/v${BAT_VERSION}/bat-v${BAT_VERSION}-${RUST_TARGET}.tar.gz" bat
dl_tar "https://github.com/lsd-rs/lsd/releases/download/v${LSD_VERSION}/lsd-v${LSD_VERSION}-${RUST_TARGET}.tar.gz" lsd
dl_tar "https://github.com/BurntSushi/ripgrep/releases/download/${RG_VERSION}/ripgrep-${RG_VERSION}-${RUST_TARGET}.tar.gz" rg
dl_tar "https://github.com/dandavison/delta/releases/download/${DELTA_VERSION}/delta-${DELTA_VERSION}-${RUST_TARGET}.tar.gz" delta

# --- Go tools (static) ---
dl_tar "https://github.com/junegunn/fzf/releases/download/v${FZF_VERSION}/fzf-${FZF_VERSION}-${OS}_${GOARCH}.tar.gz" fzf
dl_tar "https://github.com/jesseduffield/lazygit/releases/download/v${LAZYGIT_VERSION}/lazygit_${LAZYGIT_VERSION}_${LAZYGIT_OS}_${ARCH_GENERIC}.tar.gz" lazygit

# --- Single-binary tools ---
dl "https://github.com/direnv/direnv/releases/download/v${DIRENV_VERSION}/direnv.${OS}-${GOARCH}" "$OUT/bin/direnv"
dl "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-${JQ_OS}-${GOARCH}" "$OUT/bin/jq"

# --- bat-extras (batman is a shell script — portable) ---
tmp=$(mktemp -d)
curl -fsSL "https://github.com/eth-p/bat-extras/releases/download/v${BAT_EXTRAS_VERSION}/bat-extras-${BAT_EXTRAS_VERSION}.zip" -o "$tmp/bat-extras.zip"
unzip -q "$tmp/bat-extras.zip" -d "$tmp/bat-extras"
cp "$tmp/bat-extras/bin/batman" "$OUT/bin/batman"
rm -rf "$tmp"
echo "  batman"

# --- Neovim (self-contained tarball) ---
if [ "${SKIP_NVIM:-}" != "1" ]; then
  echo "  nvim"
  nvim_tmp=$(mktemp -d)
  curl -fsSL "https://github.com/neovim/neovim/releases/download/v${NVIM_VERSION}/nvim-${NVIM_OS}-${ARCH_GENERIC}.tar.gz" | tar xz -C "$nvim_tmp"
  mv "$nvim_tmp"/nvim-* "$OUT/nvim"
  rm -rf "$nvim_tmp"
fi

# --- Go SDK ---
echo "  go"
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.${OS}-${GOARCH}.tar.gz" | tar xz -C "$OUT"

chmod +x "$OUT/bin"/*
echo "==> Done: $(ls "$OUT/bin" | wc -l | tr -d ' ') binaries + nvim + go"
