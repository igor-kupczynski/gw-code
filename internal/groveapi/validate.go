package groveapi

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateWorkspace checks workspace fields needed for rendering.
func ValidateWorkspace(ws *Workspace) error {
	if ws == nil {
		return fmt.Errorf("workspace is nil")
	}
	if strings.TrimSpace(ws.Name) == "" {
		return fmt.Errorf("workspace name is required")
	}
	if strings.ContainsAny(ws.Name, `/\`) {
		return fmt.Errorf("workspace %s: name must not contain path separators", ws.Name)
	}
	if strings.TrimSpace(ws.Path) == "" {
		return fmt.Errorf("workspace %s: path is required", ws.Name)
	}
	if len(ws.Repos) == 0 {
		return fmt.Errorf("workspace %s: at least one repository is required", ws.Name)
	}
	seenNames := make(map[string]struct{}, len(ws.Repos))
	seenPaths := make(map[string]struct{}, len(ws.Repos))
	for _, repo := range ws.Repos {
		name := strings.TrimSpace(repo.RepoName)
		path := strings.TrimSpace(repo.WorktreePath)
		if name == "" {
			return fmt.Errorf("workspace %s: repository logical name is required", ws.Name)
		}
		if path == "" {
			return fmt.Errorf("workspace %s: repository %s: worktree_path is required", ws.Name, name)
		}
		clean := filepath.Clean(path)
		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("workspace %s: duplicate repository name %q", ws.Name, name)
		}
		seenNames[name] = struct{}{}
		if _, ok := seenPaths[clean]; ok {
			return fmt.Errorf("workspace %s: duplicate worktree_path %q", ws.Name, clean)
		}
		seenPaths[clean] = struct{}{}
	}
	return nil
}
