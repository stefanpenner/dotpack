package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
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

// buildOptions holds flags parsed from the build command line.
type buildOptions struct {
	os       string
	arch     string
	nvimHead bool // build nvim from HEAD instead of stable/pinned version
}

// Build builds a dotpack bundle for the given OS and architecture.
func Build(args []string, vers *versions.Versions, scriptDir string) error {
	opts := buildOptions{
		os:   "linux",
		arch: runtime.GOARCH,
	}
	// Normalize Go arch to uname style
	switch opts.arch {
	case "amd64":
		opts.arch = "x86_64"
	case "arm64":
		opts.arch = "aarch64"
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--os":
			i++
			if i < len(args) {
				opts.os = args[i]
			}
		case "--arch":
			i++
			if i < len(args) {
				opts.arch = args[i]
			}
		case "--nvim-head":
			opts.nvimHead = true
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	switch opts.os {
	case "darwin":
		return buildDarwin(opts, vers, scriptDir)
	case "windows":
		return buildWindows(opts.arch, vers, scriptDir)
	default:
		return buildLinux(opts.arch, scriptDir)
	}
}

func buildDarwin(opts buildOptions, vers *versions.Versions, scriptDir string) error {
	p, err := platform.New("darwin", opts.arch)
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

	// Skip nvim download — we build it from source
	if err := downloadAll(out, p, vers, map[string]bool{"nvim": true}); err != nil {
		return err
	}

	// Build tools from source for best portability
	fmt.Println("==> Building tools from source...")
	if err := buildNvim(out, vers, opts.nvimHead); err != nil {
		return err
	}
	if err := buildHtop(binDir, vers); err != nil {
		return err
	}
	if err := buildBtop(binDir, vers); err != nil {
		return err
	}

	// Create wrapper scripts in bin/ for tools with runtime dependencies
	wrappers := []wrapper{
		{"nvim", "nvim/bin/nvim", map[string]string{"VIMRUNTIME": "$PREFIX/nvim/share/nvim/runtime"}},
		{"go", "go/bin/go", map[string]string{"GOROOT": "$PREFIX/go"}},
		{"gofmt", "go/bin/gofmt", map[string]string{"GOROOT": "$PREFIX/go"}},
	}
	if err := createUnixWrappers(binDir, wrappers); err != nil {
		return err
	}

	// Copy self into bundle
	if self, err := os.Executable(); err == nil {
		fmt.Println("  dotpack")
		if err := copyFile(self, filepath.Join(binDir, "dotpack")); err != nil {
			return fmt.Errorf("copy dotpack: %w", err)
		}
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

func buildWindows(arch string, vers *versions.Versions, scriptDir string) error {
	p, err := platform.New("windows", arch)
	if err != nil {
		return err
	}

	fmt.Printf("==> Building dotpack for windows/%s...\n", p.ArchGeneric)

	out, err := os.MkdirTemp("", "dotpack-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(out)

	binDir := filepath.Join(out, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	fmt.Println("==> Downloading binaries (windows/" + p.ArchGeneric + ")")

	if err := downloadAll(out, p, vers, nil); err != nil {
		return err
	}

	// Create .cmd wrapper scripts for Windows
	wrappers := []wrapper{
		{"nvim", `nvim\bin\nvim.exe`, map[string]string{"VIMRUNTIME": `%PREFIX%\nvim\share\nvim\runtime`}},
		{"go", `go\bin\go.exe`, map[string]string{"GOROOT": `%PREFIX%\go`}},
		{"gofmt", `go\bin\gofmt.exe`, map[string]string{"GOROOT": `%PREFIX%\go`}},
	}
	if err := createWindowsWrappers(binDir, wrappers); err != nil {
		return err
	}

	// Copy self into bundle
	if self, err := os.Executable(); err == nil {
		fmt.Println("  dotpack")
		destName := "dotpack"
		if runtime.GOOS == "windows" {
			destName = "dotpack.exe"
		}
		if err := copyFile(self, filepath.Join(binDir, destName)); err != nil {
			return fmt.Errorf("copy dotpack: %w", err)
		}
	}

	// Generate checksums
	if err := generateChecksums(out); err != nil {
		return err
	}

	outputFile := filepath.Join(scriptDir, fmt.Sprintf("dotpack-windows-%s.zip", p.ArchGeneric))
	if err := archive.CreateZip(outputFile, out); err != nil {
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

// buildBtop builds btop from source using cmake.
func buildBtop(binDir string, vers *versions.Versions) error {
	ver := vers.Get("BTOP_VERSION")
	fmt.Println("  btop (building from source)")

	srcDir, err := os.MkdirTemp("", "dotpack-btop-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(srcDir)

	// Download source tarball
	url := fmt.Sprintf("https://github.com/aristocratos/btop/archive/refs/tags/v%s.tar.gz", ver)
	if err := download.TarGzFull(url, srcDir, 0); err != nil {
		return fmt.Errorf("download btop source: %w", err)
	}

	// Find extracted directory
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	var srcRoot string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "btop-") {
			srcRoot = filepath.Join(srcDir, e.Name())
			break
		}
	}
	if srcRoot == "" {
		return fmt.Errorf("btop source directory not found")
	}

	buildDir := filepath.Join(srcRoot, "build")

	// cmake configure
	configure := exec.Command("cmake", "-B", buildDir, "-S", srcRoot,
		"-DCMAKE_BUILD_TYPE=Release",
		"-DBTOP_GPU=OFF",
		"-DBTOP_LTO=ON",
	)
	configure.Stdout = os.Stdout
	configure.Stderr = os.Stderr
	if err := configure.Run(); err != nil {
		return fmt.Errorf("btop cmake configure: %w", err)
	}

	// cmake build
	build := exec.Command("cmake", "--build", buildDir, "--config", "Release")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("btop cmake build: %w", err)
	}

	// Find and copy binary — cmake may place it in build/ or build/bin/
	btopBin := filepath.Join(buildDir, "btop")
	if _, err := os.Stat(btopBin); err != nil {
		btopBin = filepath.Join(buildDir, "bin", "btop")
	}
	if err := copyFile(btopBin, filepath.Join(binDir, "btop")); err != nil {
		return fmt.Errorf("copy btop: %w", err)
	}

	return nil
}

// buildNvim builds neovim from source using cmake.
// If head is true, builds from the latest HEAD of the main branch;
// otherwise builds the stable tag (or pinned NVIM_VERSION).
func buildNvim(outDir string, vers *versions.Versions, head bool) error {
	label := "stable"
	if head {
		label = "HEAD"
	}
	fmt.Printf("  nvim (building %s from source)\n", label)

	srcDir, err := os.MkdirTemp("", "dotpack-nvim-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(srcDir)

	// Clone source
	cloneArgs := []string{"clone", "--depth", "1"}
	if head {
		cloneArgs = append(cloneArgs, "https://github.com/neovim/neovim.git")
	} else {
		cloneArgs = append(cloneArgs, "--branch", "stable", "https://github.com/neovim/neovim.git")
	}
	srcRoot := filepath.Join(srcDir, "neovim")
	cloneArgs = append(cloneArgs, srcRoot)

	clone := exec.Command("git", cloneArgs...)
	clone.Stdout = os.Stdout
	clone.Stderr = os.Stderr
	if err := clone.Run(); err != nil {
		return fmt.Errorf("nvim git clone: %w", err)
	}

	installDir := filepath.Join(outDir, "nvim")

	// cmake configure + build + install
	build := exec.Command("make",
		"CMAKE_BUILD_TYPE=Release",
		fmt.Sprintf("CMAKE_INSTALL_PREFIX=%s", installDir),
		fmt.Sprintf("-j%d", runtime.NumCPU()),
	)
	build.Dir = srcRoot
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("nvim build: %w", err)
	}

	install := exec.Command("make", "install")
	install.Dir = srcRoot
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("nvim install: %w", err)
	}

	return nil
}

// buildHtop builds htop from source using autotools.
func buildHtop(binDir string, vers *versions.Versions) error {
	ver := vers.Get("HTOP_VERSION")
	fmt.Println("  htop (building from source)")

	srcDir, err := os.MkdirTemp("", "dotpack-htop-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(srcDir)

	// Clone source at the pinned version
	srcRoot := filepath.Join(srcDir, "htop")
	clone := exec.Command("git", "clone", "--depth", "1", "--branch", ver,
		"https://github.com/htop-dev/htop.git", srcRoot)
	clone.Stdout = os.Stdout
	clone.Stderr = os.Stderr
	if err := clone.Run(); err != nil {
		return fmt.Errorf("htop git clone: %w", err)
	}

	// autogen
	autogen := exec.Command("./autogen.sh")
	autogen.Dir = srcRoot
	autogen.Stdout = os.Stdout
	autogen.Stderr = os.Stderr
	if err := autogen.Run(); err != nil {
		return fmt.Errorf("htop autogen: %w", err)
	}

	// configure
	configure := exec.Command("./configure", "CFLAGS=-Os -DNDEBUG")
	configure.Dir = srcRoot
	configure.Stdout = os.Stdout
	configure.Stderr = os.Stderr
	if err := configure.Run(); err != nil {
		return fmt.Errorf("htop configure: %w", err)
	}

	// build
	make := exec.Command("make", fmt.Sprintf("-j%d", runtime.NumCPU()))
	make.Dir = srcRoot
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("htop build: %w", err)
	}

	// copy binary
	if err := copyFile(filepath.Join(srcRoot, "htop"), filepath.Join(binDir, "htop")); err != nil {
		return fmt.Errorf("copy htop: %w", err)
	}

	return nil
}

func downloadAll(out string, p *platform.Platform, vers *versions.Versions, skip map[string]bool) error {
	binDir := filepath.Join(out, "bin")
	exe := p.ExeSuffix

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
		{"dust", "bootandy/dust", "dust-v%s-%s", true},
	}

	versionKeys := map[string]string{
		"fd": "FD_VERSION", "bat": "BAT_VERSION", "lsd": "LSD_VERSION",
		"rg": "RG_VERSION", "delta": "DELTA_VERSION", "dust": "DUST_VERSION",
	}

	archiveExt := p.RustArchiveExt()

	for _, t := range rustTools {
		if p.SkipTool(t.name) {
			continue
		}
		ver := vers.Get(versionKeys[t.name])
		target := p.RustTarget
		if t.useTargetFor {
			target = p.RustTargetFor(t.name)
		}
		archiveName := fmt.Sprintf(t.prefix, ver, target)
		vPrefix := ver
		if t.name == "fd" || t.name == "bat" || t.name == "lsd" || t.name == "dust" {
			vPrefix = "v" + ver
		}
		url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.%s", t.repo, vPrefix, archiveName, archiveExt)
		if t.name == "rg" || t.name == "delta" {
			url = fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.%s", t.repo, ver, archiveName, archiveExt)
		}
		binaryName := t.name + exe
		if archiveExt == "zip" {
			if err := download.ZipBinary(url, binDir, binaryName); err != nil {
				return fmt.Errorf("download %s: %w", t.name, err)
			}
		} else {
			if err := download.TarGzBinary(url, binDir, binaryName); err != nil {
				return fmt.Errorf("download %s: %w", t.name, err)
			}
		}
	}

	// Go tools
	fzfVer := vers.Get("FZF_VERSION")
	fzfExt := p.FzfArchiveExt()
	fzfURL := fmt.Sprintf("https://github.com/junegunn/fzf/releases/download/v%s/fzf-%s-%s_%s.%s",
		fzfVer, fzfVer, p.OS, p.GoArch, fzfExt)
	fzfBinary := "fzf" + exe
	if fzfExt == "zip" {
		if err := download.ZipBinary(fzfURL, binDir, fzfBinary); err != nil {
			return fmt.Errorf("download fzf: %w", err)
		}
	} else {
		if err := download.TarGzBinary(fzfURL, binDir, fzfBinary); err != nil {
			return fmt.Errorf("download fzf: %w", err)
		}
	}

	lazygitVer := vers.Get("LAZYGIT_VERSION")
	lazygitExt := p.LazygitArchiveExt()
	lazygitURL := fmt.Sprintf("https://github.com/jesseduffield/lazygit/releases/download/v%s/lazygit_%s_%s_%s.%s",
		lazygitVer, lazygitVer, p.LazygitOS, p.ArchGeneric, lazygitExt)
	lazygitBinary := "lazygit" + exe
	if lazygitExt == "zip" {
		if err := download.ZipBinary(lazygitURL, binDir, lazygitBinary); err != nil {
			return fmt.Errorf("download lazygit: %w", err)
		}
	} else {
		if err := download.TarGzBinary(lazygitURL, binDir, lazygitBinary); err != nil {
			return fmt.Errorf("download lazygit: %w", err)
		}
	}

	// age (encryption tool — two binaries: age, age-keygen)
	ageVer := vers.Get("AGE_VERSION")
	ageExt := "tar.gz"
	if p.IsWindows() {
		ageExt = "zip"
	}
	ageURL := fmt.Sprintf("https://github.com/FiloSottile/age/releases/download/v%s/age-v%s-%s-%s.%s",
		ageVer, ageVer, p.OS, p.GoArch, ageExt)
	for _, ageBin := range []string{"age", "age-keygen"} {
		binaryName := ageBin + exe
		if ageExt == "zip" {
			if err := download.ZipBinary(ageURL, binDir, binaryName); err != nil {
				return fmt.Errorf("download %s: %w", ageBin, err)
			}
		} else {
			if err := download.TarGzBinary(ageURL, binDir, binaryName); err != nil {
				return fmt.Errorf("download %s: %w", ageBin, err)
			}
		}
	}

	// Single-binary tools
	direnvVer := vers.Get("DIRENV_VERSION")
	direnvURL := fmt.Sprintf("https://github.com/direnv/direnv/releases/download/v%s/direnv.%s-%s",
		direnvVer, p.OS, p.GoArch)
	direnvDest := "direnv" + exe
	if err := download.File(direnvURL, filepath.Join(binDir, direnvDest)); err != nil {
		return fmt.Errorf("download direnv: %w", err)
	}

	jqVer := vers.Get("JQ_VERSION")
	jqBinary := "jq" + exe
	var jqURL string
	if p.IsWindows() {
		jqURL = fmt.Sprintf("https://github.com/jqlang/jq/releases/download/jq-%s/jq-windows-%s.exe",
			jqVer, p.GoArch)
	} else {
		jqURL = fmt.Sprintf("https://github.com/jqlang/jq/releases/download/jq-%s/jq-%s-%s",
			jqVer, p.JqOS, p.GoArch)
	}
	if err := download.File(jqURL, filepath.Join(binDir, jqBinary)); err != nil {
		return fmt.Errorf("download jq: %w", err)
	}

	// bat-extras (batman is a shell script — skip on Windows)
	if !p.SkipTool("batman") {
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
	}

	// Neovim (skip if building from source)
	if !skip["nvim"] {
		nvimVer := vers.Get("NVIM_VERSION")
		nvimArchiveName := p.NvimArchiveName(nvimVer)
		nvimExt := p.NvimArchiveExt()
		nvimURL := fmt.Sprintf("https://github.com/neovim/neovim/releases/download/v%s/%s.%s",
			nvimVer, nvimArchiveName, nvimExt)
		fmt.Println("  nvim")
		nvimDir := filepath.Join(out, "nvim")
		if nvimExt == "zip" {
			if err := download.ZipFull(nvimURL, nvimDir, 1); err != nil {
				return fmt.Errorf("download nvim: %w", err)
			}
		} else {
			if err := download.TarGzFull(nvimURL, nvimDir, 1); err != nil {
				return fmt.Errorf("download nvim: %w", err)
			}
		}
	}

	// Go SDK
	goVer := vers.Get("GO_VERSION")
	goExt := p.GoArchiveExt()
	goURL := fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.%s", goVer, p.OS, p.GoArch, goExt)
	fmt.Println("  go")
	if goExt == "zip" {
		if err := download.ZipFull(goURL, out, 0); err != nil {
			return fmt.Errorf("download go: %w", err)
		}
	} else {
		if err := download.TarGzFull(goURL, out, 0); err != nil {
			return fmt.Errorf("download go: %w", err)
		}
	}

	// fzf shell integration (not useful on Windows)
	if !p.IsWindows() {
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
	}

	// Zsh plugins (not useful on Windows)
	if !p.IsWindows() {
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
	}

	// chmod +x all binaries
	entries2, _ := os.ReadDir(binDir)
	for _, e := range entries2 {
		os.Chmod(filepath.Join(binDir, e.Name()), 0755)
	}

	fmt.Printf("==> Done: %d binaries + nvim + go", len(entries2))
	if !p.IsWindows() {
		fmt.Print(" + plugins")
	}
	fmt.Println()
	return nil
}

func generateChecksums(dir string) error {
	var files []string
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if info.Mode()&0111 != 0 {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk %s: %w", dir, err)
	}
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

// wrapper defines a tool that needs a wrapper script to set env vars before
// exec'ing the real binary.
type wrapper struct {
	name     string            // binary name in bin/ (e.g., "git")
	realPath string            // relative path from PREFIX to real binary
	envVars  map[string]string // env var name -> value template ($PREFIX or %PREFIX%)
}

// createUnixWrappers writes POSIX shell wrapper scripts into binDir.
// Each script resolves its own location, sets environment variables, and
// exec's the real binary.
func createUnixWrappers(binDir string, wrappers []wrapper) error {
	for _, w := range wrappers {
		var b strings.Builder
		b.WriteString("#!/bin/sh\n")
		b.WriteString(`PREFIX="$(cd "$(dirname "$0")/.." && pwd)"` + "\n")
		for k, v := range w.envVars {
			expanded := strings.ReplaceAll(v, "$PREFIX", `"$PREFIX"`)
			// For FPATH, append existing value
			if k == "FPATH" {
				fmt.Fprintf(&b, "export %s=%s${%s:+:$%s}\n", k, expanded, k, k)
			} else {
				fmt.Fprintf(&b, "export %s=%s\n", k, expanded)
			}
		}
		fmt.Fprintf(&b, "exec \"$PREFIX/%s\" \"$@\"\n", w.realPath)

		path := filepath.Join(binDir, w.name)
		if err := os.WriteFile(path, []byte(b.String()), 0755); err != nil {
			return fmt.Errorf("wrapper %s: %w", w.name, err)
		}
	}
	return nil
}

// createWindowsWrappers writes .cmd wrapper scripts into binDir.
func createWindowsWrappers(binDir string, wrappers []wrapper) error {
	for _, w := range wrappers {
		var b strings.Builder
		b.WriteString("@echo off\r\n")
		for k, v := range w.envVars {
			expanded := strings.ReplaceAll(v, `%PREFIX%`, `%~dp0..`)
			fmt.Fprintf(&b, "set \"%s=%s\"\r\n", k, expanded)
		}
		realPath := strings.ReplaceAll(w.realPath, `%PREFIX%`, `%~dp0..`)
		fmt.Fprintf(&b, "\"%%~dp0..\\%s\" %%*\r\n", realPath)

		path := filepath.Join(binDir, w.name+".cmd")
		if err := os.WriteFile(path, []byte(b.String()), 0755); err != nil {
			return fmt.Errorf("wrapper %s: %w", w.name, err)
		}
	}
	return nil
}

// VersionSummary returns a multi-line string of tool versions for display.
func VersionSummary(vers *versions.Versions) string {
	keys := []struct{ label, key string }{
		{"fzf", "FZF_VERSION"}, {"fd", "FD_VERSION"}, {"bat", "BAT_VERSION"},
		{"lsd", "LSD_VERSION"}, {"rg", "RG_VERSION"}, {"delta", "DELTA_VERSION"},
		{"lazygit", "LAZYGIT_VERSION"}, {"jq", "JQ_VERSION"}, {"direnv", "DIRENV_VERSION"},
		{"nvim", "NVIM_VERSION"}, {"go", "GO_VERSION"}, {"git", "GIT_VERSION"},
		{"htop", "HTOP_VERSION"}, {"btop", "BTOP_VERSION"}, {"dust", "DUST_VERSION"},
		{"age", "AGE_VERSION"},
	}
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "  %-10s %s\n", k.label, vers.Get(k.key))
	}
	return b.String()
}
