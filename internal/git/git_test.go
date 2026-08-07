package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_ExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard path",
			input:    "/home/user/projects/my-repo",
			expected: "my-repo",
		},
		{
			name:     "trailing slash handled by filepath.Base",
			input:    "/var/code/taskrefs",
			expected: "taskrefs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractRepoName(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_SanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain branch",
			input:    "main",
			expected: "main",
		},
		{
			name:     "feature branch with slash",
			input:    "feature/auth-login",
			expected: "feature-auth-login",
		},
		{
			name:     "nested branch slashes",
			input:    "user/fix/bug-123",
			expected: "user-fix-bug-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeBranchName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeBranchName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_GitOperations_InTempRepo(t *testing.T) {
	tempDir := t.TempDir()

	cmdInit := exec.Command("git", "init")
	cmdInit.Dir = tempDir
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("failed to init temp git repo: %v", err)
	}

	cmdConfigUser := exec.Command("git", "config", "user.name", "Test User")
	cmdConfigUser.Dir = tempDir
	_ = cmdConfigUser.Run()

	cmdConfigEmail := exec.Command("git", "config", "user.email", "test@example.com")
	cmdConfigEmail.Dir = tempDir
	_ = cmdConfigEmail.Run()

	dummyFile := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(dummyFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = tempDir
	_ = cmdAdd.Run()

	cmdCommit := exec.Command("git", "commit", "-m", "initial commit")
	cmdCommit.Dir = tempDir
	_ = cmdCommit.Run()

	root, err := GetGitRepoRoot(tempDir)
	if err != nil {
		t.Fatalf("GetGitRepoRoot() error = %v", err)
	}

	resolvedTempDir, _ := filepath.EvalSymlinks(tempDir)
	resolvedRoot, _ := filepath.EvalSymlinks(root)

	if resolvedRoot != resolvedTempDir {
		t.Errorf("GetGitRepoRoot() = %q, want %q", resolvedRoot, resolvedTempDir)
	}

	branch, err := GetGitBranch(tempDir)
	if err != nil {
		t.Fatalf("GetGitBranch() error = %v", err)
	}

	if branch == "" {
		t.Error("GetGitBranch() returned empty branch name")
	}
}
