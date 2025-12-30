package api

import (
	"encoding/json"
	"net/http"

	"github.com/matperez/group-photo-reaction-bot/admin/internal/service"
)

// StatsHandler обрабатывает запросы к статистике.
type StatsHandler struct {
	service *service.Service
}

// NewStatsHandler создает новый handler статистики.
func NewStatsHandler(svc *service.Service) *StatsHandler {
	return &StatsHandler{
		service: svc,
	}
}

// Get обрабатывает GET /api/stats.
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
