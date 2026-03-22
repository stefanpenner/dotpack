#!/bin/bash
# Regenerate the VERSIONS dict in tools/versions.bzl from versions.env.
# Used by update-check workflow to keep both files in sync.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/versions.env"
BZL_FILE="$REPO_ROOT/tools/versions.bzl"

while IFS='=' read -r key value; do
    [[ -z "$key" || "$key" =~ ^# ]] && continue
    short_key="${key%_VERSION}"
    # Update the version in versions.bzl
    sed -i.bak "s/\"${short_key}\": \"[^\"]*\"/\"${short_key}\": \"${value}\"/" "$BZL_FILE"
done < "$ENV_FILE"

rm -f "${BZL_FILE}.bak"
echo "tools/versions.bzl updated from versions.env"
