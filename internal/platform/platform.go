package platform

import "fmt"

// Platform holds normalized OS/arch information and target triples.
type Platform struct {
	OS           string // "linux" or "darwin"
	Arch         string // original arch string
	RustArch     string // "x86_64" or "aarch64"
	GoArch       string // "amd64" or "arm64"
	ArchGeneric  string // "x86_64" or "arm64"
	RustTarget   string // e.g. "x86_64-unknown-linux-musl"
	RustTargetGNU string // for ripgrep/delta aarch64 fallback
	LazygitOS    string // "Linux" or "Darwin"
	NvimOS       string // "linux" or "macos"
	JqOS         string // "linux" or "macos"
	DockerPlatform string // e.g. "linux/amd64"
}

// New creates a Platform with all fields normalized.
func New(os, arch string) (*Platform, error) {
	p := &Platform{OS: os, Arch: arch}

	switch arch {
	case "x86_64", "amd64":
		p.RustArch = "x86_64"
		p.GoArch = "amd64"
		p.ArchGeneric = "x86_64"
	case "aarch64", "arm64":
		p.RustArch = "aarch64"
		p.GoArch = "arm64"
		p.ArchGeneric = "arm64"
	default:
		return nil, fmt.Errorf("unsupported arch: %s", arch)
	}

	switch os {
	case "linux":
		p.RustTarget = p.RustArch + "-unknown-linux-musl"
		p.RustTargetGNU = p.RustArch + "-unknown-linux-gnu"
		p.LazygitOS = "Linux"
		p.NvimOS = "linux"
		p.JqOS = "linux"
	case "darwin":
		p.RustTarget = p.RustArch + "-apple-darwin"
		p.RustTargetGNU = p.RustTarget // not needed on Mac
		p.LazygitOS = "Darwin"
		p.NvimOS = "macos"
		p.JqOS = "macos"
	default:
		return nil, fmt.Errorf("unsupported OS: %s", os)
	}

	p.DockerPlatform = "linux/" + p.GoArch

	return p, nil
}

// RustTargetFor returns the appropriate rust target triple for a project.
// ripgrep and delta don't ship aarch64 musl builds.
func (p *Platform) RustTargetFor(project string) string {
	switch project {
	case "ripgrep", "delta":
		if p.OS == "linux" && p.RustArch == "aarch64" {
			return p.RustTargetGNU
		}
		return p.RustTarget
	default:
		return p.RustTarget
	}
}
