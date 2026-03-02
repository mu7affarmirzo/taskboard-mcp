package telegram_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	infratelegram "telegram-trello-bot/internal/infrastructure/telegram"
	"telegram-trello-bot/internal/usecase/port"
)

func TestBuildBoardKeyboard(t *testing.T) {
	boards := []port.BoardInfo{
		{ID: "b1", Name: "Work"},
		{ID: "b2", Name: "Personal"},
	}

	kb := infratelegram.BuildBoardKeyboard(boards)

	assert.Len(t, kb.InlineKeyboard, 2)
	assert.Equal(t, "Work", kb.InlineKeyboard[0][0].Text)
	assert.Equal(t, "board:b1", *kb.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "Personal", kb.InlineKeyboard[1][0].Text)
	assert.Equal(t, "board:b2", *kb.InlineKeyboard[1][0].CallbackData)
}

func TestBuildListKeyboard(t *testing.T) {
	lists := []port.ListInfo{
		{ID: "l1", Name: "To Do"},
		{ID: "l2", Name: "Done"},
	}

	kb := infratelegram.BuildListKeyboard(lists)

	assert.Len(t, kb.InlineKeyboard, 2)
	assert.Equal(t, "To Do", kb.InlineKeyboard[0][0].Text)
	assert.Equal(t, "list:l1", *kb.InlineKeyboard[0][0].CallbackData)
}

func TestBuildConfirmKeyboard(t *testing.T) {
	kb := infratelegram.BuildConfirmKeyboard()

	assert.Len(t, kb.InlineKeyboard, 1)
	assert.Len(t, kb.InlineKeyboard[0], 3)
	assert.Equal(t, "Create", kb.InlineKeyboard[0][0].Text)
	assert.Equal(t, "confirm:create", *kb.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "Edit", kb.InlineKeyboard[0][1].Text)
	assert.Equal(t, "Cancel", kb.InlineKeyboard[0][2].Text)
}

func TestBuildBoardKeyboard_Empty(t *testing.T) {
	kb := infratelegram.BuildBoardKeyboard(nil)
	assert.Empty(t, kb.InlineKeyboard)
}
