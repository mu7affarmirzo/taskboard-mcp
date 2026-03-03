package controller

import (
	"context"

	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
)

type MiniAppSettingsController struct {
	deleteCard    *usecase.DeleteCardUseCase
	addComment    *usecase.AddCommentUseCase
	selectBoard   *usecase.SelectBoardUseCase
	selectList    *usecase.SelectListUseCase
	connectTrello *usecase.ConnectTrelloUseCase
}

func NewMiniAppSettingsController(
	deleteCard *usecase.DeleteCardUseCase,
	addComment *usecase.AddCommentUseCase,
	selectBoard *usecase.SelectBoardUseCase,
	selectList *usecase.SelectListUseCase,
	connectTrello *usecase.ConnectTrelloUseCase,
) *MiniAppSettingsController {
	return &MiniAppSettingsController{
		deleteCard:    deleteCard,
		addComment:    addComment,
		selectBoard:   selectBoard,
		selectList:    selectList,
		connectTrello: connectTrello,
	}
}

func (c *MiniAppSettingsController) HandleDeleteCard(ctx context.Context, telegramID int64, cardID string) error {
	return c.deleteCard.Execute(ctx, telegramID, cardID)
}

func (c *MiniAppSettingsController) HandleAddComment(ctx context.Context, telegramID int64, cardID string, text string) error {
	return c.addComment.Execute(ctx, telegramID, cardID, text)
}

func (c *MiniAppSettingsController) HandleUpdateSettings(ctx context.Context, telegramID int64, boardID string, listID string) error {
	if boardID != "" {
		if err := c.selectBoard.Execute(ctx, telegramID, boardID); err != nil {
			return err
		}
	}
	if listID != "" {
		if err := c.selectList.Execute(ctx, telegramID, listID); err != nil {
			return err
		}
	}
	return nil
}

func (c *MiniAppSettingsController) HandleConnectTrello(ctx context.Context, telegramID int64, token string) (*dto.ConnectTrelloOutput, error) {
	return c.connectTrello.Execute(ctx, dto.ConnectTrelloInput{TelegramID: telegramID, Token: token})
}
