package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/stefanpenner/dotpack/internal/ssh"
)

const statusScript = `
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
`

// Status shows installed tool versions on a host, or locally if no host given.
func Status(host string) error {
	if host == "" {
		fmt.Println("==> Versions (local):")
		cmd := exec.Command("sh", "-c", statusScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Printf("==> Versions on %s:\n", host)
	return ssh.RunInteractive(host, statusScript)
}
