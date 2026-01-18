package syncclip

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port  string   `yaml:"port"`
	URL   string   `yaml:"url"`
	Peers []string `yaml:"peers"`
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

func normalizeConfig(cfg *Config) {
	normalizePort(cfg)
	normalizeURL(cfg)
	normalizePeers(cfg)
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

func normalizePeers(cfg *Config) {
	for i, addr := range cfg.Peers {
		if strings.Contains(addr, "://") {
			continue
		}

		normalized := addr
		if !strings.HasSuffix(normalized, "/ws") {
			normalized = strings.TrimSuffix(normalized, "/") + "/ws"
		}

		cfg.Peers[i] = "ws://" + normalized
	}
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
