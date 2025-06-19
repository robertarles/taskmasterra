package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertarles/taskmasterra/v2/pkg/task"
)

// TaskStats contains statistics about tasks
type TaskStats struct {
	TotalTasks     int
	CompletedTasks int
	ActiveTasks    int
	BlockedTasks   int
	WorkedTasks    int
	PriorityStats  map[string]int
	EffortStats    map[int]int
	Date           time.Time
}

// NewTaskStats creates a new TaskStats instance
func NewTaskStats() *TaskStats {
	return &TaskStats{
		PriorityStats: make(map[string]int),
		EffortStats:   make(map[int]int),
		Date:          time.Now(),
	}
}

// AnalyzeFile analyzes a markdown file and returns task statistics
func AnalyzeFile(filePath string) (*TaskStats, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	stats := NewTaskStats()
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if !task.IsTask(line) {
			continue
		}

		stats.TotalTasks++

		// Count by status
		if task.IsCompleted(line) {
			stats.CompletedTasks++
		} else if task.IsActive(line) {
			stats.ActiveTasks++
		} else {
			// Check for specific status markers
			if strings.Contains(line, "[b]") || strings.Contains(line, "[B]") {
				stats.BlockedTasks++
			} else if strings.Contains(line, "[w]") || strings.Contains(line, "[W]") {
				stats.WorkedTasks++
			}
		}

		// Count by priority
		taskInfo := task.ParseTaskInfo(line)
		if taskInfo != nil {
			priority := taskInfo.Priority.String()
			stats.PriorityStats[priority]++

			if taskInfo.Effort > 0 {
				stats.EffortStats[taskInfo.Effort]++
			}
		}
	}

	return stats, nil
}

// GenerateReport generates a formatted report from task statistics
func GenerateReport(stats *TaskStats) string {
	var report strings.Builder

	report.WriteString("# Task Statistics Report\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", stats.Date.Format("2006-01-02 15:04:05")))

	// Overall statistics
	report.WriteString("## Overall Statistics\n")
	report.WriteString(fmt.Sprintf("- Total Tasks: %d\n", stats.TotalTasks))
	report.WriteString(fmt.Sprintf("- Completed: %d (%.1f%%)\n", stats.CompletedTasks, percentage(stats.CompletedTasks, stats.TotalTasks)))
	report.WriteString(fmt.Sprintf("- Active: %d (%.1f%%)\n", stats.ActiveTasks, percentage(stats.ActiveTasks, stats.TotalTasks)))
	report.WriteString(fmt.Sprintf("- Blocked: %d (%.1f%%)\n", stats.BlockedTasks, percentage(stats.BlockedTasks, stats.TotalTasks)))
	report.WriteString(fmt.Sprintf("- Worked On: %d (%.1f%%)\n", stats.WorkedTasks, percentage(stats.WorkedTasks, stats.TotalTasks)))
	report.WriteString("\n")

	// Priority breakdown
	if len(stats.PriorityStats) > 0 {
		report.WriteString("## Priority Breakdown\n")
		for priority, count := range stats.PriorityStats {
			if priority != "None" {
				report.WriteString(fmt.Sprintf("- %s: %d (%.1f%%)\n", priority, count, percentage(count, stats.TotalTasks)))
			}
		}
		report.WriteString("\n")
	}

	// Effort breakdown
	if len(stats.EffortStats) > 0 {
		report.WriteString("## Effort Breakdown\n")
		for effort, count := range stats.EffortStats {
			report.WriteString(fmt.Sprintf("- Effort %d: %d tasks\n", effort, count))
		}
		report.WriteString("\n")
	}

	// Progress summary
	completionRate := percentage(stats.CompletedTasks, stats.TotalTasks)
	report.WriteString("## Progress Summary\n")
	report.WriteString(fmt.Sprintf("- Completion Rate: %.1f%%\n", completionRate))
	
	if stats.TotalTasks > 0 {
		activeRate := percentage(stats.ActiveTasks, stats.TotalTasks)
		report.WriteString(fmt.Sprintf("- Active Rate: %.1f%%\n", activeRate))
	}

	return report.String()
}

// percentage calculates percentage with proper handling of zero values
func percentage(part, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(part) / float64(total) * 100
}

// SaveReport saves a statistics report to a file
func SaveReport(report string, outputPath string) error {
	// Ensure directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
} 