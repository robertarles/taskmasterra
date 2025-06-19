package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskProcessing(t *testing.T) {
	// Test cases for task status checking
	tests := []struct {
		name     string
		line     string
		isTask   bool
		touched  bool
		active   bool
		complete bool
	}{
		{"Worked touched uppercase", "- [W] task", true, true, false, false},
		{"Worked not-touched lowercase", "- [w] task", true, false, false, false},
		{"Blocked touched uppercase", "- [B] task", true, true, false, false},
		{"Blocked not-touched lowercase", "- [b] task", true, false, false, false},
		{"Completed touched uppercase", "- [X] task", true, true, false, true},
		{"Completed not-touched lowercase", "- [x] task", true, false, false, true},
		{"Empty status", "- [ ] task", true, false, false, false},
		{"No status", "- task", false, false, false, false},
		{"Active task", "- [w] !! task", true, false, true, false},
		{"Subtask", "  - [W] subtask", false, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTask(tt.line); got != tt.isTask {
				t.Errorf("IsTask() = %v, want %v for %v", got, tt.isTask, tt.line)
			}
			if got := IsTouched(tt.line); got != tt.touched {
				t.Errorf("IsTouched() = %v, want %v for %v", got, tt.touched, tt.line)
			}
			if got := IsActive(tt.line); got != tt.active {
				t.Errorf("IsActive() = %v, want %v for %v", got, tt.active, tt.line)
			}
			if got := IsCompleted(tt.line); got != tt.complete {
				t.Errorf("IsCompleted() = %v, want %v for %v", got, tt.complete, tt.line)
			}
		})
	}
}

func TestTaskStatusConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Convert uppercase W", "- [W] task", "- [w] task"},
		{"Convert uppercase B", "- [B] task", "- [b] task"},
		{"Convert uppercase X", "- [X] task", "- [x] task"},
		{"No conversion needed", "- [w] task", "- [w] task"},
		{"No conversion for empty", "- [ ] task", "- [ ] task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertActiveToTouched(tt.input); got != tt.expected {
				t.Errorf("ConvertActiveToTouched() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func setupTestFiles(t *testing.T) (string, func()) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "taskmasterra-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test todo.md
	todoContent := `# "Test TODO"

## divorce Harness

- [W] 1 worked touched (LOWERCASE-W) (ARCHIVE-FAIL,JOURNAL-PASS )
  - 1 sub 1 (ARCHIVE-FAIL, JOURNAL-PASS)
- [B] 2 blocked touched, (LOWERCASE-B) (ARCHIVE-FAIL, JOURNAL-PASS)
- [w] 3 nada (ARCHIVE-FAIL,JOURNAL-FAIL)
- [x] 4 archive (ARCHIVE-PASS,JOURNAL-FAIL)
- [ ] 5 nada (ARCHIVE-FAIL,JOURNAL-FAIL)
- 6 nada just another item (ARCHIVE-FAIL,JOURNAL-FAIL)
  - 6 sub 1 nada of item (ARCHIVE-FAIL,JOURNAL-FAIL)
- 7 nada some item (ARCHIVE-FAIL,JOURNAL-FAIL)
  - 7 sub 1 nada (ARCHIVE-FAIL,JOURNAL-FAIL)
  - 7 sub 2 nada (ARCHIVE-FAIL,JOURNAL-FAIL)
- [X] 8 worked touched (ARCHIVE-PASS,JOURNAL-PASS)

## ACTIVE

- [W] 9 worked touched, four children {LOWERCASE-W) (ARCHIVE-FAIL,JOURNAL-PASS)
  - 9 sub 1 child, sub 2 should be removed (ARCHIVE-FAIL,JOURNAL-PASS)
  - [x] 9 sub 2 child completed (ARCHIVE-FAIL,JOURNAL-PASS)
  - [b] 9 sub 3 child blocked (ARCHIVE-FAIL,JOURNAL-PASS)
  - [ ] 9 sub 4 child no status (ARCHIVE-FAIL,JOURNAL-PASS)
  - [W] 9 sub 5 worked, should be journalled and lowercase W (ARCHIVE-FAIL,JOURNAL-PASS)
- [b] 10 blocked  (ARCHIVE-FAIL,JOURNAL-FAIL)`

	todoPath := filepath.Join(tmpDir, "todo.md")
	if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write todo.md: %v", err)
	}

	// Create empty journal and archive files
	journalPath := filepath.Join(tmpDir, "todo.xjournal.md")
	archivePath := filepath.Join(tmpDir, "todo.xarchive.md")
	for _, path := range []string{journalPath, archivePath} {
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create %s: %v", path, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestTaskFileProcessing(t *testing.T) {
	tmpDir, cleanup := setupTestFiles(t)
	defer cleanup()

	todoPath := filepath.Join(tmpDir, "todo.md")
	journalPath := filepath.Join(tmpDir, "todo.xjournal.md")
	archivePath := filepath.Join(tmpDir, "todo.xarchive.md")

	// Process the tasks (you'll need to implement this)
	if err := ProcessTasks(todoPath); err != nil {
		t.Fatalf("Failed to process tasks: %v", err)
	}

	// Read and verify the results
	journalContent, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatalf("Failed to read journal: %v", err)
	}
	archiveContent, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("Failed to read archive: %v", err)
	}
	updatedTodoContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated todo: %v", err)
	}

	// Verify journal entries
	journalLines := strings.Split(string(journalContent), "\n")
	for _, line := range journalLines {
		if strings.Contains(line, "JOURNAL-FAIL") {
			t.Errorf("Found incorrectly journaled line: %s", line)
		}
	}

	// Verify archive entries
	archiveLines := strings.Split(string(archiveContent), "\n")
	for _, line := range archiveLines {
		if !strings.Contains(line, "[x]") && !strings.Contains(line, "[X]") && line != "" {
			t.Errorf("Found non-completed task in archive: %s", line)
		}
	}

	// Verify updated todo file
	todoLines := strings.Split(string(updatedTodoContent), "\n")
	for _, line := range todoLines {
		if IsTask(line) && (strings.Contains(line, "[W]") || strings.Contains(line, "[B]") || strings.Contains(line, "[X]")) {
			t.Errorf("Found uppercase status in updated todo: %s", line)
		}
		if IsTask(line) && (strings.Contains(line, "[x]") || strings.Contains(line, "[X]")) {
			t.Errorf("Found completed task in updated todo: %s", line)
		}
	}
}

func TestIsTaskDetailAndIsSubTask(t *testing.T) {
	tests := []struct {
		line         string
		isDetail     bool
		isSubTask    bool
	}{
		{"  - [ ] detail line", true, true},
		{"\t- [x] subtask with tab", true, true},
		{"    - [w] deeply indented", true, true},
		{"- [ ] not a detail", false, false},
		{"- [x] not a subtask", false, false},
		{"random text", false, false},
		{"", false, false},
		{"- [ ]", false, false},
		{"  - detail without brackets", true, false},
	}

	for _, tt := range tests {
		if got := IsTaskDetail(tt.line); got != tt.isDetail {
			t.Errorf("IsTaskDetail(%q) = %v, want %v", tt.line, got, tt.isDetail)
		}
		if got := IsSubTask(tt.line); got != tt.isSubTask {
			t.Errorf("IsSubTask(%q) = %v, want %v", tt.line, got, tt.isSubTask)
		}
	}
} 