package task

import (
	"regexp"
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
	if !IsTask(line) {
		return false
	}
	return regexp.MustCompile(`^\s*- \[[BWX]\]`).MatchString(line)
}

// IsTask checks if a line represents a task
func IsTask(line string) bool {
	return regexp.MustCompile(`^- \[`).MatchString(line)
}

// IsSubTask checks if a line represents a subtask
func IsSubTask(line string) bool {
	return regexp.MustCompile(`^\s+- \[`).MatchString(line)
}

// IsTaskDetail checks if a line is an indented detail line
func IsTaskDetail(line string) bool {
	return regexp.MustCompile(`^\s+- `).MatchString(line)
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