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
  version                         Print dotpack version
  versions                        Print bundled tool versions

Options:
  OS defaults to linux for Docker builds, current OS for local
  ARCH defaults to current machine architecture
  DOTPACK_PREFIX env var controls install location (default ~/.local)

Examples:
  dotpack build                   Build for linux/current-arch
  dotpack build --os darwin       Build for darwin/current-arch
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
