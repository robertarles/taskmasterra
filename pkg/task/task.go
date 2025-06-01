package task

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robertarles/taskmasterra/v2/pkg/journal"
)

// Task represents a task item with its status and details
type Task struct {
	Line string
}

// IsCompleted checks if a task is marked as completed
func IsCompleted(line string) bool {
	if !IsTask(line) {
		return false
	}
	return regexp.MustCompile(`^\s*- \[[Xx]\]`).MatchString(line)
}

// IsActive checks if a task is marked as active (needs attention today)
func IsActive(line string) bool {
	if !IsTask(line) {
		return false
	}
	return regexp.MustCompile(`^\s*- \[.\] !! `).MatchString(line)
}

// IsTouched checks if a task has been touched/worked on
func IsTouched(line string) bool {
	if !IsTask(line) && !IsSubTask(line) {
		return false
	}
	return regexp.MustCompile(`(^- \[[BWX]\]|^\s+- \[[BWX]\])`).MatchString(line)
}

// IsTask checks if a line represents a task
func IsTask(line string) bool {
	return regexp.MustCompile(`^- \[`).MatchString(line)
}

// IsSubTask checks if a line represents a subtask
func IsSubTask(line string) bool {
	return regexp.MustCompile(`^[ \t]+- \[`).MatchString(line)
}

// IsTaskDetail checks if a line is an indented detail line
func IsTaskDetail(line string) bool {
	return regexp.MustCompile(`^[ \t]+- `).MatchString(line)
}

// ReplaceStatus replaces the task status marker
func ReplaceStatus(line string, oldMarker, newMarker rune) string {
	if regexp.MustCompile(`- \[` + string(oldMarker) + `\]`).MatchString(line) {
		return regexp.MustCompile(`- \[` + string(oldMarker) + `\]`).ReplaceAllString(line, "- [" + string(newMarker) + "]")
	}
	return line
}

// ConvertActiveToTouched converts active task status to touched status
func ConvertActiveToTouched(line string) string {
	line = ReplaceStatus(line, 'B', 'b')
	line = ReplaceStatus(line, 'W', 'w')
	line = ReplaceStatus(line, 'X', 'x')
	return line
}

// ProcessTasks processes a todo file, moving completed tasks to archive and touched tasks to journal
func ProcessTasks(filePath string) error {
	// Read the original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	jm := journal.NewManager(filePath)
	timestamp := journal.FormatTimestamp()

	var journalEntries, archiveEntries, updatedLines []string
	
	for i := 0; i < len(lines); {
		line := lines[i]
		nextLine := i + 1

		if IsTouched(line) || IsActive(line) {
			entry := fmt.Sprintf("%s %s", timestamp, line)
			journalEntries = append(journalEntries, entry)

			if !IsCompleted(line) {
				modifiedLine := ConvertActiveToTouched(line)
				updatedLines = append(updatedLines, modifiedLine)
			} else {
				archiveEntries = append(archiveEntries, line)
			}

			// Process child items
			for j := nextLine; j < len(lines); j++ {
				if IsTaskDetail(lines[j]) {
					journalEntries = append(journalEntries, lines[j])
					if !IsCompleted(line) {
						updatedLines = append(updatedLines, lines[j])
					}
					nextLine = j + 1
				} else {
					break
				}
			}
		} else if IsCompleted(line) {
			entry := fmt.Sprintf("%s %s", timestamp, line)
			archiveEntries = append(archiveEntries, entry)

			// Process child items
			for j := nextLine; j < len(lines); j++ {
				if IsTaskDetail(lines[j]) {
					archiveEntries = append(archiveEntries, lines[j])
					nextLine = j + 1
				} else {
					break
				}
			}
		} else {
			updatedLines = append(updatedLines, line)
		}

		i = nextLine
	}

	// Write to journal and archive
	if err := jm.WriteToJournal(journalEntries); err != nil {
		return fmt.Errorf("error writing to journal: %w", err)
	}

	if err := jm.WriteToArchive(archiveEntries); err != nil {
		return fmt.Errorf("error writing to archive: %w", err)
	}

	// Update original file
	if err := os.WriteFile(filePath, []byte(strings.Join(updatedLines, "\n")), 0644); err != nil {
		return fmt.Errorf("error updating original file: %w", err)
	}

	return nil
} 