package port

import "context"

type MemberInfo struct {
	ID       string
	Username string
	FullName string
}

type MemberResolver interface {
	GetMembers(ctx context.Context, token string, boardID string) ([]MemberInfo, error)
	MatchMembers(ctx context.Context, token string, boardID string, names []string) ([]string, error)
}
