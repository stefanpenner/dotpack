package cmd

import (
	"fmt"

	"github.com/stefanpenner/dotpack/internal/ssh"
)

// Status shows installed tool versions on a remote host.
func Status(host string) error {
	if host == "" {
		return fmt.Errorf("usage: dotpack status <host>")
	}

	fmt.Printf("==> Versions on %s:\n", host)
	return ssh.RunInteractive(host, `
    _dp="${DOTPACK_PREFIX:-$HOME/.local}"
    export PATH="$_dp/bin:$_dp/git/bin:$_dp/zsh/bin:$_dp/go/bin:$PATH"
    for cmd in zsh git nvim go fzf fd bat rg lsd delta jq direnv lazygit htop dotpack; do
      if command -v "$cmd" > /dev/null 2>&1; then
        case "$cmd" in
          zsh)     ver=$(zsh --version 2>&1) ;;
          git)     ver=$(git --version 2>&1) ;;
          nvim)    ver=$(nvim --version 2>&1 | head -1) ;;
          go)      ver=$(go version 2>&1) ;;
          htop)    ver=$(htop --version 2>&1 | head -1) ;;
          lazygit) ver=$(lazygit --version 2>&1 | head -1) ;;
          dotpack) ver=$(dotpack version 2>&1) ;;
          *)       ver=$("$cmd" --version 2>&1 | head -1) ;;
        esac
        printf "  %-10s %s\n" "$cmd" "$ver"
      else
        printf "  %-10s (not installed)\n" "$cmd"
      fi
    done
  `)
}
