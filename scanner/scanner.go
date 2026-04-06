package scanner

import (
	"os"
	"path/filepath"
	"projector/config"
	"strings"
)

func ScanProjects(codeFolder string) (map[string]config.ProjectDetails, error) {
	projects := make(map[string]config.ProjectDetails)

	entries, err := os.ReadDir(codeFolder)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			projectPath := filepath.Join(codeFolder, entry.Name())
			details := config.ProjectDetails{
				Show: true,
			}

			// Check for git
			if _, err := os.Stat(filepath.Join(projectPath, ".git")); err == nil {
				details.Git = true
			}

			// Check for README
			files, _ := os.ReadDir(projectPath)
			for _, f := range files {
				if !f.IsDir() {
					name := strings.ToUpper(f.Name())
					if name == "README" || name == "README.MD" || name == "README.TXT" {
						details.Readme = true
					}
				}
			}

			// ... existing detection logic ...
			// Check for starred: assuming if a .starred file exists in root
			if _, err := os.Stat(filepath.Join(projectPath, ".starred")); err == nil {
				details.Starred = true
			} else {
				details.Starred = false
			}

			// Check for build-test-install scripts
			scriptDir := filepath.Join(projectPath, "script")
			scripts := []string{"build", "test", "install", "build-test-install"}
			foundCount := 0
			for _, s := range scripts {
				if _, err := os.Stat(filepath.Join(scriptDir, s)); err == nil {
					foundCount++
				}
			}

			if foundCount == len(scripts) {
				details.BuildTestInstall = "true"
			} else if foundCount > 0 {
				details.BuildTestInstall = "partial"
			} else {
				details.BuildTestInstall = "false"
			}

			// Detect languages (simplistic)
			languages := make(map[string]bool)
			filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					ext := strings.ToLower(filepath.Ext(path))
					switch ext {
					case ".rb":
						languages["Ruby"] = true
					case ".html":
						languages["HTML"] = true
					case ".go":
						languages["Go"] = true
					case ".c", ".h", ".cpp":
						languages["C/C++"] = true
					case ".sh":
						languages["Shell"] = true
					case ".js", ".ts":
						languages["JS/TS"] = true
					case ".css", ".sass":
						languages["CSS/Sass"] = true
					case ".java":
						languages["Java"] = true
					case ".py":
						languages["Python"] = true
					}
				}
				return nil
			})
			for lang := range languages {
				details.Languages = append(details.Languages, lang)
			}

			projects[entry.Name()] = details
		}
	}

	return projects, nil
}
