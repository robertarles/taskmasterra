package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Line    int
	Message string
	Level   ErrorLevel
}

// ErrorLevel represents the severity of a validation error
type ErrorLevel int

const (
	LevelInfo ErrorLevel = iota
	LevelWarning
	LevelError
)

// String returns the string representation of error level
func (l ErrorLevel) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ValidationResult contains validation results
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
	Info     []ValidationError
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Info:     []ValidationError{},
	}
}

// AddError adds an error to the validation result
func (r *ValidationResult) AddError(line int, message string) {
	r.Errors = append(r.Errors, ValidationError{Line: line, Message: message, Level: LevelError})
}

// AddWarning adds a warning to the validation result
func (r *ValidationResult) AddWarning(line int, message string) {
	r.Warnings = append(r.Warnings, ValidationError{Line: line, Message: message, Level: LevelWarning})
}

// AddInfo adds an info message to the validation result
func (r *ValidationResult) AddInfo(line int, message string) {
	r.Info = append(r.Info, ValidationError{Line: line, Message: message, Level: LevelInfo})
}

// HasErrors returns true if there are any errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ValidateFile validates a markdown task file
func ValidateFile(content string) *ValidationResult {
	result := NewValidationResult()
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1
		validateLine(line, lineNum, result)
	}

	// Global validations
	validateGlobal(content, result)

	return result
}

// validateLine validates a single line
func validateLine(line string, lineNum int, result *ValidationResult) {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return
	}

	// Check for valid task format
	if strings.HasPrefix(strings.TrimSpace(line), "- [") {
		validateTaskLine(line, lineNum, result)
	} else if strings.HasPrefix(strings.TrimSpace(line), "#") {
		validateHeaderLine(line, lineNum, result)
	} else if strings.HasPrefix(strings.TrimSpace(line), "- ") {
		validateDetailLine(line, lineNum, result)
	}
}

// validateTaskLine validates a task line
func validateTaskLine(line string, lineNum int, result *ValidationResult) {
	// Check for valid task status format
	taskRe := regexp.MustCompile(`^\s*- \[([^\]]+)\]\s*(.*)`)
	matches := taskRe.FindStringSubmatch(line)
	if len(matches) < 3 {
		result.AddError(lineNum, "Invalid task format")
		return
	}

	status := matches[1]
	title := strings.TrimSpace(matches[2])

	// Validate status
	validStatuses := []string{" ", "x", "X", "w", "W", "b", "B"}
	isValidStatus := false
	for _, valid := range validStatuses {
		if status == valid {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		result.AddWarning(lineNum, fmt.Sprintf("Unknown status '%s'", status))
	}

	// Check for empty title
	if title == "" {
		result.AddWarning(lineNum, "Task has no title")
	}

	// Check for active marker
	if strings.Contains(line, "!!") {
		if !strings.Contains(status, " ") && !strings.Contains(status, "w") && !strings.Contains(status, "W") {
			result.AddWarning(lineNum, "Active task (!!) should have empty or 'w' status")
		}
	}

	// Check for priority and effort format
	priorityRe := regexp.MustCompile(`\b([A-Z])(\d+)\b`)
	priorityMatches := priorityRe.FindStringSubmatch(line)
	if len(priorityMatches) >= 3 {
		priority := priorityMatches[1]
		effort := priorityMatches[2]

		// Validate priority letter
		validPriorities := []string{"A", "B", "C", "D"}
		isValidPriority := false
		for _, valid := range validPriorities {
			if priority == valid {
				isValidPriority = true
				break
			}
		}

		if !isValidPriority {
			result.AddWarning(lineNum, fmt.Sprintf("Unknown priority '%s'", priority))
		}

		// Validate effort number (should be fibonacci-like)
		fibonacciNumbers := []string{"1", "2", "3", "5", "8", "13", "21", "34", "55", "89"}
		isValidEffort := false
		for _, valid := range fibonacciNumbers {
			if effort == valid {
				isValidEffort = true
				break
			}
		}

		if !isValidEffort {
			result.AddInfo(lineNum, fmt.Sprintf("Effort '%s' is not a standard fibonacci number", effort))
		}
	}
}

// validateHeaderLine validates a header line
func validateHeaderLine(line string, lineNum int, result *ValidationResult) {
	// Check for proper header format
	headerRe := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	matches := headerRe.FindStringSubmatch(line)
	if len(matches) < 3 {
		result.AddWarning(lineNum, "Invalid header format")
		return
	}

	level := len(matches[1])
	title := strings.TrimSpace(matches[2])

	if title == "" {
		result.AddWarning(lineNum, "Header has no title")
	}

	if level > 3 {
		result.AddInfo(lineNum, "Consider using fewer header levels for better organization")
	}
}

// validateDetailLine validates a detail line
func validateDetailLine(line string, lineNum int, result *ValidationResult) {
	// Check for proper indentation
	if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
		result.AddWarning(lineNum, "Detail line should be indented")
	}

	// Check for empty content
	content := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "  "), "\t"))
	if content == "" {
		result.AddWarning(lineNum, "Detail line has no content")
	}
}

// validateGlobal performs global validations
func validateGlobal(content string, result *ValidationResult) {
	lines := strings.Split(content, "\n")

	// Check for file structure
	hasHeader := false
	taskCount := 0
	completedCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			hasHeader = true
		}
		if strings.HasPrefix(trimmed, "- [") {
			taskCount++
			if strings.Contains(trimmed, "[x]") || strings.Contains(trimmed, "[X]") {
				completedCount++
			}
		}
	}

	if !hasHeader {
		result.AddInfo(1, "Consider adding a header to organize your tasks")
	}

	if taskCount == 0 {
		result.AddWarning(1, "No tasks found in file")
	}

	if taskCount > 0 && completedCount == taskCount {
		result.AddInfo(1, "All tasks are completed - consider archiving or creating new tasks")
	}
}

// FormatValidationResult formats validation results for display
func FormatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if len(result.Errors) == 0 && len(result.Warnings) == 0 && len(result.Info) == 0 {
		output.WriteString("✅ No issues found\n")
		return output.String()
	}

	// Print errors
	if len(result.Errors) > 0 {
		output.WriteString(fmt.Sprintf("❌ %d errors:\n", len(result.Errors)))
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", err.Line, err.Message))
		}
		output.WriteString("\n")
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		output.WriteString(fmt.Sprintf("⚠️  %d warnings:\n", len(result.Warnings)))
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", warning.Line, warning.Message))
		}
		output.WriteString("\n")
	}

	// Print info
	if len(result.Info) > 0 {
		output.WriteString(fmt.Sprintf("ℹ️  %d suggestions:\n", len(result.Info)))
		for _, info := range result.Info {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", info.Line, info.Message))
		}
	}

	return output.String()
} 