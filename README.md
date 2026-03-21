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
dotpack build --os darwin --nvim-head  # Build with nvim from HEAD
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
| git | yes | yes | yes |
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
| btop | yes | yes | — |
| dust | yes | yes | yes |
| age | yes | yes | yes |
| batman | yes | yes | — |
| dotpack | yes | yes | yes |

Also bundles zsh plugins (autosuggestions, fast-syntax-highlighting, history-substring-search, powerlevel10k) and fzf shell integration on Linux/macOS.

### Portability

On **Linux**, all binaries are statically linked against musl libc — they run on any Linux distribution with no shared library dependencies.

On **macOS**, nvim, htop, and btop are compiled from source for best portability. Rust and Go tools are downloaded as pre-built releases (already statically linked). All macOS binaries are **best-effort hermetic** — they link against `libSystem.dylib` (always present) but have no other external dependencies. Apple does not support fully static linking, so this is the best achievable.

On **Windows**, binaries are downloaded from upstream releases. Git uses [MinGit](https://github.com/git-for-windows/git) — a portable, self-contained distribution. They don't require package managers but depend on system DLLs.

Some tools require supporting files at runtime (git needs `libexec/`, nvim needs `share/nvim/runtime/`, go needs its SDK, zsh needs function files). These are handled transparently via wrapper scripts in `bin/` that set the correct environment variables (`GIT_EXEC_PATH`, `VIMRUNTIME`, `GOROOT`, `FPATH`) before exec'ing the real binary — no manual configuration needed beyond PATH.

All other tools are single-binary and fully self-contained (on Linux, statically linked; on macOS/Windows, best-effort).

## Shell integration

Wrapper scripts in `bin/` handle `GOROOT`, `GIT_EXEC_PATH`, `VIMRUNTIME`, and `FPATH` automatically, so you only need to add `bin/` to your PATH.

**Bash** (`~/.bashrc` or `~/.bash_profile`):
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**Zsh** (`~/.zshrc`):
```zsh
export PATH="$HOME/.local/bin:$PATH"

# Bundled plugins (optional)
source "$HOME/.local/share/zsh-autosuggestions/zsh-autosuggestions.zsh"
source "$HOME/.local/share/zsh-fast-syntax-highlighting/fast-syntax-highlighting.plugin.zsh"
source "$HOME/.local/share/zsh-history-substring-search/zsh-history-substring-search.zsh"
source "$HOME/.local/share/powerlevel10k/powerlevel10k.zsh-theme"

# fzf keybindings and completion (optional)
source "$HOME/.local/share/fzf/key-bindings.zsh"
source "$HOME/.local/share/fzf/completion.zsh"
```

**Windows (PowerShell profile)**:
```powershell
$env:PATH = "$env:LOCALAPPDATA\dotpack\bin;$env:PATH"
```

**direnv** — if using direnv, also add the hook:
```bash
# bash
eval "$(direnv hook bash)"

# zsh
eval "$(direnv hook zsh)"
```

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

Requires Go 1.21+ and Docker (for Linux builds only). macOS builds compile nvim, htop, and btop from source and require cmake, ninja, autoconf, automake, and libtool (install via `brew install cmake ninja autoconf automake libtool`).
