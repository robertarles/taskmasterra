// Package validator provides functionality for validating markdown-based todo files.
// It checks task syntax, priority/effort formats, active marker positioning, and provides
// helpful suggestions for improving task organization and formatting.
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Precompiled regex patterns for better performance
var (
	taskLineRegex      = regexp.MustCompile(`^\s*- \[([^\]]+)\]\s*(.*)`)
	activeMarkerRegex  = regexp.MustCompile(`^\s*- \[[^\]]+\] !! `)
	priorityEffortRegex = regexp.MustCompile(`\b([A-Z])(\d+)\b`)
	headerRegex        = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
)

// ValidationError represents a validation error with line number, message, and severity level.
type ValidationError struct {
	Line    int
	Message string
	Level   ErrorLevel
}

// ErrorLevel represents the severity of a validation error.
// INFO = suggestions, WARNING = potential issues, ERROR = must fix.
type ErrorLevel int

const (
	LevelInfo ErrorLevel = iota
	LevelWarning
	LevelError
)

// String returns the string representation of error level.
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

// ValidationResult contains validation results organized by severity level.
// Errors must be fixed, warnings should be addressed, and info provides suggestions.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
	Info     []ValidationError
}

// NewValidationResult creates a new validation result with empty slices.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Info:     []ValidationError{},
	}
}

// AddError adds an error to the validation result.
// Errors indicate issues that must be fixed for proper functionality.
func (r *ValidationResult) AddError(line int, message string) {
	r.Errors = append(r.Errors, ValidationError{Line: line, Message: message, Level: LevelError})
}

// AddWarning adds a warning to the validation result.
// Warnings indicate potential issues that should be addressed.
func (r *ValidationResult) AddWarning(line int, message string) {
	r.Warnings = append(r.Warnings, ValidationError{Line: line, Message: message, Level: LevelWarning})
}

// AddInfo adds an info message to the validation result.
// Info messages provide suggestions for improvement.
func (r *ValidationResult) AddInfo(line int, message string) {
	r.Info = append(r.Info, ValidationError{Line: line, Message: message, Level: LevelInfo})
}

// HasErrors returns true if there are any errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ValidateFile validates a markdown task file and returns validation results.
// This is the main entry point for file validation. It processes each line and
// performs both line-specific and global validations.
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

// validateLine validates a single line based on its type.
// Routes to appropriate validation functions based on line content.
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

// validateTaskLine validates a task line for proper format and content.
// Checks status validity, active marker positioning, priority/effort format, and more.
func validateTaskLine(line string, lineNum int, result *ValidationResult) {
	// Check for valid task status format
	matches := taskLineRegex.FindStringSubmatch(line)
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

	// Check for active marker position
	if strings.Contains(line, "!!") {
		// Find the expected position for !! (immediately after status bracket)
		if activeMarkerRegex.MatchString(line) {
			// Ensure there are no other !! in the rest of the line
			idxs := activeMarkerRegex.FindStringIndex(line)
			if idxs != nil {
				rest := line[idxs[1]:]
				if strings.Contains(rest, "!!") {
					result.AddError(lineNum, "Multiple active markers (!!) are not allowed")
				}
			}
		} else {
			result.AddError(lineNum, "Active marker (!!) must come immediately after the status bracket and before any priority/effort markers")
		}
		if !strings.Contains(status, " ") && !strings.Contains(status, "w") && !strings.Contains(status, "W") {
			result.AddWarning(lineNum, "Active task (!!) should have empty or 'w' status")
		}
	}

	// Check for priority and effort format
	priorityMatches := priorityEffortRegex.FindStringSubmatch(line)
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

// validateHeaderLine validates a header line for proper markdown format.
// Checks header level, title presence, and provides organization suggestions.
func validateHeaderLine(line string, lineNum int, result *ValidationResult) {
	// Check for proper header format
	matches := headerRegex.FindStringSubmatch(line)
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

// validateDetailLine validates a detail line for proper indentation and content.
// Checks that detail lines are properly indented and have content.
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

// validateGlobal performs global validations across the entire file content.
// Checks for overall file structure, task presence, and provides general suggestions.
func validateGlobal(content string, result *ValidationResult) {
	lines := strings.Split(content, "\n")
	
	// Check for tasks
	hasTasks := false
	allCompleted := true
	hasHeaders := false
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmedLine, "- [") {
			hasTasks = true
			if !strings.Contains(trimmedLine, "[x]") && !strings.Contains(trimmedLine, "[X]") {
				allCompleted = false
			}
		} else if strings.HasPrefix(trimmedLine, "#") {
			hasHeaders = true
		}
	}
	
	// Add global suggestions
	if !hasTasks {
		result.AddWarning(1, "No tasks found in file")
	}
	
	if !hasHeaders {
		result.AddInfo(1, "Consider adding a header to organize your tasks")
	}
	
	if hasTasks && allCompleted {
		result.AddInfo(1, "All tasks are completed - consider archiving or creating new tasks")
	}
}

// FormatValidationResult formats validation results for display.
// Returns a user-friendly string representation of all validation issues found.
func FormatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if len(result.Errors) == 0 && len(result.Warnings) == 0 && len(result.Info) == 0 {
		output.WriteString("✅ No issues found\n")
		return output.String()
	}

	// Format errors
	if len(result.Errors) > 0 {
		output.WriteString(fmt.Sprintf("❌ %d errors:\n", len(result.Errors)))
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", err.Line, err.Message))
		}
	}

	// Format warnings
	if len(result.Warnings) > 0 {
		output.WriteString(fmt.Sprintf("⚠️  %d warnings:\n", len(result.Warnings)))
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", warning.Line, warning.Message))
		}
	}

	// Format info messages
	if len(result.Info) > 0 {
		output.WriteString(fmt.Sprintf("ℹ️  %d suggestions:\n", len(result.Info)))
		for _, info := range result.Info {
			output.WriteString(fmt.Sprintf("  Line %d: %s\n", info.Line, info.Message))
		}
	}

	return output.String()
} 