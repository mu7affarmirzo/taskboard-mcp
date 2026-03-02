package valueobject

import "testing"

func TestNewTaskID(t *testing.T) {
	tid := NewTaskID("task-123")
	if tid.String() != "task-123" {
		t.Errorf("expected 'task-123', got %q", tid.String())
	}
}

func TestTaskID_Empty(t *testing.T) {
	tid := NewTaskID("")
	if tid.String() != "" {
		t.Errorf("expected empty string, got %q", tid.String())
	}
}
