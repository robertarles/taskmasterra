package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func isCompletedTask(line string) bool {
	// Check if the line starts with "- [√]", "- [x]", or "- [X]"
	return strings.HasPrefix(line, "- [√]") || strings.HasPrefix(line, "- [x]") || strings.HasPrefix(line, "- [X]")
}

/**
 * active can be either active(.) or active and touched(:)
 **/
func isActiveTask(line string) bool {
	// Check if the line starts with "- ["
	if !strings.HasPrefix(line, "- [") {
		return false
	}

	// Find the position of the closing bracket ']'
	closingBracketIndex := strings.Index(line, "]")
	if closingBracketIndex == -1 {
		return false
	}

	// Ensure there's enough length for the expected pattern
	if len(line) <= closingBracketIndex+3 {
		return false
	}

	// Check for space after closing bracket
	if line[closingBracketIndex+1] != ' ' {
		return false
	}

	// Check if the next character is '.' or ':'
	marker := line[closingBracketIndex+2]
	if marker != '.' && marker != ':' {
		return false
	}

	// Check for space after the marker
	if line[closingBracketIndex+3] != ' ' {
		return false
	}

	return true
}

/**
 * active AND touched is a :
 * this does not cover active (.) not yet touched (:)
 **/
func isTouchedTask(line string) bool {
	// Check if the line starts with "- ["
	if !strings.HasPrefix(line, "- [") {
		return false
	}

	// Find the position of the closing bracket ']'
	closingBracketIndex := strings.Index(line, "]")
	if closingBracketIndex == -1 {
		return false
	}

	// Ensure there's enough length for the expected pattern
	if len(line) <= closingBracketIndex+3 {
		return false
	}

	// Check for space after closing bracket
	if line[closingBracketIndex+1] != ' ' {
		return false
	}

	// Check if the next character is ':'
	marker := line[closingBracketIndex+2]
	if marker != ':' {
		return false
	}

	// Check for space after the marker
	if line[closingBracketIndex+3] != ' ' {
		return false
	}

	return true
}

func isTaskForToday(line string) bool {
	// Determine if a task is to be worked on today based on your criteria
	// we'll consider tasks marked with an active task
	return isActiveTask(line)
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	} else if strings.HasPrefix(path, "$HOME") {
		homeDir, exists := os.LookupEnv("HOME")
		if !exists {
			return "", fmt.Errorf("environment variable HOME not set")
		}
		path = filepath.Join(homeDir, path[len("$HOME"):])
	}
	return path, nil
}

