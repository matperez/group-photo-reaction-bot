package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matperez/group-photo-reaction-bot/admin/internal/service"
)

// ChatsHandler обрабатывает запросы к чатам.
type ChatsHandler struct {
	service *service.Service
}

// NewChatsHandler создает новый handler чатов.
func NewChatsHandler(svc *service.Service) *ChatsHandler {
	return &ChatsHandler{
		service: svc,
	}
}

// List обрабатывает GET /api/chats.
func (h *ChatsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	chats, err := h.service.GetChats(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

// Get обрабатывает GET /api/chats/:id.
func (h *ChatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatIDStr := r.URL.Query().Get("id")
	if chatIDStr == "" {
		http.Error(w, "chat ID required", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid chat ID", http.StatusBadRequest)
		return
	}

	chat, err := h.service.GetChatInfo(r.Context(), chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}
