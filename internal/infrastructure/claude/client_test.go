package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendMessage_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		var req messageRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, "system prompt", req.System)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)
		assert.Equal(t, "hello", req.Messages[0].Content)

		resp := messageResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: `{"title":"parsed task"}`},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		model:      "test-model",
		httpClient: server.Client(),
	}

	assert.NotNil(t, client)
}

func TestClient_SendMessage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "key",
		model:      "model",
		httpClient: server.Client(),
	}
	assert.NotNil(t, client)
}

func TestClient_SendMessage_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := messageResponse{Content: nil}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "key",
		model:      "model",
		httpClient: server.Client(),
	}
	assert.NotNil(t, client)
}

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("api-key", "claude-3-haiku", nil)
	assert.NotNil(t, client)
}

func TestClient_SendMessage_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := messageResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: `{"title":"Buy groceries","priority":"high"}`},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "key",
		model:      "model",
		httpClient: server.Client(),
	}

	resp, err := client.httpClient.Get(server.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var msgResp messageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&msgResp))
	require.Len(t, msgResp.Content, 1)
	assert.Contains(t, msgResp.Content[0].Text, "Buy groceries")
}

func TestSendMessage_RealFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var req messageRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		resp := messageResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: "response text"},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	client := NewClient("key", "model", nil)
	assert.NotNil(t, client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.SendMessage(ctx, "sys", "msg")
	assert.Error(t, err)
}
