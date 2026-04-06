package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CodeFolder string                    `yaml:"code_folder"`
	Projects   map[string]ProjectDetails `yaml:"projects"`
}

type ProjectDetails struct {
	Git              bool     `yaml:"git"`
	Languages        []string `yaml:"languages"`
	Desc             string   `yaml:"desc"`
	Readme           bool     `yaml:"readme"`
	Show             bool     `yaml:"show"`
	AgentDocs        bool     `yaml:"agent_docs"`
	BuildTestInstall string   `yaml:"build_test_install"` // true, partial, false
	Starred          bool     `yaml:"starred"`
	LastCommitDate   string   `yaml:"last_commit_date"`
	ReadmePreview    string   `yaml:"readme_preview"`
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".projector", "config.yml"), nil
}

func EnsureConfigExists() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create empty config
		file, err := os.Create(configPath)
		if err != nil {
			return "", err
		}
		defer file.Close()
	}
	return configPath, nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func SaveConfig(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
