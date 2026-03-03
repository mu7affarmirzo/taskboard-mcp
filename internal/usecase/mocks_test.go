package usecase_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type MockParser struct{ mock.Mock }

func (m *MockParser) Parse(ctx context.Context, msg string) (*entity.Task, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

type MockBoard struct{ mock.Mock }

func (m *MockBoard) GetBoards(ctx context.Context, token string) ([]port.BoardInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.BoardInfo), args.Error(1)
}

func (m *MockBoard) GetLists(ctx context.Context, token string, boardID string) ([]port.ListInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.ListInfo), args.Error(1)
}

func (m *MockBoard) GetLabels(ctx context.Context, token string, boardID string) ([]port.LabelInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.LabelInfo), args.Error(1)
}

func (m *MockBoard) MatchLabels(ctx context.Context, token string, boardID string, names []string) ([]string, error) {
	args := m.Called(ctx, token, boardID, names)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockBoard) CreateCard(ctx context.Context, token string, p port.CreateCardParams) (*port.CardResult, error) {
	args := m.Called(ctx, token, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.CardResult), args.Error(1)
}

func (m *MockBoard) SearchCards(ctx context.Context, token string, boardID string, query string) ([]port.CardResult, error) {
	args := m.Called(ctx, token, boardID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.CardResult), args.Error(1)
}

func (m *MockBoard) GetCards(ctx context.Context, token string, listID string) ([]port.CardResult, error) {
	args := m.Called(ctx, token, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.CardResult), args.Error(1)
}

func (m *MockBoard) CreateList(ctx context.Context, token string, boardID string, name string) (*port.ListInfo, error) {
	args := m.Called(ctx, token, boardID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ListInfo), args.Error(1)
}

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) Save(ctx context.Context, user *entity.User) error {
	return m.Called(ctx, user).Error(0)
}

type MockMemberResolver struct{ mock.Mock }

func (m *MockMemberResolver) GetMembers(ctx context.Context, token string, boardID string) ([]port.MemberInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.MemberInfo), args.Error(1)
}

func (m *MockMemberResolver) MatchMembers(ctx context.Context, token string, boardID string, names []string) ([]string, error) {
	args := m.Called(ctx, token, boardID, names)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

type MockTaskLog struct{ mock.Mock }

func (m *MockTaskLog) Log(ctx context.Context, entry port.TaskLogEntry) error {
	return m.Called(ctx, entry).Error(0)
}

type MockIntentParser struct{ mock.Mock }

func (m *MockIntentParser) ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	args := m.Called(ctx, rawMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.IntentOutput), args.Error(1)
}

type MockCardManager struct{ mock.Mock }

func (m *MockCardManager) GetCard(ctx context.Context, token string, cardID string) (*port.CardInfo, error) {
	args := m.Called(ctx, token, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.CardInfo), args.Error(1)
}

func (m *MockCardManager) UpdateCard(ctx context.Context, token string, cardID string, params port.UpdateCardParams) error {
	return m.Called(ctx, token, cardID, params).Error(0)
}

func (m *MockCardManager) ArchiveCard(ctx context.Context, token string, cardID string) error {
	return m.Called(ctx, token, cardID).Error(0)
}

func (m *MockCardManager) DeleteCard(ctx context.Context, token string, cardID string) error {
	return m.Called(ctx, token, cardID).Error(0)
}

func (m *MockCardManager) AddComment(ctx context.Context, token string, cardID string, text string) error {
	return m.Called(ctx, token, cardID, text).Error(0)
}
