package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertarles/taskmasterra/v2/pkg/config"
	"github.com/robertarles/taskmasterra/v2/pkg/journal"
	"github.com/robertarles/taskmasterra/v2/pkg/reminder"
	"github.com/robertarles/taskmasterra/v2/pkg/stats"
	"github.com/robertarles/taskmasterra/v2/pkg/task"
	"github.com/robertarles/taskmasterra/v2/pkg/utils"
	"github.com/robertarles/taskmasterra/v2/pkg/validator"
)

// Build information. Populated at build-time.
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
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
	content, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", expandedPath, err)
	}

	// Validate the file and log warnings/errors
	result := validator.ValidateFile(content)
	if result.HasErrors() || result.HasWarnings() {
		fmt.Fprintf(os.Stderr, "⚠️  Validation issues found in %s:\n", expandedPath)
		fmt.Fprint(os.Stderr, validator.FormatValidationResult(result))
		if result.HasErrors() {
			fmt.Fprintf(os.Stderr, "⚠️  Continuing with recordkeep despite validation errors...\n")
		}
	}

	lines := strings.Split(content, "\n")
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
				entry := fmt.Sprintf("%s %s", timestamp, line)
				archiveEntries = append(archiveEntries, entry)
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
	if err := utils.WriteFileContent(expandedPath, strings.Join(updatedLines, "\n")); err != nil {
		return fmt.Errorf("error updating original file '%s': %w", expandedPath, err)
	}

	return nil
}

func updateCalendar(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("error expanding file path: %w", err)
	}

	// Read the file content for validation
	content, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", expandedPath, err)
	}

	// Validate the file and log warnings/errors
	result := validator.ValidateFile(content)
	if result.HasErrors() || result.HasWarnings() {
		fmt.Fprintf(os.Stderr, "⚠️  Validation issues found in %s:\n", expandedPath)
		fmt.Fprint(os.Stderr, validator.FormatValidationResult(result))
		if result.HasErrors() {
			fmt.Fprintf(os.Stderr, "⚠️  Continuing with updatereminders despite validation errors...\n")
		}
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
	var currentLine string

	for scanner.Scan() {
		line := scanner.Text()

		if task.IsTask(line) && !task.IsCompleted(line) {
			// If we have a previous task, add it with its notes
			if currentTask != "" {
				if err := rs.AddReminder(currentTask, task.IsActive(currentLine), strings.Join(notes, "\n")); err != nil {
					return fmt.Errorf("failed to add reminder: %w", err)
				}
			}

			// Start new task
			closingBracketIndex := strings.Index(line, "]")
			if closingBracketIndex != -1 {
				currentTask = strings.TrimSpace(line[closingBracketIndex+1:])
				currentLine = line
				notes = nil
			}
		} else if currentTask != "" && task.IsTaskDetail(line) {
			// Collect notes for current task
			notes = append(notes, line)
		}
	}

	// Add the last task if there is one
	if currentTask != "" {
		if err := rs.AddReminder(currentTask, task.IsActive(currentLine), strings.Join(notes, "\n")); err != nil {
			return fmt.Errorf("failed to add reminder: %w", err)
		}
	}

	return nil
}

func printHelp() {
	fmt.Println("Usage: taskmasterra <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  updatereminders Update the calendar with today's tasks")
	fmt.Println("  recordkeep    Record tasks to journal and archive files")
	fmt.Println("  stats         Generate task statistics report")
	fmt.Println("  validate      Validate task file format")
	fmt.Println("  config        Manage configuration")
	fmt.Println("  version       Show the version of taskmasterra")
	fmt.Println("  help          Show this help message")
}

func generateStats(filePath string, outputPath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("error expanding file path: %w", err)
	}

	statsData, err := stats.AnalyzeFile(expandedPath)
	if err != nil {
		return fmt.Errorf("error analyzing file: %w", err)
	}

	report := stats.GenerateReport(statsData)

	if outputPath != "" {
		expandedOutputPath, err := expandPath(outputPath)
		if err != nil {
			return fmt.Errorf("error expanding output path: %w", err)
		}
		if err := stats.SaveReport(report, expandedOutputPath); err != nil {
			return fmt.Errorf("error saving report: %w", err)
		}
		fmt.Printf("Statistics report saved to: %s\n", expandedOutputPath)
	} else {
		fmt.Println(report)
	}

	return nil
}

func validateFile(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("error expanding file path: %w", err)
	}

	content, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", expandedPath, err)
	}

	result := validator.ValidateFile(content)
	fmt.Print(validator.FormatValidationResult(result))

	if result.HasErrors() {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

func manageConfig(configPath string, show bool, init bool) error {
	if init {
		cfg := config.DefaultConfig()
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid default config: %w", err)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}
		defaultConfigPath := filepath.Join(homeDir, ".taskmasterra", "config.json")
		if err := config.SaveConfig(cfg, defaultConfigPath); err != nil {
			return fmt.Errorf("error creating config file: %w", err)
		}
		fmt.Printf("Configuration file created at: %s\n", defaultConfigPath)
		return nil
	}

	if show {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling config: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "updatereminders", "updatecal":
		updateCalCmd := flag.NewFlagSet("updatereminders", flag.ExitOnError)
		inputFilePath := updateCalCmd.String("i", "", "Path to the markdown input file")
		if err := updateCalCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for updatereminders command. Use -i to specify the path.")
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
		if err := recordKeepCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for recordkeep command. Use -i to specify the path.")
			recordKeepCmd.Usage()
			return
		}

		if err := recordKeep(*inputFilePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "stats":
		statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)
		inputFilePath := statsCmd.String("i", "", "Path to the markdown input file")
		outputFilePath := statsCmd.String("o", "", "Path to the output statistics report file")
		if err := statsCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for stats command. Use -i to specify the path.")
			statsCmd.Usage()
			return
		}

		if err := generateStats(*inputFilePath, *outputFilePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "validate":
		validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
		inputFilePath := validateCmd.String("i", "", "Path to the markdown input file")
		if err := validateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *inputFilePath == "" {
			fmt.Println("Error: Input file path is required for validate command. Use -i to specify the path.")
			validateCmd.Usage()
			return
		}

		if err := validateFile(*inputFilePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "config":
		configCmd := flag.NewFlagSet("config", flag.ExitOnError)
		configFilePath := configCmd.String("c", "", "Path to the configuration file")
		show := configCmd.Bool("show", false, "Show the configuration")
		init := configCmd.Bool("init", false, "Initialize a new configuration")
		if err := configCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *init && *show {
			fmt.Println("Error: Cannot use both -init and -show flags together")
			configCmd.Usage()
			return
		}

		if err := manageConfig(*configFilePath, *show, *init); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Println(getVersionString())
		return

	default:
		printHelp()
	}
} 