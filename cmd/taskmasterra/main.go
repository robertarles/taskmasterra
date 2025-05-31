package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertarles/taskmasterra/v2/pkg/journal"
	"github.com/robertarles/taskmasterra/v2/pkg/reminder"
	"github.com/robertarles/taskmasterra/v2/pkg/task"
)

// Build information. Populated at build-time.
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

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

func recordKeep(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("error expanding file path: %w", err)
	}

	// Read the original file
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	jm := journal.NewManager(expandedPath)
	timestamp := journal.FormatTimestamp()

	var journalEntries, archiveEntries, updatedLines []string
	
	for i := 0; i < len(lines); {
		line := lines[i]
		nextLine := i + 1

		if task.IsTouched(line) || task.IsActive(line) {
			entry := fmt.Sprintf("%s %s", timestamp, line)
			journalEntries = append(journalEntries, entry)

			if !task.IsCompleted(line) {
				modifiedLine := task.ConvertActiveToTouched(line)
				updatedLines = append(updatedLines, modifiedLine)
			}else{
				archiveEntries = append(archiveEntries, line)
			}

			// Process child items
			for j := nextLine; j < len(lines); j++ {
				if task.IsTaskDetail(lines[j]) {
					journalEntries = append(journalEntries, lines[j])
					if !task.IsCompleted(line) {
						updatedLines = append(updatedLines, lines[j])
					}
					nextLine = j + 1
				} else {
					break
				}
			}
		} else if task.IsCompleted(line) {
			entry := fmt.Sprintf("%s %s", timestamp, line)
			archiveEntries = append(archiveEntries, entry)

			// Process child items
			for j := nextLine; j < len(lines); j++ {
				if task.IsTaskDetail(lines[j]) {
					archiveEntries = append(archiveEntries, lines[j])
					nextLine = j + 1
				} else {
					break
				}
			}
		} else {
			updatedLines = append(updatedLines, line)
		}

		i = nextLine
	}

	// Write to journal and archive
	if err := jm.WriteToJournal(journalEntries); err != nil {
		return fmt.Errorf("error writing to journal: %w", err)
	}

	if err := jm.WriteToArchive(archiveEntries); err != nil {
		return fmt.Errorf("error writing to archive: %w", err)
	}

	// Update original file
	if err := os.WriteFile(expandedPath, []byte(strings.Join(updatedLines, "\n")), 0644); err != nil {
		return fmt.Errorf("error updating original file: %w", err)
	}

	return nil
}

func updateCalendar(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("error expanding file path: %w", err)
	}

	baseFileName := filepath.Base(expandedPath)
	listName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))

	rs := reminder.NewService(listName)
	if err := rs.ClearList(); err != nil {
		return fmt.Errorf("failed to clear reminders list: %w", err)
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentTask string
	var notes []string

	for scanner.Scan() {
		line := scanner.Text()

		if task.IsTask(line) && !task.IsCompleted(line) {
			// If we have a previous task, add it with its notes
			if currentTask != "" {
				if err := rs.AddReminder(currentTask, task.IsActive(line), strings.Join(notes, "\n")); err != nil {
					return fmt.Errorf("failed to add reminder: %w", err)
				}
			}

			// Start new task
			closingBracketIndex := strings.Index(line, "]")
			if closingBracketIndex != -1 {
				currentTask = strings.TrimSpace(line[closingBracketIndex+1:])
				notes = nil
			}
		} else if currentTask != "" && task.IsTaskDetail(line) {
			// Collect notes for current task
			notes = append(notes, line)
		}
	}

	// Add the last task if there is one
	if currentTask != "" {
		if err := rs.AddReminder(currentTask, false, strings.Join(notes, "\n")); err != nil {
			return fmt.Errorf("failed to add reminder: %w", err)
		}
	}

	return nil
}

func printHelp() {
	fmt.Println("Usage: taskmasterra <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  updatecal     Update the calendar with today's tasks")
	fmt.Println("  recordkeep    Record tasks to journal and archive files")
	fmt.Println("  version       Show the version of taskmasterra")
	fmt.Println("  help          Show this help message")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
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

		if err := updateCalendar(*inputFilePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "recordkeep":
		recordKeepCmd := flag.NewFlagSet("recordkeep", flag.ExitOnError)
		inputFilePath := recordKeepCmd.String("i", "", "Path to the markdown input file")
		recordKeepCmd.Parse(os.Args[2:])

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for recordkeep command. Use -i to specify the path.")
			recordKeepCmd.Usage()
			return
		}

		if err := recordKeep(*inputFilePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("taskmasterra %s (commit: %s, built at: %s)\n", Version, Commit, BuildTime)

	default:
		printHelp()
	}
} 