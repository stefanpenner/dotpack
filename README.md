# dotpack

Hermetic, self-contained tool bundle for Linux, macOS, and Windows. All binaries are statically linked — no system dependencies required. dotpack itself is a single Go binary.

## Quick start

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/stefanpenner/dotpack/master/scripts/install.sh | bash
source ~/.profile
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/stefanpenner/dotpack/master/scripts/install.ps1 | iex
```

Or download the CLI directly and build:

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
dotpack build --os windows       # Build Windows bundle
dotpack push nas                 # Deploy to remote host via SSH
dotpack status                   # Check installed versions locally
dotpack status nas               # Check installed versions on host
dotpack upgrade                  # Download and install latest release
dotpack install                  # Install bundle locally
dotpack clean                    # Remove build artifacts
dotpack version                  # Print dotpack version
dotpack versions                 # Print bundled tool versions
```

## What's included

| Tool | Linux | macOS | Windows |
|------|:-----:|:-----:|:-------:|
| zsh | yes | yes | — |
| git | yes | yes | — |
| nvim | yes | yes | yes |
| go | yes | yes | yes |
| fzf | yes | yes | yes |
| fd | yes | yes | yes |
| bat | yes | yes | yes |
| rg (ripgrep) | yes | yes | yes |
| lsd | yes | yes | yes |
| delta | yes | yes | yes |
| jq | yes | yes | yes |
| direnv | yes | yes | yes |
| lazygit | yes | yes | yes |
| htop | yes | yes | — |
| batman | yes | yes | — |
| dotpack | yes | yes | yes |

Also bundles zsh plugins (autosuggestions, fast-syntax-highlighting, history-substring-search, powerlevel10k) and fzf shell integration on Linux/macOS.

## Deploy to a remote host

```bash
dotpack build                    # Builds linux bundle via Docker
dotpack push nas                 # Deploys to nas:~/.local/
dotpack status nas               # Verify everything works
```

## With dotfiles

dotpack provides the tools. [dotfiles](https://github.com/stefanpenner/dotfiles) provides the config.

```bash
curl -fsSL https://raw.githubusercontent.com/stefanpenner/dotfiles/master/bootstrap.sh | sh
```

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `DOTPACK_PREFIX` | `~/.local` (unix) / `%LOCALAPPDATA%\dotpack` (Windows) | Install location for all tools |
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
- SLSA build provenance via `actions/attest-build-provenance`
- SHA256 checksums for every release artifact
- Dependabot keeps GHA actions updated
- Weekly automated checks for new tool versions (opens PRs)
- All tool versions pinned in [`versions.env`](versions.env)

## Building from source

```bash
make build-go          # Build the dotpack CLI
make build             # Build linux bundle (Docker)
make build-darwin      # Build macOS bundle
make build-windows     # Build Windows bundle
make install           # Install locally
make test              # Run tests on all platforms
```

Requires Go 1.21+ and Docker (for Linux builds only).
