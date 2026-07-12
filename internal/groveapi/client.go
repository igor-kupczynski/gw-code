package groveapi

import "context"

// Client resolves Grove workspaces without importing Grove internals.
// ShowWorkspace returns either a non-nil workspace or an error.
type Client interface {
	ShowWorkspace(ctx context.Context, name string) (*Workspace, error)
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
}
