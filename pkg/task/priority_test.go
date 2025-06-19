package task

import (
	"testing"
)

func TestPriority_String(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityNone, "None"},
		{PriorityLow, "Low"},
		{PriorityMedium, "Medium"},
		{PriorityHigh, "High"},
		{PriorityCritical, "Critical"},
		{Priority(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.priority.String()
			if result != tt.expected {
				t.Errorf("Priority.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected Priority
	}{
		{
			name:     "Critical priority A",
			line:     "- [ ] A1 !! Critical task",
			expected: PriorityCritical,
		},
		{
			name:     "High priority B",
			line:     "- [w] B2 High priority task",
			expected: PriorityHigh,
		},
		{
			name:     "Medium priority C",
			line:     "- [b] C3 Medium priority task",
			expected: PriorityMedium,
		},
		{
			name:     "Low priority D",
			line:     "- [x] D5 Low priority task",
			expected: PriorityLow,
		},
		{
			name:     "No priority",
			line:     "- [ ] Regular task without priority",
			expected: PriorityNone,
		},
		{
			name:     "Invalid priority E",
			line:     "- [ ] E1 Invalid priority",
			expected: PriorityNone,
		},
		{
			name:     "Priority in middle of line",
			line:     "- [ ] Task with A1 priority in the middle",
			expected: PriorityCritical,
		},
		{
			name:     "Multiple priorities (should use first)",
			line:     "- [ ] A1 B2 C3 Multiple priorities",
			expected: PriorityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePriority(tt.line)
			if result != tt.expected {
				t.Errorf("ParsePriority() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseEffort(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{
			name:     "Effort 1",
			line:     "- [ ] A1 !! Task with effort 1",
			expected: 1,
		},
		{
			name:     "Effort 2",
			line:     "- [w] B2 Task with effort 2",
			expected: 2,
		},
		{
			name:     "Effort 3",
			line:     "- [b] C3 Task with effort 3",
			expected: 3,
		},
		{
			name:     "Effort 5",
			line:     "- [x] D5 Task with effort 5",
			expected: 5,
		},
		{
			name:     "Effort 8",
			line:     "- [ ] E8 Task with effort 8",
			expected: 8,
		},
		{
			name:     "Effort 13",
			line:     "- [ ] F13 Task with effort 13",
			expected: 13,
		},
		{
			name:     "No effort",
			line:     "- [ ] Regular task without effort",
			expected: 0,
		},
		{
			name:     "Effort in middle of line",
			line:     "- [ ] Task with A1 effort in the middle",
			expected: 1,
		},
		{
			name:     "Multiple efforts (should use first)",
			line:     "- [ ] A1 B2 C3 Multiple efforts",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEffort(tt.line)
			if result != tt.expected {
				t.Errorf("ParseEffort() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestParseTaskInfo(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *TaskInfo
	}{
		{
			name: "Complete task info",
			line: "- [w] A1 !! Task with priority and effort",
			expected: &TaskInfo{
				Line:     "- [w] A1 !! Task with priority and effort",
				Priority: PriorityCritical,
				Effort:   1,
				Status:   "w",
				Title:    "A1 !! Task with priority and effort",
			},
		},
		{
			name: "Completed task",
			line: "- [x] B2 Completed task",
			expected: &TaskInfo{
				Line:     "- [x] B2 Completed task",
				Priority: PriorityHigh,
				Effort:   2,
				Status:   "x",
				Title:    "B2 Completed task",
			},
		},
		{
			name: "Active task",
			line: "- [ ] !! Active task",
			expected: &TaskInfo{
				Line:     "- [ ] !! Active task",
				Priority: PriorityNone,
				Effort:   0,
				Status:   " ",
				Title:    "!! Active task",
			},
		},
		{
			name: "Task without priority or effort",
			line: "- [b] Regular task",
			expected: &TaskInfo{
				Line:     "- [b] Regular task",
				Priority: PriorityNone,
				Effort:   0,
				Status:   "b",
				Title:    "Regular task",
			},
		},
		{
			name:     "Not a task line",
			line:     "This is not a task line",
			expected: nil,
		},
		{
			name:     "Empty task",
			line:     "- [ ]",
			expected: &TaskInfo{
				Line:     "- [ ]",
				Priority: PriorityNone,
				Effort:   0,
				Status:   " ",
				Title:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTaskInfo(tt.line)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("ParseTaskInfo() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("ParseTaskInfo() = nil, want %v", tt.expected)
				return
			}

			if result.Line != tt.expected.Line {
				t.Errorf("ParseTaskInfo().Line = %s, want %s", result.Line, tt.expected.Line)
			}

			if result.Priority != tt.expected.Priority {
				t.Errorf("ParseTaskInfo().Priority = %v, want %v", result.Priority, tt.expected.Priority)
			}

			if result.Effort != tt.expected.Effort {
				t.Errorf("ParseTaskInfo().Effort = %d, want %d", result.Effort, tt.expected.Effort)
			}

			if result.Status != tt.expected.Status {
				t.Errorf("ParseTaskInfo().Status = %s, want %s", result.Status, tt.expected.Status)
			}

			if result.Title != tt.expected.Title {
				t.Errorf("ParseTaskInfo().Title = %s, want %s", result.Title, tt.expected.Title)
			}
		})
	}
}

func TestFormatTaskInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     *TaskInfo
		expected string
	}{
		{
			name:     "Nil info",
			info:     nil,
			expected: "",
		},
		{
			name: "Complete info",
			info: &TaskInfo{
				Line:     "- [w] A1 !! Task",
				Priority: PriorityCritical,
				Effort:   1,
				Status:   "w",
				Title:    "A1 !! Task",
			},
			expected: "Critical | Effort: 1 | Status: w | A1 !! Task",
		},
		{
			name: "No priority",
			info: &TaskInfo{
				Line:     "- [ ] Regular task",
				Priority: PriorityNone,
				Effort:   0,
				Status:   " ",
				Title:    "Regular task",
			},
			expected: "Status:   | Regular task",
		},
		{
			name: "No effort",
			info: &TaskInfo{
				Line:     "- [x] B2 Completed",
				Priority: PriorityHigh,
				Effort:   0,
				Status:   "x",
				Title:    "B2 Completed",
			},
			expected: "High | Status: x | B2 Completed",
		},
		{
			name: "No status",
			info: &TaskInfo{
				Line:     "- [ ] Task",
				Priority: PriorityMedium,
				Effort:   3,
				Status:   "",
				Title:    "Task",
			},
			expected: "Medium | Effort: 3 | Task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTaskInfo(tt.info)
			if result != tt.expected {
				t.Errorf("FormatTaskInfo() = %s, want %s", result, tt.expected)
			}
		})
	}
} 