package valueobject

import "telegram-trello-bot/internal/domain/domainerror"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

func NewPriority(s string) (Priority, error) {
	switch s {
	case "low":
		return PriorityLow, nil
	case "medium", "":
		return PriorityMedium, nil
	case "high":
		return PriorityHigh, nil
	default:
		return "", domainerror.ErrInvalidPriority
	}
}

func (p Priority) String() string { return string(p) }
