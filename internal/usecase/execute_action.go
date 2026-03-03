package usecase

import (
	"context"
	"fmt"
	"strings"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ExecuteActionUseCase struct {
	board       port.TaskBoard
	cardManager port.CardManager
	members     port.MemberResolver
	userRepo    port.UserRepository
	taskLog     port.TaskLogRepository
}

func NewExecuteActionUseCase(
	board port.TaskBoard,
	cardManager port.CardManager,
	members port.MemberResolver,
	userRepo port.UserRepository,
	taskLog port.TaskLogRepository,
) *ExecuteActionUseCase {
	return &ExecuteActionUseCase{
		board:       board,
		cardManager: cardManager,
		members:     members,
		userRepo:    userRepo,
		taskLog:     taskLog,
	}
}

func (uc *ExecuteActionUseCase) Execute(
	ctx context.Context,
	telegramID int64,
	intent *dto.IntentOutput,
) (*dto.ActionOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	token := user.TrelloToken()
	boardID := user.DefaultBoard()

	action := valueobject.Action(intent.Action)

	if action.NeedsCard() && intent.CardID == "" {
		cardID, err := uc.resolveCard(ctx, token, boardID, intent.CardName)
		if err != nil {
			return nil, err
		}
		intent.CardID = cardID
	}

	switch action {
	case valueobject.ActionMoveCard:
		return uc.moveCard(ctx, token, boardID, intent)
	case valueobject.ActionUpdateCard:
		return uc.updateCard(ctx, token, intent)
	case valueobject.ActionGetCard:
		return uc.getCard(ctx, token, intent)
	case valueobject.ActionListCards:
		return uc.listCards(ctx, token, boardID, intent)
	case valueobject.ActionListLists:
		return uc.listLists(ctx, token, boardID)
	case valueobject.ActionListLabels:
		return uc.listLabels(ctx, token, boardID)
	case valueobject.ActionCreateList:
		return uc.createList(ctx, token, boardID, intent)
	case valueobject.ActionArchiveCard:
		return uc.archiveCard(ctx, token, intent)
	case valueobject.ActionDeleteCard:
		return uc.deleteCard(ctx, token, intent)
	case valueobject.ActionAddComment:
		return uc.addComment(ctx, token, intent)
	case valueobject.ActionAssignCard:
		return uc.assignCard(ctx, token, boardID, intent)
	case valueobject.ActionSearchCards:
		return uc.searchCards(ctx, token, boardID, intent)
	case valueobject.ActionAddLabel:
		return uc.addLabel(ctx, token, boardID, intent)
	case valueobject.ActionSetDueDate:
		return uc.setDueDate(ctx, token, intent)
	default:
		return nil, fmt.Errorf("%w: %s", domainerror.ErrUnknownAction, intent.Action)
	}
}

func (uc *ExecuteActionUseCase) resolveCard(ctx context.Context, token, boardID, cardName string) (string, error) {
	if cardName == "" {
		return "", domainerror.ErrCardNotFound
	}
	cards, err := uc.board.SearchCards(ctx, token, boardID, cardName)
	if err != nil {
		return "", fmt.Errorf("search cards: %w", err)
	}
	if len(cards) == 0 {
		return "", fmt.Errorf("%w: %q", domainerror.ErrCardNotFound, cardName)
	}
	nameLower := strings.ToLower(cardName)
	for _, c := range cards {
		if strings.ToLower(c.Title) == nameLower {
			return c.CardID, nil
		}
	}
	return cards[0].CardID, nil
}

func (uc *ExecuteActionUseCase) moveCard(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	lists, err := uc.board.GetLists(ctx, token, boardID)
	if err != nil {
		return nil, fmt.Errorf("get lists: %w", err)
	}

	var targetListID string
	listNameLower := strings.ToLower(intent.ListName)
	for _, l := range lists {
		if strings.ToLower(l.Name) == listNameLower || strings.Contains(strings.ToLower(l.Name), listNameLower) {
			targetListID = l.ID
			break
		}
	}
	if targetListID == "" {
		return nil, fmt.Errorf("%w: list %q not found", domainerror.ErrActionFailed, intent.ListName)
	}

	err = uc.cardManager.UpdateCard(ctx, token, intent.CardID, port.UpdateCardParams{
		IDList: &targetListID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("Card moved to %s", intent.ListName),
	}, nil
}

func (uc *ExecuteActionUseCase) updateCard(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	params := port.UpdateCardParams{}
	if intent.Title != "" {
		params.Name = &intent.Title
	}
	if intent.Description != "" {
		params.Desc = &intent.Description
	}

	err := uc.cardManager.UpdateCard(ctx, token, intent.CardID, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: "Card updated",
	}, nil
}

func (uc *ExecuteActionUseCase) getCard(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	card, err := uc.cardManager.GetCard(ctx, token, intent.CardID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: card.Title,
		CardURL: card.URL,
		Items: []dto.ActionItem{{
			ID:    card.ID,
			Name:  card.Title,
			URL:   card.URL,
			Extra: card.Description,
		}},
	}, nil
}

func (uc *ExecuteActionUseCase) listCards(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	lists, err := uc.board.GetLists(ctx, token, boardID)
	if err != nil {
		return nil, fmt.Errorf("get lists: %w", err)
	}

	var targetListID string
	listNameLower := strings.ToLower(intent.ListName)
	for _, l := range lists {
		if strings.ToLower(l.Name) == listNameLower || strings.Contains(strings.ToLower(l.Name), listNameLower) {
			targetListID = l.ID
			break
		}
	}
	if targetListID == "" {
		return nil, fmt.Errorf("%w: list %q not found", domainerror.ErrActionFailed, intent.ListName)
	}

	cards, err := uc.board.GetCards(ctx, token, targetListID)
	if err != nil {
		return nil, fmt.Errorf("get cards: %w", err)
	}

	items := make([]dto.ActionItem, len(cards))
	for i, c := range cards {
		items[i] = dto.ActionItem{
			ID:   c.CardID,
			Name: c.Title,
			URL:  c.CardURL,
		}
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("Cards in %s", intent.ListName),
		Items:   items,
	}, nil
}

func (uc *ExecuteActionUseCase) listLists(ctx context.Context, token, boardID string) (*dto.ActionOutput, error) {
	lists, err := uc.board.GetLists(ctx, token, boardID)
	if err != nil {
		return nil, fmt.Errorf("get lists: %w", err)
	}

	items := make([]dto.ActionItem, len(lists))
	for i, l := range lists {
		items[i] = dto.ActionItem{
			ID:   l.ID,
			Name: l.Name,
		}
	}

	return &dto.ActionOutput{
		Action:  string(valueobject.ActionListLists),
		Message: "Board lists",
		Items:   items,
	}, nil
}

func (uc *ExecuteActionUseCase) listLabels(ctx context.Context, token, boardID string) (*dto.ActionOutput, error) {
	labels, err := uc.board.GetLabels(ctx, token, boardID)
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}

	items := make([]dto.ActionItem, len(labels))
	for i, l := range labels {
		items[i] = dto.ActionItem{
			ID:    l.ID,
			Name:  l.Name,
			Extra: l.Color,
		}
	}

	return &dto.ActionOutput{
		Action:  string(valueobject.ActionListLabels),
		Message: "Board labels",
		Items:   items,
	}, nil
}

func (uc *ExecuteActionUseCase) createList(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	list, err := uc.board.CreateList(ctx, token, boardID, intent.ListName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("List %q created", list.Name),
	}, nil
}

func (uc *ExecuteActionUseCase) archiveCard(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	err := uc.cardManager.ArchiveCard(ctx, token, intent.CardID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: "Card archived",
	}, nil
}

func (uc *ExecuteActionUseCase) deleteCard(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	err := uc.cardManager.DeleteCard(ctx, token, intent.CardID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: "Card deleted",
	}, nil
}

func (uc *ExecuteActionUseCase) addComment(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	err := uc.cardManager.AddComment(ctx, token, intent.CardID, intent.CommentText)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: "Comment added",
	}, nil
}

