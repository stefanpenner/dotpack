"""Pinned tool versions for Bazel builds.

These MUST match versions.env (the single source of truth, embedded by Go CLI).
Run `bazel test //tools:versions_sync_test` to verify they're in sync.
"""

VERSIONS = {
    "FZF": "0.70.0",
    "FD": "10.4.2",
    "BAT": "0.26.1",
    "EZA": "0.23.4",
    "RG": "15.1.0",
    "DELTA": "0.18.2",
    "LAZYGIT": "0.60.0",
    "BAT_EXTRAS": "2024.08.24",
    "JQ": "1.8.1",
    "DIRENV": "2.37.1",
    "NVIM": "0.11.6",
    "GO": "1.26.1",
    "GIT": "2.53.0",
    "GIT_WINDOWS": "2.53.0.2",
    "ZSH": "5.9",
    "HTOP": "3.4.1",
    "BTOP": "1.4.6",
    "DUST": "1.2.4",
    "AGE": "1.3.1",
    "ZIG": "0.15.2",
    "MAKE": "4.4.1",
    "NCURSES": "6.5",
    # Zsh plugins
    "ZSH_AUTOSUGGESTIONS": "v0.7.1",
    "FAST_SYNTAX_HIGHLIGHTING": "v1.56",
    "ZSH_HISTORY_SUBSTRING_SEARCH": "v1.1.0",
    "POWERLEVEL10K": "v1.20.0",
}

# Rust target triples per platform
RUST_TARGETS = {
    "linux_amd64": "x86_64-unknown-linux-musl",
    "linux_arm64": "aarch64-unknown-linux-musl",
    "linux_arm64_gnu": "aarch64-unknown-linux-gnu",
    "darwin_amd64": "x86_64-apple-darwin",
    "darwin_arm64": "aarch64-apple-darwin",
    "windows_amd64": "x86_64-pc-windows-msvc",
}

# Platforms we build for
PLATFORMS = ["linux_amd64", "linux_arm64", "darwin_arm64"]
