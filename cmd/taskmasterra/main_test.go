package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertarles/taskmasterra/v2/pkg/reminder"
)

// Save the original exec.Command
var execCommand = exec.Command

// Mock exec.Command for testing
func fakeExecCommand(command string, args ...string) *exec.Cmd {
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

	// Mock different command behaviors
	switch args[0] {
	case "osascript":
		if len(args) < 3 || args[1] != "-e" {
			fmt.Fprintf(os.Stderr, "invalid osascript arguments")
			os.Exit(1)
		}
		script := args[2]
		if strings.Contains(script, "error-test") {
			if strings.Contains(script, "delete reminders") {
				fmt.Fprintf(os.Stderr, "osascript error: failed to clear list")
				os.Exit(1)
			}
			if strings.Contains(script, "make new reminder") {
				fmt.Fprintf(os.Stderr, "osascript error: failed to add reminder")
				os.Exit(1)
			}
		}
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

func TestExpandPath(t *testing.T) {
	// Save original environment and restore after test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	homeDir := "/home/testuser"
	os.Setenv("HOME", homeDir)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "Absolute path",
			path:    "/absolute/path/file.txt",
			want:    "/absolute/path/file.txt",
			wantErr: false,
		},
		{
			name:    "Path with tilde",
			path:    "~/documents/file.txt",
			want:    filepath.Join(homeDir, "documents/file.txt"),
			wantErr: false,
		},
		{
			name:    "Path with $HOME",
			path:    "$HOME/documents/file.txt",
			want:    filepath.Join(homeDir, "documents/file.txt"),
			wantErr: false,
		},
		{
			name:    "Relative path",
			path:    "relative/path/file.txt",
			want:    "relative/path/file.txt",
			wantErr: false,
		},
		{
			name:    "Empty path",
			path:    "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid environment variable",
			path:    "$INVALID_VAR/file.txt",
			want:    "$INVALID_VAR/file.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordKeep(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "recordkeep-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		todoContent string
		wantErr     bool
	}{
		{
			name: "Normal operation",
			todoContent: `# Test TODO

- [W] Task 1 (touched)
  - Sub task 1
- [w] Task 2 (not touched)
- [x] Task 3 (completed)
  - Sub task 3
- [ ] Task 4 (no status)
`,
			wantErr: false,
		},
		{
			name:        "Empty file",
			todoContent: "",
			wantErr:    false,
		},
		{
			name: "Invalid task format",
			todoContent: `# Test TODO
- Invalid task format
- [x] Valid task
`,
			wantErr: false,
		},
		{
			name: "Only completed tasks",
			todoContent: `# Test TODO
- [x] Task 1
- [x] Task 2
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoPath := filepath.Join(tmpDir, "todo.md")
			if err := os.WriteFile(todoPath, []byte(tt.todoContent), 0644); err != nil {
				t.Fatalf("Failed to write todo file: %v", err)
			}

			err := recordKeep(todoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("recordKeep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that the journal and archive files were created
			if _, err := os.Stat(filepath.Join(tmpDir, "todo.xjournal.md")); err != nil && !tt.wantErr {
				t.Errorf("Journal file was not created: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tmpDir, "todo.xarchive.md")); err != nil && !tt.wantErr {
				t.Errorf("Archive file was not created: %v", err)
			}
		})
	}
}

func TestUpdateCalendar(t *testing.T) {
	// Save original execCommand and restore after test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()
	execCommand = fakeExecCommand

	// Override reminder.ExecCommand as well
	origReminderExecCommand := reminder.ExecCommand
	defer func() { reminder.ExecCommand = origReminderExecCommand }()
	reminder.ExecCommand = fakeExecCommand

	tmpDir, err := os.MkdirTemp("", "updatecal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		todoContent string
		wantErr     bool
	}{
		{
			name: "Normal operation",
			todoContent: `# Test TODO
- [ ] !! Task 1 (active)
- [w] Task 2 (not active)
- [x] Task 3 (completed)
- [ ] Task 4 (no status)
`,
			wantErr: false,
		},
		{
			name: "No active tasks",
			todoContent: `# Test TODO
- [ ] Task 1
- [w] Task 2
- [x] Task 3
`,
			wantErr: false,
		},
		{
			name: "Error case",
			todoContent: `# Test TODO
- [ ] !! error-test task
`,
			wantErr: true,
		},
		{
			name:        "Empty file",
			todoContent: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoPath := filepath.Join(tmpDir, "todo.md")
			if err := os.WriteFile(todoPath, []byte(tt.todoContent), 0644); err != nil {
				t.Fatalf("Failed to write todo file: %v", err)
			}

			err := updateCalendar(todoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateCalendar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{
		"Usage:",
		"recordkeep",
		"updatereminders",
		"version",
		"help",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("printHelp() output does not contain %q", expected)
		}
	}
}

func TestMain_NoArgs(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"taskmasterra"}
	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Error("main() with no args should print usage")
	}
}

func TestMain_InvalidCommand(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"taskmasterra", "invalid"}
	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Error("main() with invalid command should print help")
	}
}

func TestMain_Commands(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save original execCommand and restore after test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()
	execCommand = fakeExecCommand

	// Override reminder.ExecCommand as well
	origReminderExecCommand := reminder.ExecCommand
	defer func() { reminder.ExecCommand = origReminderExecCommand }()
	reminder.ExecCommand = fakeExecCommand

	tmpDir, err := os.MkdirTemp("", "main-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	todoPath := filepath.Join(tmpDir, "todo.md")
	todoContent := `# Test TODO
- [ ] !! Task 1
- [x] Task 2
`
	if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "recordkeep command",
			args:    []string{"taskmasterra", "recordkeep", "-i", todoPath},
			wantErr: false,
		},
		{
			name:    "updatereminders command",
			args:    []string{"taskmasterra", "updatereminders", "-i", todoPath},
			wantErr: false,
		},
		{
			name:    "updatecal command (alias)",
			args:    []string{"taskmasterra", "updatecal", "-i", todoPath},
			wantErr: false,
		},
		{
			name:    "recordkeep without input",
			args:    []string{"taskmasterra", "recordkeep"},
			wantErr: true,
		},
		{
			name:    "updatereminders without input",
			args:    []string{"taskmasterra", "updatereminders"},
			wantErr: true,
		},
		{
			name:    "version command",
			args:    []string{"taskmasterra", "version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Args = tt.args
			main()

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("Failed to read captured output: %v", err)
			}

			output := buf.String()
			if tt.wantErr && !strings.Contains(output, "Error:") {
				t.Error("Expected error output")
			}
		})
	}
} 