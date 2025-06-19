package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robertarles/taskmasterra/v2/pkg/utils"
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
	data, err := utils.ReadFileContent(configPath)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := utils.EnsureDirectoryExists(configDir); err != nil {
		return fmt.Errorf("failed to create config directory '%s': %w", configDir, err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := utils.WriteFileContent(configPath, string(data)); err != nil {
		return fmt.Errorf("failed to write config file '%s': %w", configPath, err)
	}

	return nil
}

// Validate checks the configuration for invalid or out-of-range values
func (c *Config) Validate() error {
	if c.DefaultDueHour < 0 || c.DefaultDueHour > 23 {
		return fmt.Errorf("default_due_hour must be between 0 and 23 (got %d)", c.DefaultDueHour)
	}
	if c.DefaultDueMinute < 0 || c.DefaultDueMinute > 59 {
		return fmt.Errorf("default_due_minute must be between 0 and 59 (got %d)", c.DefaultDueMinute)
	}
	if c.ReminderListName == "" {
		return fmt.Errorf("reminder_list_name cannot be empty")
	}
	if c.JournalSuffix == "" {
		return fmt.Errorf("journal_suffix cannot be empty")
	}
	if c.ArchiveSuffix == "" {
		return fmt.Errorf("archive_suffix cannot be empty")
	}
	if c.ActiveMarker == "" {
		return fmt.Errorf("active_marker cannot be empty")
	}
	return nil
} 