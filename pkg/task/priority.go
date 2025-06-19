package task

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Precompiled regex patterns for better performance
var (
	priorityEffortRegex = regexp.MustCompile(`\b([A-Z])(\d+)\b`)
	statusRegex         = regexp.MustCompile(`^\s*- \[([^\]]+)\]`)
	titleRegex          = regexp.MustCompile(`^\s*- \[[^\]]+\]\s*(.*)`)
)

// Priority represents task priority levels
type Priority int

const (
	PriorityNone Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// String returns the string representation of priority
func (p Priority) String() string {
	switch p {
	case PriorityNone:
		return "None"
	case PriorityLow:
		return "Low"
	case PriorityMedium:
		return "Medium"
	case PriorityHigh:
		return "High"
	case PriorityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ParsePriority extracts priority from task line
func ParsePriority(line string) Priority {
	// Look for priority markers like A1, B2, C3, etc.
	matches := priorityEffortRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return PriorityNone
	}

	priority := matches[1]

	switch priority {
	case "A":
		return PriorityCritical
	case "B":
		return PriorityHigh
	case "C":
		return PriorityMedium
	case "D":
		return PriorityLow
	default:
		return PriorityNone
	}
}

// ParseEffort extracts effort estimation from task line
func ParseEffort(line string) int {
	// Look for fibonacci effort numbers (1, 2, 3, 5, 8, 13, 21, 34, 55, 89)
	matches := priorityEffortRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return 0
	}

	effort, _ := strconv.Atoi(matches[2])
	return effort
}

// TaskInfo contains parsed task information
type TaskInfo struct {
	Line     string
	Priority Priority
	Effort   int
	Status   string
	Title    string
}

// ParseTaskInfo extracts all task information from a line
func ParseTaskInfo(line string) *TaskInfo {
	if !IsTask(line) {
		return nil
	}

	// Extract status
	statusMatches := statusRegex.FindStringSubmatch(line)
	status := ""
	if len(statusMatches) > 1 {
		status = statusMatches[1]
	}

	// Extract title (everything after the status)
	titleMatches := titleRegex.FindStringSubmatch(line)
	title := ""
	if len(titleMatches) > 1 {
		title = strings.TrimSpace(titleMatches[1])
	}

	return &TaskInfo{
		Line:     line,
		Priority: ParsePriority(line),
		Effort:   ParseEffort(line),
		Status:   status,
		Title:    title,
	}
}

// FormatTaskInfo formats task information for display
func FormatTaskInfo(info *TaskInfo) string {
	if info == nil {
		return ""
	}

	parts := []string{}
	
	// Add priority if present
	if info.Priority != PriorityNone {
		parts = append(parts, info.Priority.String())
	}
	
	// Add effort if present
	if info.Effort > 0 {
		parts = append(parts, fmt.Sprintf("Effort: %d", info.Effort))
	}
	
	// Add status
	if info.Status != "" {
		parts = append(parts, fmt.Sprintf("Status: %s", info.Status))
	}
	
	// Add title
	if info.Title != "" {
		parts = append(parts, info.Title)
	}
	
	return strings.Join(parts, " | ")
} 