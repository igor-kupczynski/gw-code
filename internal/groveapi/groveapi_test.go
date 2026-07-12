package groveapi_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/igor-kupczynski/gw-code/internal/groveapi"
)

type fakeClient struct {
	workspaces []groveapi.Workspace
}

func (c *fakeClient) ShowWorkspace(ctx context.Context, name string) (*groveapi.Workspace, error) {
	_ = ctx
	for i := range c.workspaces {
		if c.workspaces[i].Name == name {
			ws := c.workspaces[i]
			return &ws, nil
		}
	}
	return nil, fmt.Errorf("workspace %s not found", name)
}

func (c *fakeClient) ListWorkspaces(ctx context.Context) ([]groveapi.Workspace, error) {
	_ = ctx
	return c.workspaces, nil
}

func TestFindWorkspaceByPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsDir := filepath.Join(root, "otel-dev")
	repoDir := filepath.Join(wsDir, "collector-core")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	workspaces := []groveapi.Workspace{{
		Name: "otel-dev",
		Path: wsDir,
		Repos: []groveapi.RepoWorktree{{
			RepoName: "collector-core", WorktreePath: repoDir,
		}},
	}}
	found, err := groveapi.FindWorkspaceByPath(workspaces, repoDir)
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name != "otel-dev" {
		t.Fatalf("expected otel-dev, got %#v", found)
	}
}

func TestResolverPrecedence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsDir := filepath.Join(root, "demo")
	repoDir := filepath.Join(wsDir, "svc-a")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	client := &fakeClient{workspaces: []groveapi.Workspace{
		{
			Name: "alpha",
			Path: filepath.Join(root, "alpha"),
			Repos: []groveapi.RepoWorktree{{
				RepoName: "svc-a", WorktreePath: filepath.Join(root, "alpha", "svc-a"),
			}},
		},
		{
			Name: "beta",
			Path: wsDir,
			Repos: []groveapi.RepoWorktree{{
				RepoName: "svc-a", WorktreePath: repoDir,
			}},
		},
	}}
	env := groveapi.Env{Workspace: "alpha"}
	resolver := groveapi.NewResolver(client, env)
	resolver.Cwd = func() (string, error) { return repoDir, nil }

	ws, err := resolver.Resolve(t.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "alpha" {
		t.Fatalf("expected GROVE_WORKSPACE precedence, got %q", ws.Name)
	}

	ws, err = resolver.Resolve(t.Context(), "beta")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "beta" {
		t.Fatalf("expected explicit name, got %q", ws.Name)
	}

	env.Workspace = ""
	resolver = groveapi.NewResolver(client, env)
	resolver.Cwd = func() (string, error) { return repoDir, nil }
	ws, err = resolver.Resolve(t.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "beta" {
		t.Fatalf("expected cwd detection, got %q", ws.Name)
	}
}

func TestResolverIgnoresUnrelatedInvalidWorkspace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsDir := filepath.Join(root, "valid")
	repoDir := filepath.Join(wsDir, "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	client := &fakeClient{workspaces: []groveapi.Workspace{
		{Name: "stale", Path: filepath.Join(root, "stale")},
		{
			Name: "valid",
			Path: wsDir,
			Repos: []groveapi.RepoWorktree{{
				RepoName: "repo", WorktreePath: repoDir,
			}},
		},
	}}
	resolver := groveapi.NewResolver(client, groveapi.Env{})
	resolver.Cwd = func() (string, error) { return repoDir, nil }

	ws, err := resolver.Resolve(t.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "valid" {
		t.Fatalf("workspace = %q", ws.Name)
	}
}

func TestFindWorkspaceByPathPrefixFalsePositive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsDir := filepath.Join(root, "otel-dev")
	other := filepath.Join(root, "otel-dev-extra")
	for _, dir := range []string{wsDir, other} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	workspaces := []groveapi.Workspace{{
		Name: "otel-dev",
		Path: wsDir,
		Repos: []groveapi.RepoWorktree{{
			RepoName: "core", WorktreePath: filepath.Join(wsDir, "core"),
		}},
	}}
	found, err := groveapi.FindWorkspaceByPath(workspaces, other)
	if err != nil {
		t.Fatal(err)
	}
	if found != nil {
		t.Fatalf("expected nil for prefix false positive, got %#v", found)
	}
}

func TestValidateWorkspaceDuplicateRepoName(t *testing.T) {
	t.Parallel()
	err := groveapi.ValidateWorkspace(&groveapi.Workspace{
		Name: "demo",
		Path: "/tmp/demo",
		Repos: []groveapi.RepoWorktree{
			{RepoName: "a", WorktreePath: "/tmp/demo/a"},
			{RepoName: "a", WorktreePath: "/tmp/demo/b"},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate repo error")
	}
}
