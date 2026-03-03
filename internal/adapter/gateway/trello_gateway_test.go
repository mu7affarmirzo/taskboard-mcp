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

	// Compile-time check: TrelloGateway satisfies port.TaskBoard and port.MemberResolver
	var _ port.TaskBoard = gw
	var _ port.MemberResolver = gw
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

func TestTrelloGateway_MatchMembers_ExactUsername(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		members := []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			FullName string `json:"fullName"`
		}{
			{ID: "m1", Username: "john", FullName: "John Doe"},
			{ID: "m2", Username: "jane", FullName: "Jane Smith"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(members))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchMembers(context.Background(), "token", "board-1", []string{"john"})

	require.NoError(t, err)
	assert.Equal(t, []string{"m1"}, ids)
}

func TestTrelloGateway_MatchMembers_FullNameMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		members := []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			FullName string `json:"fullName"`
		}{
			{ID: "m1", Username: "jdoe", FullName: "John Doe"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(members))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchMembers(context.Background(), "token", "board-1", []string{"john doe"})

	require.NoError(t, err)
	assert.Equal(t, []string{"m1"}, ids)
}

func TestTrelloGateway_MatchMembers_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		members := []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			FullName string `json:"fullName"`
		}{
			{ID: "m1", Username: "John", FullName: "John Doe"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(members))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchMembers(context.Background(), "token", "board-1", []string{"JOHN"})

	require.NoError(t, err)
	assert.Equal(t, []string{"m1"}, ids)
}

func TestTrelloGateway_MatchMembers_NoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		members := []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			FullName string `json:"fullName"`
		}{
			{ID: "m1", Username: "john", FullName: "John Doe"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(members))
	}))
	defer server.Close()

	gw := gateway.NewTrelloGatewayWithURL(server.URL, "test-key")
	ids, err := gw.MatchMembers(context.Background(), "token", "board-1", []string{"alice"})

	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestTrelloGateway_MatchMembers_EmptyInput(t *testing.T) {
	gw := gateway.NewTrelloGateway(trello.NewClient("key", nil))
	ids, err := gw.MatchMembers(context.Background(), "token", "board-1", nil)

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
