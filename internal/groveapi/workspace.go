package groveapi

// Workspace is the narrow workspace view gw-code needs from Grove.
type Workspace struct {
	Name  string         `json:"name"`
	Path  string         `json:"path"`
	Repos []RepoWorktree `json:"repos"`
}

// RepoWorktree is one repository root inside a Grove workspace.
type RepoWorktree struct {
	RepoName     string `json:"repo_name"`
	WorktreePath string `json:"worktree_path"`
}
