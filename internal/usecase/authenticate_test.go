package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type MockInitDataValidator struct{ mock.Mock }

func (m *MockInitDataValidator) Validate(ctx context.Context, initData string) (*port.InitDataResult, error) {
	args := m.Called(ctx, initData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.InitDataResult), args.Error(1)
}

type MockSessionManager struct{ mock.Mock }

func (m *MockSessionManager) CreateToken(ctx context.Context, claims port.SessionClaims) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func (m *MockSessionManager) ValidateToken(ctx context.Context, token string) (*port.SessionClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.SessionClaims), args.Error(1)
}

func TestAuthenticate_HappyPath(t *testing.T) {
	validator := new(MockInitDataValidator)
	session := new(MockSessionManager)
	userRepo := new(MockUserRepo)

	uc := usecase.NewAuthenticateUseCase(validator, session, userRepo)

	validator.On("Validate", mock.Anything, "init-data-raw").Return(&port.InitDataResult{
		TelegramID: 12345,
		FirstName:  "John",
		Username:   "johndoe",
		AuthDate:   1700000000,
	}, nil)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	session.On("CreateToken", mock.Anything, port.SessionClaims{
		TelegramID: 12345,
		Username:   "johndoe",
	}).Return("jwt-token-abc", nil)

	output, err := uc.Execute(context.Background(), dto.AuthInput{InitData: "init-data-raw"})

	assert.NoError(t, err)
	assert.Equal(t, "jwt-token-abc", output.Token)
	assert.Equal(t, int64(12345), output.UserID)
	assert.Equal(t, "John", output.FirstName)
	assert.Equal(t, "johndoe", output.Username)
}

func TestAuthenticate_InvalidInitData(t *testing.T) {
	validator := new(MockInitDataValidator)
	session := new(MockSessionManager)
	userRepo := new(MockUserRepo)

	uc := usecase.NewAuthenticateUseCase(validator, session, userRepo)

	validator.On("Validate", mock.Anything, "bad-data").Return(nil, errors.New("invalid"))

	output, err := uc.Execute(context.Background(), dto.AuthInput{InitData: "bad-data"})

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "validate init data")
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	validator := new(MockInitDataValidator)
	session := new(MockSessionManager)
	userRepo := new(MockUserRepo)

	uc := usecase.NewAuthenticateUseCase(validator, session, userRepo)

	validator.On("Validate", mock.Anything, "init-data").Return(&port.InitDataResult{
		TelegramID: 99999,
		FirstName:  "Jane",
		Username:   "jane",
		AuthDate:   1700000000,
	}, nil)

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(99999)).Return(nil, errors.New("not found"))

	output, err := uc.Execute(context.Background(), dto.AuthInput{InitData: "init-data"})

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "find user")
}

func TestAuthenticate_TokenCreationFails(t *testing.T) {
	validator := new(MockInitDataValidator)
	session := new(MockSessionManager)
	userRepo := new(MockUserRepo)

	uc := usecase.NewAuthenticateUseCase(validator, session, userRepo)

	validator.On("Validate", mock.Anything, "init-data").Return(&port.InitDataResult{
		TelegramID: 12345,
		FirstName:  "John",
		Username:   "johndoe",
		AuthDate:   1700000000,
	}, nil)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	session.On("CreateToken", mock.Anything, port.SessionClaims{
		TelegramID: 12345,
		Username:   "johndoe",
	}).Return("", errors.New("signing failed"))

	output, err := uc.Execute(context.Background(), dto.AuthInput{InitData: "init-data"})

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "create token")
}
