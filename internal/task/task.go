// Package task manages branch-specific task contexts, storage directories, and symlinks.
package task

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"tskr/internal/config"
	"tskr/internal/git"
)

// Start initializes task files and symlinks for current git repo and branch.
func Start(w io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working dir: %w", err)
	}

	repoRoot, err := git.GetGitRepoRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	branch, err := git.GetGitBranch(repoRoot)
	if err != nil {
		return fmt.Errorf("get git branch: %w", err)
	}

	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := config.FindConfigFile(configDir)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	repoName := git.ExtractRepoName(repoRoot)
	safeBranch := git.SanitizeBranchName(branch)

	storageBranchDir := filepath.Join(cfg.StorageDir, repoName, safeBranch)
	if err := os.MkdirAll(storageBranchDir, 0755); err != nil {
		return fmt.Errorf("create branch storage dir: %w", err)
	}

	taskFiles := cfg.GetTaskFiles()
	for _, fileName := range taskFiles {
		filePath := filepath.Join(storageBranchDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			content := fmt.Sprintf("# %s\n\nTask context for %s (%s).\n", fileName, repoName, branch)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("create task file %s: %w", fileName, err)
			}
		}
	}

	taskRefsDir := filepath.Join(repoRoot, ".task-refs")
	if _, err := os.Stat(taskRefsDir); err == nil {
		fmt.Fprintf(w, "note: directory .task-refs already exists in %s\n", repoRoot)
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(taskRefsDir, 0755); err != nil {
			return fmt.Errorf("create .task-refs dir: %w", err)
		}
	}

	if err := createSymlinks(storageBranchDir, taskRefsDir, taskFiles); err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := config.LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref := config.AddOrUpdateReference(refs, repoName, repoRoot, branch, storageBranchDir)
	if err := config.SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr started for %s [%s] (ID: %d)\nSymlinks created in .task-refs/\n", repoName, branch, ref.ID)
	return nil
}

// Switch updates symlinks in .task-refs to point to the active branch storage.
func Switch(w io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working dir: %w", err)
	}

	repoRoot, err := git.GetGitRepoRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	branch, err := git.GetGitBranch(repoRoot)
	if err != nil {
		return fmt.Errorf("get git branch: %w", err)
	}

	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := config.FindConfigFile(configDir)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	repoName := git.ExtractRepoName(repoRoot)
	safeBranch := git.SanitizeBranchName(branch)

	storageBranchDir := filepath.Join(cfg.StorageDir, repoName, safeBranch)
	if err := os.MkdirAll(storageBranchDir, 0755); err != nil {
		return fmt.Errorf("create branch storage dir: %w", err)
	}

	taskFiles := cfg.GetTaskFiles()
	for _, fileName := range taskFiles {
		filePath := filepath.Join(storageBranchDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			content := fmt.Sprintf("# %s\n\nTask context for %s (%s).\n", fileName, repoName, branch)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("create task file %s: %w", fileName, err)
			}
		}
	}

	taskRefsDir := filepath.Join(repoRoot, ".task-refs")
	if err := os.MkdirAll(taskRefsDir, 0755); err != nil {
		return fmt.Errorf("create .task-refs dir: %w", err)
	}

	if err := removeSymlinks(taskRefsDir, taskFiles); err != nil {
		return err
	}

	if err := createSymlinks(storageBranchDir, taskRefsDir, taskFiles); err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := config.LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref := config.AddOrUpdateReference(refs, repoName, repoRoot, branch, storageBranchDir)
	if err := config.SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr switched to %s [%s] (ID: %d)\n", repoName, branch, ref.ID)
	return nil
}

// Sync synchronizes symlinks for the current repository and active branch.
func Sync(w io.Writer) error {
	return Switch(w)
}

// List prints formatted table of all tracked repository references.
func List(w io.Writer) error {
	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := config.FindConfigFile(configDir)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := config.LoadReferences(refsPath)
	if err != nil {
		return err
	}

	if len(refs.References) == 0 {
		fmt.Fprintln(w, "No task references found.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "ID\tREPOSITORY\tBRANCH\tREPO PATH")
	fmt.Fprintln(tw, "--\t----------\t------\t---------")

	for _, r := range refs.References {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", r.ID, r.RepoName, r.Branch, r.RepoPath)
	}

	return tw.Flush()
}

// Go prints navigation command string for specified reference ID.
func Go(id int, w io.Writer) error {
	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := config.FindConfigFile(configDir)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := config.LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref, err := config.FindReferenceByID(refs, id)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "cd %q && git checkout %q && tskr switch\n", ref.RepoPath, ref.Branch)
	return nil
}

// createSymlinks creates symlinks in targetDir pointing to files in storageDir.
func createSymlinks(storageDir, targetDir string, files []string) error {
	for _, fileName := range files {
		src := filepath.Join(storageDir, fileName)
		dest := filepath.Join(targetDir, fileName)

		_ = os.Remove(dest)
		if err := os.Symlink(src, dest); err != nil {
			return fmt.Errorf("create symlink from %s to %s: %w", src, dest, err)
		}
	}
	return nil
}

// removeSymlinks deletes symlinks from targetDir.
func removeSymlinks(targetDir string, files []string) error {
	for _, fileName := range files {
		dest := filepath.Join(targetDir, fileName)
		if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove symlink %s: %w", dest, err)
		}
	}

	entries, err := os.ReadDir(targetDir)
	if err == nil {
		for _, entry := range entries {
			info, err := entry.Info()
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				_ = os.Remove(filepath.Join(targetDir, entry.Name()))
			}
		}
	}

	return nil
}
