package entity

import (
	"testing"

	"telegram-trello-bot/internal/domain/valueobject"
)

func TestNewUser_Defaults(t *testing.T) {
	tid := valueobject.NewTelegramID(12345)
	u := NewUser(tid)

	if u.TelegramID() != tid {
		t.Errorf("expected telegram ID %v, got %v", tid, u.TelegramID())
	}
	if u.TrelloToken() != "" {
		t.Errorf("expected empty trello token, got %q", u.TrelloToken())
	}
	if u.HasBoardConfigured() {
		t.Error("expected HasBoardConfigured=false for new user")
	}
	if u.HasListConfigured() {
		t.Error("expected HasListConfigured=false for new user")
	}
	if !u.UseLLM() {
		t.Error("expected UseLLM=true by default")
	}
}

func TestUser_Setters(t *testing.T) {
	u := NewUser(valueobject.NewTelegramID(1))

	u.SetTrelloToken("token123")
	if u.TrelloToken() != "token123" {
		t.Errorf("expected trello token 'token123', got %q", u.TrelloToken())
	}

	u.SetDefaultBoard("board-abc")
	if u.DefaultBoard() != "board-abc" {
		t.Errorf("expected board 'board-abc', got %q", u.DefaultBoard())
	}
	if !u.HasBoardConfigured() {
		t.Error("expected HasBoardConfigured=true after setting board")
	}

	u.SetDefaultList("list-xyz")
	if u.DefaultList() != "list-xyz" {
		t.Errorf("expected list 'list-xyz', got %q", u.DefaultList())
	}
	if !u.HasListConfigured() {
		t.Error("expected HasListConfigured=true after setting list")
	}

	u.SetUseLLM(false)
	if u.UseLLM() {
		t.Error("expected UseLLM=false after SetUseLLM(false)")
	}
}
