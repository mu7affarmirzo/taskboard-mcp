package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type AuthenticateUseCase struct {
	validator  port.InitDataValidator
	session    port.SessionManager
	userRepo   port.UserRepository
}

func NewAuthenticateUseCase(
	validator port.InitDataValidator,
	session port.SessionManager,
	userRepo port.UserRepository,
) *AuthenticateUseCase {
	return &AuthenticateUseCase{
		validator: validator,
		session:   session,
		userRepo:  userRepo,
	}
}

func (uc *AuthenticateUseCase) Execute(ctx context.Context, input dto.AuthInput) (*dto.AuthOutput, error) {
	result, err := uc.validator.Validate(ctx, input.InitData)
	if err != nil {
		return nil, fmt.Errorf("validate init data: %w", err)
	}

	// Find or create user
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(result.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	token, err := uc.session.CreateToken(ctx, port.SessionClaims{
		TelegramID: result.TelegramID,
		Username:   result.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &dto.AuthOutput{
		Token:     token,
		UserID:    user.TelegramID().Int64(),
		FirstName: result.FirstName,
		Username:  result.Username,
	}, nil
}
