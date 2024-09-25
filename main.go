package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isTaskForToday(line string) bool {
	// Check if the line starts with "- [" and contains " T "
	return strings.HasPrefix(line, "- [") && strings.Contains(line, " T ")
}

func addToReminders(task, listName string) error {
	// Create the AppleScript command to add the reminder
	appleScript := fmt.Sprintf(`
		tell application "Reminders"
			if not (exists list "%s") then
				make new list with properties {name:"%s"}
			end if
			set dueDate to current date
			set hours of dueDate to 16
			set minutes of dueDate to 0
			set seconds of dueDate to 0
			tell list "%s"
				make new reminder with properties {name:"%s", due date:dueDate}
			end tell
		end tell
	`, listName, listName, listName, task)

	// Execute the AppleScript using osascript
	cmd := exec.Command("osascript", "-e", appleScript)

	// Capture the standard output and error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("osascript error: %v: %s", err, stderr.String())
	}

	return nil
}

func main() {
	// Define the path to the markdown file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		return
	}
	filePath := filepath.Join(homeDir, "plaintext.robert/notes/disney/todo.disney.md")

	// Extract the file name without the extension to use as the Reminders list name
	listName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var tasksForToday []string

	for scanner.Scan() {
		line := scanner.Text()
		if isTaskForToday(line) {
			// Extract the actual task description by removing the markdown checkbox and the " T "
			taskDescription := strings.TrimSpace(line[strings.Index(line, "]")+2:])
			tasksForToday = append(tasksForToday, taskDescription)

			// Add the task to macOS Reminders with today's date at 4:00 PM
			if err := addToReminders(taskDescription, listName); err != nil {
				fmt.Printf("Failed to add task '%s' to Reminders: %v\n", taskDescription, err)
			} else {
				fmt.Printf("Added task '%s' to Reminders list '%s'.\n", taskDescription, listName)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Print tasks that are marked for today
	if len(tasksForToday) > 0 {
		fmt.Println("\nTasks to be worked on today:")
		for _, task := range tasksForToday {
			fmt.Println(task)
		}
	} else {
		fmt.Println("No tasks marked for today.")
	}
}
