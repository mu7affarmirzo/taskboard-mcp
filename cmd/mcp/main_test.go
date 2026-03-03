package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/infrastructure/trello"
)

func setupTestServer(t *testing.T, handler http.Handler) (*mcpServer, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	client := trello.NewClientWithURL(ts.URL, "test-key", nil)
	s := &mcpServer{
		client:  client,
		token:   "test-token",
		boardID: "board-1",
		listID:  "default-list",
	}
	return s, ts
}

func callTool(t *testing.T, s *mcpServer, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	raw, err := json.Marshal(args)
	require.NoError(t, err)
	req.Params.Arguments = make(map[string]any)
	require.NoError(t, json.Unmarshal(raw, &req.Params.Arguments))
	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	return result
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, result.Content)
	tc, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])
	return tc.Text
}

func TestCreateCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/1/cards" {
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardResponse{
				ID:       "card-1",
				ShortURL: "https://trello.com/c/abc",
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleCreateCard, map[string]any{
		"title":       "New feature",
		"description": "Implement login",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "card-1")
	assert.Contains(t, text, "https://trello.com/c/abc")
}

func TestCreateCard_WithListName(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/1/boards/board-1/lists":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.ListResponse{
				{ID: "list-todo", Name: "To Do"},
				{ID: "list-done", Name: "Done"},
			}))
		case r.Method == "POST" && r.URL.Path == "/1/cards":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "list-todo", r.FormValue("idList"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardResponse{
				ID:       "card-2",
				ShortURL: "https://trello.com/c/def",
			}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleCreateCard, map[string]any{
		"title":     "Task",
		"list_name": "To Do",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "card-2")
}

func TestCreateCard_MissingTitle(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	result := callTool(t, s, s.handleCreateCard, map[string]any{})
	assert.True(t, result.IsError)
}

func TestCreateCard_NoListConfigured(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	s.listID = ""

	result := callTool(t, s, s.handleCreateCard, map[string]any{
		"title": "No list",
	})
	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "no list specified")
}

func TestMoveCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/1/boards/board-1/lists":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.ListResponse{
				{ID: "list-testing", Name: "Testing"},
			}))
		case r.Method == "PUT" && r.URL.Path == "/1/cards/card-1":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "list-testing", r.FormValue("idList"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{
				ID:     "card-1",
				IDList: "list-testing",
			}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleMoveCard, map[string]any{
		"card_id":   "card-1",
		"list_name": "Testing",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Testing")
}

func TestMoveCard_ListNotFound(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/boards/board-1/lists" {
			require.NoError(t, json.NewEncoder(w).Encode([]trello.ListResponse{
				{ID: "list-1", Name: "To Do"},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleMoveCard, map[string]any{
		"card_id":   "card-1",
		"list_name": "Nonexistent",
	})

	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "not found")
}

func TestUpdateCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/1/cards/card-1" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "Updated title", r.FormValue("name"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{
				ID:       "card-1",
				Name:     "Updated title",
				ShortURL: "https://trello.com/c/abc",
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleUpdateCard, map[string]any{
		"card_id": "card-1",
		"title":   "Updated title",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Updated title")
}

func TestUpdateCard_NoFields(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	result := callTool(t, s, s.handleUpdateCard, map[string]any{
		"card_id": "card-1",
	})

	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "no fields to update")
}

func TestGetCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/cards/card-1" {
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{
				ID:       "card-1",
				Name:     "My Card",
				Desc:     "Description here",
				ShortURL: "https://trello.com/c/abc",
				IDList:   "list-1",
				Due:      "2025-04-01",
				Labels:   []trello.LabelResponse{{ID: "lb1", Name: "Bug"}},
				Members:  []trello.MemberResponse{{ID: "m1", FullName: "John Doe"}},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleGetCard, map[string]any{
		"card_id": "card-1",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "My Card")
	assert.Contains(t, text, "Bug")
	assert.Contains(t, text, "John Doe")
}

func TestGetCard_APIError(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleGetCard, map[string]any{
		"card_id": "bad-id",
	})

	assert.True(t, result.IsError)
}

func TestListCards_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/1/boards/board-1/lists":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.ListResponse{
				{ID: "list-1", Name: "To Do"},
			}))
		case r.URL.Path == "/1/lists/list-1/cards":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.CardDetailResponse{
				{ID: "c1", Name: "Card 1", ShortURL: "https://trello.com/c/1"},
				{ID: "c2", Name: "Card 2", ShortURL: "https://trello.com/c/2"},
			}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleListCards, map[string]any{
		"list_name": "To Do",
	})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Card 1")
	assert.Contains(t, text, "Card 2")
}

func TestListLists_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/boards/board-1/lists" {
			require.NoError(t, json.NewEncoder(w).Encode([]trello.ListResponse{
				{ID: "l1", Name: "To Do"},
				{ID: "l2", Name: "In Progress"},
				{ID: "l3", Name: "Done"},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleListLists, map[string]any{})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "To Do")
	assert.Contains(t, text, "In Progress")
	assert.Contains(t, text, "Done")
}

func TestListLabels_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/boards/board-1/labels" {
			require.NoError(t, json.NewEncoder(w).Encode([]trello.LabelResponse{
				{ID: "lb1", Name: "Bug", Color: "red"},
				{ID: "lb2", Name: "Feature", Color: "green"},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleListLabels, map[string]any{})

	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Bug")
	assert.Contains(t, text, "Feature")
}

func TestListLabels_APIError(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleListLabels, map[string]any{})
	assert.True(t, result.IsError)
}

func TestCreateList_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/1/lists" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "Sprint 5", r.FormValue("name"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.ListResponse{
				ID: "list-new", Name: "Sprint 5",
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleCreateList, map[string]any{"name": "Sprint 5"})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Sprint 5")
	assert.Contains(t, text, "list-new")
}

