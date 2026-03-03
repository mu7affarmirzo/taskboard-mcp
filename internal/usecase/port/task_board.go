package port

import "context"

type BoardInfo struct {
	ID   string
	Name string
}

type ListInfo struct {
	ID   string
	Name string
}

type LabelInfo struct {
	ID    string
	Name  string
	Color string
}

type CardResult struct {
	CardID  string
	CardURL string
	Title   string
	ListID  string
}

type CreateCardParams struct {
	ListID      string
	Title       string
	Description string
	DueDate     *string
	LabelIDs    []string
	MemberIDs   []string
	Position    string
}

type TaskBoard interface {
	GetBoards(ctx context.Context, token string) ([]BoardInfo, error)
	GetLists(ctx context.Context, token string, boardID string) ([]ListInfo, error)
	GetLabels(ctx context.Context, token string, boardID string) ([]LabelInfo, error)
	MatchLabels(ctx context.Context, token string, boardID string, names []string) ([]string, error)
	CreateCard(ctx context.Context, token string, params CreateCardParams) (*CardResult, error)
	SearchCards(ctx context.Context, token string, boardID string, query string) ([]CardResult, error)
	GetCards(ctx context.Context, token string, listID string) ([]CardResult, error)
	CreateList(ctx context.Context, token string, boardID string, name string) (*ListInfo, error)
}
