package main

import (
	"path/filepath"
	"testing"
)

func Test_Config_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	tests := []struct {
		name       string
		saveConfig *Config
		wantErr    bool
	}{
		{
			name:       "valid config save and reload",
			saveConfig: &Config{StorageDir: filepath.Join(tempDir, "custom-tskr")},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})
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
