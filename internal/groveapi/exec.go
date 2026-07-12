package groveapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecClient resolves workspaces by invoking `gw ws show <name> --json`.
type ExecClient struct {
	Env Env
}

// NewExecClient creates a Grove CLI client.
func NewExecClient(env Env) *ExecClient {
	return &ExecClient{Env: env}
}

func (c *ExecClient) ShowWorkspace(ctx context.Context, name string) (*Workspace, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("workspace name is required")
	}
	out, err := c.run(ctx, "ws", "show", name, "--json")
	if err != nil {
		return nil, fmt.Errorf("gw ws show %s --json: %w", name, err)
	}
	var ws Workspace
	if err := json.Unmarshal(out, &ws); err != nil {
		return nil, fmt.Errorf("decode workspace JSON: %w", err)
	}
	return &ws, nil
}

func (c *ExecClient) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	out, err := c.run(ctx, "ws", "list", "--json")
	if err != nil {
		return nil, fmt.Errorf("gw ws list --json: %w", err)
	}
	var workspaces []Workspace
	if err := json.Unmarshal(out, &workspaces); err != nil {
		return nil, fmt.Errorf("decode workspace list JSON: %w", err)
	}
	return workspaces, nil
}

func (c *ExecClient) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gw", args...)
	cmd.Env = c.withGroveEnv()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	if msg := strings.TrimSpace(stderr.String()); msg != "" {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	return nil, err
}

func (c *ExecClient) withGroveEnv() []string {
	env := append(os.Environ(),
		"GROVE_DIR="+c.Env.Dir,
		"GROVE_CONFIG="+c.Env.Config,
		"GROVE_STATE="+c.Env.State,
	)
	if c.Env.Workspace != "" {
		env = append(env, "GROVE_WORKSPACE="+c.Env.Workspace)
	}
	return env
}