func clearRemindersList(listName string) error {
	// AppleScript to delete all reminders in the specified list
	appleScript := fmt.Sprintf(`
        tell application "Reminders"
            if exists list "%s" then
                tell list "%s"
                    delete reminders
                end tell
            end if
        end tell
    `, listName, listName)

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

func recordKeep(filePath string) {
	// Expand ~ and $HOME in filePath
	expandedFilePath, err := expandPath(filePath)
	if err != nil {
		fmt.Println("Error expanding file path:", err)
		return
	}

	// Prepare the output file paths
	baseFileName := filepath.Base(expandedFilePath)                          // e.g., "todo.disney.md"
	baseName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName)) // e.g., "todo.disney"
	dirPath := filepath.Dir(expandedFilePath)
	xjournalPath := filepath.Join(dirPath, baseName+".xjournal.md")
	xarchivePath := filepath.Join(dirPath, baseName+".xarchive.md")

	// Read existing content of xjournal and xarchive files
	var xjournalContent []byte
	if _, err := os.Stat(xjournalPath); err == nil {
		xjournalContent, err = ioutil.ReadFile(xjournalPath)
		if err != nil {
			fmt.Println("Error reading xjournal file:", err)
			return
		}
	}

	var xarchiveContent []byte
	if _, err := os.Stat(xarchivePath); err == nil {
		xarchiveContent, err = ioutil.ReadFile(xarchivePath)
		if err != nil {
			fmt.Println("Error reading xarchive file:", err)
			return
		}
	}

	// Read the original markdown file content
	originalContentBytes, err := ioutil.ReadFile(expandedFilePath)
	if err != nil {
		fmt.Println("Error reading markdown file:", err)
		return
	}
	originalLines := strings.Split(string(originalContentBytes), "\n")

	// Prepare variables for processing
	var xjournalEntries []string
	var xarchiveEntries []string
	var updatedLines []string

	// Get the current time in UTC
	currentTime := time.Now().UTC()
	timestamp := currentTime.Format("[2006-01-02 15:04:05 UTC]") // Format as [YYYY-MM-DD HH:MM:SS UTC]

	// Process each line
	for _, line := range originalLines {
		modifiedLine := ""
		// if "touched", journal item
		if isTouchedTask(line) {
			// Prepare entry for xjournal file with timestamp
			entry := fmt.Sprintf("%s %s", timestamp, line)
			xjournalEntries = append(xjournalEntries, entry)
			fmt.Printf("Recording touched task to journal: %s\n", entry)
			// Modify the line, replacing ':' with '.'
			modifiedLine = replaceMarker(line, ':', '.')
		}
		// if completed, archive it and remove from current, but do not add to updated file
		if isCompletedTask(line) {
			// Prepare entry for xarchive file with timestamp
			entry := fmt.Sprintf("%s %s", timestamp, line)
			xarchiveEntries = append(xarchiveEntries, entry)
			fmt.Printf("Recording completed task to archive and removing from markdown: %s\n", entry)
			// Do not add the line to updatedLines
		} else {
			if len(modifiedLine) > 0 {
				updatedLines = append(updatedLines, modifiedLine)
			} else {
				updatedLines = append(updatedLines, line)
			}
		}
	}

	// Write to the xjournal file
	if len(xjournalEntries) > 0 {
		newXjournalContent := []byte(strings.Join(xjournalEntries, "\n") + "\n")
		newXjournalContent = append(newXjournalContent, xjournalContent...)
		err = ioutil.WriteFile(xjournalPath, newXjournalContent, 0644)
		if err != nil {
			fmt.Println("Error writing to xjournal file:", err)
			return
		}
	}

	// Write to the xarchive file
	if len(xarchiveEntries) > 0 {
		newXarchiveContent := []byte(strings.Join(xarchiveEntries, "\n") + "\n")
		newXarchiveContent = append(newXarchiveContent, xarchiveContent...)
		err = ioutil.WriteFile(xarchivePath, newXarchiveContent, 0644)
		if err != nil {
			fmt.Println("Error writing to xarchive file:", err)
			return
		}

		fmt.Println("Removed archived tasks from the markdown file.")
	}

	// Write the updated lines back to the original markdown file
	err = ioutil.WriteFile(expandedFilePath, []byte(strings.Join(updatedLines, "\n")), 0644)
	if err != nil {
		fmt.Println("Error writing to markdown file:", err)
		return
	}

	fmt.Println("Updated markdown file with modified tasks.")
	fmt.Println("Tasks have been recorded successfully.")
}

func replaceMarker(line string, oldMarker, newMarker rune) string {
	// Find the position of the closing bracket ']'
	closingBracketIndex := strings.Index(line, "]")
	if closingBracketIndex == -1 {
		return line // Return the line unmodified if no closing bracket
	}

	// Ensure there's enough length for the expected pattern
	if len(line) <= closingBracketIndex+3 {
		return line
	}

	// Check that the marker is oldMarker
	if rune(line[closingBracketIndex+2]) == oldMarker {
		// Build a new line with the marker replaced
		newLine := line[:closingBracketIndex+2] + string(newMarker) + line[closingBracketIndex+3:]
		return newLine
	}
	return line // Return unmodified if the marker is not oldMarker
}

func isTask(line string) bool {
	// Check if the line starts with "- ["
	return strings.HasPrefix(line, "- [")
}

