package scanner

import (
	"bufio"
	"os"
	"os/exec"
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

				// Git Date
				cmd := exec.Command("git", "-C", projectPath, "log", "-1", "--format=%cd", "--date=short")
				out, err := cmd.Output()
				if err == nil {
					details.LastCommitDate = strings.TrimSpace(string(out))
				}
			}

			// README Preview
			readmeNames := []string{"README.md", "README", "README.txt"}
			for _, name := range readmeNames {
				path := filepath.Join(projectPath, name)
				if _, err := os.Stat(path); err == nil {
					file, err := os.Open(path)
					if err == nil {
						scanner := bufio.NewScanner(file)
						var lines []string
						for i := 0; i < 10 && scanner.Scan(); i++ {
							lines = append(lines, scanner.Text())
						}
						details.ReadmePreview = strings.Join(lines, "\n")
						file.Close()
						break
					}
				}
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

			// Check for starred
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

			// Detect languages
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
					case ".kt":
						languages["Kotlin"] = true
					case ".swift":
						languages["Swift"] = true
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
