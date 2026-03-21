package ssh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a command on a remote host and returns stdout.
func Run(host, command string) (string, error) {
	cmd := exec.Command("ssh", host, command)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ssh %s: %w", host, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunInteractive executes a command on a remote host with stdout/stderr connected.
func RunInteractive(host, command string) error {
	cmd := exec.Command("ssh", host, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PipeFile pipes a local file's contents to a remote command via SSH stdin.
func PipeFile(host, localPath, remoteCmd string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("ssh", host, remoteCmd)
	cmd.Stdin = f
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PipeBytes pipes bytes to a remote command via SSH stdin.
func PipeBytes(host string, data []byte, remoteCmd string) error {
	cmd := exec.Command("ssh", host, remoteCmd)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
