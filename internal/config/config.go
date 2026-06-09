package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConnectionConfig struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Protocol    string `yaml:"protocol" json:"protocol"`
	Host        string `yaml:"host,omitempty" json:"host,omitempty"`
	Port        int    `yaml:"port,omitempty" json:"port,omitempty"`
	Bucket      string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	Region      string `yaml:"region,omitempty" json:"region,omitempty"`
	PathStyle   bool   `yaml:"path_style,omitempty" json:"path_style,omitempty"`
	Secure      bool   `yaml:"secure,omitempty" json:"secure,omitempty"`
	Username    string `yaml:"username,omitempty" json:"username,omitempty"`
	KeychainKey string `yaml:"keychain_key,omitempty" json:"keychain_key,omitempty"`
}

type Config struct {
	Connections []ConnectionConfig `yaml:"connections"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "goexplore")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Connections: []ConnectionConfig{}}, nil
	} else if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
