# dotpack

Hermetic, self-contained tool bundle for Linux and macOS. All binaries are statically linked — no system dependencies required. dotpack itself is a single Go binary.

## Quick start

```bash
# Install dotpack + tools
curl -fsSL https://github.com/stefanpenner/dotpack/releases/latest/download/install.sh | bash
source ~/.profile
```

Or download the dotpack CLI directly:

```bash
# Download the CLI (pick your platform)
curl -fsSL https://github.com/stefanpenner/dotpack/releases/latest/download/dotpack-darwin-arm64 -o dotpack
chmod +x dotpack

# Build and install the tool bundle
./dotpack build --os darwin
./dotpack install
```

## Usage

```bash
dotpack build                    # Build linux bundle (Docker)
dotpack build --os darwin        # Build macOS bundle
dotpack push nas                 # Deploy to remote host via SSH
dotpack status nas               # Check installed versions on host
dotpack install                  # Install bundle locally
dotpack clean                    # Remove build artifacts
dotpack version                  # Print dotpack version
dotpack versions                 # Print bundled tool versions
```

## Deploy to a remote host

```bash
dotpack build                    # Builds linux bundle via Docker
dotpack push nas                 # Deploys to nas:~/.local/
dotpack status nas               # Verify everything works
```

## What's included

| Tool | Description |
|------|-------------|
| zsh | Shell (static, with modules) |
| git | Version control (static, with HTTPS) |
| nvim | Neovim editor (static) |
| go | Go SDK |
| fzf | Fuzzy finder |
| fd | Better find |
| bat | Better cat |
| rg | ripgrep |
| lsd | Better ls |
| delta | Better diff |
| jq | JSON processor |
| direnv | Directory environments |
| lazygit | Git TUI |
| htop | Process viewer (static) |
| batman | Man pages with syntax highlighting |
| dotpack | This tool (self-updating) |

Also bundles zsh plugins (autosuggestions, fast-syntax-highlighting, history-substring-search, powerlevel10k) and fzf shell integration.

## With dotfiles

dotpack provides the tools. [dotfiles](https://github.com/stefanpenner/dotfiles) provides the config.

```bash
# 1. Install tools
dotpack build --os darwin && dotpack install

# 2. Install config
git clone git@github.com:stefanpenner/dotfiles.git ~/.dotfiles
cd ~/.dotfiles && make install
```

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `DOTPACK_PREFIX` | `~/.local` | Install location for all tools |
| `XDG_DATA_HOME` | `~/.local/share` | Plugin/data search path (used by zsh config) |

## Updating tool versions

Edit `versions.env`, commit, push, and create a release:

```bash
vim versions.env
git commit -am "bump fd to 10.5.0"
gh release create v0.2.0
```

A weekly GHA workflow also checks for new upstream versions and opens PRs automatically.

## Supply chain security

- All GHA actions pinned by commit SHA
- SLSA build provenance via `actions/attest-build-provenance` (public repos)
- SHA256 checksums for every release artifact
- Dependabot keeps GHA actions updated
- Weekly automated checks for new tool versions (opens PRs)
- All tool versions pinned in [`versions.env`](versions.env)

## Building from source

```bash
make build-go          # Build the dotpack CLI
make build             # Build linux bundle (Docker)
make build-darwin      # Build macOS bundle
make install           # Install locally
```

Requires Go 1.21+ and Docker (for Linux builds).
