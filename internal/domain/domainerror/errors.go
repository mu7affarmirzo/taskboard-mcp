package domainerror

import "errors"

var (
	ErrEmptyTaskTitle  = errors.New("task title cannot be empty")
	ErrInvalidPriority = errors.New("invalid priority value")
	ErrBoardNotSet     = errors.New("default board not configured")
	ErrListNotSet      = errors.New("default list not configured")
	ErrUserNotFound    = errors.New("user not found")
	ErrParsingFailed      = errors.New("failed to parse message into task")
	ErrCardCreation       = errors.New("failed to create card on board")
	ErrEmptyTrelloToken   = errors.New("trello token cannot be empty")
	ErrTrelloNotConnected = errors.New("trello account not connected")
)
