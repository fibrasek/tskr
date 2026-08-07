package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tskr/internal/cli"
	"tskr/internal/config"
)

func setupTestGitRepo(t *testing.T, repoDir string) {
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

func Test_runCLI_Table(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	repoDir := filepath.Join(tempHome, "my-repo")
	setupTestGitRepo(t, repoDir)

	tests := []struct {
		name          string
		workingDir    string
		args          []string
		wantErr       bool
		expectedOut   string
		checkArtifact func(t *testing.T)
	}{
		{
			name:        "subcommand init",
			workingDir:  tempHome,
			args:        []string{"tskr", "init"},
			wantErr:     false,
			expectedOut: "tskr initialized successfully",
			checkArtifact: func(t *testing.T) {
				cfgFile := filepath.Join(tempHome, ".config", "tskr", "config.json")
				if _, err := os.Stat(cfgFile); err != nil {
					t.Errorf("expected config file %s to exist", cfgFile)
				}
			},
		},
		{
			name:        "subcommand start in git repo",
			workingDir:  repoDir,
			args:        []string{"tskr", "start"},
			wantErr:     false,
			expectedOut: "tskr started for my-repo",
			checkArtifact: func(t *testing.T) {
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
			},
		},
		{
			name:        "subcommand list",
			workingDir:  repoDir,
			args:        []string{"tskr", "list"},
			wantErr:     false,
			expectedOut: "my-repo",
		},
		{
			name:        "subcommand go with valid id",
			workingDir:  tempHome,
			args:        []string{"tskr", "go", "1"},
			wantErr:     false,
			expectedOut: "cd",
		},
		{
			name:       "subcommand go with invalid id",
			workingDir: tempHome,
			args:       []string{"tskr", "go", "999"},
			wantErr:    true,
		},
		{
			name:        "subcommand config show",
			workingDir:  tempHome,
			args:        []string{"tskr", "config", "show"},
			wantErr:     false,
			expectedOut: "Active Config File:",
		},
		{
			name:        "subcommand config set files.naming",
			workingDir:  tempHome,
			args:        []string{"tskr", "config", "set", "files.naming", "fileA.md,fileB.md"},
			wantErr:     false,
			expectedOut: "Configuration files.naming updated successfully",
		},
		{
			name:        "subcommand config convert yaml",
			workingDir:  tempHome,
			args:        []string{"tskr", "config", "convert", "yaml"},
			wantErr:     false,
			expectedOut: "Configuration converted to yaml",
		},
		{
			name:        "subcommand install with custom dir",
			workingDir:  tempHome,
			args:        []string{"tskr", "install", "--dir", filepath.Join(tempHome, "custombin")},
			wantErr:     false,
			expectedOut: "tskr installed to",
			checkArtifact: func(t *testing.T) {
				installedBin := filepath.Join(tempHome, "custombin", "tskr")
				if _, err := os.Stat(installedBin); err != nil {
					t.Errorf("expected installed binary %s to exist", installedBin)
				}
			},
		},
		{
			name:        "subcommand completion bash",
			workingDir:  tempHome,
			args:        []string{"tskr", "completion", "bash"},
			wantErr:     false,
			expectedOut: "bash completion for tskr",
		},
		{
			name:        "subcommand completion zsh",
			workingDir:  tempHome,
			args:        []string{"tskr", "completion", "zsh"},
			wantErr:     false,
			expectedOut: "#compdef tskr",
		},
		{
			name:        "subcommand completion fish",
			workingDir:  tempHome,
			args:        []string{"tskr", "completion", "fish"},
			wantErr:     false,
			expectedOut: "fish completion for tskr",
		},
		{
			name:       "unknown subcommand",
			workingDir: tempHome,
			args:       []string{"tskr", "foobar"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origWd, _ := os.Getwd()
			_ = os.Chdir(tt.workingDir)
			defer func() { _ = os.Chdir(origWd) }()

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := cli.Run(tt.args, &stdout, &stderr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("runCLI() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectedOut != "" && !strings.Contains(stdout.String(), tt.expectedOut) {
				t.Errorf("stdout = %q, want substring %q", stdout.String(), tt.expectedOut)
			}

			if tt.checkArtifact != nil {
				tt.checkArtifact(t)
			}
		})
	}
}

func Test_switch_newBranch(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	repoDir := filepath.Join(tempHome, "switch-repo")
	setupTestGitRepo(t, repoDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(repoDir)
	defer func() { _ = os.Chdir(origWd) }()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := cli.Run([]string{"tskr", "start"}, &stdout, &stderr); err != nil {
		t.Fatalf("tskr start failed: %v", err)
	}

	cmdCheckout := exec.Command("git", "checkout", "-b", "feature/login")
	cmdCheckout.Dir = repoDir
	if err := cmdCheckout.Run(); err != nil {
		t.Fatalf("failed to checkout new branch: %v", err)
	}

	stdout.Reset()
	if err := cli.Run([]string{"tskr", "switch"}, &stdout, &stderr); err != nil {
		t.Fatalf("tskr switch failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "tskr switched to switch-repo") {
		t.Errorf("unexpected output: %s", stdout.String())
	}

	// Verify symlinks point to feature-login directory
	taskRefsDir := filepath.Join(repoDir, ".task-refs")
	linkPath := filepath.Join(taskRefsDir, "task-requirements.md")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink %s: %v", linkPath, err)
	}

	if !strings.Contains(target, "feature-login") {
		t.Errorf("symlink target = %q, want it to contain 'feature-login'", target)
	}
}
