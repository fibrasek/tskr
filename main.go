package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"tskr/internal/completion"
	"tskr/internal/config"
	"tskr/internal/git"
	"tskr/internal/installer"
)

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
	case "config":
		return handleConfig(args[2:], stdout)
	case "install":
		return installer.Install(args[2:], stdout)
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
	case "completion":
		return completion.Generate(args[2:], stdout)
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
	fmt.Fprintln(w, "  tskr init                  Initialize configuration and storage directory")
	fmt.Fprintln(w, "  tskr config [show]         Show active configuration")
	fmt.Fprintln(w, "  tskr config set <k> <v>    Set configuration value (storage_dir, files.naming)")
	fmt.Fprintln(w, "  tskr config convert <fmt>  Convert configuration file to json, yaml, or toml")
	fmt.Fprintln(w, "  tskr install               Install tskr binary to $HOME/.local/bin")
	fmt.Fprintln(w, "  tskr start                 Start task tracking in current repository")
	fmt.Fprintln(w, "  tskr switch                Switch symlinks to current branch task context")
	fmt.Fprintln(w, "  tskr sync                  Synchronize symlinks for current repository and branch")
	fmt.Fprintln(w, "  tskr list                  List all tracked task references")
	fmt.Fprintln(w, "  tskr go <id>               Print command to navigate to repository and switch branch")
	fmt.Fprintln(w, "  tskr completion <sh>       Generate shell completion script (bash, zsh, fish)")
}

// handleInit initializes user config and storage directories.
func handleInit(w io.Writer) error {
	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := config.FindConfigFile(configDir)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	refsPath := filepath.Join(cfg.StorageDir, "task-references.json")
	refs, err := config.LoadReferences(refsPath)
	if err != nil {
		return err
	}

	if err := config.SaveReferences(refsPath, refs); err != nil {
		return err
	}

	fmt.Fprintf(w, "tskr initialized successfully.\nConfig: %s\nStorage: %s\n", configPath, cfg.StorageDir)
	return nil
}

// handleConfig handles configuration inspection, setting, and conversion commands.
func handleConfig(args []string, w io.Writer) error {
	configDir, err := config.GetDefaultConfigDir()
	if err != nil {
		return err
	}

	if len(args) == 0 || args[0] == "show" {
		configPath := config.FindConfigFile(configDir)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Active Config File: %s\n", configPath)
		fmt.Fprintf(w, "Storage Directory: %s\n", cfg.StorageDir)
		fmt.Fprintf(w, "Task Files:\n")
		for _, f := range cfg.GetTaskFiles() {
			fmt.Fprintf(w, "  - %s\n", f)
		}
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: tskr config set <key> <value>")
		}
		key := strings.ToLower(args[1])
		val := args[2]

		configPath := config.FindConfigFile(configDir)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return err
		}

		switch key {
		case "storage_dir", "storagedir":
			cfg.StorageDir = val
		case "files.naming", "files_naming", "naming", "task_files":
			parts := strings.Split(val, ",")
			var names []string
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					names = append(names, trimmed)
				}
			}
			cfg.Files.Naming = names
		default:
			return fmt.Errorf("unknown config key %q (supported: storage_dir, files.naming)", key)
		}

		if err := config.SaveConfig(configPath, cfg); err != nil {
			return err
		}
		fmt.Fprintf(w, "Configuration %s updated successfully.\n", key)
		return nil

	case "convert":
		if len(args) < 2 {
			return fmt.Errorf("usage: tskr config convert <json|yaml|toml>")
		}
		targetFmt := args[1]
		if targetFmt == "--to" {
			if len(args) < 3 {
				return fmt.Errorf("usage: tskr config convert --to <json|yaml|toml>")
			}
			targetFmt = args[2]
		}

		newPath, err := config.ConvertConfig(configDir, targetFmt)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Configuration converted to %s: %s\n", targetFmt, newPath)
		return nil

	default:
		return fmt.Errorf("unknown config subcommand %q (usage: tskr config [show|set|convert])", subcommand)
	}
}

// handleStart initializes task files and symlinks for current git repo and branch.
func handleStart(w io.Writer) error {
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

// handleSwitch updates symlinks in .task-refs to point to the active branch storage.
func handleSwitch(w io.Writer) error {
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

// handleSync synchronizes symlinks for the current repository and active branch.
func handleSync(w io.Writer) error {
	return handleSwitch(w)
}

// handleList prints formatted table of all tracked repository references.
func handleList(w io.Writer) error {
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

// handleGo prints navigation command string for specified reference ID.
func handleGo(w io.Writer, id int) error {
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
