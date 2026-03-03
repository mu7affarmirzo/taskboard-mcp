package usecase

import (
	"context"
	"errors"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type RegisterUserUseCase struct {
	userRepo    port.UserRepository
	trelloAPIKey string
}

func NewRegisterUserUseCase(userRepo port.UserRepository, trelloAPIKey string) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:    userRepo,
		trelloAPIKey: trelloAPIKey,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input dto.RegisterUserInput) (*dto.RegisterUserOutput, error) {
	tid := valueobject.TelegramID(input.TelegramID)

	isNew := false
	_, err := uc.userRepo.FindByTelegramID(ctx, tid)
	if err != nil {
		if !errors.Is(err, domainerror.ErrUserNotFound) {
			return nil, fmt.Errorf("find user: %w", err)
		}
		user := entity.NewUser(tid)
		if err := uc.userRepo.Save(ctx, user); err != nil {
			return nil, fmt.Errorf("save user: %w", err)
		}
		isNew = true
	}

	authURL := fmt.Sprintf(
		"https://trello.com/1/authorize?expiration=never&scope=read,write&response_type=token&key=%s&name=Telegram+Trello+Bot",
		uc.trelloAPIKey,
	)

	return &dto.RegisterUserOutput{
		IsNewUser:    isNew,
		TrelloAuthURL: authURL,
	}, nil
}