func TestCreateList_MissingName(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	result := callTool(t, s, s.handleCreateList, map[string]any{})
	assert.True(t, result.IsError)
}

func TestArchiveCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/1/cards/c1" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "true", r.FormValue("closed"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1", Closed: true}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleArchiveCard, map[string]any{"card_id": "c1"})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "archived")
}

func TestArchiveCard_APIError(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleArchiveCard, map[string]any{"card_id": "bad"})
	assert.True(t, result.IsError)
}

func TestDeleteCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/1/cards/c1" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleDeleteCard, map[string]any{"card_id": "c1"})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "deleted")
}

func TestDeleteCard_APIError(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleDeleteCard, map[string]any{"card_id": "bad"})
	assert.True(t, result.IsError)
}

func TestAddComment_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/1/cards/c1/actions/comments" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "Deployed to staging", r.FormValue("text"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CommentResponse{
				ID:   "comment-1",
				Data: trello.CommentDataResponse{Text: "Deployed to staging"},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAddComment, map[string]any{
		"card_id": "c1",
		"text":    "Deployed to staging",
	})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Deployed to staging")
	assert.Contains(t, text, "comment-1")
}

func TestAddComment_MissingText(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	result := callTool(t, s, s.handleAddComment, map[string]any{"card_id": "c1"})
	assert.True(t, result.IsError)
}

func TestAssignCard_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/1/boards/board-1/members":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.MemberResponse{
				{ID: "m1", Username: "john", FullName: "John Doe"},
				{ID: "m2", Username: "jane", FullName: "Jane Smith"},
			}))
		case r.Method == "PUT" && r.URL.Path == "/1/cards/c1":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "m1", r.FormValue("idMembers"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1"}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAssignCard, map[string]any{
		"card_id":      "c1",
		"member_names": "john",
	})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "m1")
}

func TestAssignCard_NoMatch(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/boards/board-1/members" {
			require.NoError(t, json.NewEncoder(w).Encode([]trello.MemberResponse{
				{ID: "m1", Username: "john", FullName: "John Doe"},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAssignCard, map[string]any{
		"card_id":      "c1",
		"member_names": "nonexistent",
	})
	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "no matching members")
}

func TestSearchCards_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/search" {
			require.NoError(t, json.NewEncoder(w).Encode(trello.SearchResponse{
				Cards: []trello.CardDetailResponse{
					{ID: "c1", Name: "Payment integration", ShortURL: "https://trello.com/c/1"},
					{ID: "c2", Name: "Payment bug", ShortURL: "https://trello.com/c/2"},
				},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleSearchCards, map[string]any{"query": "payment"})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "Payment integration")
	assert.Contains(t, text, "Payment bug")
}

func TestSearchCards_NoResults(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/search" {
			require.NoError(t, json.NewEncoder(w).Encode(trello.SearchResponse{Cards: []trello.CardDetailResponse{}}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleSearchCards, map[string]any{"query": "nonexistent"})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "[]")
}

func TestAddLabel_HappyPath(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/1/cards/c1":
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{
				ID:     "c1",
				Labels: []trello.LabelResponse{{ID: "lb1", Name: "Feature"}},
			}))
		case r.URL.Path == "/1/boards/board-1/labels":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.LabelResponse{
				{ID: "lb1", Name: "Feature", Color: "green"},
				{ID: "lb2", Name: "Bug", Color: "red"},
			}))
		case r.Method == "PUT" && r.URL.Path == "/1/cards/c1":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "lb1,lb2", r.FormValue("idLabels"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1"}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAddLabel, map[string]any{
		"card_id":    "c1",
		"label_name": "Bug",
	})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "added")
	assert.Contains(t, text, "lb2")
}

func TestAddLabel_NotFound(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/1/cards/c1":
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1"}))
		case r.URL.Path == "/1/boards/board-1/labels":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.LabelResponse{
				{ID: "lb1", Name: "Bug", Color: "red"},
			}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAddLabel, map[string]any{
		"card_id":    "c1",
		"label_name": "Nonexistent",
	})
	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "not found")
}

func TestAddLabel_AlreadyAssigned(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/1/cards/c1":
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{
				ID:     "c1",
				Labels: []trello.LabelResponse{{ID: "lb1", Name: "Bug"}},
			}))
		case r.URL.Path == "/1/boards/board-1/labels":
			require.NoError(t, json.NewEncoder(w).Encode([]trello.LabelResponse{
				{ID: "lb1", Name: "Bug", Color: "red"},
			}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleAddLabel, map[string]any{
		"card_id":    "c1",
		"label_name": "Bug",
	})
	assert.True(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "already assigned")
}

func TestSetDueDate_Set(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/1/cards/c1" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "2026-03-15", r.FormValue("due"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1"}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleSetDueDate, map[string]any{
		"card_id":  "c1",
		"due_date": "2026-03-15",
	})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "2026-03-15")
}

func TestSetDueDate_Clear(t *testing.T) {
	s, ts := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/1/cards/c1" {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "null", r.FormValue("due"))
			require.NoError(t, json.NewEncoder(w).Encode(trello.CardDetailResponse{ID: "c1"}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	result := callTool(t, s, s.handleSetDueDate, map[string]any{
		"card_id": "c1",
	})
	assert.False(t, result.IsError)
	text := resultText(t, result)
	assert.Contains(t, text, "cleared")
}

func TestToolRegistration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	client := trello.NewClientWithURL(ts.URL, "key", nil)
	srv, _ := newMCPServerWithClient(client, "token", "board-1", "list-1")
	assert.NotNil(t, srv)
}
