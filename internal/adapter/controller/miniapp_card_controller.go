package controller

import (
	"context"

	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
)

type MemberResult struct {
	Members []MemberResultItem `json:"members"`
}

type MemberResultItem struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type CardListResult struct {
	Cards []CardResultItem `json:"cards"`
}

type CardResultItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	ListID string `json:"list_id"`
}

type CardDetailResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	ListID      string   `json:"list_id"`
	Due         string   `json:"due"`
	Labels      []string `json:"labels"`
	Members     []string `json:"members"`
}

type CreateCardResult struct {
	CardID  string `json:"card_id"`
	CardURL string `json:"card_url"`
	Title   string `json:"title"`
}

type CreateCardRequest struct {
	TelegramID  int64
	ListID      string
	Title       string
	Description string
	DueDate     string
	LabelIDs    []string
	MemberIDs   []string
}

type UpdateCardRequest struct {
	Title       *string
	Description *string
	ListID      *string
	Due         *string
	LabelIDs    *string
	MemberIDs   *string
}

type MiniAppCardController struct {
	listMembers *usecase.ListMembersUseCase
	listCards   *usecase.ListCardsUseCase
	getCard     *usecase.GetCardUseCase
	createCard  *usecase.CreateCardFormUseCase
	updateCard  *usecase.UpdateCardUseCase
}

func NewMiniAppCardController(
	listMembers *usecase.ListMembersUseCase,
	listCards *usecase.ListCardsUseCase,
	getCard *usecase.GetCardUseCase,
	createCard *usecase.CreateCardFormUseCase,
	updateCard *usecase.UpdateCardUseCase,
) *MiniAppCardController {
	return &MiniAppCardController{
		listMembers: listMembers,
		listCards:   listCards,
		getCard:     getCard,
		createCard:  createCard,
		updateCard:  updateCard,
	}
}

func (c *MiniAppCardController) HandleListMembers(ctx context.Context, telegramID int64, boardID string) (*MemberResult, error) {
	output, err := c.listMembers.Execute(ctx, telegramID, boardID)
	if err != nil {
		return nil, err
	}
	items := make([]MemberResultItem, len(output.Members))
	for i, m := range output.Members {
		items[i] = MemberResultItem{ID: m.ID, Username: m.Username, FullName: m.FullName}
	}
	return &MemberResult{Members: items}, nil
}

func (c *MiniAppCardController) HandleListCards(ctx context.Context, telegramID int64, listID string) (*CardListResult, error) {
	output, err := c.listCards.Execute(ctx, telegramID, listID)
	if err != nil {
		return nil, err
	}
	items := make([]CardResultItem, len(output.Cards))
	for i, card := range output.Cards {
		items[i] = CardResultItem{ID: card.ID, Title: card.Title, URL: card.URL, ListID: card.ListID}
	}
	return &CardListResult{Cards: items}, nil
}

func (c *MiniAppCardController) HandleGetCard(ctx context.Context, telegramID int64, cardID string) (*CardDetailResult, error) {
	output, err := c.getCard.Execute(ctx, telegramID, cardID)
	if err != nil {
		return nil, err
	}
	return &CardDetailResult{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		URL:         output.URL,
		ListID:      output.ListID,
		Due:         output.Due,
		Labels:      output.Labels,
		Members:     output.Members,
	}, nil
}

func (c *MiniAppCardController) HandleCreateCard(ctx context.Context, req CreateCardRequest) (*CreateCardResult, error) {
	output, err := c.createCard.Execute(ctx, dto.CreateCardFormInput{
		TelegramID:  req.TelegramID,
		ListID:      req.ListID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		LabelIDs:    req.LabelIDs,
		MemberIDs:   req.MemberIDs,
	})
	if err != nil {
		return nil, err
	}
	return &CreateCardResult{
		CardID:  output.CardID,
		CardURL: output.CardURL,
		Title:   output.Title,
	}, nil
}

func (c *MiniAppCardController) HandleUpdateCard(ctx context.Context, telegramID int64, cardID string, req UpdateCardRequest) error {
	return c.updateCard.Execute(ctx, telegramID, cardID, dto.UpdateCardInput{
		Title:       req.Title,
		Description: req.Description,
		ListID:      req.ListID,
		Due:         req.Due,
		LabelIDs:    req.LabelIDs,
		MemberIDs:   req.MemberIDs,
	})
}
