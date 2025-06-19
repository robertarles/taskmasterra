// Package main provides the taskmasterra CLI application for managing markdown-based todo lists.
// It supports journaling, archiving, validation, statistics, and macOS Reminders integration.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertarles/taskmasterra/v2/pkg/config"
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

// expandPath expands environment variables and home directory shortcuts in file paths.
// Supports ~ for home directory and $HOME environment variable.
func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory for path expansion: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	} else if strings.HasPrefix(path, "$HOME") {
		homeDir, exists := os.LookupEnv("HOME")
		if !exists {
			return "", fmt.Errorf("environment variable HOME not set, cannot expand path: %s", path)
		}
		path = filepath.Join(homeDir, path[len("$HOME"):])
	}
	return path, nil
}

// recordKeep processes a todo file, moving completed tasks to archive and touched tasks to journal.
// It validates the file first and continues processing even if validation issues are found.
func recordKeep(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to expand file path '%s': %w", filePath, err)
	}

	// Read the original file
	content, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", expandedPath, err)
	}

	// Validate the file and log warnings/errors
	result := validator.ValidateFile(content)
	if result.HasErrors() || result.HasWarnings() {
		fmt.Fprintf(os.Stderr, "⚠️  Validation issues found in %s:\n", expandedPath)
		fmt.Fprint(os.Stderr, validator.FormatValidationResult(result))
		if result.HasErrors() {
			fmt.Fprintf(os.Stderr, "⚠️  Continuing with recordkeep despite validation errors\n")
		}
	}

	// Process the tasks
	if err := task.ProcessTasks(expandedPath); err != nil {
		return fmt.Errorf("failed to process tasks in file '%s': %w", expandedPath, err)
	}

	fmt.Printf("✅ Successfully processed tasks in %s\n", expandedPath)
	return nil
}

// updateCalendar syncs active tasks from a todo file to macOS Reminders.app.
// Only tasks marked with !! (active marker) are added to reminders.
func updateCalendar(filePath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to expand file path '%s': %w", filePath, err)
	}

	// Read the file content for validation
	content, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", expandedPath, err)
	}

	// Validate the file and log warnings/errors
	result := validator.ValidateFile(content)
	if result.HasErrors() || result.HasWarnings() {
		fmt.Fprintf(os.Stderr, "⚠️  Validation issues found in %s:\n", expandedPath)
		fmt.Fprint(os.Stderr, validator.FormatValidationResult(result))
		if result.HasErrors() {
			fmt.Fprintf(os.Stderr, "⚠️  Continuing with updatereminders despite validation errors\n")
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create reminder service
	service := reminder.NewService(cfg.ReminderListName)

	// Clear existing reminders
	if err := service.ClearList(); err != nil {
		return fmt.Errorf("failed to clear reminder list '%s': %w", cfg.ReminderListName, err)
	}

	// Read file content for processing
	fileContent, err := utils.ReadFileContent(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s' for reminder processing: %w", expandedPath, err)
	}

	lines := strings.Split(fileContent, "\n")
	activeCount := 0

	for i, line := range lines {
		lineNum := i + 1
		if task.IsActive(line) {
			activeCount++
			taskInfo := task.ParseTaskInfo(line)
			if taskInfo == nil {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Could not parse task info on line %d: %s\n", lineNum, line)
				continue
			}

			// Create reminder with due date if specified
			withDueDate := taskInfo.Priority == task.PriorityCritical || taskInfo.Priority == task.PriorityHigh
			note := fmt.Sprintf("Priority: %s", taskInfo.Priority.String())
			if taskInfo.Effort > 0 {
				note += fmt.Sprintf(", Effort: %d", taskInfo.Effort)
			}

			if err := service.AddReminder(taskInfo.Title, withDueDate, note); err != nil {
				return fmt.Errorf("failed to add reminder for task on line %d: %w", lineNum, err)
			}
		}
	}

	if activeCount == 0 {
		fmt.Printf("ℹ️  No active tasks found in %s\n", expandedPath)
	} else {
		fmt.Printf("✅ Successfully added %d active tasks to reminder list '%s'\n", activeCount, cfg.ReminderListName)
	}

	return nil
}

// printHelp displays comprehensive help information for the taskmasterra CLI.
func printHelp() {
	fmt.Println("Taskmasterra - Markdown-based task management with journaling and Reminders integration")
	fmt.Println()
	fmt.Println("Usage: taskmasterra <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  recordkeep      Process tasks: archive completed, journal touched tasks")
	fmt.Println("                  Example: taskmasterra recordkeep -i todo.md")
	fmt.Println()
	fmt.Println("  updatereminders Sync active tasks (marked with !!) to macOS Reminders.app")
	fmt.Println("                  Example: taskmasterra updatereminders -i todo.md")
	fmt.Println()
	fmt.Println("  stats           Generate comprehensive task statistics report")
	fmt.Println("                  Example: taskmasterra stats -i todo.md -o report.md")
	fmt.Println()
	fmt.Println("  validate        Check todo file format and get improvement suggestions")
	fmt.Println("                  Example: taskmasterra validate -i todo.md")
	fmt.Println()
	fmt.Println("  config          Manage application configuration")
	fmt.Println("                  Examples:")
	fmt.Println("                    taskmasterra config -init    # Initialize default config")
	fmt.Println("                    taskmasterra config -show    # Show current config")
	fmt.Println()
	fmt.Println("  version         Show version information")
	fmt.Println("  help            Show this help message")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/robertarles/taskmasterra")
}

// generateStats creates a comprehensive statistics report from a todo file.
func generateStats(filePath string, outputPath string) error {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to expand file path '%s': %w", filePath, err)
	}

	// Analyze the file
	statsData, err := stats.AnalyzeFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to analyze file '%s': %w", expandedPath, err)
	}

	// Generate report
	report := stats.GenerateReport(statsData)

	// Save report
	if err := stats.SaveReport(report, outputPath); err != nil {
		return fmt.Errorf("failed to save report to '%s': %w", outputPath, err)
	}

	fmt.Printf("✅ Statistics report generated and saved to: %s\n", outputPath)
	return nil
}

