package reminder

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// helperCommand returns a fake exec.Cmd for testing
func helperCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess is not a real test. It's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Get the command and arguments that were passed to exec.Command
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		os.Exit(1)
	}

	// Mock the osascript command
	if args[0] == "osascript" {
		os.Exit(0)
	}
	os.Exit(1)
}

// TestEscapeAppleScriptString tests the escaping of special characters
func TestEscapeAppleScriptString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String with quotes",
			input:    `Hello "World"`,
			expected: `Hello \"World\"`,
		},
		{
			name:     "String with backslashes",
			input:    `Hello\World`,
			expected: `Hello\\World`,
		},
		{
			name:     "String with quotes and backslashes",
			input:    `Hello \"World\"`,
			expected: `Hello \\\"World\\\"`,
		},
		{
			name:     "Complex string with special characters",
			input:    `Task with "quotes", \backslashes\, and "nested \"quotes\""`,
			expected: `Task with \"quotes\", \\backslashes\\, and \"nested \\\"quotes\\\"\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAppleScriptString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeAppleScriptString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestClearList tests the ClearList function
func TestClearList(t *testing.T) {
	// Save the original exec.Command and restore it after the test
	originalExecCommand := ExecCommand
	defer func() { ExecCommand = originalExecCommand }()

	tests := []struct {
		name        string
		listName    string
		expectError bool
	}{
		{
			name:        "Simple list name",
			listName:    "Todo",
			expectError: false,
		},
		{
			name:        "List name with special characters",
			listName:    `Todo "List" with \special\ chars`,
			expectError: false,
		},
		{
			name:        "Empty list name",
			listName:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock command
			ExecCommand = func(command string, args ...string) *exec.Cmd {
				if command != "osascript" {
					t.Errorf("Expected command 'osascript', got %q", command)
				}
				if len(args) != 2 || args[0] != "-e" {
					t.Errorf("Expected args ['-e', '<script>'], got %v", args)
				}
				script := args[1]
				if !strings.Contains(script, "tell application \"Reminders\"") {
					t.Errorf("Script doesn't contain expected Reminders application command")
				}
				if !strings.Contains(script, escapeAppleScriptString(tt.listName)) {
					t.Errorf("Script doesn't contain escaped list name")
				}
				return helperCommand(command, args...)
			}

			service := NewService(tt.listName)
			err := service.ClearList()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestAddReminder tests the AddReminder function
func TestAddReminder(t *testing.T) {
	// Save the original exec.Command and restore it after the test
	originalExecCommand := ExecCommand
	defer func() { ExecCommand = originalExecCommand }()

	tests := []struct {
		name        string
		listName    string
		task        string
		withDueDate bool
		note        string
		expectError bool
	}{
		{
			name:        "Simple task",
			listName:    "Todo",
			task:        "Buy groceries",
			withDueDate: false,
			note:        "",
			expectError: false,
		},
		{
			name:        "Task with due date",
			listName:    "Todo",
			task:        "Pay bills",
			withDueDate: true,
			note:        "",
			expectError: false,
		},
		{
			name:        "Task with note",
			listName:    "Todo",
			task:        "Call John",
			withDueDate: false,
			note:        "Discuss project timeline",
			expectError: false,
		},
		{
			name:        "Complex task with special characters",
			listName:    `Work "Tasks"`,
			task:        `Review "code" with \special\ chars`,
			withDueDate: true,
			note:        `Contains "quotes" and \backslashes\`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock command
			ExecCommand = func(command string, args ...string) *exec.Cmd {
				if command != "osascript" {
					t.Errorf("Expected command 'osascript', got %q", command)
				}
				if len(args) != 2 || args[0] != "-e" {
					t.Errorf("Expected args ['-e', '<script>'], got %v", args)
				}
				script := args[1]

				// Verify script contains required elements
				if !strings.Contains(script, "tell application \"Reminders\"") {
					t.Errorf("Script doesn't contain expected Reminders application command")
				}
				if !strings.Contains(script, escapeAppleScriptString(tt.listName)) {
					t.Errorf("Script doesn't contain escaped list name")
				}
				if !strings.Contains(script, escapeAppleScriptString(tt.task)) {
					t.Errorf("Script doesn't contain escaped task")
				}
				if tt.note != "" && !strings.Contains(script, escapeAppleScriptString(tt.note)) {
					t.Errorf("Script doesn't contain escaped note")
				}
				if tt.withDueDate && !strings.Contains(script, "due date:dueDate") {
					t.Errorf("Script doesn't contain due date setup")
				}

				return helperCommand(command, args...)
			}

			service := NewService(tt.listName)
			err := service.AddReminder(tt.task, tt.withDueDate, tt.note)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
} 