package trello

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_DoGet_Success(t *testing.T) {
	expected := []BoardResponse{
		{ID: "b1", Name: "Work"},
		{ID: "b2", Name: "Personal"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client()}

	var boards []BoardResponse
	err := client.doGet(context.Background(), server.URL, &boards)
	require.NoError(t, err)
	assert.Len(t, boards, 2)
	assert.Equal(t, "b1", boards[0].ID)
	assert.Equal(t, "Work", boards[0].Name)
}

func TestClient_DoGet_Lists(t *testing.T) {
	expected := []ListResponse{
		{ID: "l1", Name: "To Do"},
		{ID: "l2", Name: "Done"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}

	var lists []ListResponse
	err := client.doGet(context.Background(), server.URL, &lists)
	require.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "To Do", lists[0].Name)
}

func TestClient_DoGet_Labels(t *testing.T) {
	expected := []LabelResponse{
		{ID: "lb1", Name: "Bug", Color: "red"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}

	var labels []LabelResponse
	err := client.doGet(context.Background(), server.URL, &labels)
	require.NoError(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, "Bug", labels[0].Name)
	assert.Equal(t, "red", labels[0].Color)
}

func TestClient_DoGet_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}

	var result []BoardResponse
	err := client.doGet(context.Background(), server.URL, &result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_DoGet_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("not json"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}

	var result []BoardResponse
	err := client.doGet(context.Background(), server.URL, &result)
	assert.Error(t, err)
}

func TestClient_DoGet_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode([]BoardResponse{}))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result []BoardResponse
	err := client.doGet(ctx, server.URL, &result)
	assert.Error(t, err)
}

func TestNewClient(t *testing.T) {
	client := NewClient("my-key", nil)
	assert.NotNil(t, client)
}

func TestNewClientWithURL(t *testing.T) {
	client := NewClientWithURL("http://localhost:8080", "my-key", nil)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
}

func TestClient_GetBoards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/1/members/me/boards")
		boards := []BoardResponse{{ID: "b1", Name: "Work"}}
		require.NoError(t, json.NewEncoder(w).Encode(boards))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	boards, err := client.GetBoards(context.Background(), "token")
	require.NoError(t, err)
	assert.Len(t, boards, 1)
	assert.Equal(t, "Work", boards[0].Name)
}

func TestClient_GetLists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/1/boards/b1/lists")
		lists := []ListResponse{{ID: "l1", Name: "To Do"}}
		require.NoError(t, json.NewEncoder(w).Encode(lists))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	lists, err := client.GetLists(context.Background(), "token", "b1")
	require.NoError(t, err)
	assert.Len(t, lists, 1)
	assert.Equal(t, "To Do", lists[0].Name)
}

func TestClient_GetLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/1/boards/b1/labels")
		labels := []LabelResponse{{ID: "lb1", Name: "Bug", Color: "red"}}
		require.NoError(t, json.NewEncoder(w).Encode(labels))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	labels, err := client.GetLabels(context.Background(), "token", "b1")
	require.NoError(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, "Bug", labels[0].Name)
}

func TestClient_CreateCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/1/cards")
		card := CardResponse{ID: "c1", ShortURL: "https://trello.com/c/abc"}
		require.NoError(t, json.NewEncoder(w).Encode(card))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	card, err := client.CreateCard(context.Background(), "token", CreateCardRequest{
		Name:   "Test card",
		ListID: "list-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "c1", card.ID)
	assert.Equal(t, "https://trello.com/c/abc", card.ShortURL)
}

func TestClient_GetMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/1/boards/b1/members")
		members := []MemberResponse{
			{ID: "m1", Username: "john", FullName: "John Doe"},
			{ID: "m2", Username: "jane", FullName: "Jane Smith"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(members))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	members, err := client.GetMembers(context.Background(), "token", "b1")
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "john", members[0].Username)
	assert.Equal(t, "John Doe", members[0].FullName)
}