// validateFile validates a todo file and displays any issues found.
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

// manageConfig handles configuration file operations (initialize, show).
func manageConfig(configPath string, show bool, init bool) error {
	if init {
		cfg := config.DefaultConfig()
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("default configuration is invalid: %w", err)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory for config initialization: %w", err)
		}
		defaultConfigPath := filepath.Join(homeDir, ".taskmasterra", "config.json")
		if err := config.SaveConfig(cfg, defaultConfigPath); err != nil {
			return fmt.Errorf("failed to create configuration file at '%s': %w", defaultConfigPath, err)
		}
		fmt.Printf("✅ Configuration file created at: %s\n", defaultConfigPath)
		return nil
	}

	if show {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration from '%s': %w", configPath, err)
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		configJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
		}
		fmt.Println(string(configJSON))
		return nil
	}

	return fmt.Errorf("no action specified for config command")
}

// suggestCommand returns the closest matching command for a given input.
func suggestCommand(input string, commands []string) string {
	input = strings.ToLower(input)
	minDist := 100
	closest := ""
	for _, cmd := range commands {
		dist := levenshtein(input, cmd)
		if dist < minDist {
			minDist = dist
			closest = cmd
		}
	}
	if minDist <= 3 && closest != "" {
		return closest
	}
	return ""
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

func main() {
	validCommands := []string{"updatereminders", "updatecal", "recordkeep", "stats", "validate", "config", "version", "help"}

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "updatereminders", "updatecal":
		updateCalCmd := flag.NewFlagSet("updatereminders", flag.ExitOnError)
		inputFilePath := updateCalCmd.String("i", "", "Path to the markdown input file")
		updateCalCmd.Usage = func() {
			fmt.Println("\nUsage: taskmasterra updatereminders -i <inputfile>")
			fmt.Println("Sync active tasks (marked with !!) to macOS Reminders.app")
			updateCalCmd.PrintDefaults()
		}
		if err := updateCalCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			updateCalCmd.Usage()
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
		recordKeepCmd.Usage = func() {
			fmt.Println("\nUsage: taskmasterra recordkeep -i <inputfile>")
			fmt.Println("Process tasks: archive completed, journal touched tasks")
			recordKeepCmd.PrintDefaults()
		}
		if err := recordKeepCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			recordKeepCmd.Usage()
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
		statsCmd.Usage = func() {
			fmt.Println("\nUsage: taskmasterra stats -i <inputfile> -o <outputfile>")
			fmt.Println("Generate comprehensive task statistics report")
			statsCmd.PrintDefaults()
		}
		if err := statsCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			statsCmd.Usage()
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
		validateCmd.Usage = func() {
			fmt.Println("\nUsage: taskmasterra validate -i <inputfile>")
			fmt.Println("Check todo file format and get improvement suggestions")
			validateCmd.PrintDefaults()
		}
		if err := validateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			validateCmd.Usage()
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
		configCmd.Usage = func() {
			fmt.Println("\nUsage: taskmasterra config -init | -show [-c <configfile>]")
			fmt.Println("Manage application configuration")
			configCmd.PrintDefaults()
		}
		if err := configCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Error parsing flags: %v\n", err)
			configCmd.Usage()
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

	case "help":
		printHelp()
		return

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'.\n", command)
		suggestion := suggestCommand(command, validCommands)
		if suggestion != "" {
			fmt.Fprintf(os.Stderr, "Did you mean '%s'?\n", suggestion)
		}
		printHelp()
		os.Exit(1)
	}
} 