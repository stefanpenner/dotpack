#!/bin/sh
# Build GNU make as a static binary.
# Expects MAKE_VERSION env var.
set -euo pipefail

exec 3>&1  # save stdout for final tar output
exec 1>&2  # redirect all build output to stderr

curl -fsSL "https://ftp.gnu.org/gnu/make/make-${MAKE_VERSION}.tar.gz" | tar xz
cd make-${MAKE_VERSION}
./configure CFLAGS="-Os -DNDEBUG" LDFLAGS="-static"
make -j$(nproc)
strip make
tar czf - make >&3  # write tarball to original stdout
