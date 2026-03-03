package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"telegram-trello-bot/internal/infrastructure/trello"
)

type mcpServer struct {
	client  *trello.Client
	token   string
	boardID string
	listID  string
}

func main() {
	apiKey := os.Getenv("TRELLO_API_KEY")
	token := os.Getenv("TRELLO_TOKEN")
	boardID := os.Getenv("TRELLO_BOARD_ID")
	listID := os.Getenv("TRELLO_LIST_ID")

	if apiKey == "" || token == "" || boardID == "" {
		log.Fatal("TRELLO_API_KEY, TRELLO_TOKEN, and TRELLO_BOARD_ID are required")
	}

	client := trello.NewClient(apiKey, &http.Client{Timeout: 30 * time.Second})
	s := &mcpServer{client: client, token: token, boardID: boardID, listID: listID}

	srv := server.NewMCPServer("trello", "1.0.0",
		server.WithToolCapabilities(false),
	)

	registerTools(srv, s)

	if err := server.ServeStdio(srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func newMCPServerWithClient(client *trello.Client, token, boardID, listID string) (*server.MCPServer, *mcpServer) {
	s := &mcpServer{client: client, token: token, boardID: boardID, listID: listID}
	srv := server.NewMCPServer("trello", "1.0.0",
		server.WithToolCapabilities(false),
	)
	registerTools(srv, s)
	return srv, s
}

func registerTools(srv *server.MCPServer, s *mcpServer) {
	srv.AddTool(mcp.NewTool("create_card",
		mcp.WithDescription("Create a new Trello card on the configured board"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Card title")),
		mcp.WithString("description", mcp.Description("Card description")),
		mcp.WithString("list_name", mcp.Description("List name to create the card in (uses default list if omitted)")),
		mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format")),
	), s.handleCreateCard)

	srv.AddTool(mcp.NewTool("move_card",
		mcp.WithDescription("Move a Trello card to a different list"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("list_name", mcp.Required(), mcp.Description("Target list name")),
	), s.handleMoveCard)

	srv.AddTool(mcp.NewTool("update_card",
		mcp.WithDescription("Update a Trello card's fields"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("title", mcp.Description("New card title")),
		mcp.WithString("description", mcp.Description("New card description")),
		mcp.WithString("due_date", mcp.Description("New due date in YYYY-MM-DD format")),
	), s.handleUpdateCard)

	srv.AddTool(mcp.NewTool("get_card",
		mcp.WithDescription("Get details of a Trello card"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
	), s.handleGetCard)

	srv.AddTool(mcp.NewTool("list_cards",
		mcp.WithDescription("List all cards in a Trello list"),
		mcp.WithString("list_name", mcp.Required(), mcp.Description("List name")),
	), s.handleListCards)

	srv.AddTool(mcp.NewTool("list_lists",
		mcp.WithDescription("List all lists on the configured Trello board"),
	), s.handleListLists)

	srv.AddTool(mcp.NewTool("list_labels",
		mcp.WithDescription("List all labels on the configured Trello board"),
	), s.handleListLabels)

	srv.AddTool(mcp.NewTool("create_list",
		mcp.WithDescription("Create a new list on the configured Trello board"),
		mcp.WithString("name", mcp.Required(), mcp.Description("List name")),
	), s.handleCreateList)

	srv.AddTool(mcp.NewTool("archive_card",
		mcp.WithDescription("Archive a Trello card (soft delete)"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
	), s.handleArchiveCard)

	srv.AddTool(mcp.NewTool("delete_card",
		mcp.WithDescription("Permanently delete a Trello card"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
	), s.handleDeleteCard)

	srv.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to a Trello card"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Comment text")),
	), s.handleAddComment)

	srv.AddTool(mcp.NewTool("assign_card",
		mcp.WithDescription("Assign or unassign members on a Trello card by username"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("member_names", mcp.Required(), mcp.Description("Comma-separated member usernames or full names")),
	), s.handleAssignCard)

	srv.AddTool(mcp.NewTool("search_cards",
		mcp.WithDescription("Search for cards on the board by keyword"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	), s.handleSearchCards)

	srv.AddTool(mcp.NewTool("add_label",
		mcp.WithDescription("Add a label to a Trello card by label name"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("label_name", mcp.Required(), mcp.Description("Label name to add")),
	), s.handleAddLabel)

	srv.AddTool(mcp.NewTool("set_due_date",
		mcp.WithDescription("Set or clear the due date on a Trello card"),
		mcp.WithString("card_id", mcp.Required(), mcp.Description("Trello card ID")),
		mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format (omit or empty to clear)")),
	), s.handleSetDueDate)
}

func (s *mcpServer) findListByName(ctx context.Context, name string) (string, error) {
	lists, err := s.client.GetLists(ctx, s.token, s.boardID)
	if err != nil {
		return "", fmt.Errorf("get lists: %w", err)
	}
	nameLower := strings.ToLower(name)
	for _, l := range lists {
		if strings.ToLower(l.Name) == nameLower {
			return l.ID, nil
		}
	}
	return "", fmt.Errorf("list %q not found on board", name)
}

func (s *mcpServer) handleCreateCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := req.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	description := req.GetString("description", "")
	listName := req.GetString("list_name", "")
	dueDate := req.GetString("due_date", "")

	listID := s.listID
	if listName != "" {
		id, err := s.findListByName(ctx, listName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		listID = id
	}
	if listID == "" {
		return mcp.NewToolResultError("no list specified and TRELLO_LIST_ID not configured"), nil
	}

	cardReq := trello.CreateCardRequest{
		Name:        title,
		Description: description,
		ListID:      listID,
	}
	if dueDate != "" {
		cardReq.Due = dueDate
	}

	card, err := s.client.CreateCard(ctx, s.token, cardReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create card failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id":  card.ID,
		"card_url": card.ShortURL,
		"title":    title,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleMoveCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	listName, err := req.RequireString("list_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	listID, err := s.findListByName(ctx, listName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	_, err = s.client.UpdateCard(ctx, s.token, cardID, trello.UpdateCardRequest{IDList: &listID})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("move card failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id":   cardID,
		"moved_to":  listName,
		"list_id":   listID,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleUpdateCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateReq := trello.UpdateCardRequest{}
	hasUpdate := false

	if title := req.GetString("title", ""); title != "" {
		updateReq.Name = &title
		hasUpdate = true
	}
	if desc := req.GetString("description", ""); desc != "" {
		updateReq.Desc = &desc
		hasUpdate = true
	}
	if due := req.GetString("due_date", ""); due != "" {
		updateReq.Due = &due
		hasUpdate = true
	}

	if !hasUpdate {
		return mcp.NewToolResultError("no fields to update"), nil
	}

	card, err := s.client.UpdateCard(ctx, s.token, cardID, updateReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("update card failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id":  card.ID,
		"title":    card.Name,
		"card_url": card.ShortURL,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleGetCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	card, err := s.client.GetCard(ctx, s.token, cardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get card failed: %v", err)), nil
	}

	labels := make([]string, len(card.Labels))
	for i, l := range card.Labels {
		labels[i] = l.Name
	}
	members := make([]string, len(card.Members))
	for i, m := range card.Members {
		members[i] = m.FullName
	}

	result := map[string]any{
		"card_id":     card.ID,
		"title":       card.Name,
		"description": card.Desc,
		"card_url":    card.ShortURL,
		"list_id":     card.IDList,
		"due":         card.Due,
		"labels":      labels,
		"members":     members,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleListCards(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	listName, err := req.RequireString("list_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	listID, err := s.findListByName(ctx, listName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cards, err := s.client.GetCards(ctx, s.token, listID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list cards failed: %v", err)), nil
	}

	items := make([]map[string]string, len(cards))
	for i, c := range cards {
		items[i] = map[string]string{
			"card_id":  c.ID,
			"title":    c.Name,
			"card_url": c.ShortURL,
		}
	}
	return jsonResult(items)
}

func (s *mcpServer) handleListLists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lists, err := s.client.GetLists(ctx, s.token, s.boardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list lists failed: %v", err)), nil
	}

	items := make([]map[string]string, len(lists))
	for i, l := range lists {
		items[i] = map[string]string{
			"list_id": l.ID,
			"name":    l.Name,
		}
	}
	return jsonResult(items)
}

func (s *mcpServer) handleListLabels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	labels, err := s.client.GetLabels(ctx, s.token, s.boardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list labels failed: %v", err)), nil
	}

	items := make([]map[string]string, len(labels))
	for i, l := range labels {
		items[i] = map[string]string{
			"label_id": l.ID,
			"name":     l.Name,
			"color":    l.Color,
		}
	}
	return jsonResult(items)
}

func (s *mcpServer) handleCreateList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	list, err := s.client.CreateList(ctx, s.token, s.boardID, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create list failed: %v", err)), nil
	}

	result := map[string]string{
		"list_id": list.ID,
		"name":    list.Name,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleArchiveCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.client.ArchiveCard(ctx, s.token, cardID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("archive card failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id": cardID,
		"status":  "archived",
	}
	return jsonResult(result)
}

func (s *mcpServer) handleDeleteCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.client.DeleteCard(ctx, s.token, cardID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("delete card failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id": cardID,
		"status":  "deleted",
	}
	return jsonResult(result)
}

func (s *mcpServer) handleAddComment(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	comment, err := s.client.AddComment(ctx, s.token, cardID, text)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("add comment failed: %v", err)), nil
	}

	result := map[string]string{
		"comment_id": comment.ID,
		"card_id":    cardID,
		"text":       comment.Data.Text,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleAssignCard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	memberNamesStr, err := req.RequireString("member_names")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	names := strings.Split(memberNamesStr, ",")
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
	}

	members, err := s.client.GetMembers(ctx, s.token, s.boardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get members failed: %v", err)), nil
	}

	var matchedIDs []string
	for _, name := range names {
		nameLower := strings.ToLower(name)
		for _, m := range members {
			if strings.ToLower(m.Username) == nameLower || strings.ToLower(m.FullName) == nameLower ||
				strings.Contains(strings.ToLower(m.FullName), nameLower) {
				matchedIDs = append(matchedIDs, m.ID)
				break
			}
		}
	}

	if len(matchedIDs) == 0 {
		return mcp.NewToolResultError("no matching members found"), nil
	}

	idMembers := strings.Join(matchedIDs, ",")
	_, err = s.client.UpdateCard(ctx, s.token, cardID, trello.UpdateCardRequest{IDMembers: &idMembers})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("assign card failed: %v", err)), nil
	}

	result := map[string]any{
		"card_id":        cardID,
		"assigned_count": len(matchedIDs),
		"member_ids":     matchedIDs,
	}
	return jsonResult(result)
}

func (s *mcpServer) handleSearchCards(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cards, err := s.client.SearchCards(ctx, s.token, s.boardID, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	items := make([]map[string]string, len(cards))
	for i, c := range cards {
		items[i] = map[string]string{
			"card_id":  c.ID,
			"title":    c.Name,
			"card_url": c.ShortURL,
			"list_id":  c.IDList,
		}
	}
	return jsonResult(items)
}

func (s *mcpServer) handleAddLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	labelName, err := req.RequireString("label_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get current card to preserve existing labels
	card, err := s.client.GetCard(ctx, s.token, cardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get card failed: %v", err)), nil
	}

	// Find label ID by name
	labels, err := s.client.GetLabels(ctx, s.token, s.boardID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get labels failed: %v", err)), nil
	}

	var labelID string
	nameLower := strings.ToLower(labelName)
	for _, l := range labels {
		if strings.ToLower(l.Name) == nameLower {
			labelID = l.ID
			break
		}
	}
	if labelID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("label %q not found on board", labelName)), nil
	}

	// Build label list: existing + new
	existingIDs := make([]string, len(card.Labels))
	for i, l := range card.Labels {
		existingIDs[i] = l.ID
	}
	for _, id := range existingIDs {
		if id == labelID {
			return mcp.NewToolResultError("label already assigned to card"), nil
		}
	}
	allIDs := strings.Join(append(existingIDs, labelID), ",")

	_, err = s.client.UpdateCard(ctx, s.token, cardID, trello.UpdateCardRequest{IDLabels: &allIDs})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("add label failed: %v", err)), nil
	}

	result := map[string]string{
		"card_id":    cardID,
		"label_name": labelName,
		"label_id":   labelID,
		"status":     "added",
	}
	return jsonResult(result)
}

func (s *mcpServer) handleSetDueDate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cardID, err := req.RequireString("card_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	dueDate := req.GetString("due_date", "")
	if dueDate == "" {
		dueDate = "null"
	}

	_, err = s.client.UpdateCard(ctx, s.token, cardID, trello.UpdateCardRequest{Due: &dueDate})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("set due date failed: %v", err)), nil
	}

	status := "set to " + dueDate
	if dueDate == "null" {
		status = "cleared"
	}

	result := map[string]string{
		"card_id": cardID,
		"status":  status,
	}
	return jsonResult(result)
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
