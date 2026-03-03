package miniapp

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/usecase/port"
)

func NewAPIRouter(
	ctrl *controller.MiniAppController,
	cardCtrl *controller.MiniAppCardController,
	settingsCtrl *controller.MiniAppSettingsController,
	sessionMgr port.SessionManager,
	logger *slog.Logger,
	frontendFS fs.FS,
) http.Handler {
	mux := http.NewServeMux()

	// Auth (no auth required)
	mux.HandleFunc("POST /api/auth", handleAuth(ctrl, logger))

	// User settings
	mux.HandleFunc("GET /api/user/settings", handleGetSettings(ctrl, logger))
	mux.HandleFunc("PUT /api/user/settings", handleUpdateSettings(settingsCtrl, logger))
	mux.HandleFunc("PUT /api/user/token", handleConnectTrello(settingsCtrl, logger))

	// Boards
	mux.HandleFunc("GET /api/boards", handleListBoards(ctrl, logger))
	mux.HandleFunc("GET /api/boards/{id}/lists", handleListLists(ctrl, logger))
	mux.HandleFunc("GET /api/boards/{id}/labels", handleListLabels(ctrl, logger))
	mux.HandleFunc("GET /api/boards/{id}/members", handleListMembers(cardCtrl, logger))

	// Lists / Cards
	mux.HandleFunc("GET /api/lists/{id}/cards", handleListCards(cardCtrl, logger))
	mux.HandleFunc("POST /api/cards", handleCreateCard(cardCtrl, logger))
	mux.HandleFunc("GET /api/cards/{id}", handleGetCard(cardCtrl, logger))
	mux.HandleFunc("PUT /api/cards/{id}", handleUpdateCard(cardCtrl, logger))
	mux.HandleFunc("DELETE /api/cards/{id}", handleDeleteCard(settingsCtrl, logger))
	mux.HandleFunc("POST /api/cards/{id}/comments", handleAddComment(settingsCtrl, logger))

	// SPA fallback — serve frontend for non-API, non-healthz paths
	if frontendFS != nil {
		spaHandler := newSPAHandler(frontendFS)
		mux.Handle("/", spaHandler)
	}

	// Middleware chain: CORS → Auth → JSON → mux
	var handler http.Handler = mux
	handler = JSONMiddleware(handler)
	handler = AuthMiddleware(sessionMgr)(handler)
	handler = CORSMiddleware(handler)

	return handler
}

// SPA handler serves static files and falls back to index.html
type spaHandler struct {
	fileServer http.Handler
	fs         fs.FS
}

func newSPAHandler(frontendFS fs.FS) *spaHandler {
	return &spaHandler{
		fileServer: http.FileServerFS(frontendFS),
		fs:         frontendFS,
	}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try to serve the file directly
	if _, err := fs.Stat(h.fs, path); err == nil {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	// Fallback to index.html for SPA routing
	r.URL.Path = "/"
	h.fileServer.ServeHTTP(w, r)
}

// --- Handler Functions ---

func handleAuth(ctrl *controller.MiniAppController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InitData string `json:"init_data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		output, err := ctrl.HandleAuth(r.Context(), req.InitData)
		if err != nil {
			logger.Error("auth failed", "error", err)
			writeError(w, "authentication failed", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleGetSettings(ctrl *controller.MiniAppController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleGetSettings(r.Context(), telegramID)
		if err != nil {
			logger.Error("get settings failed", "error", err)
			writeError(w, "failed to get settings", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleUpdateSettings(ctrl *controller.MiniAppSettingsController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BoardID string `json:"board_id"`
			ListID  string `json:"list_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		telegramID := TelegramIDFromContext(r.Context())
		if err := ctrl.HandleUpdateSettings(r.Context(), telegramID, req.BoardID, req.ListID); err != nil {
			logger.Error("update settings failed", "error", err)
			writeError(w, "failed to update settings", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleConnectTrello(ctrl *controller.MiniAppSettingsController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleConnectTrello(r.Context(), telegramID, req.Token)
		if err != nil {
			logger.Error("connect trello failed", "error", err)
			writeError(w, "failed to connect trello", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleListBoards(ctrl *controller.MiniAppController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleListBoards(r.Context(), telegramID)
		if err != nil {
			logger.Error("list boards failed", "error", err)
			writeError(w, "failed to list boards", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleListLists(ctrl *controller.MiniAppController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleListLists(r.Context(), telegramID, boardID)
		if err != nil {
			logger.Error("list lists failed", "error", err)
			writeError(w, "failed to list lists", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleListLabels(ctrl *controller.MiniAppController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleListLabels(r.Context(), telegramID, boardID)
		if err != nil {
			logger.Error("list labels failed", "error", err)
			writeError(w, "failed to list labels", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleListMembers(ctrl *controller.MiniAppCardController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleListMembers(r.Context(), telegramID, boardID)
		if err != nil {
			logger.Error("list members failed", "error", err)
			writeError(w, "failed to list members", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleListCards(ctrl *controller.MiniAppCardController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleListCards(r.Context(), telegramID, listID)
		if err != nil {
			logger.Error("list cards failed", "error", err)
			writeError(w, "failed to list cards", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleCreateCard(ctrl *controller.MiniAppCardController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ListID      string   `json:"list_id"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			DueDate     string   `json:"due_date"`
			LabelIDs    []string `json:"label_ids"`
			MemberIDs   []string `json:"member_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleCreateCard(r.Context(), controller.CreateCardRequest{
			TelegramID:  telegramID,
			ListID:      req.ListID,
			Title:       req.Title,
			Description: req.Description,
			DueDate:     req.DueDate,
			LabelIDs:    req.LabelIDs,
			MemberIDs:   req.MemberIDs,
		})
		if err != nil {
			logger.Error("create card failed", "error", err)
			writeError(w, "failed to create card", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, output)
	}
}

func handleGetCard(ctrl *controller.MiniAppCardController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		output, err := ctrl.HandleGetCard(r.Context(), telegramID, cardID)
		if err != nil {
			logger.Error("get card failed", "error", err)
			writeError(w, "failed to get card", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func handleUpdateCard(ctrl *controller.MiniAppCardController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		var req struct {
			Title       *string `json:"title"`
			Description *string `json:"description"`
			ListID      *string `json:"list_id"`
			Due         *string `json:"due"`
			LabelIDs    *string `json:"label_ids"`
			MemberIDs   *string `json:"member_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		telegramID := TelegramIDFromContext(r.Context())
		if err := ctrl.HandleUpdateCard(r.Context(), telegramID, cardID, controller.UpdateCardRequest{
			Title:       req.Title,
			Description: req.Description,
			ListID:      req.ListID,
			Due:         req.Due,
			LabelIDs:    req.LabelIDs,
			MemberIDs:   req.MemberIDs,
		}); err != nil {
			logger.Error("update card failed", "error", err)
			writeError(w, "failed to update card", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleDeleteCard(ctrl *controller.MiniAppSettingsController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		telegramID := TelegramIDFromContext(r.Context())
		if err := ctrl.HandleDeleteCard(r.Context(), telegramID, cardID); err != nil {
			logger.Error("delete card failed", "error", err)
			writeError(w, "failed to delete card", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleAddComment(ctrl *controller.MiniAppSettingsController, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		telegramID := TelegramIDFromContext(r.Context())
		if err := ctrl.HandleAddComment(r.Context(), telegramID, cardID, req.Text); err != nil {
			logger.Error("add comment failed", "error", err)
			writeError(w, "failed to add comment", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
