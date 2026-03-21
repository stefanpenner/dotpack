package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stefanpenner/dotpack/internal/ssh"
)

// Push deploys a dotpack bundle to a remote host via SSH.
func Push(host, scriptDir string) error {
	if host == "" {
		return fmt.Errorf("usage: dotpack push <host>")
	}

	// Get remote architecture
	arch, err := ssh.Run(host, "uname -m")
	if err != nil {
		return fmt.Errorf("failed to detect remote arch: %w", err)
	}

	tarball := filepath.Join(scriptDir, fmt.Sprintf("dotpack-linux-%s.tar.gz", arch))
	if _, err := os.Stat(tarball); os.IsNotExist(err) {
		return fmt.Errorf("no bundle found at %s\nRun: dotpack build --arch %s", tarball, arch)
	}

	// Get remote prefix
	remotePrefix, err := ssh.Run(host, `echo "${DOTPACK_PREFIX:-$HOME/.local}"`)
	if err != nil {
		remotePrefix = "$HOME/.local"
	}

	fmt.Printf("==> Deploying to %s (%s)...\n", host, remotePrefix)

	// Create directory and extract
	if err := ssh.RunInteractive(host, fmt.Sprintf("mkdir -p '%s/bin'", remotePrefix)); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := ssh.PipeFile(host, tarball, fmt.Sprintf("tar xzf - -C '%s'", remotePrefix)); err != nil {
		return fmt.Errorf("tar extract: %w", err)
	}

	// Ensure PATH and env are configured (idempotent)
	hasProfile, _ := ssh.Run(host, `grep -q "dotpack managed PATH" ~/.profile 2>/dev/null && echo yes || echo no`)
	if hasProfile != "yes" {
		profileBlock := fmt.Sprintf(`
# dotpack managed PATH
export DOTPACK_PREFIX="${DOTPACK_PREFIX:-%s}"
export PATH="$DOTPACK_PREFIX/bin:$DOTPACK_PREFIX/git/bin:$DOTPACK_PREFIX/zsh/bin:$DOTPACK_PREFIX/go/bin:$PATH"
export GIT_EXEC_PATH="$DOTPACK_PREFIX/git/libexec/git-core"
# SSL certs — pick the first existing cert bundle
for _cert in /etc/ssl/certs/ca-certificates.crt /etc/ssl/cert.pem /etc/pki/tls/certs/ca-bundle.crt; do
  [ -f "$_cert" ] && export GIT_SSL_CAINFO="$_cert" && break
done
unset _cert
`, remotePrefix)
		if err := ssh.PipeBytes(host, []byte(profileBlock), "cat >> ~/.profile"); err != nil {
			return fmt.Errorf("profile update: %w", err)
		}
	}

	fmt.Println()
	if err := Status(host); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("==> Deploy complete. Start a new shell or: source ~/.profile")
	fmt.Printf("    prefix: %s\n", remotePrefix)
	return nil
}
