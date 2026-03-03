package controller

import (
	"context"

	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
)

type AuthResult struct {
	Token     string `json:"token"`
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type SettingsResult struct {
	TrelloConnected bool   `json:"trello_connected"`
	DefaultBoardID  string `json:"default_board_id"`
	DefaultListID   string `json:"default_list_id"`
}

type BoardResult struct {
	Boards []BoardResultItem `json:"boards"`
}

type BoardResultItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListResult struct {
	Lists []ListResultItem `json:"lists"`
}

type ListResultItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LabelResult struct {
	Labels []LabelResultItem `json:"labels"`
}

type LabelResultItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type MiniAppController struct {
	auth            *usecase.AuthenticateUseCase
	getUserSettings *usecase.GetUserSettingsUseCase
	listBoards      *usecase.ListBoardsUseCase
	listLists       *usecase.ListListsUseCase
	listLabels      *usecase.ListLabelsUseCase
}

func NewMiniAppController(
	auth *usecase.AuthenticateUseCase,
	getUserSettings *usecase.GetUserSettingsUseCase,
	listBoards *usecase.ListBoardsUseCase,
	listLists *usecase.ListListsUseCase,
	listLabels *usecase.ListLabelsUseCase,
) *MiniAppController {
	return &MiniAppController{
		auth:            auth,
		getUserSettings: getUserSettings,
		listBoards:      listBoards,
		listLists:       listLists,
		listLabels:      listLabels,
	}
}

func (c *MiniAppController) HandleAuth(ctx context.Context, initData string) (*AuthResult, error) {
	output, err := c.auth.Execute(ctx, dto.AuthInput{InitData: initData})
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		Token:     output.Token,
		UserID:    output.UserID,
		FirstName: output.FirstName,
		Username:  output.Username,
	}, nil
}

func (c *MiniAppController) HandleGetSettings(ctx context.Context, telegramID int64) (*SettingsResult, error) {
	output, err := c.getUserSettings.Execute(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	return &SettingsResult{
		TrelloConnected: output.TrelloConnected,
		DefaultBoardID:  output.DefaultBoardID,
		DefaultListID:   output.DefaultListID,
	}, nil
}

func (c *MiniAppController) HandleListBoards(ctx context.Context, telegramID int64) (*BoardResult, error) {
	output, err := c.listBoards.Execute(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	items := make([]BoardResultItem, len(output.Boards))
	for i, b := range output.Boards {
		items[i] = BoardResultItem{ID: b.ID, Name: b.Name}
	}
	return &BoardResult{Boards: items}, nil
}

func (c *MiniAppController) HandleListLists(ctx context.Context, telegramID int64, boardID string) (*ListResult, error) {
	output, err := c.listLists.Execute(ctx, telegramID, boardID)
	if err != nil {
		return nil, err
	}
	items := make([]ListResultItem, len(output.Lists))
	for i, l := range output.Lists {
		items[i] = ListResultItem{ID: l.ID, Name: l.Name}
	}
	return &ListResult{Lists: items}, nil
}

func (c *MiniAppController) HandleListLabels(ctx context.Context, telegramID int64, boardID string) (*LabelResult, error) {
	output, err := c.listLabels.Execute(ctx, telegramID, boardID)
	if err != nil {
		return nil, err
	}
	items := make([]LabelResultItem, len(output.Labels))
	for i, l := range output.Labels {
		items[i] = LabelResultItem{ID: l.ID, Name: l.Name, Color: l.Color}
	}
	return &LabelResult{Labels: items}, nil
}
