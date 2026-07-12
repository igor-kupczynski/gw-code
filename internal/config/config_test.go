package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/igor-kupczynski/gw-code/internal/config"
)

func TestLoadGlobalEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gw-code.toml")
	if err := os.WriteFile(path, []byte("editor = \"cursor\"\nargs = [\"--classic\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Editor != "cursor" {
		t.Fatalf("editor = %q", cfg.Editor)
	}
	if len(cfg.Args) != 1 || cfg.Args[0] != "--classic" {
		t.Fatalf("args = %#v", cfg.Args)
	}
}

func TestLoadRejectsUnsupportedOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gw-code.toml")
	if err := os.WriteFile(path, []byte("editor = \"code\"\nwindow = \"reuse\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected unsupported option error")
	}
}

func TestGlobalPath(t *testing.T) {
	if got := config.GlobalPath("/tmp/.grove"); got != "/tmp/.grove/gw-code.toml" {
		t.Fatalf("GlobalPath = %q", got)
	}
}
