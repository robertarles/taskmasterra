package stats

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTaskStats(t *testing.T) {
	stats := NewTaskStats()

	if stats.TotalTasks != 0 {
		t.Errorf("Expected TotalTasks to be 0, got %d", stats.TotalTasks)
	}

	if stats.CompletedTasks != 0 {
		t.Errorf("Expected CompletedTasks to be 0, got %d", stats.CompletedTasks)
	}

	if stats.ActiveTasks != 0 {
		t.Errorf("Expected ActiveTasks to be 0, got %d", stats.ActiveTasks)
	}

	if stats.BlockedTasks != 0 {
		t.Errorf("Expected BlockedTasks to be 0, got %d", stats.BlockedTasks)
	}

	if stats.WorkedTasks != 0 {
		t.Errorf("Expected WorkedTasks to be 0, got %d", stats.WorkedTasks)
	}

	if len(stats.PriorityStats) != 0 {
		t.Errorf("Expected PriorityStats to be empty, got %d items", len(stats.PriorityStats))
	}

	if len(stats.EffortStats) != 0 {
		t.Errorf("Expected EffortStats to be empty, got %d items", len(stats.EffortStats))
	}
}

func TestAnalyzeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stats-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testContent := `# Test TODO

- [ ] A1 !! Active task with priority A and effort 1
- [w] B2 Worked task with priority B and effort 2
- [b] C3 Blocked task with priority C and effort 3
- [x] D5 Completed task with priority D and effort 5
- [ ] !! Another active task
- [W] Worked task without priority
- [B] Blocked task without priority
- [X] Completed task without priority
`

	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	stats, err := AnalyzeFile(filePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Check total tasks
	expectedTotal := 8
	if stats.TotalTasks != expectedTotal {
		t.Errorf("Expected TotalTasks to be %d, got %d", expectedTotal, stats.TotalTasks)
	}

	// Check completed tasks (only [x] and [X] count as completed)
	expectedCompleted := 2
	if stats.CompletedTasks != expectedCompleted {
		t.Errorf("Expected CompletedTasks to be %d, got %d", expectedCompleted, stats.CompletedTasks)
	}

	// Check active tasks (only [ ] !! with space before !! counts as active)
	expectedActive := 1
	if stats.ActiveTasks != expectedActive {
		t.Errorf("Expected ActiveTasks to be %d, got %d", expectedActive, stats.ActiveTasks)
	}

	// Check blocked tasks
	expectedBlocked := 2
	if stats.BlockedTasks != expectedBlocked {
		t.Errorf("Expected BlockedTasks to be %d, got %d", expectedBlocked, stats.BlockedTasks)
	}

	// Check worked tasks
	expectedWorked := 2
	if stats.WorkedTasks != expectedWorked {
		t.Errorf("Expected WorkedTasks to be %d, got %d", expectedWorked, stats.WorkedTasks)
	}

	// Check priority stats
	expectedPriorities := map[string]int{
		"Critical": 1, // A
		"High":     1, // B
		"Medium":   1, // C
		"Low":      1, // D
	}

	for priority, expected := range expectedPriorities {
		if stats.PriorityStats[priority] != expected {
			t.Errorf("Expected PriorityStats[%s] to be %d, got %d", priority, expected, stats.PriorityStats[priority])
		}
	}

	// Check effort stats
	expectedEfforts := map[int]int{
		1: 1, // A1
		2: 1, // B2
		3: 1, // C3
		5: 1, // D5
	}

	for effort, expected := range expectedEfforts {
		if stats.EffortStats[effort] != expected {
			t.Errorf("Expected EffortStats[%d] to be %d, got %d", effort, expected, stats.EffortStats[effort])
		}
	}
}

func TestGenerateReport(t *testing.T) {
	stats := NewTaskStats()
	stats.TotalTasks = 10
	stats.CompletedTasks = 6
	stats.ActiveTasks = 2
	stats.BlockedTasks = 1
	stats.WorkedTasks = 1
	stats.PriorityStats["High"] = 3
	stats.PriorityStats["Medium"] = 2
	stats.EffortStats[5] = 2
	stats.EffortStats[8] = 1

	report := GenerateReport(stats)

	// Check that report contains expected sections
	expectedSections := []string{
		"# Task Statistics Report",
		"## Overall Statistics",
		"## Priority Breakdown",
		"## Effort Breakdown",
		"## Progress Summary",
	}

	for _, section := range expectedSections {
		if !strings.Contains(report, section) {
			t.Errorf("Report should contain section: %s", section)
		}
	}

	// Check that report contains expected data
	expectedData := []string{
		"Total Tasks: 10",
		"Completed: 6 (60.0%)",
		"Active: 2 (20.0%)",
		"Blocked: 1 (10.0%)",
		"Worked On: 1 (10.0%)",
		"High: 3 (30.0%)",
		"Medium: 2 (20.0%)",
		"Effort 5: 2 tasks",
		"Effort 8: 1 tasks",
		"Completion Rate: 60.0%",
	}

	for _, data := range expectedData {
		if !strings.Contains(report, data) {
			t.Errorf("Report should contain: %s", data)
		}
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name     string
		part     int
		total    int
		expected float64
	}{
		{"Zero total", 5, 0, 0.0},
		{"Zero part", 0, 10, 0.0},
		{"Half", 5, 10, 50.0},
		{"Quarter", 1, 4, 25.0},
		{"Full", 10, 10, 100.0},
		{"Small percentage", 1, 100, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentage(tt.part, tt.total)
			if result != tt.expected {
				t.Errorf("percentage(%d, %d) = %f, want %f", tt.part, tt.total, result, tt.expected)
			}
		})
	}
}

func TestSaveReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "report-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	report := "# Test Report\nThis is a test report."
	outputPath := filepath.Join(tmpDir, "report.md")

	err = SaveReport(report, outputPath)
	if err != nil {
		t.Fatalf("SaveReport failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected report file to be created")
	}

	// Verify file content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read saved report: %v", err)
	}

	if string(content) != report {
		t.Errorf("Saved report content doesn't match. Expected: %s, Got: %s", report, string(content))
	}
} 