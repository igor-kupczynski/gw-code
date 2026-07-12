package groveapi_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/igor-kupczynski/gw-code/internal/groveapi"
)

func TestExecClientPreservesStderr(t *testing.T) {
	root := t.TempDir()
	fakeGW := filepath.Join(root, "gw")
	script := `#!/bin/sh
echo "Workspace not found: missing" >&2
exit 1
`
	if err := os.WriteFile(fakeGW, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	client := groveapi.NewExecClient(groveapi.Env{Dir: filepath.Join(root, ".grove")})
	_, err := client.ShowWorkspace(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Workspace not found: missing") {
		t.Fatalf("err = %v", err)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected wrapped exit status 1, err = %v", err)
	}
}
