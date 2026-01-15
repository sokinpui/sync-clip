package syncclip

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port string `yaml:"port"`
	URL  string `yaml:"url"`
}

func LoadConfig(explicitPath string, defaultFilename string) (*Config, error) {
	cfg := &Config{
		Port: ":2352",
		URL:  "http://localhost:2352",
	}

	configPath := explicitPath
	if configPath == "" {
		configPath = GetDefaultConfigPath(defaultFilename)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		_ = initDefaultConfig(configPath, cfg)
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	normalizeConfig(cfg)
	return cfg, nil
}

func initDefaultConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func normalizeConfig(cfg *Config) {
	normalizePort(cfg)
	normalizeURL(cfg)
}

func normalizePort(cfg *Config) {
	if cfg.Port == "" || strings.Contains(cfg.Port, ":") {
		return
	}
	cfg.Port = ":" + cfg.Port
}

func normalizeURL(cfg *Config) {
	if cfg.URL == "" || strings.Contains(cfg.URL, "://") {
		return
	}

	if !strings.Contains(cfg.URL, ":") {
		cfg.URL = "localhost:" + cfg.URL
	}
	cfg.URL = "http://" + cfg.URL
}

func GetDefaultConfigPath(filename string) string {
	configDir := getBaseConfigDir()
	return filepath.Join(configDir, "sync-clip", filename)
}

func getBaseConfigDir() string {
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config")
		}
	}

	if dir, err := os.UserConfigDir(); err == nil {
		return dir
	}

	return "."
}
