package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const claudeBaseURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: httpClient,
	}
}

type messageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messageResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) SendMessage(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	reqBody := messageRequest{
		Model:     c.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []message{
			{Role: "user", Content: userMessage},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", claudeBaseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API returned status %d", resp.StatusCode)
	}

	var msgResp messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(msgResp.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}

	return msgResp.Content[0].Text, nil
}
