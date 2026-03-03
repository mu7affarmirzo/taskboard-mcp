package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ListMembersUseCase struct {
	memberResolver port.MemberResolver
	userRepo       port.UserRepository
}

func NewListMembersUseCase(memberResolver port.MemberResolver, userRepo port.UserRepository) *ListMembersUseCase {
	return &ListMembersUseCase{memberResolver: memberResolver, userRepo: userRepo}
}

func (uc *ListMembersUseCase) Execute(ctx context.Context, telegramID int64, boardID string) (*dto.ListMembersOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	members, err := uc.memberResolver.GetMembers(ctx, user.TrelloToken(), boardID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	items := make([]dto.MemberItem, len(members))
	for i, m := range members {
		items[i] = dto.MemberItem{ID: m.ID, Username: m.Username, FullName: m.FullName}
	}
	return &dto.ListMembersOutput{Members: items}, nil
}
