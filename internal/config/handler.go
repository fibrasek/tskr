package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Init initializes user config and storage directories.
func Init(w io.Writer) error {
	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}
	configPath := FindConfigFile(configDir)

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

// HandleCLI handles configuration inspection, setting, and conversion commands.
func HandleCLI(args []string, w io.Writer) error {
	configDir, err := GetDefaultConfigDir()
	if err != nil {
		return err
	}

	if len(args) == 0 || args[0] == "show" {
		configPath := FindConfigFile(configDir)
		cfg, err := LoadConfig(configPath)
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

		configPath := FindConfigFile(configDir)
		cfg, err := LoadConfig(configPath)
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

		if err := SaveConfig(configPath, cfg); err != nil {
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

		newPath, err := ConvertConfig(configDir, targetFmt)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Configuration converted to %s: %s\n", targetFmt, newPath)
		return nil

	default:
		return fmt.Errorf("unknown config subcommand %q (usage: tskr config [show|set|convert])", subcommand)
	}
}
