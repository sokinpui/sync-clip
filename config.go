package syncclip

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port string `yaml:"port"`
	URL  string `yaml:"url"`
}

func LoadConfig(explicitPath string, defaultFilename string) (*Config, error) {
	configPath := explicitPath
	if configPath == "" {
		configPath = GetDefaultConfigPath(defaultFilename)
	}

	cfg := &Config{
		Port: ":8080",
		URL:  "http://localhost:8080",
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Port != "" && !strings.Contains(cfg.Port, ":") {
		cfg.Port = ":" + cfg.Port
	}

	if cfg.URL != "" {
		if !strings.Contains(cfg.URL, "://") {
			if !strings.Contains(cfg.URL, ":") {
				cfg.URL = "localhost:" + cfg.URL
			}
			cfg.URL = "http://" + cfg.URL
		}
	}

	return cfg, nil
}

func GetDefaultConfigPath(filename string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filename
	}

	return filepath.Join(configDir, "sync-clip", filename)
}
