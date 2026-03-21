package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/stefanpenner/dotpack/cmd"
	"github.com/stefanpenner/dotpack/internal/versions"
)

//go:embed versions.env
var versionsEnv string

//go:embed Dockerfile
var dockerfile string

//go:embed scripts/download-binaries.sh
var downloadScript string

// Version is set at build time via -ldflags.
var Version = "dev"

func usage() {
	fmt.Print(`Usage: dotpack <command> [options]

Commands:
  build [--os OS] [--arch ARCH]   Build bundle (default: linux/current arch)
  push <host>                     Deploy bundle to remote host via SSH
  status <host>                   Show installed tool versions on host
  install                         Install bundle locally (DOTPACK_PREFIX, default ~/.local)
  clean                           Remove build artifacts and Docker image
  upgrade                         Download and install the latest release
  version                         Print dotpack version
  versions                        Print bundled tool versions

Options:
  OS: linux, darwin, or windows (default: linux for Docker builds)
  ARCH defaults to current machine architecture
  DOTPACK_PREFIX env var controls install location
    (default: ~/.local on unix, %LOCALAPPDATA%\dotpack on Windows)
  DOTPACK_NO_AUTOUPDATE=1 disables automatic update checks

Examples:
  dotpack build                   Build for linux/current-arch
  dotpack build --os darwin       Build for darwin/current-arch
  dotpack build --os windows      Build for windows/current-arch
  dotpack push nas                Deploy to NAS
  dotpack status nas              Check versions on NAS
  dotpack install                 Install locally
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	vers := versions.Parse(versionsEnv)

	scriptDir, isTmp, err := cmd.FindScriptDir(dockerfile, versionsEnv, downloadScript)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if isTmp {
		defer os.RemoveAll(scriptDir)
	}

	// Auto-update before running the command (skip for upgrade/version).
	switch os.Args[1] {
	case "upgrade", "version", "versions":
		// no auto-update
	default:
		cmd.AutoUpdate(Version)
	}

	switch os.Args[1] {
	case "build":
		err = cmd.Build(os.Args[2:], vers, scriptDir)
	case "push":
		host := ""
		if len(os.Args) > 2 {
			host = os.Args[2]
		}
		err = cmd.Push(host, scriptDir)
	case "status":
		host := ""
		if len(os.Args) > 2 {
			host = os.Args[2]
		}
		err = cmd.Status(host)
	case "install":
		err = cmd.Install(scriptDir)
	case "clean":
		err = cmd.Clean(scriptDir)
	case "upgrade":
		err = cmd.Upgrade(Version)
	case "version":
		fmt.Println(Version)
	case "versions":
		fmt.Print(cmd.VersionSummary(vers))
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
