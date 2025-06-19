package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	// Reminder settings
	DefaultDueHour   int    `json:"default_due_hour"`
	DefaultDueMinute int    `json:"default_due_minute"`
	ReminderListName string `json:"reminder_list_name"`

	// Journal settings
	JournalSuffix string `json:"journal_suffix"`
	ArchiveSuffix string `json:"archive_suffix"`

	// File settings
	DefaultFilePermissions os.FileMode `json:"default_file_permissions"`

	// Task settings
	ActiveMarker string `json:"active_marker"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultDueHour:        16,
		DefaultDueMinute:      0,
		ReminderListName:      "Taskmasterra",
		JournalSuffix:         ".xjournal.md",
		ArchiveSuffix:         ".xarchive.md",
		DefaultFilePermissions: 0644,
		ActiveMarker:          "!!",
	}
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		// Try to find config in default location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return DefaultConfig(), nil
		}
		configPath = filepath.Join(homeDir, ".taskmasterra", "config.json")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return config, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
} 