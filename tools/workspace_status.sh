#!/bin/sh
# Provides version stamp for Bazel builds
echo "STABLE_VERSION $(git describe --tags --always --dirty 2>/dev/null || echo dev)"