func updateCalendar(filePath string) {
	expandedFilePath, err := expandPath(filePath)
	if err != nil {
		fmt.Println("Error expanding file path:", err)
		return
	}

	baseFileName := filepath.Base(expandedFilePath)
	listName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))

	err = clearRemindersList(listName)
	if err != nil {
		fmt.Printf("Failed to clear Reminders list '%s': %v\n", listName, err)
		return
	}
	fmt.Printf("Cleared Reminders list '%s'.\n", listName)

	file, err := os.Open(expandedFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var tasksForToday []string
	var tasksWithoutDueDate []string

	for scanner.Scan() {
		line := scanner.Text()
		if isTask(line) && !isCompletedTask(line) {
			closingBracketIndex := strings.Index(line, "]")
			if closingBracketIndex != -1 {
				taskDescription := strings.TrimSpace(line[closingBracketIndex+1:])
				withDueDate := isTaskForToday(line)

				// Check for child notes directly below the task
				var note string
				if scanner.Scan() {
					nextLine := scanner.Text()
					if strings.HasPrefix(nextLine, "  -") { // Indented bullet point for a note
						note = strings.TrimSpace(nextLine[2:])
					} else {
						// Move scanner back by one line if not a note
						scanner = bufio.NewScanner(file)
					}
				}

				// Add the task to macOS Reminders with the note if found
				if err := addToReminders(taskDescription, listName, withDueDate, note); err != nil {
					fmt.Printf("Failed to add task '%s' to Reminders: %v\n", taskDescription, err)
				} else {
					if withDueDate {
						tasksForToday = append(tasksForToday, taskDescription)
						fmt.Printf("Added task '%s' to Reminders list '%s' with due date.\n", taskDescription, listName)
					} else {
						tasksWithoutDueDate = append(tasksWithoutDueDate, taskDescription)
						fmt.Printf("Added task '%s' to Reminders list '%s' without due date.\n", taskDescription, listName)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	if len(tasksForToday) > 0 {
		fmt.Println("\nTasks with due date:")
		for _, task := range tasksForToday {
			fmt.Println(task)
		}
	} else {
		fmt.Println("No tasks with due date.")
	}

	if len(tasksWithoutDueDate) > 0 {
		fmt.Println("\nTasks without due date:")
		for _, task := range tasksWithoutDueDate {
			fmt.Println(task)
		}
	}
}

func addToReminders(task, listName string, withDueDate bool, note string) error {
	properties := fmt.Sprintf(`{name:"%s"`, task)
	if note != "" {
		properties += fmt.Sprintf(`, body:"%s"`, note)
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

	appleScript := fmt.Sprintf(`
        tell application "Reminders"
            if not (exists list "%s") then
                make new list with properties {name:"%s"}
            end if%s
            tell list "%s"
                make new reminder with properties %s
            end tell
        end tell
    `, listName, listName, dueDateSetup, listName, properties)

	cmd := exec.Command("osascript", "-e", appleScript)

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
	if len(os.Args) < 2 {
		fmt.Println("Error: Command is required.")
		fmt.Println("Usage: taskmasterra <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  updatecal     Update the calendar with today's tasks")
		fmt.Println("  recordkeep    Record tasks to journal and archive files")
		return
	}

	command := os.Args[1]

	switch command {
	case "updatecal":
		updateCalCmd := flag.NewFlagSet("updatecal", flag.ExitOnError)
		inputFilePath := updateCalCmd.String("i", "", "Path to the markdown input file")
		updateCalCmd.Parse(os.Args[2:])

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for updatecal command. Use -i to specify the path.")
			updateCalCmd.Usage()
			return
		}

		updateCalendar(*inputFilePath)

	case "recordkeep":
		recordKeepCmd := flag.NewFlagSet("recordkeep", flag.ExitOnError)
		inputFilePath := recordKeepCmd.String("i", "", "Path to the markdown input file")
		recordKeepCmd.Parse(os.Args[2:])

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for recordkeep command. Use -i to specify the path.")
			recordKeepCmd.Usage()
			return
		}

		recordKeep(*inputFilePath)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: taskmasterra <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  updatecal     Update the calendar with today's tasks")
		fmt.Println("  recordkeep    Record tasks to journal and archive files")
	}
}