func (uc *ExecuteActionUseCase) assignCard(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	memberIDs, err := uc.members.MatchMembers(ctx, token, boardID, intent.Members)
	if err != nil {
		return nil, fmt.Errorf("match members: %w", err)
	}
	if len(memberIDs) == 0 {
		return nil, fmt.Errorf("%w: no matching members found", domainerror.ErrActionFailed)
	}

	idMembers := strings.Join(memberIDs, ",")
	err = uc.cardManager.UpdateCard(ctx, token, intent.CardID, port.UpdateCardParams{
		IDMembers: &idMembers,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("Card assigned to %s", strings.Join(intent.Members, ", ")),
	}, nil
}

func (uc *ExecuteActionUseCase) searchCards(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	cards, err := uc.board.SearchCards(ctx, token, boardID, intent.SearchQuery)
	if err != nil {
		return nil, fmt.Errorf("search cards: %w", err)
	}

	items := make([]dto.ActionItem, len(cards))
	for i, c := range cards {
		items[i] = dto.ActionItem{
			ID:   c.CardID,
			Name: c.Title,
			URL:  c.CardURL,
		}
	}

	msg := fmt.Sprintf("Found %d card(s)", len(cards))
	if len(cards) == 0 {
		msg = "No cards found"
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: msg,
		Items:   items,
	}, nil
}

func (uc *ExecuteActionUseCase) addLabel(ctx context.Context, token, boardID string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	labelIDs, err := uc.board.MatchLabels(ctx, token, boardID, []string{intent.LabelName})
	if err != nil {
		return nil, fmt.Errorf("match labels: %w", err)
	}
	if len(labelIDs) == 0 {
		return nil, fmt.Errorf("%w: label %q not found", domainerror.ErrActionFailed, intent.LabelName)
	}

	card, err := uc.cardManager.GetCard(ctx, token, intent.CardID)
	if err != nil {
		return nil, fmt.Errorf("get card: %w", err)
	}

	allLabels := append(card.Labels, labelIDs...)
	idLabels := strings.Join(allLabels, ",")
	err = uc.cardManager.UpdateCard(ctx, token, intent.CardID, port.UpdateCardParams{
		IDLabels: &idLabels,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("Label %q added", intent.LabelName),
	}, nil
}

func (uc *ExecuteActionUseCase) setDueDate(ctx context.Context, token string, intent *dto.IntentOutput) (*dto.ActionOutput, error) {
	if intent.DueDate == nil {
		return nil, fmt.Errorf("%w: no due date provided", domainerror.ErrActionFailed)
	}
	due := intent.DueDate.Format("2006-01-02")
	err := uc.cardManager.UpdateCard(ctx, token, intent.CardID, port.UpdateCardParams{
		Due: &due,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrActionFailed, err)
	}

	return &dto.ActionOutput{
		Action:  intent.Action,
		Message: fmt.Sprintf("Due date set to %s", intent.DueDate.Format("Jan 2, 2006")),
	}, nil
}
