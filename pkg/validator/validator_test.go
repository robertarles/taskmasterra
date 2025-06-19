package validator

import (
	"strings"
	"testing"
)

func TestNewValidationResult(t *testing.T) {
	result := NewValidationResult()

	if len(result.Errors) != 0 {
		t.Errorf("Expected Errors to be empty, got %d items", len(result.Errors))
	}

	if len(result.Warnings) != 0 {
		t.Errorf("Expected Warnings to be empty, got %d items", len(result.Warnings))
	}

	if len(result.Info) != 0 {
		t.Errorf("Expected Info to be empty, got %d items", len(result.Info))
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()
	result.AddError(5, "Test error")

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Line != 5 {
		t.Errorf("Expected error line to be 5, got %d", result.Errors[0].Line)
	}

	if result.Errors[0].Message != "Test error" {
		t.Errorf("Expected error message to be 'Test error', got %s", result.Errors[0].Message)
	}

	if result.Errors[0].Level != LevelError {
		t.Errorf("Expected error level to be LevelError, got %v", result.Errors[0].Level)
	}
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := NewValidationResult()
	result.AddWarning(10, "Test warning")

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if result.Warnings[0].Line != 10 {
		t.Errorf("Expected warning line to be 10, got %d", result.Warnings[0].Line)
	}

	if result.Warnings[0].Message != "Test warning" {
		t.Errorf("Expected warning message to be 'Test warning', got %s", result.Warnings[0].Message)
	}

	if result.Warnings[0].Level != LevelWarning {
		t.Errorf("Expected warning level to be LevelWarning, got %v", result.Warnings[0].Level)
	}
}

func TestValidationResult_AddInfo(t *testing.T) {
	result := NewValidationResult()
	result.AddInfo(15, "Test info")

	if len(result.Info) != 1 {
		t.Errorf("Expected 1 info, got %d", len(result.Info))
	}

	if result.Info[0].Line != 15 {
		t.Errorf("Expected info line to be 15, got %d", result.Info[0].Line)
	}

	if result.Info[0].Message != "Test info" {
		t.Errorf("Expected info message to be 'Test info', got %s", result.Info[0].Message)
	}

	if result.Info[0].Level != LevelInfo {
		t.Errorf("Expected info level to be LevelInfo, got %v", result.Info[0].Level)
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := NewValidationResult()

	if result.HasErrors() {
		t.Error("Expected HasErrors to return false for empty result")
	}

	result.AddError(1, "Test error")
	if !result.HasErrors() {
		t.Error("Expected HasErrors to return true when errors exist")
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	result := NewValidationResult()

	if result.HasWarnings() {
		t.Error("Expected HasWarnings to return false for empty result")
	}

	result.AddWarning(1, "Test warning")
	if !result.HasWarnings() {
		t.Error("Expected HasWarnings to return true when warnings exist")
	}
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		hasError bool
	}{
		{
			name: "Valid file",
			content: `# Test TODO
- [ ] A1 !! Valid task with priority and effort
- [w] B2 Worked task
- [x] C3 Completed task
`,
			hasError: false,
		},
		{
			name: "Invalid task format",
			content: `# Test TODO
- [ Invalid task format
- [ ] Valid task
`,
			hasError: true,
		},
		{
			name: "Empty title",
			content: `# Test TODO
- [ ] 
- [w] Valid task
`,
			hasError: false, // This should be a warning, not an error
		},
		{
			name: "Unknown status",
			content: `# Test TODO
- [q] Unknown status task
- [ ] Valid task
`,
			hasError: false, // This should be a warning, not an error
		},
		{
			name: "No tasks",
			content: `# Test TODO
This is just a header with no tasks.
`,
			hasError: false, // This should be a warning, not an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFile(tt.content)

			if tt.hasError && !result.HasErrors() {
				t.Error("Expected validation to have errors")
			}

			if !tt.hasError && result.HasErrors() {
				t.Errorf("Expected validation to pass, but got %d errors", len(result.Errors))
			}
		})
	}
}

func TestFormatValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *ValidationResult
		expected []string
	}{
		{
			name:     "No issues",
			result:   NewValidationResult(),
			expected: []string{"✅ No issues found"},
		},
		{
			name: "With errors",
			result: func() *ValidationResult {
				r := NewValidationResult()
				r.AddError(1, "Test error 1")
				r.AddError(5, "Test error 2")
				return r
			}(),
			expected: []string{"❌ 2 errors:", "Line 1: Test error 1", "Line 5: Test error 2"},
		},
		{
			name: "With warnings",
			result: func() *ValidationResult {
				r := NewValidationResult()
				r.AddWarning(2, "Test warning 1")
				r.AddWarning(8, "Test warning 2")
				return r
			}(),
			expected: []string{"⚠️  2 warnings:", "Line 2: Test warning 1", "Line 8: Test warning 2"},
		},
		{
			name: "With info",
			result: func() *ValidationResult {
				r := NewValidationResult()
				r.AddInfo(3, "Test info 1")
				r.AddInfo(10, "Test info 2")
				return r
			}(),
			expected: []string{"ℹ️  2 suggestions:", "Line 3: Test info 1", "Line 10: Test info 2"},
		},
		{
			name: "Mixed issues",
			result: func() *ValidationResult {
				r := NewValidationResult()
				r.AddError(1, "Error")
				r.AddWarning(2, "Warning")
				r.AddInfo(3, "Info")
				return r
			}(),
			expected: []string{"❌ 1 errors:", "⚠️  1 warnings:", "ℹ️  1 suggestions:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatValidationResult(tt.result)

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but got: %s", expected, output)
				}
			}
		})
	}
}

func TestErrorLevel_String(t *testing.T) {
	tests := []struct {
		level    ErrorLevel
		expected string
	}{
		{LevelInfo, "INFO"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{ErrorLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("ErrorLevel.String() = %s, want %s", result, tt.expected)
			}
		})
	}
} 