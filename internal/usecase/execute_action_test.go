package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

func setupExecuteAction() (*usecase.ExecuteActionUseCase, *MockBoard, *MockCardManager, *MockMemberResolver, *MockUserRepo, *MockTaskLog) {
	board := new(MockBoard)
	cardMgr := new(MockCardManager)
	members := new(MockMemberResolver)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	user := entity.NewUser(valueobject.TelegramID(123))
	user.SetTrelloToken("tok")
	user.SetDefaultBoard("board1")
	user.SetDefaultList("list1")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(123)).Return(user, nil)

	uc := usecase.NewExecuteActionUseCase(board, cardMgr, members, userRepo, taskLog)
	return uc, board, cardMgr, members, userRepo, taskLog
}

func TestExecuteAction_MoveCard(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "login card").
		Return([]port.CardResult{{CardID: "c1", Title: "login card"}}, nil)
	board.On("GetLists", mock.Anything, "tok", "board1").
		Return([]port.ListInfo{{ID: "l2", Name: "Done"}}, nil)
	cardMgr.On("UpdateCard", mock.Anything, "tok", "c1", mock.Anything).Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "move_card",
		CardName: "login card",
		ListName: "Done",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Card moved to Done", result.Message)
}

func TestExecuteAction_GetCard(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "auth").
		Return([]port.CardResult{{CardID: "c1", Title: "auth"}}, nil)
	cardMgr.On("GetCard", mock.Anything, "tok", "c1").
		Return(&port.CardInfo{
			ID:    "c1",
			Title: "auth",
			URL:   "https://trello.com/c/c1",
		}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "get_card",
		CardName: "auth",
	})

	assert.NoError(t, err)
	assert.Equal(t, "auth", result.Message)
	assert.Equal(t, "https://trello.com/c/c1", result.CardURL)
}

func TestExecuteAction_ListCards(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("GetLists", mock.Anything, "tok", "board1").
		Return([]port.ListInfo{{ID: "l1", Name: "Testing"}}, nil)
	board.On("GetCards", mock.Anything, "tok", "l1").
		Return([]port.CardResult{
			{CardID: "c1", Title: "card1", CardURL: "url1"},
			{CardID: "c2", Title: "card2", CardURL: "url2"},
		}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "list_cards",
		ListName: "Testing",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Cards in Testing", result.Message)
	assert.Len(t, result.Items, 2)
}

func TestExecuteAction_SearchCards(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "auth").
		Return([]port.CardResult{
			{CardID: "c1", Title: "auth card", CardURL: "url1"},
		}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:      "search_cards",
		SearchQuery: "auth",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Found 1 card(s)", result.Message)
	assert.Len(t, result.Items, 1)
}

func TestExecuteAction_AddComment(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "payment").
		Return([]port.CardResult{{CardID: "c1", Title: "payment"}}, nil)
	cardMgr.On("AddComment", mock.Anything, "tok", "c1", "API keys updated").Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:      "add_comment",
		CardName:    "payment",
		CommentText: "API keys updated",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Comment added", result.Message)
}

func TestExecuteAction_ArchiveCard(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "old task").
		Return([]port.CardResult{{CardID: "c1", Title: "old task"}}, nil)
	cardMgr.On("ArchiveCard", mock.Anything, "tok", "c1").Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "archive_card",
		CardName: "old task",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Card archived", result.Message)
}

func TestExecuteAction_DeleteCard(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "old task").
		Return([]port.CardResult{{CardID: "c1", Title: "old task"}}, nil)
	cardMgr.On("DeleteCard", mock.Anything, "tok", "c1").Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "delete_card",
		CardName: "old task",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Card deleted", result.Message)
}

func TestExecuteAction_AssignCard(t *testing.T) {
	uc, board, cardMgr, members, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "auth card").
		Return([]port.CardResult{{CardID: "c1", Title: "auth card"}}, nil)
	members.On("MatchMembers", mock.Anything, "tok", "board1", []string{"john"}).
		Return([]string{"m1"}, nil)
	cardMgr.On("UpdateCard", mock.Anything, "tok", "c1", mock.Anything).Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "assign_card",
		CardName: "auth card",
		Members:  []string{"john"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Card assigned to john", result.Message)
}

func TestExecuteAction_SetDueDate(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "my card").
		Return([]port.CardResult{{CardID: "c1", Title: "my card"}}, nil)

	due := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	cardMgr.On("UpdateCard", mock.Anything, "tok", "c1", mock.Anything).Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "set_due_date",
		CardName: "my card",
		DueDate:  &due,
	})

	assert.NoError(t, err)
	assert.Contains(t, result.Message, "Mar 15, 2026")
}

