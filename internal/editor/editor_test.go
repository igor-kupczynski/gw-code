package editor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/igor-kupczynski/gw-code/internal/editor"
)

func TestBuildArgsOrder(t *testing.T) {
	t.Parallel()
	workspace := "/tmp/work space';rm -rf /.code-workspace"
	argv := editor.BuildArgs([]string{"--classic"}, workspace)
	if len(argv) != 3 {
		t.Fatalf("argv = %#v", argv)
	}
	if argv[0] != "--classic" || argv[1] != "--new-window" || argv[2] != workspace {
		t.Fatalf("argv = %#v", argv)
	}
}

func TestExecutableNameDefaultsToCode(t *testing.T) {
	t.Parallel()
	if got := editor.ExecutableName(""); got != "code" {
		t.Fatalf("ExecutableName(\"\") = %q", got)
	}
	if got := editor.ExecutableName("cursor"); got != "cursor" {
		t.Fatalf("ExecutableName(\"cursor\") = %q", got)
	}
}

func TestResolveExecutableUsesPATH(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	binary := filepath.Join(dir, "code")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	lookPath := func(name string) (string, error) {
		if name == "code" {
			return binary, nil
		}
		return "", os.ErrNotExist
	}
	got, err := editor.ResolveExecutable("code", lookPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != binary {
		t.Fatalf("ResolveExecutable = %q", got)
	}
}
