package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecutableName returns the configured executable or `code`.
func ExecutableName(configured string) string {
	if strings.TrimSpace(configured) == "" {
		return "code"
	}
	return configured
}

// ResolveExecutable finds the editor binary on PATH or validates an absolute path.
func ResolveExecutable(name string, lookPath func(string) (string, error)) (string, error) {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if filepath.IsAbs(name) || strings.Contains(name, string(filepath.Separator)) {
		if _, err := os.Stat(name); err != nil {
			return "", fmt.Errorf("editor executable %q not found", name)
		}
		return name, nil
	}
	path, err := lookPath(name)
	if err != nil {
		return "", fmt.Errorf("editor %q not found on PATH", name)
	}
	return path, nil
}

// BuildArgs constructs configured args + --new-window + workspace.
func BuildArgs(args []string, workspace string) []string {
	argv := append([]string(nil), args...)
	return append(argv, "--new-window", workspace)
}

// Launch starts the editor without a shell and releases the child process.
func Launch(executable string, args []string, workspace string, lookPath func(string) (string, error)) error {
	name := ExecutableName(executable)
	binary, err := ResolveExecutable(name, lookPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(binary, BuildArgs(args, workspace)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
