// Package task provides functionality for parsing, processing, and managing markdown-based todo tasks.
// It supports task status detection, priority parsing, effort estimation, and task lifecycle management.
package task

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/robertarles/taskmasterra/v2/pkg/journal"
	"github.com/robertarles/taskmasterra/v2/pkg/utils"
)

// Precompiled regex patterns for better performance
var (
	completedTaskRegex = regexp.MustCompile(`^\s*- \[[Xx]\]`)
	activeTaskRegex    = regexp.MustCompile(`^\s*- \[.\] !! `)
	touchedTaskRegex   = regexp.MustCompile(`(^- \[[BWX]\]|^\s+- \[[BWX]\])`)
	taskRegex          = regexp.MustCompile(`^- \[`)
	subTaskRegex       = regexp.MustCompile(`^[ \t]+- \[`)
	taskDetailRegex    = regexp.MustCompile(`^[ \t]+- `)
)

// Task represents a task item with its status and details.
// This is a simple wrapper around the raw line content.
type Task struct {
	Line string
}

// IsCompleted checks if a task is marked as completed.
// Returns true if the line represents a task with [x] or [X] status.
func IsCompleted(line string) bool {
	if !IsTask(line) {
		return false
	}
	return completedTaskRegex.MatchString(line)
}

// IsActive checks if a task is marked as active (needs attention today).
// A task is active if it has the !! marker immediately after the status bracket.
// Returns false if there are multiple !! markers in the line.
func IsActive(line string) bool {
	if !IsTask(line) {
		return false
	}
	// Check for the correct prefix
	if !activeTaskRegex.MatchString(line) {
		return false
	}
	// Ensure there are no other !! in the rest of the line
	idxs := activeTaskRegex.FindStringIndex(line)
	if idxs == nil {
		return false
	}
	rest := line[idxs[1]:]
	return !strings.Contains(rest, "!!")
}

// IsTouched checks if a task has been touched/worked on.
// A task is touched if it has uppercase status markers [B], [W], or [X].
func IsTouched(line string) bool {
	if !IsTask(line) && !IsSubTask(line) {
		return false
	}
	return touchedTaskRegex.MatchString(line)
}

// IsTask checks if a line represents a task.
// Returns true if the line starts with "- [" (task list item).
func IsTask(line string) bool {
	return taskRegex.MatchString(line)
}

// IsSubTask checks if a line represents a subtask.
// Returns true if the line is indented and starts with "- [".
func IsSubTask(line string) bool {
	return subTaskRegex.MatchString(line)
}

// IsTaskDetail checks if a line is an indented detail line.
// Returns true if the line is indented and starts with "- " (not a task).
func IsTaskDetail(line string) bool {
	return taskDetailRegex.MatchString(line)
}

// ReplaceStatus replaces the task status marker in a line.
// Creates a new regex pattern for the specific marker and performs the replacement.
func ReplaceStatus(line string, oldMarker, newMarker rune) string {
	// Create regex pattern for the specific marker
	oldPattern := regexp.MustCompile(`- \[` + string(oldMarker) + `\]`)
	if oldPattern.MatchString(line) {
		return oldPattern.ReplaceAllString(line, "- [" + string(newMarker) + "]")
	}
	return line
}

// ConvertActiveToTouched converts active task status to touched status.
// Converts uppercase status markers (B, W, X) to lowercase (b, w, x).
func ConvertActiveToTouched(line string) string {
	line = ReplaceStatus(line, 'B', 'b')
	line = ReplaceStatus(line, 'W', 'w')
	line = ReplaceStatus(line, 'X', 'x')
	return line
}

// ProcessTasks processes a todo file, moving completed tasks to archive and touched tasks to journal.
// This is the main workflow function that:
// - Reads the todo file
// - Processes each task line
// - Moves completed tasks to archive with timestamps
// - Moves touched/active tasks to journal with timestamps
// - Updates the original file with converted status markers
func ProcessTasks(filePath string) error {
	// Read the original file
	content, err := utils.ReadFileContent(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	lines := strings.Split(content, "\n")
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
				// Archive parent line with timestamp
				archiveEntries = append(archiveEntries, fmt.Sprintf("%s %s", timestamp, line))
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
			// Archive parent line with timestamp
			archiveEntries = append(archiveEntries, fmt.Sprintf("%s %s", timestamp, line))

			// Process child items
			for j := nextLine; j < len(lines); j++ {
				if IsTaskDetail(lines[j]) {
					// Archive child detail line with timestamp
					archiveEntries = append(archiveEntries, fmt.Sprintf("%s %s", timestamp, lines[j]))
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
		return fmt.Errorf("failed to write journal entries for file '%s': %w", filePath, err)
	}

	if err := jm.WriteToArchive(archiveEntries); err != nil {
		return fmt.Errorf("failed to write archive entries for file '%s': %w", filePath, err)
	}

	// Update original file
	if err := utils.WriteFileContent(filePath, strings.Join(updatedLines, "\n")); err != nil {
		return fmt.Errorf("failed to update original file '%s': %w", filePath, err)
	}

	return nil
} 