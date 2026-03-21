package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/stefanpenner/dotpack/internal/archive"
	"github.com/stefanpenner/dotpack/internal/docker"
	"github.com/stefanpenner/dotpack/internal/download"
	"github.com/stefanpenner/dotpack/internal/platform"
	"github.com/stefanpenner/dotpack/internal/versions"
)

// Build builds a dotpack bundle for the given OS and architecture.
func Build(args []string, vers *versions.Versions, scriptDir string) error {
	targetOS := "linux"
	targetArch := runtime.GOARCH
	// Normalize Go arch to uname style
	switch targetArch {
	case "amd64":
		targetArch = "x86_64"
	case "arm64":
		targetArch = "aarch64"
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--os":
			i++
			if i < len(args) {
				targetOS = args[i]
			}
		case "--arch":
			i++
			if i < len(args) {
				targetArch = args[i]
			}
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	if targetOS == "darwin" {
		return buildDarwin(targetArch, vers, scriptDir)
	}
	return buildLinux(targetArch, scriptDir)
}

func buildDarwin(arch string, vers *versions.Versions, scriptDir string) error {
	p, err := platform.New("darwin", arch)
	if err != nil {
		return err
	}

	// Use ArchGeneric for display and filenames (arm64, not aarch64, for darwin)
	displayArch := p.ArchGeneric
	fmt.Printf("==> Building dotpack for darwin/%s...\n", displayArch)

	out, err := os.MkdirTemp("", "dotpack-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(out)

	binDir := filepath.Join(out, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	fmt.Println("==> Downloading binaries (darwin/" + displayArch + ")")

	if err := downloadAll(out, p, vers); err != nil {
		return err
	}

	// Copy self into bundle
	self, err := os.Executable()
	if err == nil {
		fmt.Println("  dotpack")
		copyFile(self, filepath.Join(binDir, "dotpack"))
	}

	// Generate checksums
	if err := generateChecksums(out); err != nil {
		return err
	}

	outputFile := filepath.Join(scriptDir, fmt.Sprintf("dotpack-darwin-%s.tar.gz", displayArch))
	if err := archive.CreateTarGz(outputFile, out); err != nil {
		return err
	}

	docker.PrintSize(outputFile)
	fmt.Println("==> Build complete.")
	return nil
}

func buildLinux(arch, scriptDir string) error {
	fmt.Printf("==> Building dotpack for linux/%s (Docker)...\n", arch)

	p, err := platform.New("linux", arch)
	if err != nil {
		return err
	}

	if err := docker.Build(p.DockerPlatform, "dotpack", scriptDir); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}

	outputFile := filepath.Join(scriptDir, fmt.Sprintf("dotpack-linux-%s.tar.gz", arch))
	if err := docker.RunToFile("dotpack", outputFile); err != nil {
		return fmt.Errorf("docker run: %w", err)
	}

	docker.PrintSize(outputFile)
	fmt.Println("==> Build complete.")
	return nil
}

func downloadAll(out string, p *platform.Platform, vers *versions.Versions) error {
	binDir := filepath.Join(out, "bin")

	// Rust tools
	rustTools := []struct {
		name, repo, prefix string
		useTargetFor       bool
	}{
		{"fd", "sharkdp/fd", "fd-v%s-%s", false},
		{"bat", "sharkdp/bat", "bat-v%s-%s", false},
		{"lsd", "lsd-rs/lsd", "lsd-v%s-%s", false},
		{"rg", "BurntSushi/ripgrep", "ripgrep-%s-%s", true},
		{"delta", "dandavison/delta", "delta-%s-%s", true},
	}

	versionKeys := map[string]string{
		"fd": "FD_VERSION", "bat": "BAT_VERSION", "lsd": "LSD_VERSION",
		"rg": "RG_VERSION", "delta": "DELTA_VERSION",
	}

	for _, t := range rustTools {
		ver := vers.Get(versionKeys[t.name])
		target := p.RustTarget
		if t.useTargetFor {
			target = p.RustTargetFor(t.name)
		}
		archiveName := fmt.Sprintf(t.prefix, ver, target)
		vPrefix := ver
		if t.name == "fd" || t.name == "bat" || t.name == "lsd" {
			vPrefix = "v" + ver
		}
		url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.tar.gz", t.repo, vPrefix, archiveName)
		if t.name == "rg" || t.name == "delta" {
			url = fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.tar.gz", t.repo, ver, archiveName)
		}
		if err := download.TarGzBinary(url, binDir, t.name); err != nil {
			return fmt.Errorf("download %s: %w", t.name, err)
		}
	}

	// Go tools
	fzfVer := vers.Get("FZF_VERSION")
	fzfURL := fmt.Sprintf("https://github.com/junegunn/fzf/releases/download/v%s/fzf-%s-%s_%s.tar.gz",
		fzfVer, fzfVer, p.OS, p.GoArch)
	if err := download.TarGzBinary(fzfURL, binDir, "fzf"); err != nil {
		return fmt.Errorf("download fzf: %w", err)
	}

	lazygitVer := vers.Get("LAZYGIT_VERSION")
	lazygitURL := fmt.Sprintf("https://github.com/jesseduffield/lazygit/releases/download/v%s/lazygit_%s_%s_%s.tar.gz",
		lazygitVer, lazygitVer, p.LazygitOS, p.ArchGeneric)
	if err := download.TarGzBinary(lazygitURL, binDir, "lazygit"); err != nil {
		return fmt.Errorf("download lazygit: %w", err)
	}

	// Single-binary tools
	direnvVer := vers.Get("DIRENV_VERSION")
	direnvURL := fmt.Sprintf("https://github.com/direnv/direnv/releases/download/v%s/direnv.%s-%s",
		direnvVer, p.OS, p.GoArch)
	if err := download.File(direnvURL, filepath.Join(binDir, "direnv")); err != nil {
		return fmt.Errorf("download direnv: %w", err)
	}

	jqVer := vers.Get("JQ_VERSION")
	jqURL := fmt.Sprintf("https://github.com/jqlang/jq/releases/download/jq-%s/jq-%s-%s",
		jqVer, p.JqOS, p.GoArch)
	if err := download.File(jqURL, filepath.Join(binDir, "jq")); err != nil {
		return fmt.Errorf("download jq: %w", err)
	}

	// bat-extras (batman is a shell script)
	batExtrasVer := vers.Get("BAT_EXTRAS_VERSION")
	batExtrasURL := fmt.Sprintf("https://github.com/eth-p/bat-extras/releases/download/v%s/bat-extras-%s.zip",
		batExtrasVer, batExtrasVer)
	if err := download.ZipFiles(batExtrasURL, map[string]string{
		fmt.Sprintf("bat-extras-%s/bin/batman", batExtrasVer): filepath.Join(binDir, "batman"),
	}); err != nil {
		// Try alternate zip layout
		if err2 := download.ZipFiles(batExtrasURL, map[string]string{
			"bin/batman": filepath.Join(binDir, "batman"),
		}); err2 != nil {
			return fmt.Errorf("download batman: %w", err)
		}
	}
	fmt.Println("  batman")

	// Neovim
	nvimVer := vers.Get("NVIM_VERSION")
	nvimURL := fmt.Sprintf("https://github.com/neovim/neovim/releases/download/v%s/nvim-%s-%s.tar.gz",
		nvimVer, p.NvimOS, p.ArchGeneric)
	fmt.Println("  nvim")
	nvimDir := filepath.Join(out, "nvim")
	if err := download.TarGzFull(nvimURL, nvimDir, 1); err != nil {
		return fmt.Errorf("download nvim: %w", err)
	}

	// Go SDK
	goVer := vers.Get("GO_VERSION")
	goURL := fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.tar.gz", goVer, p.OS, p.GoArch)
	fmt.Println("  go")
	if err := download.TarGzFull(goURL, out, 0); err != nil {
		return fmt.Errorf("download go: %w", err)
	}

	// fzf shell integration
	fmt.Println("  fzf shell integration")
	fzfShellDir := filepath.Join(out, "share", "fzf")
	if err := os.MkdirAll(fzfShellDir, 0755); err != nil {
		return err
	}
	fzfSrcURL := fmt.Sprintf("https://github.com/junegunn/fzf/archive/refs/tags/v%s.tar.gz", fzfVer)
	fzfTmp, err := os.MkdirTemp("", "dotpack-fzf-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(fzfTmp)
	if err := download.TarGzFull(fzfSrcURL, fzfTmp, 0); err != nil {
		return fmt.Errorf("download fzf shell: %w", err)
	}
	// Find and copy shell files
	entries, err := os.ReadDir(fzfTmp)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			shellDir := filepath.Join(fzfTmp, e.Name(), "shell")
			if _, err := os.Stat(shellDir); err == nil {
				for _, name := range []string{"key-bindings.zsh", "completion.zsh"} {
					src := filepath.Join(shellDir, name)
					if _, err := os.Stat(src); err == nil {
						copyFile(src, filepath.Join(fzfShellDir, name))
					}
				}
				break
			}
		}
	}

	// Zsh plugins
	shareDir := filepath.Join(out, "share")
	plugins := []struct {
		name, repo, versionKey string
	}{
		{"zsh-autosuggestions", "zsh-users/zsh-autosuggestions", "ZSH_AUTOSUGGESTIONS_VERSION"},
		{"zsh-fast-syntax-highlighting", "zdharma-continuum/fast-syntax-highlighting", "FAST_SYNTAX_HIGHLIGHTING_VERSION"},
		{"zsh-history-substring-search", "zsh-users/zsh-history-substring-search", "ZSH_HISTORY_SUBSTRING_SEARCH_VERSION"},
		{"powerlevel10k", "romkatv/powerlevel10k", "POWERLEVEL10K_VERSION"},
	}
	for _, plug := range plugins {
		ver := vers.Get(plug.versionKey)
		url := fmt.Sprintf("https://github.com/%s/archive/refs/tags/%s.tar.gz", plug.repo, ver)
		fmt.Printf("  %s\n", plug.name)
		destDir := filepath.Join(shareDir, plug.name)
		if err := download.TarGzToDir(url, destDir); err != nil {
			return fmt.Errorf("download %s: %w", plug.name, err)
		}
	}

	// chmod +x all binaries
	entries2, _ := os.ReadDir(binDir)
	for _, e := range entries2 {
		os.Chmod(filepath.Join(binDir, e.Name()), 0755)
	}

	fmt.Printf("==> Done: %d binaries + nvim + go + plugins\n", len(entries2))
	return nil
}

func generateChecksums(dir string) error {
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if info.Mode()&0111 != 0 {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)

	f, err := os.Create(filepath.Join(dir, "SHA256SUMS"))
	if err != nil {
		return err
	}
	defer f.Close()

	for _, path := range files {
		hash, err := sha256File(path)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(dir, path)
		fmt.Fprintf(f, "%s  %s\n", hash, rel)
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode()|0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// FindScriptDir returns the directory containing the dotpack source files.
// It checks for the repo checkout first (Dockerfile exists), then falls back
// to a temp dir with embedded files.
func FindScriptDir(embeddedDockerfile, embeddedVersionsEnv, embeddedDownloadScript string) (string, bool, error) {
	// Check if we're in a repo checkout
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
			return dir, false, nil
		}
	}

	// Check current working directory
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "Dockerfile")); err == nil {
			return cwd, false, nil
		}
	}

	// Fall back to temp dir with embedded files
	tmp, err := os.MkdirTemp("", "dotpack-context-*")
	if err != nil {
		return "", false, err
	}

	if err := os.WriteFile(filepath.Join(tmp, "Dockerfile"), []byte(embeddedDockerfile), 0644); err != nil {
		return "", true, err
	}
	if err := os.WriteFile(filepath.Join(tmp, "versions.env"), []byte(embeddedVersionsEnv), 0644); err != nil {
		return "", true, err
	}
	scriptsDir := filepath.Join(tmp, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return "", true, err
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "download-binaries.sh"), []byte(embeddedDownloadScript), 0755); err != nil {
		return "", true, err
	}

	return tmp, true, nil
}

// VersionSummary returns a multi-line string of tool versions for display.
func VersionSummary(vers *versions.Versions) string {
	keys := []struct{ label, key string }{
		{"fzf", "FZF_VERSION"}, {"fd", "FD_VERSION"}, {"bat", "BAT_VERSION"},
		{"lsd", "LSD_VERSION"}, {"rg", "RG_VERSION"}, {"delta", "DELTA_VERSION"},
		{"lazygit", "LAZYGIT_VERSION"}, {"jq", "JQ_VERSION"}, {"direnv", "DIRENV_VERSION"},
		{"nvim", "NVIM_VERSION"}, {"go", "GO_VERSION"}, {"git", "GIT_VERSION"},
		{"htop", "HTOP_VERSION"},
	}
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "  %-10s %s\n", k.label, vers.Get(k.key))
	}
	return b.String()
}
