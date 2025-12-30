package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Middleware проверяет JWT токен в запросе.
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем публичные endpoints
			if r.URL.Path == "/api/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Извлекаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Валидируем токен
			claims, err := ValidateToken(tokenString, secret)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			// Добавляем claims в контекст
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID извлекает user ID из контекста.
func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value("user_id").(int)
	return userID, ok
}

// GetUsername извлекает username из контекста.
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// respondError отправляет JSON ошибку.
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CORS middleware для поддержки будущего веб-интерфейса.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
