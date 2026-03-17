# dotpack

Hermetic, self-contained tool bundle for Linux and macOS. All binaries are statically linked — no system dependencies required.

## Install

```bash
curl -fsSL https://github.com/stefanpenner/dotpack/releases/latest/download/install.sh | bash
```

Then add to your shell profile:

```bash
export PATH="$HOME/.local/bin:$HOME/.local/git/bin:$HOME/.local/zsh/bin:$HOME/.local/go/bin:$PATH"
```

### Verify provenance

```bash
gh attestation verify dotpack-linux-x86_64.tar.gz --repo stefanpenner/dotpack
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

## With dotfiles

dotpack provides the tools. [dotfiles](https://github.com/stefanpenner/dotfiles) provides the config.

```bash
# 1. Install tools
curl -fsSL https://github.com/stefanpenner/dotpack/releases/latest/download/install.sh | bash
export PATH="$HOME/.local/bin:$HOME/.local/git/bin:$HOME/.local/zsh/bin:$HOME/.local/go/bin:$PATH"

# 2. Install config
git clone git@github.com:stefanpenner/dotfiles.git ~/.dotfiles
cd ~/.dotfiles && make install
```

## NAS / remote host deployment

```bash
# Build (once, locally)
./dotpack build

# Deploy to any host
./dotpack push nas
./dotpack status nas
```

## Build locally

```bash
# Linux (via Docker)
./dotpack build

# macOS (native)
./dotpack build --os darwin

# Install the built bundle
./dotpack install
```

## Release

Create a GitHub release to trigger builds for all platforms:
- `dotpack-linux-x86_64.tar.gz`
- `dotpack-linux-aarch64.tar.gz`
- `dotpack-darwin-x86_64.tar.gz`
- `dotpack-darwin-arm64.tar.gz`

Each artifact includes SHA256 checksums and SLSA build provenance.

## Supply chain security

- All GHA actions pinned by commit SHA
- SLSA build provenance via `actions/attest-build-provenance`
- SHA256 checksums for every release artifact
- Dependabot keeps GHA actions updated
- Weekly automated checks for new tool versions (opens PRs)
- All tool versions pinned in [`versions.env`](versions.env)
