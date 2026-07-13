package workspacefile_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/igor-kupczynski/gw-code/internal/groveapi"
	"github.com/igor-kupczynski/gw-code/internal/workspacefile"
)

func TestGoldenWorkspace(t *testing.T) {
	t.Parallel()
	fixtureRoot := filepath.Join("..", "..", "testdata", "grove-v1.1.9")
	wsData, err := os.ReadFile(filepath.Join(fixtureRoot, "workspace-show.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ws groveapi.Workspace
	if err := json.Unmarshal(wsData, &ws); err != nil {
		t.Fatal(err)
	}
	doc, err := workspacefile.Render(ws)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workspacefile.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(fixtureRoot, "golden", "otel-dev.code-workspace"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("golden mismatch:\n%s", string(got))
	}
}

func TestFolderPathsResolveFromWorkspaceFileDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsDir := filepath.Join(root, "otel-dev")
	repoPaths := []string{
		filepath.Join(wsDir, "collector-core"),
		filepath.Join(wsDir, "collector-contrib"),
	}
	for _, repo := range repoPaths {
		if err := os.MkdirAll(repo, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	ws := groveapi.Workspace{
		Name: "otel-dev",
		Path: wsDir,
		Repos: []groveapi.RepoWorktree{
			{RepoName: "collector-core", WorktreePath: repoPaths[0]},
			{RepoName: "collector-contrib", WorktreePath: repoPaths[1]},
		},
	}
	doc, err := workspacefile.Render(ws)
	if err != nil {
		t.Fatal(err)
	}
	workspaceFile := workspacefile.PathIn(wsDir, ws.Name)
	for i, folder := range doc.Folders {
		resolved := filepath.Clean(filepath.Join(filepath.Dir(workspaceFile), filepath.FromSlash(folder.Path)))
		want, err := filepath.EvalSymlinks(repoPaths[i])
		if err != nil {
			t.Fatal(err)
		}
		got, err := filepath.EvalSymlinks(resolved)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("folder %q resolved to %q, want %q", folder.Path, got, want)
		}
	}
}

func TestWriteAtomicPreservesMtime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gw-code.code-workspace")
	content := []byte("{\n  \"folders\": []\n}\n")
	changed, err := workspacefile.WriteAtomic(path, content)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected first write to report changed")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	firstMtime := info.ModTime()
	time.Sleep(10 * time.Millisecond)
	changed, err = workspacefile.WriteAtomic(path, content)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected unchanged write to skip")
	}
	info, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(firstMtime) {
		t.Fatalf("mtime changed on identical content: %v -> %v", firstMtime, info.ModTime())
	}
}

func TestExternalFolderResolvesFromWorkspaceFileDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wsPath := filepath.Join(root, "ws")
	external := filepath.Join(root, "external-repo")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	ws := groveapi.Workspace{
		Name: "demo",
		Path: wsPath,
		Repos: []groveapi.RepoWorktree{{
			RepoName:     "external",
			WorktreePath: external,
		}},
	}
	doc, err := workspacefile.Render(ws)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Folders) != 1 {
		t.Fatalf("folders = %#v", doc.Folders)
	}
	workspaceFile := workspacefile.PathIn(wsPath, ws.Name)
	resolved := filepath.Clean(filepath.Join(filepath.Dir(workspaceFile), filepath.FromSlash(doc.Folders[0].Path)))
	want, err := filepath.EvalSymlinks(external)
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("folder %q resolved to %q, want %q", doc.Folders[0].Path, got, want)
	}
}

func TestPathIn(t *testing.T) {
	t.Parallel()
	if got := workspacefile.PathIn("/tmp/ws", "otel-dev"); got != filepath.Join("/tmp/ws", "otel-dev.code-workspace") {
		t.Fatalf("PathIn = %q", got)
	}
}

func TestRemoveLegacy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	legacy := workspacefile.LegacyPathIn(dir)
	if err := os.WriteFile(legacy, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := workspacefile.RemoveLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy file still exists: %v", err)
	}
	if err := workspacefile.RemoveLegacy(dir); err != nil {
		t.Fatal("expected missing legacy file to be ignored")
	}
}
