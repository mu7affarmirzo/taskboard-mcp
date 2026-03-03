package port

import "context"

type CardInfo struct {
	ID          string
	Title       string
	Description string
	URL         string
	ListID      string
	Due         string
	Labels      []string
	Members     []string
}

type UpdateCardParams struct {
	Name      *string
	Desc      *string
	IDList    *string
	Due       *string
	IDLabels  *string
	IDMembers *string
}

type CardManager interface {
	GetCard(ctx context.Context, token string, cardID string) (*CardInfo, error)
	UpdateCard(ctx context.Context, token string, cardID string, params UpdateCardParams) error
	ArchiveCard(ctx context.Context, token string, cardID string) error
	DeleteCard(ctx context.Context, token string, cardID string) error
	AddComment(ctx context.Context, token string, cardID string, text string) error
}
