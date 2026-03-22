#!/bin/bash
# Verify that tools/versions.bzl and versions.env contain the same versions.
# This prevents the two files from drifting apart.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/versions.env"
BZL_FILE="$REPO_ROOT/tools/versions.bzl"

errors=0

# Parse versions.env into KEY=VALUE pairs (skip comments/blanks)
while IFS='=' read -r key value; do
    [[ -z "$key" || "$key" =~ ^# ]] && continue
    key="${key%_VERSION}"  # Strip _VERSION suffix
    # Check if versions.bzl contains this version
    if ! grep -q "\"$key\": \"$value\"" "$BZL_FILE"; then
        echo "MISMATCH: $key=$value in versions.env but not in versions.bzl"
        errors=$((errors + 1))
    fi
done < "$ENV_FILE"

# Check for versions in .bzl that aren't in .env
while IFS= read -r line; do
    # Match lines like: "KEY": "value",
    if [[ "$line" =~ \"([A-Z_]+)\":\ \"([^\"]+)\" ]]; then
        key="${BASH_REMATCH[1]}"
        value="${BASH_REMATCH[2]}"
        env_key="${key}_VERSION"
        if ! grep -q "^${env_key}=" "$ENV_FILE"; then
            echo "MISMATCH: $key=$value in versions.bzl but ${env_key} not in versions.env"
            errors=$((errors + 1))
        fi
    fi
done < "$BZL_FILE"

if [ "$errors" -gt 0 ]; then
    echo "FAIL: $errors version mismatches between versions.env and tools/versions.bzl"
    exit 1
fi

echo "OK: versions.env and tools/versions.bzl are in sync"
