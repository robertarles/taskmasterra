package reminder

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

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

	cmd := exec.Command("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("osascript error: %w: %s", err, stderr.String())
	}

	return nil
}

// AddReminder adds a new reminder to the list
func (s *Service) AddReminder(task string, withDueDate bool, note string) error {
	// Escape special characters in task and note
	escapedTask := escapeAppleScriptString(task)
	escapedNote := escapeAppleScriptString(note)
	escapedListName := escapeAppleScriptString(s.ListName)

	properties := fmt.Sprintf(`{name:"%s"`, escapedTask)
	if note != "" {
		properties += fmt.Sprintf(`, body:"%s"`, escapedNote)
	}
	if withDueDate {
		properties += `, due date:dueDate`
	}
	properties += `}`

	dueDateSetup := ""
	if withDueDate {
		dueDateSetup = `
			set dueDate to current date
			set hours of dueDate to 16
			set minutes of dueDate to 0
			set seconds of dueDate to 0`
	}

	script := fmt.Sprintf(`
		tell application "Reminders"
			if not (exists list "%s") then
				make new list with properties {name:"%s"}
			end if%s
			tell list "%s"
				make new reminder with properties %s
			end tell
		end tell
	`, escapedListName, escapedListName, dueDateSetup, escapedListName, properties)

	cmd := exec.Command("osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("osascript error: %w: %s", err, stderr.String())
	}

	return nil
} 