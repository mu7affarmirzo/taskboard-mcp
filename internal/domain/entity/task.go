package entity

import (
	"time"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
)

type Task struct {
	title       string
	description string
	dueDate     *time.Time
	priority    valueobject.Priority
	labels      []string
	checklist   []string
	members     []string
}

func NewTask(title string, opts ...TaskOption) (*Task, error) {
	if title == "" {
		return nil, domainerror.ErrEmptyTaskTitle
	}
	t := &Task{
		title:    title,
		priority: valueobject.PriorityMedium,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t, nil
}

type TaskOption func(*Task)

func WithDescription(desc string) TaskOption {
	return func(t *Task) { t.description = desc }
}

func WithDueDate(d time.Time) TaskOption {
	return func(t *Task) { t.dueDate = &d }
}

func WithPriority(p valueobject.Priority) TaskOption {
	return func(t *Task) { t.priority = p }
}

func WithLabels(labels []string) TaskOption {
	return func(t *Task) { t.labels = labels }
}

func WithChecklist(items []string) TaskOption {
	return func(t *Task) { t.checklist = items }
}

func WithMembers(members []string) TaskOption {
	return func(t *Task) { t.members = members }
}

func (t *Task) Title() string                  { return t.title }
func (t *Task) Description() string            { return t.description }
func (t *Task) DueDate() *time.Time            { return t.dueDate }
func (t *Task) Priority() valueobject.Priority { return t.priority }
func (t *Task) Labels() []string               { return t.labels }
func (t *Task) Checklist() []string            { return t.checklist }
func (t *Task) Members() []string              { return t.members }
func (t *Task) IsHighPriority() bool           { return t.priority == valueobject.PriorityHigh }
