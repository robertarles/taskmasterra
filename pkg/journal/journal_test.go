package journal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	filePath := "/tmp/test-todo.md"
	jm := NewManager(filePath)
	if !strings.HasSuffix(jm.JournalPath, ".xjournal.md") {
		t.Errorf("Expected JournalPath to end with .xjournal.md, got %s", jm.JournalPath)
	}
	if !strings.HasSuffix(jm.ArchivePath, ".xarchive.md") {
		t.Errorf("Expected ArchivePath to end with .xarchive.md, got %s", jm.ArchivePath)
	}
	if jm.OriginalPath != filePath {
		t.Errorf("Expected OriginalPath to be %s, got %s", filePath, jm.OriginalPath)
	}
}

func TestWriteToJournal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		entries      []string
		existingData string
		wantContains []string
		expectError  bool
	}{
		{
			name:         "Write to empty journal",
			entries:      []string{"Entry 1", "Entry 2"},
			existingData: "",
			wantContains: []string{"Entry 1", "Entry 2"},
			expectError:  false,
		},
		{
			name:         "Write to existing journal",
			entries:      []string{"New Entry 1", "New Entry 2"},
			existingData: "Existing Entry\n",
			wantContains: []string{"New Entry 1", "New Entry 2", "Existing Entry"},
			expectError:  false,
		},
		{
			name:         "Write empty entries",
			entries:      []string{},
			existingData: "Existing Entry\n",
			wantContains: []string{"Existing Entry"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoPath := filepath.Join(tmpDir, "todo.md")
			manager := NewManager(todoPath)

			// Create journal file with existing data if any
			if tt.existingData != "" {
				if err := os.WriteFile(manager.JournalPath, []byte(tt.existingData), 0644); err != nil {
					t.Fatalf("Failed to write existing data: %v", err)
				}
			}

			err := manager.WriteToJournal(tt.entries)
			if (err != nil) != tt.expectError {
				t.Errorf("WriteToJournal() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if len(tt.entries) > 0 || tt.existingData != "" {
				content, err := os.ReadFile(manager.JournalPath)
				if err != nil {
					t.Fatalf("Failed to read journal file: %v", err)
				}

				for _, want := range tt.wantContains {
					if !strings.Contains(string(content), want) {
						t.Errorf("Journal content does not contain %q", want)
					}
				}
			}
		})
	}
}

func TestWriteToArchive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		entries      []string
		existingData string
		wantContains []string
		expectError  bool
	}{
		{
			name:         "Write to empty archive",
			entries:      []string{"Entry 1", "Entry 2"},
			existingData: "",
			wantContains: []string{"Entry 1", "Entry 2"},
			expectError:  false,
		},
		{
			name:         "Write to existing archive",
			entries:      []string{"New Entry 1", "New Entry 2"},
			existingData: "Existing Entry\n",
			wantContains: []string{"New Entry 1", "New Entry 2", "Existing Entry"},
			expectError:  false,
		},
		{
			name:         "Write empty entries",
			entries:      []string{},
			existingData: "Existing Entry\n",
			wantContains: []string{"Existing Entry"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoPath := filepath.Join(tmpDir, "todo.md")
			manager := NewManager(todoPath)

			// Create archive file with existing data if any
			if tt.existingData != "" {
				if err := os.WriteFile(manager.ArchivePath, []byte(tt.existingData), 0644); err != nil {
					t.Fatalf("Failed to write existing data: %v", err)
				}
			}

			err := manager.WriteToArchive(tt.entries)
			if (err != nil) != tt.expectError {
				t.Errorf("WriteToArchive() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if len(tt.entries) > 0 || tt.existingData != "" {
				content, err := os.ReadFile(manager.ArchivePath)
				if err != nil {
					t.Fatalf("Failed to read archive file: %v", err)
				}

				for _, want := range tt.wantContains {
					if !strings.Contains(string(content), want) {
						t.Errorf("Archive content does not contain %q", want)
					}
				}
			}
		})
	}
}

func TestWriteToJournalAndArchive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "todo.md")
	jm := NewManager(filePath)

	entries1 := []string{"entry1", "entry2"}
	entries2 := []string{"entry3"}

	// Write first entries
	if err := jm.WriteToJournal(entries1); err != nil {
		t.Fatalf("WriteToJournal failed: %v", err)
	}
	if err := jm.WriteToArchive(entries1); err != nil {
		t.Fatalf("WriteToArchive failed: %v", err)
	}

	// Write second entries (should prepend)
	if err := jm.WriteToJournal(entries2); err != nil {
		t.Fatalf("WriteToJournal failed: %v", err)
	}
	if err := jm.WriteToArchive(entries2); err != nil {
		t.Fatalf("WriteToArchive failed: %v", err)
	}

	// Check that entries2 is before entries1
	journalContent, err := os.ReadFile(jm.JournalPath)
	if err != nil {
		t.Fatalf("Failed to read journal: %v", err)
	}
	archiveContent, err := os.ReadFile(jm.ArchivePath)
	if err != nil {
		t.Fatalf("Failed to read archive: %v", err)
	}
	journalLines := strings.Split(string(journalContent), "\n")
	archiveLines := strings.Split(string(archiveContent), "\n")
	if len(journalLines) < 3 || journalLines[0] != "entry3" || journalLines[1] != "entry1" {
		t.Errorf("Journal entries not prepended correctly: %v", journalLines)
	}
	if len(archiveLines) < 3 || archiveLines[0] != "entry3" || archiveLines[1] != "entry1" {
		t.Errorf("Archive entries not prepended correctly: %v", archiveLines)
	}
}

func TestWriteToJournal_Error(t *testing.T) {
	// Use a directory as the file path to force a write error
	dir, err := os.MkdirTemp("", "journal-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	jm := NewManager(dir) // JournalPath will be a directory
	entries := []string{"entry"}
	// Create a directory at the journal path
	if err := os.Mkdir(jm.JournalPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := jm.WriteToJournal(entries); err == nil {
		t.Error("Expected error when writing to a directory, got nil")
	}
}

func TestWriteToArchive_Error(t *testing.T) {
	// Use a directory as the file path to force a write error
	dir, err := os.MkdirTemp("", "archive-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	jm := NewManager(dir) // ArchivePath will be a directory
	entries := []string{"entry"}
	// Create a directory at the archive path
	if err := os.Mkdir(jm.ArchivePath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := jm.WriteToArchive(entries); err == nil {
		t.Error("Expected error when writing to a directory, got nil")
	}
}

func TestFormatTimestamp(t *testing.T) {
	ts := FormatTimestamp()
	if !strings.HasPrefix(ts, "[") || !strings.HasSuffix(ts, "UTC]") {
		t.Errorf("Timestamp format invalid: %s", ts)
	}
	// Check that it parses as a time
	trimmed := strings.Trim(ts, "[]UTC ")
	if _, err := time.Parse("2006-01-02 15:04:05", trimmed[:19]); err != nil {
		t.Errorf("Timestamp does not parse as time: %v", err)
	}
} 