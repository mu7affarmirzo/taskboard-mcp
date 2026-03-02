package presenter_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/usecase/dto"
)

func TestFormatTaskCreated_Full(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	due := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	output := &dto.CreateTaskOutput{
		CardURL:   "https://trello.com/c/abc",
		TaskTitle: "Fix bug",
		DueDate:   &due,
		Priority:  "high",
		Labels:    []string{"backend", "urgent"},
	}

	result := p.FormatTaskCreated(output)

	assert.Contains(t, result, "Task Created!")
	assert.Contains(t, result, "Fix bug")
	assert.Contains(t, result, "high")
	assert.Contains(t, result, "Mar 15, 2025")
	assert.Contains(t, result, "backend, urgent")
	assert.Contains(t, result, "https://trello.com/c/abc")
}

func TestFormatTaskCreated_Minimal(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	output := &dto.CreateTaskOutput{
		CardURL:   "url",
		TaskTitle: "Simple task",
		Priority:  "medium",
	}

	result := p.FormatTaskCreated(output)

	assert.Contains(t, result, "Simple task")
	assert.Contains(t, result, "medium")
	assert.NotContains(t, result, "Due:")
	assert.NotContains(t, result, "Labels:")
}

func TestFormatTaskPreview(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	output := &dto.ParseTaskOutput{
		TaskTitle: "Review PR",
		Priority:  "low",
	}

	result := p.FormatTaskPreview(output)

	assert.Contains(t, result, "Create this task?")
	assert.Contains(t, result, "Review PR")
	assert.Contains(t, result, "low")
}

func TestFormatTaskPreview_WithChecklist(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	output := &dto.ParseTaskOutput{
		TaskTitle: "Deploy app",
		Priority:  "high",
		Checklist: []string{"run tests", "update docs"},
	}

	result := p.FormatTaskPreview(output)

	assert.Contains(t, result, "Checklist:")
	assert.Contains(t, result, "run tests")
	assert.Contains(t, result, "update docs")
}

func TestFormatBoardList(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	output := &dto.ListBoardsOutput{
		Boards: []dto.BoardItem{
			{ID: "1", Name: "Work"},
			{ID: "2", Name: "Personal"},
		},
	}

	result := p.FormatBoardList(output)

	assert.Contains(t, result, "Your Trello Boards:")
	assert.Contains(t, result, "1. Work")
	assert.Contains(t, result, "2. Personal")
	assert.Contains(t, result, "Tap a board")
}

func TestFormatBoardSelected(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	result := p.FormatBoardSelected("board-123")

	assert.Contains(t, result, "board-123")
	assert.Contains(t, result, "select a list")
}

func TestFormatListSelected(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	result := p.FormatListSelected("list-456")

	assert.Contains(t, result, "list-456")
	assert.Contains(t, result, "all set")
}

func TestFormatError(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	result := p.FormatError(errors.New("connection refused"))

	assert.Contains(t, result, "Something went wrong")
	assert.Contains(t, result, "connection refused")
}

func TestFormatHelp(t *testing.T) {
	p := presenter.NewTelegramPresenter()
	result := p.FormatHelp()

	assert.Contains(t, result, "Trello Bot")
	assert.Contains(t, result, "/boards")
	assert.Contains(t, result, "/help")
}
