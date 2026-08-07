// Package cli parses CLI arguments, routes subcommands, and formats help usage.
package cli

import (
	"fmt"
	"io"
	"strconv"

	"tskr/internal/completion"
	"tskr/internal/config"
	"tskr/internal/installer"
	"tskr/internal/task"
)

// Run parses command line arguments and routes execution to the requested subcommand.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 {
		PrintUsage(stdout)
		return nil
	}

	subcommand := args[1]
	switch subcommand {
	case "init":
		return config.Init(stdout)
	case "config":
		return config.HandleCLI(args[2:], stdout)
	case "install":
		return installer.Install(args[2:], stdout)
	case "start":
		return task.Start(stdout)
	case "switch":
		return task.Switch(stdout)
	case "sync":
		return task.Sync(stdout)
	case "list":
		return task.List(stdout)
	case "go":
		if len(args) < 3 {
			return fmt.Errorf("usage: tskr go <id>")
		}
		id, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[2], err)
		}
		return task.Go(id, stdout)
	case "completion":
		return completion.Generate(args[2:], stdout)
	case "-h", "--help", "help":
		PrintUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", subcommand)
	}
}

// PrintUsage writes CLI help text to w.
func PrintUsage(w io.Writer) {
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
