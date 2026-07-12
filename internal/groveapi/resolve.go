package groveapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver determines which Grove workspace a command targets.
type Resolver struct {
	Client Client
	Env    Env
	Cwd    func() (string, error)
}

// NewResolver creates a workspace resolver with production defaults.
func NewResolver(client Client, env Env) *Resolver {
	return &Resolver{
		Client: client,
		Env:    env,
		Cwd:    os.Getwd,
	}
}

// Resolve applies precedence: explicit name, GROVE_WORKSPACE, then cwd detection.
func (r *Resolver) Resolve(ctx context.Context, explicit string) (*Workspace, error) {
	name := strings.TrimSpace(explicit)
	if name == "" {
		name = strings.TrimSpace(r.Env.Workspace)
	}
	if name != "" {
		ws, err := r.Client.ShowWorkspace(ctx, name)
		if err != nil {
			return nil, err
		}
		if err := ValidateWorkspace(ws); err != nil {
			return nil, err
		}
		return ws, nil
	}

	cwd, err := r.Cwd()
	if err != nil {
		return nil, fmt.Errorf("cannot determine working directory: %w", err)
	}
	workspaces, err := r.Client.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	ws, err := FindWorkspaceByPath(workspaces, cwd)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("not inside a workspace; provide a workspace name or cd into one")
	}
	if err := ValidateWorkspace(ws); err != nil {
		return nil, err
	}
	return ws, nil
}

// FindWorkspaceByPath returns the workspace containing path using prefix matching with symlink resolution.
func FindWorkspaceByPath(workspaces []Workspace, path string) (*Workspace, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	resolved = filepath.Clean(resolved)
	var matches []*Workspace
	for i := range workspaces {
		wsPath, err := filepath.EvalSymlinks(workspaces[i].Path)
		if err != nil {
			wsPath = workspaces[i].Path
		}
		wsPath = filepath.Clean(wsPath)
		if resolved == wsPath {
			matches = append(matches, &workspaces[i])
			continue
		}
		if strings.HasPrefix(resolved, wsPath+string(filepath.Separator)) {
			matches = append(matches, &workspaces[i])
		}
	}
	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		names := make([]string, len(matches))
		for i, ws := range matches {
			names[i] = ws.Name
		}
		return nil, fmt.Errorf("ambiguous workspace match for %s: %s", path, strings.Join(names, ", "))
	}
}
