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

func isTaskForToday(line string) bool {
	// Determine if a task is to be worked on today based on your criteria
	return isTouchedTask(line)
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

	// Read the markdown file into lines
	fileContent, err := ioutil.ReadFile(expandedFilePath)
	if err != nil {
		fmt.Println("Error reading markdown file:", err)
		return
	}
	lines := strings.Split(string(fileContent), "\n")

	// Get the current time in UTC
	currentTime := time.Now().UTC()
	timestamp := currentTime.Format("[2006-01-02 15:04:05 UTC]") // Format as [YYYY-MM-DD HH:MM:SS UTC]

	var xjournalEntries []string
	var xarchiveEntries []string
	linesToRemove := make(map[int]bool)

	for idx, line := range lines {
		if isTouchedTask(line) {
			// Prepare entry for xjournal file with timestamp
			entry := fmt.Sprintf("%s %s", timestamp, line)
			xjournalEntries = append(xjournalEntries, entry)
			fmt.Printf("Recording touched task to journal: %s\n", entry)
			// Do not modify the markdown file for touched tasks
		}
		if isCompletedTask(line) {
			// Prepare entry for xarchive file with timestamp
			entry := fmt.Sprintf("%s %s", timestamp, line)
			xarchiveEntries = append(xarchiveEntries, entry)
			fmt.Printf("Recording completed task to archive: %s\n", entry)
			// Mark the line for removal after successful addition to archive
			linesToRemove[idx] = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading markdown file:", err)
		return
	}

	// Write to xarchive file first
	if len(xarchiveEntries) > 0 {
		newXarchiveContent := []byte(strings.Join(xarchiveEntries, "\n") + "\n")
		newXarchiveContent = append(newXarchiveContent, xarchiveContent...)
		err = ioutil.WriteFile(xarchivePath, newXarchiveContent, 0644)
		if err != nil {
			fmt.Println("Error writing to xarchive file:", err)
			fmt.Println("Aborting removal of completed tasks from markdown file due to error.")
			return
		}

		// Remove the lines that were successfully added to the archive
		var updatedLines []string
		for idx, line := range lines {
			if !linesToRemove[idx] {
				updatedLines = append(updatedLines, line)
			} else {
				fmt.Printf("Removing completed task from markdown file: %s\n", line)
			}
		}
		// Write back the updated markdown file
		err = ioutil.WriteFile(expandedFilePath, []byte(strings.Join(updatedLines, "\n")), 0644)
		if err != nil {
			fmt.Println("Error writing updated markdown file:", err)
			return
		}
	}

	// Write to xjournal file
	if len(xjournalEntries) > 0 {
		newXjournalContent := []byte(strings.Join(xjournalEntries, "\n") + "\n")
		newXjournalContent = append(newXjournalContent, xjournalContent...)
		err = ioutil.WriteFile(xjournalPath, newXjournalContent, 0644)
		if err != nil {
			fmt.Println("Error writing to xjournal file:", err)
			return
		}
	}

	fmt.Println("Tasks have been recorded successfully.")
}

func updateCalendar(filePath string) {
	// Existing updateCalendar function code
	// ... (omitted for brevity)
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
