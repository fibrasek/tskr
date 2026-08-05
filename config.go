// Package main implements the tskr CLI tool.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds global configuration for tskr.
type Config struct {
	StorageDir string `json:"storage_dir"`
}

// Reference tracks task context metadata for a specific repository and branch.
type Reference struct {
	ID          int    `json:"id"`
	RepoName    string `json:"repo_name"`
	RepoPath    string `json:"repo_path"`
	Branch      string `json:"branch"`
	StoragePath string `json:"storage_path"`
}

// ReferencesMap contains the list of all tracked repository references.
type ReferencesMap struct {
	References []Reference `json:"references"`
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

// LoadConfig reads Config from configPath or creates default config if missing.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			defaultStorage, err := GetDefaultStorageDir()
			if err != nil {
				return nil, err
			}
			return &Config{StorageDir: defaultStorage}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json: %w", err)
	}

	if cfg.StorageDir == "" {
		defaultStorage, err := GetDefaultStorageDir()
		if err != nil {
			return nil, err
		}
		cfg.StorageDir = defaultStorage
	}

	return &cfg, nil
}

// SaveConfig writes Config to configPath in JSON format.
func SaveConfig(configPath string, cfg *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// LoadReferences reads ReferencesMap from refsPath or returns empty map if missing.
func LoadReferences(refsPath string) (*ReferencesMap, error) {
	data, err := os.ReadFile(refsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ReferencesMap{References: []Reference{}}, nil
		}
		return nil, fmt.Errorf("read references file: %w", err)
	}

	var refs ReferencesMap
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("parse references json: %w", err)
	}

	if refs.References == nil {
		refs.References = []Reference{}
	}

	return &refs, nil
}

// SaveReferences writes ReferencesMap to refsPath in JSON format.
func SaveReferences(refsPath string, refs *ReferencesMap) error {
	dir := filepath.Dir(refsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create references dir: %w", err)
	}

	data, err := json.MarshalIndent(refs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal references: %w", err)
	}

	if err := os.WriteFile(refsPath, data, 0644); err != nil {
		return fmt.Errorf("write references file: %w", err)
	}

	return nil
}

// AddOrUpdateReference adds a new reference or updates an existing repository/branch entry.
func AddOrUpdateReference(refs *ReferencesMap, repoName, repoPath, branch, storagePath string) Reference {
	for i, r := range refs.References {
		if r.RepoPath == repoPath && r.Branch == branch {
			refs.References[i].RepoName = repoName
			refs.References[i].StoragePath = storagePath
			return refs.References[i]
		}
	}

	maxID := 0
	for _, r := range refs.References {
		if r.ID > maxID {
			maxID = r.ID
		}
	}

	newRef := Reference{
		ID:          maxID + 1,
		RepoName:    repoName,
		RepoPath:    repoPath,
		Branch:      branch,
		StoragePath: storagePath,
	}

	refs.References = append(refs.References, newRef)
	return newRef
}

// FindReferenceByID returns Reference with specified ID or an error if not found.
func FindReferenceByID(refs *ReferencesMap, id int) (*Reference, error) {
	for _, r := range refs.References {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("reference id %d not found", id)
}

// FindReferenceByRepoAndBranch returns Reference matching repoPath and branch or nil.
func FindReferenceByRepoAndBranch(refs *ReferencesMap, repoPath, branch string) *Reference {
	for _, r := range refs.References {
		if r.RepoPath == repoPath && r.Branch == branch {
			return &r
		}
	}
	return nil
}
