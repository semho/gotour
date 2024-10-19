package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("User-ID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid User-ID", http.StatusUnauthorized)
			return
		}

		// userID в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
