package entity

import (
	"testing"
	"time"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
)

func TestNewTask_EmptyTitle_ReturnsError(t *testing.T) {
	task, err := NewTask("")
	if err != domainerror.ErrEmptyTaskTitle {
		t.Errorf("expected ErrEmptyTaskTitle, got %v", err)
	}
	if task != nil {
		t.Error("expected nil task for empty title")
	}
}

func TestNewTask_ValidTitle(t *testing.T) {
	task, err := NewTask("Buy groceries")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title() != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", task.Title())
	}
	if task.Priority() != valueobject.PriorityMedium {
		t.Errorf("expected default priority Medium, got %q", task.Priority())
	}
	if task.Description() != "" {
		t.Errorf("expected empty description, got %q", task.Description())
	}
	if task.DueDate() != nil {
		t.Error("expected nil due date")
	}
	if len(task.Labels()) != 0 {
		t.Errorf("expected no labels, got %v", task.Labels())
	}
	if len(task.Checklist()) != 0 {
		t.Errorf("expected no checklist, got %v", task.Checklist())
	}
}

func TestNewTask_WithOptions(t *testing.T) {
	due := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	labels := []string{"urgent", "work"}
	checklist := []string{"step 1", "step 2"}

	task, err := NewTask("Deploy app",
		WithDescription("Deploy to production"),
		WithDueDate(due),
		WithPriority(valueobject.PriorityHigh),
		WithLabels(labels),
		WithChecklist(checklist),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title() != "Deploy app" {
		t.Errorf("expected title 'Deploy app', got %q", task.Title())
	}
	if task.Description() != "Deploy to production" {
		t.Errorf("expected description 'Deploy to production', got %q", task.Description())
	}
	if task.DueDate() == nil || !task.DueDate().Equal(due) {
		t.Errorf("expected due date %v, got %v", due, task.DueDate())
	}
	if task.Priority() != valueobject.PriorityHigh {
		t.Errorf("expected priority High, got %q", task.Priority())
	}
	if len(task.Labels()) != 2 || task.Labels()[0] != "urgent" {
		t.Errorf("expected labels [urgent work], got %v", task.Labels())
	}
	if len(task.Checklist()) != 2 || task.Checklist()[0] != "step 1" {
		t.Errorf("expected checklist [step 1, step 2], got %v", task.Checklist())
	}
}

func TestTask_IsHighPriority(t *testing.T) {
	high, _ := NewTask("Urgent task", WithPriority(valueobject.PriorityHigh))
	if !high.IsHighPriority() {
		t.Error("expected IsHighPriority=true for high priority task")
	}

	medium, _ := NewTask("Normal task")
	if medium.IsHighPriority() {
		t.Error("expected IsHighPriority=false for medium priority task")
	}

	low, _ := NewTask("Low task", WithPriority(valueobject.PriorityLow))
	if low.IsHighPriority() {
		t.Error("expected IsHighPriority=false for low priority task")
	}
}
