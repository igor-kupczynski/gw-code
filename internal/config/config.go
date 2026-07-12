package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type File struct {
	Editor string   `toml:"editor"`
	Args   []string `toml:"args"`
}

func GlobalPath(groveDir string) string {
	return filepath.Join(groveDir, "gw-code.toml")
}

func Load(path string) (File, error) {
	cfg, ok, err := loadFile(path)
	if err != nil {
		return File{}, err
	}
	if !ok {
		return File{}, nil
	}
	return cfg, nil
}

func loadFile(path string) (File, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, false, nil
		}
		return File{}, false, err
	}
	var cfg File
	metadata, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return File{}, false, fmt.Errorf("parse %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return File{}, false, fmt.Errorf("unsupported config option %q in %s", undecoded[0], path)
	}
	return cfg, true, nil
}
