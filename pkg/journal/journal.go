package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertarles/taskmasterra/v2/pkg/utils"
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

	var existingContent string
	if _, err := os.Stat(m.JournalPath); err == nil {
		existingContent, err = utils.ReadFileContent(m.JournalPath)
		if err != nil {
			return fmt.Errorf("failed to read existing journal file '%s': %w", m.JournalPath, err)
		}
	}

	newContent := strings.Join(entries, "\n") + "\n" + existingContent
	if err := utils.WriteFileContent(m.JournalPath, newContent); err != nil {
		return fmt.Errorf("failed to write journal entries to '%s': %w", m.JournalPath, err)
	}

	return nil
}

// WriteToArchive writes entries to the archive file
func (m *Manager) WriteToArchive(entries []string) error {
	if len(entries) == 0 {
		return nil
	}

	var existingContent string
	if _, err := os.Stat(m.ArchivePath); err == nil {
		existingContent, err = utils.ReadFileContent(m.ArchivePath)
		if err != nil {
			return fmt.Errorf("failed to read existing archive file '%s': %w", m.ArchivePath, err)
		}
	}

	newContent := strings.Join(entries, "\n") + "\n" + existingContent
	if err := utils.WriteFileContent(m.ArchivePath, newContent); err != nil {
		return fmt.Errorf("failed to write archive entries to '%s': %w", m.ArchivePath, err)
	}

	return nil
}

// FormatTimestamp returns a formatted UTC timestamp
func FormatTimestamp() string {
	currentTime := time.Now().UTC()
	return currentTime.Format("[2006-01-02 15:04:05 UTC]")
} 