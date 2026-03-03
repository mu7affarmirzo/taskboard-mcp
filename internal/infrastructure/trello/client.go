package trello

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.trello.com"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func NewClientWithURL(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *Client) GetBoards(ctx context.Context, token string) ([]BoardResponse, error) {
	endpoint := fmt.Sprintf("%s/1/members/me/boards?key=%s&token=%s&fields=id,name",
		c.baseURL, c.apiKey, token)
	var boards []BoardResponse
	if err := c.doGet(ctx, endpoint, &boards); err != nil {
		return nil, fmt.Errorf("get boards: %w", err)
	}
	return boards, nil
}

func (c *Client) GetLists(ctx context.Context, token string, boardID string) ([]ListResponse, error) {
	endpoint := fmt.Sprintf("%s/1/boards/%s/lists?key=%s&token=%s&fields=id,name",
		c.baseURL, boardID, c.apiKey, token)
	var lists []ListResponse
	if err := c.doGet(ctx, endpoint, &lists); err != nil {
		return nil, fmt.Errorf("get lists: %w", err)
	}
	return lists, nil
}

func (c *Client) GetLabels(ctx context.Context, token string, boardID string) ([]LabelResponse, error) {
	endpoint := fmt.Sprintf("%s/1/boards/%s/labels?key=%s&token=%s",
		c.baseURL, boardID, c.apiKey, token)
	var labels []LabelResponse
	if err := c.doGet(ctx, endpoint, &labels); err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}
	return labels, nil
}

func (c *Client) GetMembers(ctx context.Context, token string, boardID string) ([]MemberResponse, error) {
	endpoint := fmt.Sprintf("%s/1/boards/%s/members?key=%s&token=%s&fields=id,username,fullName",
		c.baseURL, boardID, c.apiKey, token)
	var members []MemberResponse
	if err := c.doGet(ctx, endpoint, &members); err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	return members, nil
}

func (c *Client) CreateCard(ctx context.Context, token string, req CreateCardRequest) (*CardResponse, error) {
	body := url.Values{}
	body.Set("key", c.apiKey)
	body.Set("token", token)
	body.Set("name", req.Name)
	body.Set("desc", req.Description)
	body.Set("idList", req.ListID)
	if req.Due != "" {
		body.Set("due", req.Due)
	}
	if len(req.LabelIDs) > 0 {
		body.Set("idLabels", strings.Join(req.LabelIDs, ","))
	}
	if len(req.MemberIDs) > 0 {
		body.Set("idMembers", strings.Join(req.MemberIDs, ","))
	}
	if req.Position != "" {
		body.Set("pos", req.Position)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/1/cards", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("trello API returned status %d", resp.StatusCode)
	}

	var card CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &card, nil
}

func (c *Client) GetCard(ctx context.Context, token string, cardID string) (*CardDetailResponse, error) {
	endpoint := fmt.Sprintf("%s/1/cards/%s?key=%s&token=%s&fields=id,name,desc,shortUrl,url,idList,due&members=true&member_fields=id,username,fullName",
		c.baseURL, cardID, c.apiKey, token)
	var card CardDetailResponse
	if err := c.doGet(ctx, endpoint, &card); err != nil {
		return nil, fmt.Errorf("get card: %w", err)
	}
	return &card, nil
}

func (c *Client) UpdateCard(ctx context.Context, token string, cardID string, req UpdateCardRequest) (*CardDetailResponse, error) {
	body := url.Values{}
	body.Set("key", c.apiKey)
	body.Set("token", token)
	if req.Name != nil {
		body.Set("name", *req.Name)
	}
	if req.Desc != nil {
		body.Set("desc", *req.Desc)
	}
	if req.IDList != nil {
		body.Set("idList", *req.IDList)
	}
	if req.Due != nil {
		body.Set("due", *req.Due)
	}
	if req.IDLabels != nil {
		body.Set("idLabels", *req.IDLabels)
	}
	if req.IDMembers != nil {
		body.Set("idMembers", *req.IDMembers)
	}

	var card CardDetailResponse
	if err := c.doPut(ctx, fmt.Sprintf("%s/1/cards/%s", c.baseURL, cardID), body, &card); err != nil {
		return nil, fmt.Errorf("update card: %w", err)
	}
	return &card, nil
}

func (c *Client) GetCards(ctx context.Context, token string, listID string) ([]CardDetailResponse, error) {
	endpoint := fmt.Sprintf("%s/1/lists/%s/cards?key=%s&token=%s&fields=id,name,desc,shortUrl,url,idList,due&members=true&member_fields=id,username,fullName",
		c.baseURL, listID, c.apiKey, token)
	var cards []CardDetailResponse
	if err := c.doGet(ctx, endpoint, &cards); err != nil {
		return nil, fmt.Errorf("get cards: %w", err)
	}
	return cards, nil
}

func (c *Client) doPut(ctx context.Context, rawURL string, body url.Values, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", rawURL, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) doGet(ctx context.Context, rawURL string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
