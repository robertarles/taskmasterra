package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultDueHour != 16 {
		t.Errorf("Expected DefaultDueHour to be 16, got %d", config.DefaultDueHour)
	}

	if config.DefaultDueMinute != 0 {
		t.Errorf("Expected DefaultDueMinute to be 0, got %d", config.DefaultDueMinute)
	}

	if config.ReminderListName != "Taskmasterra" {
		t.Errorf("Expected ReminderListName to be 'Taskmasterra', got %s", config.ReminderListName)
	}

	if config.JournalSuffix != ".xjournal.md" {
		t.Errorf("Expected JournalSuffix to be '.xjournal.md', got %s", config.JournalSuffix)
	}

	if config.ArchiveSuffix != ".xarchive.md" {
		t.Errorf("Expected ArchiveSuffix to be '.xarchive.md', got %s", config.ArchiveSuffix)
	}

	if config.DefaultFilePermissions != 0644 {
		t.Errorf("Expected DefaultFilePermissions to be 0644, got %o", config.DefaultFilePermissions)
	}

	if config.ActiveMarker != "!!" {
		t.Errorf("Expected ActiveMarker to be '!!', got %s", config.ActiveMarker)
	}
}

func TestLoadConfig_NewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Test loading non-existent config (should create default)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be returned")
	}

	// Verify default config was created
	if config.DefaultDueHour != 16 {
		t.Errorf("Expected DefaultDueHour to be 16, got %d", config.DefaultDueHour)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestLoadConfig_ExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create a custom config
	customConfig := &Config{
		DefaultDueHour:   10,
		DefaultDueMinute: 30,
		ReminderListName: "CustomList",
		JournalSuffix:    ".journal.md",
		ArchiveSuffix:    ".archive.md",
		ActiveMarker:     "**",
	}

	// Save the custom config
	if err := SaveConfig(customConfig, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded config matches saved config
	if loadedConfig.DefaultDueHour != customConfig.DefaultDueHour {
		t.Errorf("Expected DefaultDueHour to be %d, got %d", customConfig.DefaultDueHour, loadedConfig.DefaultDueHour)
	}

	if loadedConfig.DefaultDueMinute != customConfig.DefaultDueMinute {
		t.Errorf("Expected DefaultDueMinute to be %d, got %d", customConfig.DefaultDueMinute, loadedConfig.DefaultDueMinute)
	}

	if loadedConfig.ReminderListName != customConfig.ReminderListName {
		t.Errorf("Expected ReminderListName to be %s, got %s", customConfig.ReminderListName, loadedConfig.ReminderListName)
	}

	if loadedConfig.JournalSuffix != customConfig.JournalSuffix {
		t.Errorf("Expected JournalSuffix to be %s, got %s", customConfig.JournalSuffix, loadedConfig.JournalSuffix)
	}

	if loadedConfig.ArchiveSuffix != customConfig.ArchiveSuffix {
		t.Errorf("Expected ArchiveSuffix to be %s, got %s", customConfig.ArchiveSuffix, loadedConfig.ArchiveSuffix)
	}

	if loadedConfig.ActiveMarker != customConfig.ActiveMarker {
		t.Errorf("Expected ActiveMarker to be %s, got %s", customConfig.ActiveMarker, loadedConfig.ActiveMarker)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	config := DefaultConfig()

	// Save config
	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Verify file content can be read back
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.DefaultDueHour != config.DefaultDueHour {
		t.Errorf("Loaded config doesn't match saved config")
	}
} 