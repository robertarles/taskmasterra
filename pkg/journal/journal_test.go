package journal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		wantJournal  string
		wantArchive  string
		wantOriginal string
	}{
		{
			name:         "Simple file path",
			filePath:     "/path/to/todo.md",
			wantJournal:  "/path/to/todo.xjournal.md",
			wantArchive:  "/path/to/todo.xarchive.md",
			wantOriginal: "/path/to/todo.md",
		},
		{
			name:         "File path with spaces",
			filePath:     "/path/to/my todo.md",
			wantJournal:  "/path/to/my todo.xjournal.md",
			wantArchive:  "/path/to/my todo.xarchive.md",
			wantOriginal: "/path/to/my todo.md",
		},
		{
			name:         "File path with extension in name",
			filePath:     "/path/to/todo.2024.md",
			wantJournal:  "/path/to/todo.2024.xjournal.md",
			wantArchive:  "/path/to/todo.2024.xarchive.md",
			wantOriginal: "/path/to/todo.2024.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.filePath)

			if manager.JournalPath != tt.wantJournal {
				t.Errorf("NewManager().JournalPath = %v, want %v", manager.JournalPath, tt.wantJournal)
			}
			if manager.ArchivePath != tt.wantArchive {
				t.Errorf("NewManager().ArchivePath = %v, want %v", manager.ArchivePath, tt.wantArchive)
			}
			if manager.OriginalPath != tt.wantOriginal {
				t.Errorf("NewManager().OriginalPath = %v, want %v", manager.OriginalPath, tt.wantOriginal)
			}
		})
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

func TestFormatTimestamp(t *testing.T) {
	timestamp := FormatTimestamp()
	pattern := `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} UTC\]$`
	matched, err := regexp.MatchString(pattern, timestamp)
	if err != nil {
		t.Fatalf("Failed to match timestamp pattern: %v", err)
	}
	if !matched {
		t.Errorf("FormatTimestamp() = %v, want format [YYYY-MM-DD HH:MM:SS UTC]", timestamp)
	}
} 