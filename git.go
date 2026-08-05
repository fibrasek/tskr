package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// runGitCmd executes git command with args in target directory.
func runGitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}

	return strings.TrimSpace(out.String()), nil
}

// GetGitRepoRoot returns absolute root directory path of the git repository.
func GetGitRepoRoot(dir string) (string, error) {
	return runGitCmd(dir, "rev-parse", "--show-toplevel")
}

// GetGitBranch returns current checked out git branch name.
func GetGitBranch(dir string) (string, error) {
	return runGitCmd(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// ExtractRepoName extracts base repository directory name from path.
func ExtractRepoName(repoPath string) string {
	return filepath.Base(repoPath)
}

// SanitizeBranchName replaces slashes in branch names with hyphens for safe filesystem directory naming.
func SanitizeBranchName(branch string) string {
	branch = strings.ReplaceAll(branch, "/", "-")
	branch = strings.ReplaceAll(branch, "\\", "-")
	return branch
}
