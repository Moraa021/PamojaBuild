package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"PamojaBuild/internal/auth"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

var authService *auth.Service

func InitAuthMiddleware(jwtSecret string) {
	authService = auth.NewService(jwtSecret)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing authorization header"})
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid authorization format"})
			return
		}

		token := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
			return
		}

		// Get user from DB
		user, err := getUserByID(claims.UserID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		}

		// Attach both user ID and user object to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireKeyholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil || !IsKeyholder(user) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: keyholders only"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func getUserByID(userID int64) (*models.User, error) {
	var user models.User
	var passwordHash string

	err := db.SQLDB.QueryRow(
		"SELECT id, phone, password_hash, role, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Phone, &passwordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	user.PasswordHash = passwordHash
	return &user, nil
}

func IsKeyholder(user *models.User) bool {
	return user.Role == "keyholder"
}