package config

import (
	"path/filepath"
	"testing"
)

func Test_Config_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		fileName   string
		saveConfig *Config
		wantErr    bool
	}{
		{
			name:       "valid json config save and reload",
			fileName:   "config.json",
			saveConfig: &Config{StorageDir: filepath.Join(tempDir, "custom-tskr"), Files: FilesConfig{Naming: []string{"req.md", "impl.md"}}},
			wantErr:    false,
		},
		{
			name:       "valid yaml config save and reload",
			fileName:   "config.yaml",
			saveConfig: &Config{StorageDir: filepath.Join(tempDir, "custom-tskr"), Files: FilesConfig{Naming: []string{"todo.md"}}},
			wantErr:    false,
		},
		{
			name:       "valid toml config save and reload",
			fileName:   "config.toml",
			saveConfig: &Config{StorageDir: filepath.Join(tempDir, "custom-tskr"), Files: FilesConfig{Naming: []string{"a.md", "b.md"}}},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, tt.fileName)
			err := SaveConfig(configPath, tt.saveConfig)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			loaded, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if loaded.StorageDir != tt.saveConfig.StorageDir {
				t.Errorf("StorageDir = %v, want %v", loaded.StorageDir, tt.saveConfig.StorageDir)
			}

			if len(loaded.GetTaskFiles()) != len(tt.saveConfig.GetTaskFiles()) {
				t.Errorf("GetTaskFiles() = %v, want %v", loaded.GetTaskFiles(), tt.saveConfig.GetTaskFiles())
			}
		})
	}
}

func Test_Config_GetTaskFiles_DefaultFallback(t *testing.T) {
	var cfg *Config
	if len(cfg.GetTaskFiles()) != 3 {
		t.Errorf("nil config expected 3 default task files, got %d", len(cfg.GetTaskFiles()))
	}

	emptyCfg := &Config{}
	if len(emptyCfg.GetTaskFiles()) != 3 {
		t.Errorf("empty config expected 3 default task files, got %d", len(emptyCfg.GetTaskFiles()))
	}

	customCfg := &Config{Files: FilesConfig{Naming: []string{"custom.md"}}}
	if len(customCfg.GetTaskFiles()) != 1 || customCfg.GetTaskFiles()[0] != "custom.md" {
		t.Errorf("custom config expected ['custom.md'], got %v", customCfg.GetTaskFiles())
	}
}

func Test_Config_ConvertConfig(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "config.json")
	initialCfg := &Config{
		StorageDir: "/tmp/storage",
		Files:      FilesConfig{Naming: []string{"task1.md", "task2.md"}},
	}

	if err := SaveConfig(jsonPath, initialCfg); err != nil {
		t.Fatalf("failed to save initial json config: %v", err)
	}

	yamlPath, err := ConvertConfig(tempDir, "yaml")
	if err != nil {
		t.Fatalf("ConvertConfig(yaml) error: %v", err)
	}

	loadedYaml, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig(yaml) error: %v", err)
	}
	if loadedYaml.StorageDir != initialCfg.StorageDir || len(loadedYaml.GetTaskFiles()) != 2 {
		t.Errorf("converted yaml config mismatch: %+v", loadedYaml)
	}

	tomlPath, err := ConvertConfig(tempDir, "toml")
	if err != nil {
		t.Fatalf("ConvertConfig(toml) error: %v", err)
	}

	loadedToml, err := LoadConfig(tomlPath)
	if err != nil {
		t.Fatalf("LoadConfig(toml) error: %v", err)
	}
	if loadedToml.StorageDir != initialCfg.StorageDir || len(loadedToml.GetTaskFiles()) != 2 {
		t.Errorf("converted toml config mismatch: %+v", loadedToml)
	}
}

func Test_References_AddUpdateFind(t *testing.T) {
	tests := []struct {
		name        string
		initialRefs []Reference
		addRepoName string
		addRepoPath string
		addBranch   string
		addStorage  string
		expectedID  int
		expectedLen int
	}{
		{
			name:        "add first reference",
			initialRefs: []Reference{},
			addRepoName: "repo1",
			addRepoPath: "/path/to/repo1",
			addBranch:   "main",
			addStorage:  "/storage/repo1/main",
			expectedID:  1,
			expectedLen: 1,
		},
		{
			name: "update existing reference",
			initialRefs: []Reference{
				{ID: 1, RepoName: "repo1", RepoPath: "/path/to/repo1", Branch: "main", StoragePath: "/old/storage"},
			},
			addRepoName: "repo1",
			addRepoPath: "/path/to/repo1",
			addBranch:   "main",
			addStorage:  "/new/storage",
			expectedID:  1,
			expectedLen: 1,
		},
		{
			name: "add second reference with incremented ID",
			initialRefs: []Reference{
				{ID: 1, RepoName: "repo1", RepoPath: "/path/to/repo1", Branch: "main", StoragePath: "/storage/1"},
			},
			addRepoName: "repo1",
			addRepoPath: "/path/to/repo1",
			addBranch:   "feature",
			addStorage:  "/storage/2",
			expectedID:  2,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refMap := &ReferencesMap{References: tt.initialRefs}
			ref := AddOrUpdateReference(refMap, tt.addRepoName, tt.addRepoPath, tt.addBranch, tt.addStorage)

			if ref.ID != tt.expectedID {
				t.Errorf("AddOrUpdateReference() ID = %v, want %v", ref.ID, tt.expectedID)
			}

			if len(refMap.References) != tt.expectedLen {
				t.Errorf("len(References) = %v, want %v", len(refMap.References), tt.expectedLen)
			}

			found, err := FindReferenceByID(refMap, tt.expectedID)
			if err != nil {
				t.Fatalf("FindReferenceByID() unexpected error: %v", err)
			}

			if found.Branch != tt.addBranch {
				t.Errorf("found.Branch = %v, want %v", found.Branch, tt.addBranch)
			}
		})
	}
}

func Test_FindReferenceByID_notFound(t *testing.T) {
	refMap := &ReferencesMap{
		References: []Reference{
			{ID: 1, RepoName: "r1", RepoPath: "/p1", Branch: "main", StoragePath: "/s1"},
		},
	}

	_, err := FindReferenceByID(refMap, 99)
	if err == nil {
		t.Error("FindReferenceByID() expected error for non-existent ID, got nil")
	}
}
