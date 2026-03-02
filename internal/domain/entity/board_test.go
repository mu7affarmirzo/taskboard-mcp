package entity

import "testing"

func TestNewBoard(t *testing.T) {
	b := NewBoard("board-1", "My Board")
	if b.ID() != "board-1" {
		t.Errorf("expected ID 'board-1', got %q", b.ID())
	}
	if b.Name() != "My Board" {
		t.Errorf("expected name 'My Board', got %q", b.Name())
	}
}

func TestNewBoardList(t *testing.T) {
	l := NewBoardList("list-1", "To Do")
	if l.ID() != "list-1" {
		t.Errorf("expected ID 'list-1', got %q", l.ID())
	}
	if l.Name() != "To Do" {
		t.Errorf("expected name 'To Do', got %q", l.Name())
	}
}
