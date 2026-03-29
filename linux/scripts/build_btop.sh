#!/bin/sh
# Build btop as a static binary.
# Expects BTOP_VERSION env var.
set -euo pipefail

exec 3>&1  # save stdout for final tar output
exec 1>&2  # redirect all build output to stderr

curl -fsSL "https://github.com/aristocratos/btop/archive/refs/tags/v${BTOP_VERSION}.tar.gz" | tar xz
cd btop-${BTOP_VERSION}
cmake -B build \
  -DCMAKE_BUILD_TYPE=Release \
  -DBTOP_STATIC=ON \
  -DBTOP_GPU=OFF \
  -DBTOP_LTO=ON
cmake --build build -j$(nproc)
find build -name btop -type f | head -1 | xargs -I{} cp {} /tmp/btop
strip /tmp/btop
tar czf - -C /tmp btop >&3  # write tarball to original stdout
