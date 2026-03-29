#!/bin/sh
# Build htop as a static binary.
# Expects HTOP_VERSION env var.
set -euo pipefail

exec 3>&1  # save stdout for final tar output
exec 1>&2  # redirect all build output to stderr

git clone --depth 1 --branch ${HTOP_VERSION} https://github.com/htop-dev/htop.git
cd htop
./autogen.sh
./configure --enable-static LDFLAGS="-static" CFLAGS="-Os -DNDEBUG"
make -j$(nproc)
strip htop
tar czf - htop >&3  # write tarball to original stdout
