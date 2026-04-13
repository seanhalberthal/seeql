package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration.
type Config struct {
	Editor      EditorConfig      `json:"editor"`
	Results     ResultsConfig     `json:"results"`
	Audit       AuditConfig       `json:"audit"`
	Connections []SavedConnection `json:"connections"`
}

// AuditConfig controls the JSON Lines audit log.
type AuditConfig struct {
	Enabled   bool   `json:"enabled"`
	Path      string `json:"path"`        // empty = ConfigDir()/audit.jsonl
	MaxSizeMB int    `json:"max_size_mb"` // 0 = no rotation
}

// EditorConfig holds editor-related settings.
type EditorConfig struct {
	TabSize         int  `json:"tab_size"`
	ShowLineNumbers bool `json:"show_line_numbers"`
}

// ResultsConfig holds result display settings.
type ResultsConfig struct {
	PageSize       int `json:"page_size"`
	MaxColumnWidth int `json:"max_column_width"`
}

// SavedConnection holds a saved database connection.
// Adapter is auto-detected from the DSN.
type SavedConnection struct {
	Name string `json:"name,omitempty"`
	DSN  string `json:"dsn"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			TabSize:         4,
			ShowLineNumbers: true,
		},
		Results: ResultsConfig{
			PageSize:       1000,
			MaxColumnWidth: 50,
		},
	}
}

// ConfigDir returns the seeql configuration directory path.
// It uses os.UserConfigDir to locate the base config directory and
// appends "seeql" to it, typically resulting in ~/.config/seeql/.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(base, "seeql"), nil
}

// Load reads a Config from the JSON file at path. If the file does not exist,
// it returns DefaultConfig without error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// LoadDefault loads configuration from the default path
// (ConfigDir()/config.json).
func LoadDefault() (*Config, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	return Load(filepath.Join(dir, "config.json"))
}

// Save writes the Config to the JSON file at path atomically, creating any
// necessary parent directories. It writes to a temp file first and renames
// to avoid corruption on crash.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	// Write to a temp file in the same directory, then rename for atomicity.
	tmp, err := os.CreateTemp(dir, ".config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

// SaveDefault writes the Config to the default path
// (ConfigDir()/config.json).
func (c *Config) SaveDefault() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return c.Save(filepath.Join(dir, "config.json"))
}
