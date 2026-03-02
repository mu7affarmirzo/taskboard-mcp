package entity

import "testing"

func TestNewLabel_Valid(t *testing.T) {
	label := NewLabel("lbl-1", "Backend", "blue")
	if label == nil {
		t.Fatal("expected non-nil label")
	}
	if label.ID() != "lbl-1" {
		t.Errorf("expected ID 'lbl-1', got %q", label.ID())
	}
	if label.Name() != "Backend" {
		t.Errorf("expected Name 'Backend', got %q", label.Name())
	}
	if label.Color() != "blue" {
		t.Errorf("expected Color 'blue', got %q", label.Color())
	}
}

func TestNewLabel_EmptyFields(t *testing.T) {
	label := NewLabel("", "", "")
	if label == nil {
		t.Fatal("expected non-nil label even with empty fields")
	}
	if label.ID() != "" {
		t.Errorf("expected empty ID, got %q", label.ID())
	}
	if label.Name() != "" {
		t.Errorf("expected empty Name, got %q", label.Name())
	}
	if label.Color() != "" {
		t.Errorf("expected empty Color, got %q", label.Color())
	}
}
