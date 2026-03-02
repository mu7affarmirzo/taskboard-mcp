package valueobject

import (
	"testing"

	"telegram-trello-bot/internal/domain/domainerror"
)

func TestNewPriority_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
	}{
		{"low", PriorityLow},
		{"medium", PriorityMedium},
		{"high", PriorityHigh},
		{"", PriorityMedium},
	}

	for _, tt := range tests {
		p, err := NewPriority(tt.input)
		if err != nil {
			t.Errorf("NewPriority(%q): unexpected error: %v", tt.input, err)
		}
		if p != tt.expected {
			t.Errorf("NewPriority(%q) = %q, want %q", tt.input, p, tt.expected)
		}
	}
}

func TestNewPriority_Invalid(t *testing.T) {
	invalid := []string{"critical", "urgent", "none", "HIGH", "Low"}

	for _, input := range invalid {
		p, err := NewPriority(input)
		if err != domainerror.ErrInvalidPriority {
			t.Errorf("NewPriority(%q): expected ErrInvalidPriority, got err=%v", input, err)
		}
		if p != "" {
			t.Errorf("NewPriority(%q): expected empty priority, got %q", input, p)
		}
	}
}

func TestPriority_String(t *testing.T) {
	if PriorityLow.String() != "low" {
		t.Errorf("expected 'low', got %q", PriorityLow.String())
	}
	if PriorityMedium.String() != "medium" {
		t.Errorf("expected 'medium', got %q", PriorityMedium.String())
	}
	if PriorityHigh.String() != "high" {
		t.Errorf("expected 'high', got %q", PriorityHigh.String())
	}
}
