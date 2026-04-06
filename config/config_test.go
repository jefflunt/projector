package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigExists(t *testing.T) {
	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "projector_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	// For tests that use os.UserHomeDir(), we might need to be careful.
	// In Go, os.UserHomeDir() on Unix systems uses $HOME.

	// Set HOME environment variable to point to tempHome
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Verify the path
	path, err := EnsureConfigExists()
	if err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(tempHome, ".projector", "config.yml")
	if path != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, path)
	}

	// Verify the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestLoadSaveConfig(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config_test.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	cfg := &Config{
		CodeFolder: "/test/folder",
		Projects: map[string]ProjectDetails{
			"test-project": {Git: true},
		},
	}

	err = SaveConfig(tempFile.Name(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	loadedCfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if loadedCfg.CodeFolder != cfg.CodeFolder {
		t.Errorf("expected %s, got %s", cfg.CodeFolder, loadedCfg.CodeFolder)
	}

	if len(loadedCfg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(loadedCfg.Projects))
	}

	if loadedCfg.Projects["test-project"].Git != cfg.Projects["test-project"].Git {
		t.Errorf("expected Git to be true, got false")
	}
}
