package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/igor-kupczynski/gw-code/internal/config"
	"github.com/igor-kupczynski/gw-code/internal/editor"
	"github.com/igor-kupczynski/gw-code/internal/groveapi"
	"github.com/igor-kupczynski/gw-code/internal/workspacefile"
)

// App wires Grove resolution, rendering, and editor launch.
type App struct {
	Resolver     *groveapi.Resolver
	GlobalConfig string
	Stderr       io.Writer
	LookPath     func(string) (string, error)
}

// RefreshResult describes a generated workspace file.
type RefreshResult struct {
	WorkspaceFilePath string
	Changed           bool
}

// Open resolves, refreshes, launches the editor, and returns the workspace file path.
func (a *App) Open(ctx context.Context, workspaceName string) (string, error) {
	result, err := a.Refresh(ctx, workspaceName)
	if err != nil {
		return "", err
	}
	cfg, err := config.Load(a.GlobalConfig)
	if err != nil {
		return result.WorkspaceFilePath, err
	}
	if err := editor.Launch(cfg.Editor, cfg.Args, result.WorkspaceFilePath, a.LookPath); err != nil {
		fmt.Fprintf(a.Stderr, "workspace file: %s\n", result.WorkspaceFilePath)
		return result.WorkspaceFilePath, err
	}
	fmt.Fprintln(a.Stderr, "opened with", editor.ExecutableName(cfg.Editor))
	return result.WorkspaceFilePath, nil
}

// Refresh generates workspace artifacts without launching an editor.
func (a *App) Refresh(ctx context.Context, workspaceName string) (RefreshResult, error) {
	ws, err := a.Resolver.Resolve(ctx, workspaceName)
	if err != nil {
		return RefreshResult{}, err
	}
	workspacePath := workspacefile.PathIn(ws.Path)
	doc, err := workspacefile.Render(*ws)
	if err != nil {
		return RefreshResult{}, err
	}
	workspaceData, err := workspacefile.Marshal(doc)
	if err != nil {
		return RefreshResult{}, err
	}
	changed, err := workspacefile.WriteAtomic(workspacePath, workspaceData)
	if err != nil {
		return RefreshResult{}, err
	}
	abs, err := filepath.Abs(workspacePath)
	if err != nil {
		return RefreshResult{}, err
	}
	return RefreshResult{
		WorkspaceFilePath: abs,
		Changed:           changed,
	}, nil
}

// Path prints the generated workspace file path without refreshing.
func (a *App) Path(ctx context.Context, workspaceName string) (string, error) {
	ws, err := a.Resolver.Resolve(ctx, workspaceName)
	if err != nil {
		return "", err
	}
	return filepath.Abs(workspacefile.PathIn(ws.Path))
}
