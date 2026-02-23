package config

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the optional promptarmor.yaml configuration file.
// All fields are optional; missing fields remain at their zero value.
type FileConfig struct {
	Target        string `yaml:"target"`
	Suite         string `yaml:"suite"`
	Concurrency   int    `yaml:"concurrency"`
	Timeout       string `yaml:"timeout"`
	PromptField   string `yaml:"prompt_field"`
	ResponseField string `yaml:"response_field"`
	APIKey        string `yaml:"api_key"`
	Provider      string `yaml:"provider"`
	Model         string `yaml:"model"`
	Output        string `yaml:"output"`
}

// Load reads a YAML configuration file at the given path.
// If the file does not exist, it returns a zero-value FileConfig and no error.
func Load(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return FileConfig{}, nil
		}
		return FileConfig{}, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, err
	}
	return cfg, nil
}
