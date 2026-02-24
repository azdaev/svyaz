package middleware

import (
	"context"
	"net/http"
	"svyaz/internal/models"
	"svyaz/internal/repo"
)

type contextKey string

const userContextKey contextKey = "user"

func Auth(r *repo.Repo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cookie, err := req.Cookie("session")
			if err != nil {
				next.ServeHTTP(w, req)
				return
			}

			userID, err := r.GetSessionUser(req.Context(), cookie.Value)
			if err != nil || userID == 0 {
				next.ServeHTTP(w, req)
				return
			}

			user, err := r.GetUser(req.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, req)
				return
			}

			ctx := context.WithValue(req.Context(), userContextKey, user)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userContextKey).(*models.User)
	return user
}
