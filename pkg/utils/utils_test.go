package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileContent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "utils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		content     string
		filePath    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid file",
			content:     "test content",
			filePath:    filepath.Join(tmpDir, "test.txt"),
			expectError: false,
		},
		{
			name:        "Empty file path",
			content:     "",
			filePath:    "",
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "Non-existent file",
			content:     "",
			filePath:    filepath.Join(tmpDir, "nonexistent.txt"),
			expectError: true,
			errorMsg:    "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file if content is provided
			if tt.content != "" {
				if err := os.WriteFile(tt.filePath, []byte(tt.content), DefaultFilePermission); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Test ReadFileContent
			result, err := ReadFileContent(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != tt.content {
					t.Errorf("Expected content '%s', got '%s'", tt.content, result)
				}
			}
		})
	}
}

func TestWriteFileContent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "utils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		content     string
		filePath    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid content",
			content:     "test content",
			filePath:    filepath.Join(tmpDir, "test.txt"),
			expectError: false,
		},
		{
			name:        "Empty file path",
			content:     "test content",
			filePath:    "",
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "Empty content",
			content:     "",
			filePath:    filepath.Join(tmpDir, "empty.txt"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test WriteFileContent
			err := WriteFileContent(tt.filePath, tt.content)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify file was written correctly
				if tt.content != "" {
					readContent, err := ReadFileContent(tt.filePath)
					if err != nil {
						t.Errorf("Failed to read written file: %v", err)
						return
					}
					if readContent != tt.content {
						t.Errorf("Expected content '%s', got '%s'", tt.content, readContent)
					}
				}
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid path",
			path:        "/valid/path",
			expectError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name:        "Path with traversal",
			path:        "/path/../malicious",
			expectError: true,
			errorMsg:    "path contains invalid traversal",
		},
		{
			name:        "Path with double dots",
			path:        "/path/../../etc/passwd",
			expectError: true,
			errorMsg:    "path contains invalid traversal",
		},
		{
			name:        "Relative path",
			path:        "relative/path",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result == "" {
					t.Errorf("Expected non-empty result")
				}
			}
		})
	}
}

func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "utils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		dirPath     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid directory",
			dirPath:     filepath.Join(tmpDir, "newdir"),
			expectError: false,
		},
		{
			name:        "Empty directory path",
			dirPath:     "",
			expectError: true,
			errorMsg:    "directory path cannot be empty",
		},
		{
			name:        "Nested directory",
			dirPath:     filepath.Join(tmpDir, "nested", "deep", "dir"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test EnsureDirectoryExists
			err := EnsureDirectoryExists(tt.dirPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify directory was created
				if tt.dirPath != "" {
					if _, err := os.Stat(tt.dirPath); os.IsNotExist(err) {
						t.Errorf("Directory was not created: %s", tt.dirPath)
					}
				}
			}
		})
	}
} 