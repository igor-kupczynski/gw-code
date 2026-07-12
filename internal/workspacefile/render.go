package workspacefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/igor-kupczynski/gw-code/internal/groveapi"
)

// Document is the generated VS Code / Cursor multi-root workspace file.
type Document struct {
	Folders []Folder `json:"folders"`
}

type Folder struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

func PathIn(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "gw-code.code-workspace")
}

// Render builds a deterministic workspace document.
func Render(workspace groveapi.Workspace) (Document, error) {
	folders := make([]Folder, 0, len(workspace.Repos))
	for _, repo := range workspace.Repos {
		folderPath, err := folderPath(workspace.Path, repo.WorktreePath)
		if err != nil {
			return Document{}, fmt.Errorf("repo %s: %w", repo.RepoName, err)
		}
		folders = append(folders, Folder{
			Path: folderPath,
			Name: repo.RepoName,
		})
	}

	doc := Document{Folders: folders}
	return doc, nil
}

func folderPath(workspaceFileDir, worktreePath string) (string, error) {
	worktreePath = filepath.Clean(worktreePath)
	base := filepath.Clean(workspaceFileDir)
	if rel, err := filepath.Rel(base, worktreePath); err == nil {
		return normalizeJSONPath(rel), nil
	}
	abs, err := filepath.Abs(worktreePath)
	if err != nil {
		return "", err
	}
	return normalizeJSONPath(abs), nil
}

func normalizeJSONPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// Marshal formats a workspace document as two-space-indented JSON with a trailing newline.
func Marshal(doc Document) ([]byte, error) {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// WriteAtomic writes content atomically, skipping unchanged files to preserve mtime.
func WriteAtomic(path string, content []byte) (bool, error) {
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) == string(content) {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return false, err
	}
	return true, nil
}
