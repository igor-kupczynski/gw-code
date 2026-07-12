package groveapi

import (
	"fmt"
	"os"
	"path/filepath"
)

// Env carries Grove plugin environment variables.
type Env struct {
	Dir       string
	Config    string
	State     string
	Workspace string
}

// EnvFromOS reads Grove plugin environment variables.
func EnvFromOS() (Env, error) {
	dir := os.Getenv("GROVE_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Env{}, fmt.Errorf("GROVE_DIR is not set and home directory is unavailable: %w", err)
		}
		dir = filepath.Join(home, ".grove")
	}
	cfg := os.Getenv("GROVE_CONFIG")
	if cfg == "" {
		cfg = filepath.Join(dir, "config.toml")
	}
	state := os.Getenv("GROVE_STATE")
	if state == "" {
		state = filepath.Join(dir, "state.json")
	}
	return Env{
		Dir:       dir,
		Config:    cfg,
		State:     state,
		Workspace: os.Getenv("GROVE_WORKSPACE"),
	}, nil
}
