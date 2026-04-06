package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanProjects(t *testing.T) {
	// Create a temporary code directory
	tempCodeFolder, err := os.MkdirTemp("", "projector_test_code")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempCodeFolder)

	// Create project structures
	// 1. Simple project
	os.Mkdir(filepath.Join(tempCodeFolder, "simple"), 0755)

	// 2. Git project
	gitDir := filepath.Join(tempCodeFolder, "git-project", ".git")
	os.MkdirAll(gitDir, 0755)

	// 3. Starred project
	os.Mkdir(filepath.Join(tempCodeFolder, "starred-project"), 0755)
	os.WriteFile(filepath.Join(tempCodeFolder, "starred-project", ".starred"), []byte(""), 0644)

	// 4. README project
	os.Mkdir(filepath.Join(tempCodeFolder, "readme-project"), 0755)
	os.WriteFile(filepath.Join(tempCodeFolder, "readme-project", "README.md"), []byte("# Hello"), 0644)

	// Scan
	projects, err := ScanProjects(tempCodeFolder)
	if err != nil {
		t.Fatal(err)
	}

	// Assertions
	if len(projects) != 4 {
		t.Errorf("expected 4 projects, got %d", len(projects))
	}

	if p, ok := projects["simple"]; !ok || p.Git || p.Starred || !p.Show {
		t.Errorf("incorrect simple project: %+v", p)
	}

	if p, ok := projects["git-project"]; !ok || !p.Git {
		t.Errorf("incorrect git project: %+v", p)
	}

	if p, ok := projects["starred-project"]; !ok || !p.Starred {
		t.Errorf("incorrect starred project: %+v", p)
	}

	if p, ok := projects["readme-project"]; !ok || !p.Readme {
		t.Errorf("incorrect readme project: %+v", p)
	}
}
