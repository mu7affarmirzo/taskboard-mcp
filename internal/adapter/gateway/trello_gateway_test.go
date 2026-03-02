package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/adapter/gateway"
	"telegram-trello-bot/internal/infrastructure/trello"
	"telegram-trello-bot/internal/usecase/port"
)

func TestTrelloGateway_ImplementsTaskBoard(t *testing.T) {
	client := trello.NewClient("test-key", nil)
	gw := gateway.NewTrelloGateway(client)

	// Compile-time check: TrelloGateway satisfies port.TaskBoard
	var _ port.TaskBoard = gw
}

func TestTrelloGateway_MatchLabels_ExactMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		}{
			{ID: "l1", Name: "Backend", Color: "blue"},
			{ID: "l2", Name: "Frontend", Color: "green"},
			{ID: "l3", Name: "Bug", Color: "red"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(labels))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchLabels(context.Background(), "token", "board-1", []string{"backend", "bug"})

	require.NoError(t, err)
	assert.Equal(t, []string{"l1", "l3"}, ids)
}

func TestTrelloGateway_MatchLabels_SubstringMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		}{
			{ID: "l1", Name: "Backend Development", Color: "blue"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(labels))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchLabels(context.Background(), "token", "board-1", []string{"backend"})

	require.NoError(t, err)
	assert.Equal(t, []string{"l1"}, ids)
}

func TestTrelloGateway_MatchLabels_NoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		}{
			{ID: "l1", Name: "Backend", Color: "blue"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(labels))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchLabels(context.Background(), "token", "board-1", []string{"design"})

	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestTrelloGateway_MatchLabels_EmptyNames(t *testing.T) {
	gw := gateway.NewTrelloGateway(trello.NewClient("key", nil))
	ids, err := gw.MatchLabels(context.Background(), "token", "board-1", nil)

	require.NoError(t, err)
	assert.Nil(t, ids)
}

func TestTrelloGateway_MatchLabels_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		}{
			{ID: "l1", Name: "URGENT", Color: "red"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(labels))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchLabels(context.Background(), "token", "board-1", []string{"urgent"})

	require.NoError(t, err)
	assert.Equal(t, []string{"l1"}, ids)
}
