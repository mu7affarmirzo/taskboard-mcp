package valueobject

import "testing"

func TestNewTelegramID(t *testing.T) {
	tid := NewTelegramID(99887766)
	if tid.Int64() != 99887766 {
		t.Errorf("expected 99887766, got %d", tid.Int64())
	}
}

func TestTelegramID_Zero(t *testing.T) {
	tid := NewTelegramID(0)
	if tid.Int64() != 0 {
		t.Errorf("expected 0, got %d", tid.Int64())
	}
}
