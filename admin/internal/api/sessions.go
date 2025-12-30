package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matperez/group-photo-reaction-bot/admin/internal/service"
)

// SessionsHandler обрабатывает запросы к сессиям.
type SessionsHandler struct {
	service *service.Service
}

// NewSessionsHandler создает новый handler сессий.
func NewSessionsHandler(svc *service.Service) *SessionsHandler {
	return &SessionsHandler{
		service: svc,
	}
}

// List обрабатывает GET /api/sessions.
func (h *SessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим параметры запроса
	limit := 50
	offset := 0
	var chatID *int64
	var isActive *bool

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

	if chatIDStr := r.URL.Query().Get("chat_id"); chatIDStr != "" {
		if id, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			chatID = &id
		}
	}

	if activeStr := r.URL.Query().Get("is_active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			isActive = &active
		}
	}

	sessions, err := h.service.GetSessions(r.Context(), limit, offset, chatID, isActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// Get обрабатывает GET /api/sessions/:id.
func (h *SessionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из пути (упрощенная версия, в реальности нужен роутер)
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "session ID required", http.StatusBadRequest)
		return
	}

	session, exists := h.service.GetSession(r.Context(), sessionID)
	if !exists {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