func TestExecuteAction_ListLists(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("GetLists", mock.Anything, "tok", "board1").
		Return([]port.ListInfo{
			{ID: "l1", Name: "To Do"},
			{ID: "l2", Name: "Done"},
		}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action: "list_lists",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Board lists", result.Message)
	assert.Len(t, result.Items, 2)
}

func TestExecuteAction_ListLabels(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("GetLabels", mock.Anything, "tok", "board1").
		Return([]port.LabelInfo{
			{ID: "lb1", Name: "Bug", Color: "red"},
		}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action: "list_labels",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Board labels", result.Message)
	assert.Len(t, result.Items, 1)
}

func TestExecuteAction_CreateList(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("CreateList", mock.Anything, "tok", "board1", "Review").
		Return(&port.ListInfo{ID: "l3", Name: "Review"}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "create_list",
		ListName: "Review",
	})

	assert.NoError(t, err)
	assert.Contains(t, result.Message, "Review")
}

func TestExecuteAction_CardNotFound(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "nonexistent").
		Return([]port.CardResult{}, nil)

	_, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "get_card",
		CardName: "nonexistent",
	})

	assert.ErrorIs(t, err, domainerror.ErrCardNotFound)
}

func TestExecuteAction_AddLabel(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "my card").
		Return([]port.CardResult{{CardID: "c1", Title: "my card"}}, nil)
	board.On("MatchLabels", mock.Anything, "tok", "board1", []string{"Bug"}).
		Return([]string{"lb1"}, nil)
	cardMgr.On("GetCard", mock.Anything, "tok", "c1").
		Return(&port.CardInfo{ID: "c1", Title: "my card", Labels: []string{}}, nil)
	cardMgr.On("UpdateCard", mock.Anything, "tok", "c1", mock.Anything).Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:    "add_label",
		CardName:  "my card",
		LabelName: "Bug",
	})

	assert.NoError(t, err)
	assert.Contains(t, result.Message, "Bug")
}

func TestExecuteAction_MoveCard_ListNotFound(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "card").
		Return([]port.CardResult{{CardID: "c1", Title: "card"}}, nil)
	board.On("GetLists", mock.Anything, "tok", "board1").
		Return([]port.ListInfo{{ID: "l1", Name: "To Do"}}, nil)

	_, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "move_card",
		CardName: "card",
		ListName: "Nonexistent",
	})

	assert.ErrorIs(t, err, domainerror.ErrActionFailed)
}

func TestExecuteAction_UpdateCard(t *testing.T) {
	uc, board, cardMgr, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "my card").
		Return([]port.CardResult{{CardID: "c1", Title: "my card"}}, nil)
	cardMgr.On("UpdateCard", mock.Anything, "tok", "c1", mock.Anything).Return(nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:      "update_card",
		CardName:    "my card",
		Title:       "new title",
		Description: "new desc",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Card updated", result.Message)
}

func TestExecuteAction_SearchCards_NoResults(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "xyz").
		Return([]port.CardResult{}, nil)

	result, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:      "search_cards",
		SearchQuery: "xyz",
	})

	assert.NoError(t, err)
	assert.Equal(t, "No cards found", result.Message)
}

func TestExecuteAction_AssignCard_NoMembers(t *testing.T) {
	uc, board, _, members, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "card").
		Return([]port.CardResult{{CardID: "c1", Title: "card"}}, nil)
	members.On("MatchMembers", mock.Anything, "tok", "board1", []string{"nobody"}).
		Return([]string{}, nil)

	_, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "assign_card",
		CardName: "card",
		Members:  []string{"nobody"},
	})

	assert.ErrorIs(t, err, domainerror.ErrActionFailed)
}

func TestExecuteAction_SetDueDate_NoDueDate(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "card").
		Return([]port.CardResult{{CardID: "c1", Title: "card"}}, nil)

	_, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "set_due_date",
		CardName: "card",
	})

	assert.ErrorIs(t, err, domainerror.ErrActionFailed)
}

func TestExecuteAction_MoveCard_SearchError(t *testing.T) {
	uc, board, _, _, _, _ := setupExecuteAction()

	board.On("SearchCards", mock.Anything, "tok", "board1", "card").
		Return(nil, errors.New("network error"))

	_, err := uc.Execute(context.Background(), 123, &dto.IntentOutput{
		Action:   "move_card",
		CardName: "card",
	})

	assert.Error(t, err)
}
