package gateway

import (
	"context"
	"strings"

	"telegram-trello-bot/internal/infrastructure/trello"
	"telegram-trello-bot/internal/usecase/port"
)

type TrelloGateway struct {
	client *trello.Client
}

func NewTrelloGateway(client *trello.Client) *TrelloGateway {
	return &TrelloGateway{client: client}
}

func NewTrelloGatewayWithURL(baseURL, apiKey string) *TrelloGateway {
	return &TrelloGateway{client: trello.NewClientWithURL(baseURL, apiKey, nil)}
}

func (g *TrelloGateway) GetBoards(ctx context.Context, token string) ([]port.BoardInfo, error) {
	boards, err := g.client.GetBoards(ctx, token)
	if err != nil {
		return nil, err
	}
	result := make([]port.BoardInfo, len(boards))
	for i, b := range boards {
		result[i] = port.BoardInfo{ID: b.ID, Name: b.Name}
	}
	return result, nil
}

func (g *TrelloGateway) GetLists(ctx context.Context, token string, boardID string) ([]port.ListInfo, error) {
	lists, err := g.client.GetLists(ctx, token, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]port.ListInfo, len(lists))
	for i, l := range lists {
		result[i] = port.ListInfo{ID: l.ID, Name: l.Name}
	}
	return result, nil
}

func (g *TrelloGateway) GetLabels(ctx context.Context, token string, boardID string) ([]port.LabelInfo, error) {
	labels, err := g.client.GetLabels(ctx, token, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]port.LabelInfo, len(labels))
	for i, l := range labels {
		result[i] = port.LabelInfo{ID: l.ID, Name: l.Name, Color: l.Color}
	}
	return result, nil
}

func (g *TrelloGateway) MatchLabels(ctx context.Context, token string, boardID string, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	labels, err := g.client.GetLabels(ctx, token, boardID)
	if err != nil {
		return nil, err
	}
	var matched []string
	for _, name := range names {
		nameLower := strings.ToLower(name)
		for _, l := range labels {
			labelLower := strings.ToLower(l.Name)
			if labelLower == nameLower || strings.Contains(labelLower, nameLower) || strings.Contains(nameLower, labelLower) {
				matched = append(matched, l.ID)
				break
			}
		}
	}
	return matched, nil
}

func (g *TrelloGateway) GetMembers(ctx context.Context, token string, boardID string) ([]port.MemberInfo, error) {
	members, err := g.client.GetMembers(ctx, token, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]port.MemberInfo, len(members))
	for i, m := range members {
		result[i] = port.MemberInfo{ID: m.ID, Username: m.Username, FullName: m.FullName}
	}
	return result, nil
}

func (g *TrelloGateway) MatchMembers(ctx context.Context, token string, boardID string, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	members, err := g.client.GetMembers(ctx, token, boardID)
	if err != nil {
		return nil, err
	}
	var matched []string
	for _, name := range names {
		nameLower := strings.ToLower(name)
		for _, m := range members {
			userLower := strings.ToLower(m.Username)
			fullLower := strings.ToLower(m.FullName)
			if userLower == nameLower || fullLower == nameLower ||
				strings.Contains(fullLower, nameLower) || strings.Contains(nameLower, userLower) {
				matched = append(matched, m.ID)
				break
			}
		}
	}
	return matched, nil
}

func (g *TrelloGateway) CreateCard(ctx context.Context, token string, params port.CreateCardParams) (*port.CardResult, error) {
	trelloParams := trello.CreateCardRequest{
		Name:        params.Title,
		Description: params.Description,
		ListID:      params.ListID,
		Position:    params.Position,
	}
	if params.DueDate != nil {
		trelloParams.Due = *params.DueDate
	}
	if len(params.LabelIDs) > 0 {
		trelloParams.LabelIDs = params.LabelIDs
	}
	if len(params.MemberIDs) > 0 {
		trelloParams.MemberIDs = params.MemberIDs
	}

	resp, err := g.client.CreateCard(ctx, token, trelloParams)
	if err != nil {
		return nil, err
	}
	return &port.CardResult{CardID: resp.ID, CardURL: resp.ShortURL}, nil
}

func (g *TrelloGateway) SearchCards(ctx context.Context, token string, boardID string, query string) ([]port.CardResult, error) {
	cards, err := g.client.SearchCards(ctx, token, boardID, query)
	if err != nil {
		return nil, err
	}
	result := make([]port.CardResult, len(cards))
	for i, c := range cards {
		result[i] = port.CardResult{
			CardID:  c.ID,
			CardURL: c.ShortURL,
			Title:   c.Name,
			ListID:  c.IDList,
		}
	}
	return result, nil
}

func (g *TrelloGateway) GetCards(ctx context.Context, token string, listID string) ([]port.CardResult, error) {
	cards, err := g.client.GetCards(ctx, token, listID)
	if err != nil {
		return nil, err
	}
	result := make([]port.CardResult, len(cards))
	for i, c := range cards {
		result[i] = port.CardResult{
			CardID:  c.ID,
			CardURL: c.ShortURL,
			Title:   c.Name,
			ListID:  c.IDList,
		}
	}
	return result, nil
}

func (g *TrelloGateway) CreateList(ctx context.Context, token string, boardID string, name string) (*port.ListInfo, error) {
	list, err := g.client.CreateList(ctx, token, boardID, name)
	if err != nil {
		return nil, err
	}
	return &port.ListInfo{ID: list.ID, Name: list.Name}, nil
}

func (g *TrelloGateway) GetCard(ctx context.Context, token string, cardID string) (*port.CardInfo, error) {
	card, err := g.client.GetCard(ctx, token, cardID)
	if err != nil {
		return nil, err
	}
	labels := make([]string, len(card.Labels))
	for i, l := range card.Labels {
		labels[i] = l.ID
	}
	members := make([]string, len(card.Members))
	for i, m := range card.Members {
		members[i] = m.ID
	}
	return &port.CardInfo{
		ID:          card.ID,
		Title:       card.Name,
		Description: card.Desc,
		URL:         card.ShortURL,
		ListID:      card.IDList,
		Due:         card.Due,
		Labels:      labels,
		Members:     members,
	}, nil
}

func (g *TrelloGateway) UpdateCard(ctx context.Context, token string, cardID string, params port.UpdateCardParams) error {
	req := trello.UpdateCardRequest{
		Name:      params.Name,
		Desc:      params.Desc,
		IDList:    params.IDList,
		Due:       params.Due,
		IDLabels:  params.IDLabels,
		IDMembers: params.IDMembers,
	}
	_, err := g.client.UpdateCard(ctx, token, cardID, req)
	return err
}

func (g *TrelloGateway) ArchiveCard(ctx context.Context, token string, cardID string) error {
	return g.client.ArchiveCard(ctx, token, cardID)
}

func (g *TrelloGateway) DeleteCard(ctx context.Context, token string, cardID string) error {
	return g.client.DeleteCard(ctx, token, cardID)
}

func (g *TrelloGateway) AddComment(ctx context.Context, token string, cardID string, text string) error {
	_, err := g.client.AddComment(ctx, token, cardID, text)
	return err
}
