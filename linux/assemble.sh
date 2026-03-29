#!/bin/bash
# Assemble the Linux devlayer bundle from source-compiled + pre-built tools.
# Usage: assemble.sh <output.tar.gz> <arch> <versions.env> <download-script> <git.tar.gz> <zsh.tar.gz> <htop.tar.gz> <btop.tar.gz> <nvim.tar.gz> <make.tar.gz>
set -euo pipefail

OUTPUT="$1"
ARCH="$2"
VERSIONS_ENV="$3"
DOWNLOAD_SCRIPT="$4"
GIT_TAR="$5"
ZSH_TAR="$6"
HTOP_TAR="$7"
BTOP_TAR="$8"
NVIM_TAR="$9"
MAKE_TAR="${10}"

STAGING=$(mktemp -d)
WORKDIR=$(mktemp -d)
trap 'rm -rf "$STAGING" "$WORKDIR"' EXIT

mkdir -p "$STAGING/bin"

# 1. Set up directory structure expected by download-binaries.sh
# (it sources ../versions.env relative to its own location)
mkdir -p "$WORKDIR/scripts"
cp "$DOWNLOAD_SCRIPT" "$WORKDIR/scripts/download-binaries.sh"
cp "$VERSIONS_ENV" "$WORKDIR/versions.env"

echo "==> Downloading pre-built binaries..."
SKIP_NVIM=1 bash "$WORKDIR/scripts/download-binaries.sh" "$STAGING" linux "$ARCH"

# 2. Extract source-compiled tools
echo "==> Adding source-compiled tools..."
mkdir -p "$STAGING/git"
tar xzf "$GIT_TAR" -C "$STAGING/git"
echo "  git"

mkdir -p "$STAGING/zsh"
tar xzf "$ZSH_TAR" -C "$STAGING/zsh"
echo "  zsh"

# Single-binary tools
tmp=$(mktemp -d)
tar xzf "$HTOP_TAR" -C "$tmp"
cp "$tmp/htop" "$STAGING/bin/htop"
chmod +x "$STAGING/bin/htop"
rm -rf "$tmp"
echo "  htop"

tmp=$(mktemp -d)
tar xzf "$BTOP_TAR" -C "$tmp"
cp "$tmp/btop" "$STAGING/bin/btop"
chmod +x "$STAGING/bin/btop"
rm -rf "$tmp"
echo "  btop"

mkdir -p "$STAGING/nvim"
tar xzf "$NVIM_TAR" -C "$STAGING/nvim"
echo "  nvim"

tmp=$(mktemp -d)
tar xzf "$MAKE_TAR" -C "$tmp"
cp "$tmp/make" "$STAGING/bin/make"
chmod +x "$STAGING/bin/make"
rm -rf "$tmp"
echo "  make"

# 3. Create wrapper scripts
echo "==> Creating wrapper scripts..."
cat > "$STAGING/bin/git" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
export GIT_EXEC_PATH="$PREFIX/git/libexec/git-core"
exec "$PREFIX/git/bin/git" "$@"
WRAPPER

cat > "$STAGING/bin/zsh" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
for d in "$PREFIX"/zsh/share/zsh/*/functions; do [ -d "$d" ] && export FPATH="$d${FPATH:+:$FPATH}" && break; done
exec "$PREFIX/zsh/bin/zsh" "$@"
WRAPPER

cat > "$STAGING/bin/nvim" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
export VIMRUNTIME="$PREFIX/nvim/share/nvim/runtime"
exec "$PREFIX/nvim/bin/nvim" "$@"
WRAPPER

cat > "$STAGING/bin/go" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
export GOROOT="$PREFIX/go"
exec "$PREFIX/go/bin/go" "$@"
WRAPPER

cat > "$STAGING/bin/gofmt" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
export GOROOT="$PREFIX/go"
exec "$PREFIX/go/bin/gofmt" "$@"
WRAPPER

cat > "$STAGING/bin/zig" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
exec "$PREFIX/zig/zig" "$@"
WRAPPER

cat > "$STAGING/bin/cc" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
exec "$PREFIX/zig/zig" cc "$@"
WRAPPER

cat > "$STAGING/bin/c++" << 'WRAPPER'
#!/bin/sh
PREFIX="$(cd "$(dirname "$0")/.." && pwd)"
exec "$PREFIX/zig/zig" c++ "$@"
WRAPPER

chmod +x "$STAGING/bin"/*

# 4. Verify
echo "==> Verifying wrappers..."
for f in git zsh nvim go gofmt zig cc c++; do
  head -1 "$STAGING/bin/$f" | grep -q '#!/bin/sh' || { echo "FAIL: bin/$f not a wrapper"; exit 1; }
  echo "  bin/$f: ok"
done

echo "==> Verifying static linkage..."
for f in git/bin/git zsh/bin/zsh bin/htop bin/btop nvim/bin/nvim bin/make; do
  echo "  $f: $(file "$STAGING/$f" | grep -o 'statically linked' || echo 'dynamically linked')"
done

# 5. Generate checksums
find "$STAGING" -type f -executable | sort | xargs sha256sum > "$STAGING/SHA256SUMS" 2>/dev/null || true

# 6. Package
echo "==> Creating bundle..."
tar czf "$OUTPUT" -C "$STAGING" .
echo "==> Bundle created: $(du -h "$OUTPUT" | cut -f1)"
