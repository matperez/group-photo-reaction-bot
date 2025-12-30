package api

import (
	"encoding/json"
	"net/http"

	"github.com/matperez/group-photo-reaction-bot/admin/internal/auth"
)

// AuthHandler обрабатывает запросы авторизации.
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler создает новый handler авторизации.
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest представляет запрос на авторизацию.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse представляет ответ на авторизацию.
type LoginResponse struct {
	Token string `json:"token"`
}

// Login обрабатывает POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, userID, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	_ = userID // Можно использовать для логирования

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}
