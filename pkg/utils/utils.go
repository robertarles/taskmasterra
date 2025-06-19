package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common constants to avoid magic numbers
const (
	DefaultFilePermission = 0644
	DefaultDirPermission  = 0755
	MaxFileSize          = 10 * 1024 * 1024 // 10MB
)

// ReadFileContent reads a file and returns its content as a string
func ReadFileContent(filePath string) (string, error) {
	// Validate file path
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file '%s' does not exist", filePath)
		}
		return "", fmt.Errorf("failed to access file '%s': %w", filePath, err)
	}

	// Check file size
	if fileInfo.Size() > MaxFileSize {
		return "", fmt.Errorf("file '%s' is too large (%d bytes, max %d bytes)", 
			filePath, fileInfo.Size(), MaxFileSize)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	return string(content), nil
}

// WriteFileContent writes content to a file with proper error handling
func WriteFileContent(filePath string, content string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, DefaultDirPermission); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), DefaultFilePermission); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", filePath, err)
	}

	return nil
}

// SanitizePath validates and cleans a file path
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check for path traversal attempts in the original path
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path contains invalid traversal: %s", path)
	}

	// Clean the path
	cleanPath := filepath.Clean(path)
	
	return cleanPath, nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	if err := os.MkdirAll(dirPath, DefaultDirPermission); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dirPath, err)
	}

	return nil
} 