func TestClient_CreateCard_WithMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "m1,m2", r.FormValue("idMembers"))
		card := CardResponse{ID: "c1", ShortURL: "https://trello.com/c/abc"}
		require.NoError(t, json.NewEncoder(w).Encode(card))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	card, err := client.CreateCard(context.Background(), "token", CreateCardRequest{
		Name:      "Test card",
		ListID:    "list-1",
		MemberIDs: []string{"m1", "m2"},
	})
	require.NoError(t, err)
	assert.Equal(t, "c1", card.ID)
}

func TestCreateCardRequest_Fields(t *testing.T) {
	req := CreateCardRequest{
		Name:        "Fix bug",
		Description: "Critical fix",
		ListID:      "list-1",
		Due:         "2025-03-15",
		LabelIDs:    []string{"l1", "l2"},
		Position:    "top",
	}
	assert.Equal(t, "Fix bug", req.Name)
	assert.Equal(t, "list-1", req.ListID)
	assert.Equal(t, "top", req.Position)
	assert.Len(t, req.LabelIDs, 2)
}

func TestClient_GetCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/1/cards/c1")
		card := CardDetailResponse{
			ID:       "c1",
			Name:     "Test card",
			Desc:     "A description",
			ShortURL: "https://trello.com/c/abc",
			URL:      "https://trello.com/c/abc/1-test",
			IDList:   "list-1",
			Due:      "2025-04-01T00:00:00.000Z",
			Labels:   []LabelResponse{{ID: "lb1", Name: "Bug", Color: "red"}},
			Members:  []MemberResponse{{ID: "m1", Username: "john", FullName: "John Doe"}},
		}
		require.NoError(t, json.NewEncoder(w).Encode(card))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	card, err := client.GetCard(context.Background(), "token", "c1")
	require.NoError(t, err)
	assert.Equal(t, "c1", card.ID)
	assert.Equal(t, "Test card", card.Name)
	assert.Equal(t, "A description", card.Desc)
	assert.Equal(t, "list-1", card.IDList)
	assert.Len(t, card.Labels, 1)
	assert.Len(t, card.Members, 1)
}

func TestClient_GetCard_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	_, err := client.GetCard(context.Background(), "token", "bad-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get card")
}

func TestClient_UpdateCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/1/cards/c1")
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "Updated title", r.FormValue("name"))
		assert.Equal(t, "list-2", r.FormValue("idList"))
		card := CardDetailResponse{
			ID:     "c1",
			Name:   "Updated title",
			IDList: "list-2",
		}
		require.NoError(t, json.NewEncoder(w).Encode(card))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	name := "Updated title"
	listID := "list-2"
	card, err := client.UpdateCard(context.Background(), "token", "c1", UpdateCardRequest{
		Name:   &name,
		IDList: &listID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated title", card.Name)
	assert.Equal(t, "list-2", card.IDList)
}

func TestClient_UpdateCard_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	name := "fail"
	_, err := client.UpdateCard(context.Background(), "token", "c1", UpdateCardRequest{Name: &name})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update card")
}

func TestClient_GetCards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/1/lists/list-1/cards")
		cards := []CardDetailResponse{
			{ID: "c1", Name: "Card 1", IDList: "list-1"},
			{ID: "c2", Name: "Card 2", IDList: "list-1"},
		}
		require.NoError(t, json.NewEncoder(w).Encode(cards))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	cards, err := client.GetCards(context.Background(), "token", "list-1")
	require.NoError(t, err)
	assert.Len(t, cards, 2)
	assert.Equal(t, "Card 1", cards[0].Name)
	assert.Equal(t, "Card 2", cards[1].Name)
}

func TestClient_GetCards_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode([]CardDetailResponse{}))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "key", nil)
	cards, err := client.GetCards(context.Background(), "token", "list-1")
	require.NoError(t, err)
	assert.Empty(t, cards)
}
