package port

import "context"

type InitDataResult struct {
	TelegramID int64
	FirstName  string
	Username   string
	AuthDate   int64
}

type InitDataValidator interface {
	Validate(ctx context.Context, initData string) (*InitDataResult, error)
}
