package controller

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/infrastructure/state"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
)

type TelegramController struct {
	createTask    *usecase.CreateTaskUseCase
	parseTask     *usecase.ParseTaskUseCase
	confirmTask   *usecase.ConfirmTaskUseCase
	listBoards    *usecase.ListBoardsUseCase
	listLists     *usecase.ListListsUseCase
	selectBoard   *usecase.SelectBoardUseCase
	selectList    *usecase.SelectListUseCase
	registerUser  *usecase.RegisterUserUseCase
	connectTrello *usecase.ConnectTrelloUseCase
	pending       *state.PendingStore
}

func NewTelegramController(
	createTask *usecase.CreateTaskUseCase,
	parseTask *usecase.ParseTaskUseCase,
	confirmTask *usecase.ConfirmTaskUseCase,
	listBoards *usecase.ListBoardsUseCase,
	listLists *usecase.ListListsUseCase,
	selectBoard *usecase.SelectBoardUseCase,
	selectList *usecase.SelectListUseCase,
	registerUser *usecase.RegisterUserUseCase,
	connectTrello *usecase.ConnectTrelloUseCase,
	pending *state.PendingStore,
) *TelegramController {
	return &TelegramController{
		createTask:    createTask,
		parseTask:     parseTask,
		confirmTask:   confirmTask,
		listBoards:    listBoards,
		listLists:     listLists,
		selectBoard:   selectBoard,
		selectList:    selectList,
		registerUser:  registerUser,
		connectTrello: connectTrello,
		pending:       pending,
	}
}

func (c *TelegramController) HandleStart(ctx context.Context, telegramID int64) (*dto.RegisterUserOutput, error) {
	return c.registerUser.Execute(ctx, dto.RegisterUserInput{TelegramID: telegramID})
}

func (c *TelegramController) HandleConnectTrello(ctx context.Context, telegramID int64, token string) (*dto.ConnectTrelloOutput, error) {
	return c.connectTrello.Execute(ctx, dto.ConnectTrelloInput{TelegramID: telegramID, Token: token})
}

func (c *TelegramController) HandleMessage(ctx context.Context, telegramID int64, text string) (*dto.CreateTaskOutput, error) {
	return c.createTask.Execute(ctx, dto.CreateTaskInput{TelegramID: telegramID, RawMessage: text})
}

func (c *TelegramController) HandleParseTask(ctx context.Context, telegramID int64, text string) (*dto.ParseTaskOutput, error) {
	output, err := c.parseTask.Execute(ctx, dto.CreateTaskInput{TelegramID: telegramID, RawMessage: text})
	if err != nil {
		return nil, err
	}

	c.pending.Set(telegramID, state.PendingTask{
		Title:       output.TaskTitle,
		Description: output.Description,
		DueDate:     output.DueDate,
		Priority:    output.Priority,
		Labels:      output.Labels,
		Checklist:   output.Checklist,
		Members:     output.Members,
		RawMessage:  text,
	})

	return output, nil
}

func (c *TelegramController) HandleConfirmTask(ctx context.Context, telegramID int64) (*dto.CreateTaskOutput, error) {
	task, ok := c.pending.Get(telegramID)
	if !ok {
		return nil, fmt.Errorf("no pending task found")
	}
	c.pending.Delete(telegramID)

	return c.confirmTask.Execute(ctx, dto.ConfirmTaskInput{
		TelegramID:  telegramID,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate,
		Priority:    task.Priority,
		Labels:      task.Labels,
		Members:     task.Members,
	})
}

func (c *TelegramController) HandleCancelTask(telegramID int64) {
	c.pending.Delete(telegramID)
}

func (c *TelegramController) HandleListBoards(ctx context.Context, telegramID int64) (*dto.ListBoardsOutput, error) {
	return c.listBoards.Execute(ctx, telegramID)
}

func (c *TelegramController) HandleListLists(ctx context.Context, telegramID int64, boardID string) (*dto.ListListsOutput, error) {
	return c.listLists.Execute(ctx, telegramID, boardID)
}

func (c *TelegramController) HandleSelectBoard(ctx context.Context, telegramID int64, boardID string) error {
	return c.selectBoard.Execute(ctx, telegramID, boardID)
}

func (c *TelegramController) HandleSelectList(ctx context.Context, telegramID int64, listID string) error {
	return c.selectList.Execute(ctx, telegramID, listID)
}
