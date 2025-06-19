package reminder

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecCommand is a variable that holds the exec.Command function.
// This allows us to replace it with a mock during testing.
var ExecCommand = exec.Command

// Service handles interactions with macOS Reminders
type Service struct {
	ListName string
}

// NewService creates a new reminder service
func NewService(listName string) *Service {
	return &Service{
		ListName: listName,
	}
}

// escapeAppleScriptString escapes special characters in a string for AppleScript
func escapeAppleScriptString(s string) string {
	// Replace backslashes first to avoid double escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Replace quotes with escaped quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// ClearList removes all reminders from the specified list
func (s *Service) ClearList() error {
	script := fmt.Sprintf(`
		tell application "Reminders"
			if exists list "%s" then
				tell list "%s"
					delete reminders
				end tell
			end if
		end tell
	`, escapeAppleScriptString(s.ListName), escapeAppleScriptString(s.ListName))

	cmd := ExecCommand("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clear reminder list '%s' via AppleScript: %w (stderr: %s)", s.ListName, err, stderr.String())
	}

	return nil
}

// AddReminder adds a new reminder to the list
func (s *Service) AddReminder(task string, withDueDate bool, note string) error {
	escapedTask := escapeAppleScriptString(task)
	escapedNote := escapeAppleScriptString(note)
	
	var script string
	if withDueDate {
		script = fmt.Sprintf(`
			tell application "Reminders"
				if exists list "%s" then
					tell list "%s"
						make new reminder with properties {name:"%s", body:"%s", due date:(current date)}
					end tell
				else
					error "List '%s' does not exist"
				end if
			end tell
		`, escapeAppleScriptString(s.ListName), escapeAppleScriptString(s.ListName), escapedTask, escapedNote, s.ListName)
	} else {
		script = fmt.Sprintf(`
			tell application "Reminders"
				if exists list "%s" then
					tell list "%s"
						make new reminder with properties {name:"%s", body:"%s"}
					end tell
				else
					error "List '%s' does not exist"
				end if
			end tell
		`, escapeAppleScriptString(s.ListName), escapeAppleScriptString(s.ListName), escapedTask, escapedNote, s.ListName)
	}

	cmd := ExecCommand("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add reminder '%s' to list '%s' via AppleScript: %w (stderr: %s)", task, s.ListName, err, stderr.String())
	}

	return nil
} 