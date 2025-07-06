package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfigDirName = ".documcp"
	ConfigFileName       = "config.json"
	IndexesDirName       = "indexes"
	ConnectionsDirName   = "connections"
	ProcessesDirName     = "processes"
	envPrefix            = "DOCUMCP_"
)

// Config holds application configuration.
type Config struct {
	AppName string `json:"app_name"`
	Version string `json:"version"`
	// ... add more as needed
}

// Validate checks if the config is valid.
func (c *Config) Validate() error {
	if c.AppName == "" {
		return errors.New("app_name must not be empty")
	}
	if c.Version == "" {
		return errors.New("version must not be empty")
	}
	// ... add more validation as needed
	return nil
}

// applyEnvOverrides overrides config fields with environment variables if set.
func (c *Config) applyEnvOverrides() {
	if v := os.Getenv(envPrefix + "APP_NAME"); v != "" {
		c.AppName = v
	}
	if v := os.Getenv(envPrefix + "VERSION"); v != "" {
		c.Version = v
	}
	// ... add more overrides as needed
}

// GetDefaultConfigDir returns the default config directory in the user's home.
func GetDefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DefaultConfigDirName), nil
}

// EnsureConfigDir ensures the config directory and subdirectories exist.
func EnsureConfigDir(dir string) error {
	subdirs := []string{
		dir,
		filepath.Join(dir, IndexesDirName),
		filepath.Join(dir, ConnectionsDirName),
		filepath.Join(dir, ProcessesDirName),
	}
	for _, d := range subdirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create config dir %s: %w", d, err)
		}
	}
	return nil
}

// LoadConfig loads config.json from the given directory, or creates a default one if missing.
func LoadConfig(dir string) (*Config, error) {
	if err := EnsureConfigDir(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Config dir error: %v\n", err)
		return nil, err
	}
	cfgPath := filepath.Join(dir, ConfigFileName)
	f, err := os.Open(cfgPath)
	if errors.Is(err, os.ErrNotExist) {
		cfg := &Config{
			AppName: "documcp",
			Version: "0.1.0",
		}
		cfg.applyEnvOverrides()
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Config validation error: %v\n", err)
			return nil, err
		}
		if err := SaveConfig(dir, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Config save error: %v\n", err)
			return nil, err
		}
		return cfg, nil
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Config open error: %v\n", err)
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config decode error: %v\n", err)
		return nil, err
	}
	cfg.applyEnvOverrides()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Config validation error: %v\n", err)
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig saves the config to config.json in the given directory.
func SaveConfig(dir string, cfg *Config) error {
	cfgPath := filepath.Join(dir, ConfigFileName)
	f, err := os.Create(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config create error: %v\n", err)
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

// GetIndexesDir returns the indexes directory path.
func GetIndexesDir(configDir string) string {
	return filepath.Join(configDir, IndexesDirName)
}

// GetConnectionsDir returns the connections directory path.
func GetConnectionsDir(configDir string) string {
	return filepath.Join(configDir, ConnectionsDirName)
}

// GetProcessesDir returns the processes directory path.
func GetProcessesDir(configDir string) string {
	return filepath.Join(configDir, ProcessesDirName)
}
