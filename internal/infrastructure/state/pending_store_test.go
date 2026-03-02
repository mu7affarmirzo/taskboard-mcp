package state

import (
	"testing"
	"time"
)

func TestPendingStore_SetAndGet(t *testing.T) {
	store := NewPendingStore()

	due := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	store.Set(12345, PendingTask{
		Title:    "Fix bug",
		Priority: "high",
		DueDate:  &due,
		Labels:   []string{"backend"},
	})

	task, ok := store.Get(12345)
	if !ok {
		t.Fatal("expected task to exist")
	}
	if task.Title != "Fix bug" {
		t.Errorf("expected title 'Fix bug', got %q", task.Title)
	}
	if task.Priority != "high" {
		t.Errorf("expected priority 'high', got %q", task.Priority)
	}
	if task.DueDate == nil || !task.DueDate.Equal(due) {
		t.Errorf("expected due date %v, got %v", due, task.DueDate)
	}
	if len(task.Labels) != 1 || task.Labels[0] != "backend" {
		t.Errorf("expected labels [backend], got %v", task.Labels)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPendingStore_GetMissing(t *testing.T) {
	store := NewPendingStore()

	_, ok := store.Get(99999)
	if ok {
		t.Error("expected task not to exist")
	}
}

func TestPendingStore_Delete(t *testing.T) {
	store := NewPendingStore()

	store.Set(12345, PendingTask{Title: "task"})
	store.Delete(12345)

	_, ok := store.Get(12345)
	if ok {
		t.Error("expected task to be deleted")
	}
}

func TestPendingStore_Overwrite(t *testing.T) {
	store := NewPendingStore()

	store.Set(12345, PendingTask{Title: "first"})
	store.Set(12345, PendingTask{Title: "second"})

	task, ok := store.Get(12345)
	if !ok {
		t.Fatal("expected task to exist")
	}
	if task.Title != "second" {
		t.Errorf("expected title 'second', got %q", task.Title)
	}
}

func TestPendingStore_MultipleUsers(t *testing.T) {
	store := NewPendingStore()

	store.Set(111, PendingTask{Title: "user1 task"})
	store.Set(222, PendingTask{Title: "user2 task"})

	t1, ok := store.Get(111)
	if !ok || t1.Title != "user1 task" {
		t.Errorf("expected user1 task, got %+v", t1)
	}

	t2, ok := store.Get(222)
	if !ok || t2.Title != "user2 task" {
		t.Errorf("expected user2 task, got %+v", t2)
	}
}
