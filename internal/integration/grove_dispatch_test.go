package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const groveVersion = "v1.1.9"

func TestGroveDispatch(t *testing.T) {
	if os.Getenv("GW_CODE_INTEGRATION") != "1" {
		t.Skip("set GW_CODE_INTEGRATION=1 to run Grove dispatch integration tests")
	}

	root := t.TempDir()
	groveDir := filepath.Join(root, ".grove")
	workspaceRoot := filepath.Join(root, "otel-dev")
	repos := []string{"collector-core", "collector-contrib"}
	for _, repo := range repos {
		if err := os.MkdirAll(filepath.Join(workspaceRoot, repo), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(groveDir, "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}

	state := []map[string]any{{
		"name":       "otel-dev",
		"path":       workspaceRoot,
		"branch":     "feature/otel",
		"created_at": "2026-06-01T10:00:00.000000",
		"repos": []map[string]string{
			{
				"repo_name":     "collector-core",
				"source_repo":   "/tmp/opentelemetry-collector",
				"worktree_path": filepath.Join(workspaceRoot, "collector-core"),
				"branch":        "feature/otel",
			},
			{
				"repo_name":     "collector-contrib",
				"source_repo":   "/tmp/opentelemetry-collector-contrib",
				"worktree_path": filepath.Join(workspaceRoot, "collector-contrib"),
				"branch":        "feature/otel",
			},
		},
	}}
	stateData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(groveDir, "state.json"), stateData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(groveDir, "config.toml"), []byte("workspace_dir = \""+filepath.Dir(workspaceRoot)+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(groveDir, "gw-code.toml"), []byte("editor = \"code\"\nargs = [\"--classic\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeCode := filepath.Join(binDir, "code")
	logPath := filepath.Join(root, "editor.log")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" >> %q\n", logPath)
	if err := os.WriteFile(fakeCode, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	gwCode := buildGWCode(t, root)
	pluginPath := filepath.Join(groveDir, "plugins", "gw-code")
	if err := copyFile(gwCode, pluginPath, 0o755); err != nil {
		t.Fatal(err)
	}

	gw := ensureGroveBinary(t, root)
	groveBin := filepath.Dir(gw)

	env := append(os.Environ(),
		"HOME="+root,
		"PATH="+groveBin+string(os.PathListSeparator)+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	t.Run("default open", func(t *testing.T) {
		_ = os.Remove(logPath)
		cmd := exec.Command(gw, "code", "otel-dev")
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gw code otel-dev: %v\n%s", err, out)
		}
		if !strings.Contains(string(out), "gw-code.code-workspace") {
			t.Fatalf("unexpected output: %s", out)
		}
		logText := waitForEditorLog(t, logPath, 2*time.Second)
		if !strings.Contains(logText, "--new-window") {
			t.Fatalf("unexpected editor argv log: %q", logText)
		}
		if !strings.Contains(logText, "--classic") {
			t.Fatalf("configured editor args missing from argv log: %q", logText)
		}
		if !strings.Contains(logText, "gw-code.code-workspace") {
			t.Fatalf("workspace file not passed to editor: %q", logText)
		}
	})

	t.Run("path flag", func(t *testing.T) {
		cmd := exec.Command(gw, "code", "--path", "otel-dev")
		cmd.Env = env
		cmd.Dir = filepath.Join(workspaceRoot, "collector-core")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gw code --path: %v\n%s", err, out)
		}
		path := strings.TrimSpace(string(out))
		if !strings.HasSuffix(path, "gw-code.code-workspace") {
			t.Fatalf("unexpected path: %q", path)
		}
	})

	t.Run("missing workspace exits error", func(t *testing.T) {
		cmd := exec.Command(gw, "code", "--path", "missing-workspace")
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error, got output: %s", out)
		}
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			t.Fatalf("err = %v output = %s", err, out)
		}
		combined := string(out)
		if strings.TrimSpace(combined) == "" {
			t.Fatalf("expected useful stderr, got empty output")
		}
	})

	t.Run("refresh flag", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(groveDir, "gw-code.toml"), []byte("unsupported = true\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(gw, "code", "--refresh", "otel-dev")
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gw code --refresh: %v\n%s", err, out)
		}
		workspaceFile := strings.TrimSpace(strings.Split(string(out), "\n")[0])
		data, err := os.ReadFile(workspaceFile)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"path": "collector-core"`) {
			t.Fatalf("workspace missing relative folder path: %s", data)
		}
	})
}

func waitForEditorLog(t *testing.T, path string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return string(data)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for editor log at %s", path)
	return ""
}

func buildGWCode(t *testing.T, root string) string {
	t.Helper()
	out := filepath.Join(root, "gw-code")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/gw-code")
	cmd.Dir = moduleRoot(t)
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build gw-code: %v\n%s", err, outBytes)
	}
	return out
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve module root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func ensureGroveBinary(t *testing.T, root string) string {
	t.Helper()
	cache := filepath.Join(moduleRoot(t), "testdata", "bin", groveVersion)
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(cache, "gw")
	if _, err := os.Stat(binary); err == nil {
		return binary
	}
	asset := groveAssetName()
	url := fmt.Sprintf("https://github.com/nicksenap/grove/releases/download/%s/%s", groveVersion, asset)
	archive := filepath.Join(root, asset)
	cmd := exec.Command("curl", "-fsSL", "-o", archive, url)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("download grove %s: %v\n%s", groveVersion, err, out)
	}
	cmd = exec.Command("tar", "-xzf", archive, "-C", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract grove: %v\n%s", err, out)
	}
	extracted := filepath.Join(root, "gw")
	if err := copyFile(extracted, binary, 0o755); err != nil {
		t.Fatal(err)
	}
	return binary
}

func groveAssetName() string {
	return fmt.Sprintf("gw_%s_%s_%s.tar.gz", strings.TrimPrefix(groveVersion, "v"), runtime.GOOS, runtime.GOARCH)
}

func copyFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}
