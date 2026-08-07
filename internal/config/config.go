// Package config handles configuration loading, saving, and task reference metadata.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var DefaultTaskFiles = []string{
	"task-requirements.md",
	"task-implementation.md",
	"task-tests.md",
}

// FilesConfig holds settings for task files.
type FilesConfig struct {
	Naming []string `json:"naming" yaml:"naming" toml:"naming"`
}

// Config holds global configuration for tskr.
type Config struct {
	StorageDir string      `json:"storage_dir" yaml:"storage_dir" toml:"storage_dir"`
	Files      FilesConfig `json:"files" yaml:"files" toml:"files"`
}

// GetTaskFiles returns the configured task file names, or DefaultTaskFiles if empty.
func (c *Config) GetTaskFiles() []string {
	if c != nil && len(c.Files.Naming) > 0 {
		return c.Files.Naming
	}
	return DefaultTaskFiles
}

// GetDefaultConfigDir returns default configuration directory ($HOME/.config/tskr).
func GetDefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "tskr"), nil
}

// GetDefaultStorageDir returns default storage directory ($HOME/.tskr).
func GetDefaultStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".tskr"), nil
}

// FindConfigFile searches configDir for config.json, config.yaml, config.yml, or config.toml.
func FindConfigFile(configDir string) string {
	candidates := []string{
		filepath.Join(configDir, "config.json"),
		filepath.Join(configDir, "config.yaml"),
		filepath.Join(configDir, "config.yml"),
		filepath.Join(configDir, "config.toml"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return filepath.Join(configDir, "config.json")
}

// LoadConfig reads Config from configPath (or directory) or creates default config if missing.
func LoadConfig(configPath string) (*Config, error) {
	targetPath := configPath
	if filepath.Ext(configPath) == "" {
		targetPath = FindConfigFile(configPath)
	} else if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		found := FindConfigFile(filepath.Dir(configPath))
		if _, err := os.Stat(found); err == nil {
			targetPath = found
		}
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			defaultStorage, err := GetDefaultStorageDir()
			if err != nil {
				return nil, err
			}
			return &Config{
				StorageDir: defaultStorage,
				Files: FilesConfig{
					Naming: append([]string(nil), DefaultTaskFiles...),
				},
			}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	ext := strings.ToLower(filepath.Ext(targetPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config yaml: %w", err)
		}
	case ".toml":
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config toml: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config json: %w", err)
		}
	}

	if cfg.StorageDir == "" {
		defaultStorage, err := GetDefaultStorageDir()
		if err != nil {
			return nil, err
		}
		cfg.StorageDir = defaultStorage
	}

	if len(cfg.Files.Naming) == 0 {
		cfg.Files.Naming = append([]string(nil), DefaultTaskFiles...)
	}

	return &cfg, nil
}

// SaveConfig writes Config to configPath in JSON, YAML, or TOML format according to file extension.
func SaveConfig(configPath string, cfg *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(configPath))
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal config yaml: %w", err)
		}
	case ".toml":
		data, err = toml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal config toml: %w", err)
		}
	default:
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal config json: %w", err)
		}
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// ConvertConfig converts existing config in configDir to targetFormat (json, yaml, yml, or toml).
func ConvertConfig(configDir string, targetFormat string) (string, error) {
	fmtClean := strings.ToLower(strings.TrimPrefix(targetFormat, "."))
	if fmtClean != "json" && fmtClean != "yaml" && fmtClean != "yml" && fmtClean != "toml" {
		return "", fmt.Errorf("unsupported format %q (must be json, yaml, yml, or toml)", targetFormat)
	}

	oldPath := FindConfigFile(configDir)
	cfg, err := LoadConfig(oldPath)
	if err != nil {
		return "", fmt.Errorf("load existing config: %w", err)
	}

	newFileName := "config." + fmtClean
	newPath := filepath.Join(configDir, newFileName)

	if err := SaveConfig(newPath, cfg); err != nil {
		return "", fmt.Errorf("save converted config: %w", err)
	}

	if filepath.Clean(oldPath) != filepath.Clean(newPath) {
		_ = os.Remove(oldPath)
	}

	return newPath, nil
}
