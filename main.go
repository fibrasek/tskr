package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
)

var defaultTaskFiles = []string{
	"task-requirements.md",
	"task-implementation.md",
	"task-tests.md",
}

func main() {
	if err := runCLI(os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// runCLI parses command line arguments and routes execution to the requested subcommand.
func runCLI(args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 {
		printUsage(stdout)
		return nil
	}

	subcommand := args[1]
	switch subcommand {
	case "init":
		return handleInit(stdout)
	case "install":
		return handleInstall(stdout)
	case "start":
		return handleStart(stdout)
	case "switch":
		return handleSwitch(stdout)
	case "sync":
		return handleSync(stdout)
	case "list":
		return handleList(stdout)
	case "go":
		if len(args) < 3 {
			return fmt.Errorf("usage: tskr go <id>")
		}
		id, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[2], err)
		}
		return handleGo(stdout, id)
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", subcommand)
	}
}

// printUsage writes CLI help text to w.
func printUsage(w io.Writer) {
	fmt.Fprintln(w, "tskr - task reference context manager")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  tskr init         Initialize configuration and storage directory")
	fmt.Fprintln(w, "  tskr install      Install tskr binary to $HOME/.local/bin")
	fmt.Fprintln(w, "  tskr start        Start task tracking in current repository")
	fmt.Fprintln(w, "  tskr switch       Switch symlinks to current branch task context")
	fmt.Fprintln(w, "  tskr sync         Synchronize symlinks for current repository and branch")
	fmt.Fprintln(w, "  tskr list         List all tracked task references")
	fmt.Fprintln(w, "  tskr go <id>      Print command to navigate to repository and switch branch")
}

// handleInit initializes user config and storage directories.
func handleInit(w io.Writer) error {
	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := LoadReferences(refsPath)
	if err != nil {
		return err
	}

	if err := SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr initialized successfully.\nConfig: %s\nStorage: %s\n", configPath, cfg.StorageDir)
	return nil
}

// handleInstall copies the binary to $HOME/.local/bin and checks PATH.
func handleInstall(w io.Writer) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get user home dir: %w", err)
	}

	installDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	destPath := filepath.Join(installDir, "tskr")

	srcFile, err := os.Open(exePath)
	if err != nil {
		return fmt.Errorf("open binary for reading: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create destination binary: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}

	fmt.Fprintf(w, "tskr installed to %s\n", destPath)

	pathEnv := os.Getenv("PATH")
	if !filepathInPATH(installDir, pathEnv) {
		fmt.Fprintf(w, "warning: %s is not in your PATH. Add it to your shell profile.\n", installDir)
	} else {
		fmt.Fprintln(w, "tskr is in your PATH and ready to use.")
	}

	return nil
}

// filepathInPATH returns true if targetDir matches any directory entry in pathEnv.
func filepathInPATH(targetDir, pathEnv string) bool {
	dirs := filepath.SplitList(pathEnv)
	targetClean := filepath.Clean(targetDir)
	for _, d := range dirs {
		if filepath.Clean(d) == targetClean {
			return true
		}
	}
	return false
}

// handleStart initializes task files and symlinks for current git repo and branch.
func handleStart(w io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working dir: %w", err)
	}

	repoRoot, err := GetGitRepoRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	branch, err := GetGitBranch(repoRoot)
	if err != nil {
		return fmt.Errorf("get git branch: %w", err)
	}

	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	repoName := ExtractRepoName(repoRoot)
	safeBranch := SanitizeBranchName(branch)

	storageBranchDir := filepath.Join(cfg.StorageDir, repoName, safeBranch)
	if err := os.MkdirAll(storageBranchDir, 0755); err != nil {
		return fmt.Errorf("create branch storage dir: %w", err)
	}

	for _, fileName := range defaultTaskFiles {
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

	if err := createSymlinks(storageBranchDir, taskRefsDir); err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref := AddOrUpdateReference(refs, repoName, repoRoot, branch, storageBranchDir)
	if err := SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr started for %s [%s] (ID: %d)\nSymlinks created in .task-refs/\n", repoName, branch, ref.ID)
	return nil
}

// handleSwitch updates symlinks in .task-refs to point to the active branch storage.
func handleSwitch(w io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working dir: %w", err)
	}

	repoRoot, err := GetGitRepoRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	branch, err := GetGitBranch(repoRoot)
	if err != nil {
		return fmt.Errorf("get git branch: %w", err)
	}

	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	repoName := ExtractRepoName(repoRoot)
	safeBranch := SanitizeBranchName(branch)

	storageBranchDir := filepath.Join(cfg.StorageDir, repoName, safeBranch)
	if err := os.MkdirAll(storageBranchDir, 0755); err != nil {
		return fmt.Errorf("create branch storage dir: %w", err)
	}

	for _, fileName := range defaultTaskFiles {
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

	if err := removeSymlinks(taskRefsDir); err != nil {
		return err
	}

	if err := createSymlinks(storageBranchDir, taskRefsDir); err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref := AddOrUpdateReference(refs, repoName, repoRoot, branch, storageBranchDir)
	if err := SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr switched to %s [%s] (ID: %d)\n", repoName, branch, ref.ID)
	return nil
}

// handleSync synchronizes symlinks for the current repository and active branch.
func handleSync(w io.Writer) error {
	return handleSwitch(w)
}

// handleList prints formatted table of all tracked repository references.
func handleList(w io.Writer) error {
	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := LoadReferences(refsPath)
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

// handleGo prints navigation command string for specified reference ID.
func handleGo(w io.Writer, id int) error {
	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := LoadReferences(refsPath)
	if err != nil {
		return err
	}

	ref, err := FindReferenceByID(refs, id)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "cd %q && git checkout %q && tskr switch\n", ref.RepoPath, ref.Branch)
	return nil
}

// createSymlinks creates symlinks in targetDir pointing to files in storageDir.
func createSymlinks(storageDir, targetDir string) error {
	for _, fileName := range defaultTaskFiles {
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
func removeSymlinks(targetDir string) error {
	for _, fileName := range defaultTaskFiles {
		dest := filepath.Join(targetDir, fileName)
		if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove symlink %s: %w", dest, err)
		}
	}
	return nil
}
