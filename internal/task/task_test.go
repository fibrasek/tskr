package task

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tskr/internal/config"
)

func setupTestRepo(t *testing.T, repoDir string) {
	t.Helper()
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	cmdInit := exec.Command("git", "init")
	cmdInit.Dir = repoDir
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	cmdUser := exec.Command("git", "config", "user.name", "Tester")
	cmdUser.Dir = repoDir
	_ = cmdUser.Run()

	cmdEmail := exec.Command("git", "config", "user.email", "test@example.com")
	cmdEmail.Dir = repoDir
	_ = cmdEmail.Run()

	dummy := filepath.Join(repoDir, "file.txt")
	if err := os.WriteFile(dummy, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = repoDir
	_ = cmdAdd.Run()

	cmdCommit := exec.Command("git", "commit", "-m", "initial commit")
	cmdCommit.Dir = repoDir
	_ = cmdCommit.Run()
}

func Test_StartAndSwitch(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	repoDir := filepath.Join(tempHome, "task-repo")
	setupTestRepo(t, repoDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(repoDir)
	defer func() { _ = os.Chdir(origWd) }()

	var stdout bytes.Buffer

	if err := Start(&stdout); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "tskr started for task-repo") {
		t.Errorf("stdout = %q, expected 'tskr started for task-repo'", stdout.String())
	}

	taskRefsDir := filepath.Join(repoDir, ".task-refs")
	for _, f := range config.DefaultTaskFiles {
		linkPath := filepath.Join(taskRefsDir, f)
		info, err := os.Lstat(linkPath)
		if err != nil {
			t.Errorf("expected symlink %s to exist: %v", linkPath, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be a symlink", linkPath)
		}
	}

	cmdCheckout := exec.Command("git", "checkout", "-b", "feature/login")
	cmdCheckout.Dir = repoDir
	if err := cmdCheckout.Run(); err != nil {
		t.Fatalf("failed to checkout new branch: %v", err)
	}

	stdout.Reset()
	if err := Switch(&stdout); err != nil {
		t.Fatalf("Switch() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "tskr switched to task-repo") {
		t.Errorf("unexpected output: %s", stdout.String())
	}

	stdout.Reset()
	if err := List(&stdout); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "task-repo") {
		t.Errorf("List() output missing repo name: %s", stdout.String())
	}

	stdout.Reset()
	if err := Go(1, &stdout); err != nil {
		t.Fatalf("Go(1) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "cd") {
		t.Errorf("Go(1) output missing cd command: %s", stdout.String())
	}
}
