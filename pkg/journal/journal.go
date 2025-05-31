package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles journal and archive operations
type Manager struct {
	JournalPath  string
	ArchivePath  string
	OriginalPath string
}

// NewManager creates a new journal manager
func NewManager(filePath string) *Manager {
	baseFileName := filepath.Base(filePath)
	baseName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))
	dirPath := filepath.Dir(filePath)
	
	return &Manager{
		JournalPath:  filepath.Join(dirPath, baseName+".xjournal.md"),
		ArchivePath:  filepath.Join(dirPath, baseName+".xarchive.md"),
		OriginalPath: filePath,
	}
}

// WriteToJournal writes entries to the journal file
func (m *Manager) WriteToJournal(entries []string) error {
	if len(entries) == 0 {
		return nil
	}

	var existingContent []byte
	if _, err := os.Stat(m.JournalPath); err == nil {
		existingContent, err = os.ReadFile(m.JournalPath)
		if err != nil {
			return fmt.Errorf("error reading journal file: %w", err)
		}
	}

	newContent := []byte(strings.Join(entries, "\n") + "\n")
	newContent = append(newContent, existingContent...)
	
	return os.WriteFile(m.JournalPath, newContent, 0644)
}

// WriteToArchive writes entries to the archive file
func (m *Manager) WriteToArchive(entries []string) error {
	if len(entries) == 0 {
		return nil
	}

	var existingContent []byte
	if _, err := os.Stat(m.ArchivePath); err == nil {
		existingContent, err = os.ReadFile(m.ArchivePath)
		if err != nil {
			return fmt.Errorf("error reading archive file: %w", err)
		}
	}

	newContent := []byte(strings.Join(entries, "\n") + "\n")
	newContent = append(newContent, existingContent...)
	
	return os.WriteFile(m.ArchivePath, newContent, 0644)
}

// FormatTimestamp returns a formatted UTC timestamp
func FormatTimestamp() string {
	currentTime := time.Now().UTC()
	return currentTime.Format("[2006-01-02 15:04:05 UTC]")
} 