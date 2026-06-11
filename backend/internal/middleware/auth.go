package middleware

import (
	"context"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "user_id", int64(1))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireKeyholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "user_id", int64(2))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}