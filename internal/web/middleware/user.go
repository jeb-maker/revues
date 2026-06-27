package middleware

import (
	"context"
	"net/http"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

type contextKey int

const userContextKey contextKey = 1

// LoadUser resolves the session cookie into a user on the request context.
func LoadUser(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := auth.SessionTokenFromRequest(r)
			if err == nil && token != "" {
				userID, err := st.UserIDByTokenHash(ctx, auth.HashToken(token))
				if err == nil {
					user, err := st.UserByID(ctx, userID)
					if err == nil {
						ctx = context.WithValue(ctx, userContextKey, user)
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext returns the authenticated user, if any.
func UserFromContext(ctx context.Context) (*store.User, bool) {
	user, ok := ctx.Value(userContextKey).(*store.User)
	return user, ok
}

// SessionTokenFromContext is a helper for handlers needing the raw cookie value.
func SessionTokenFromContext(r *http.Request) string {
	token, err := auth.SessionTokenFromRequest(r)
	if err != nil {
		return ""
	}
	return token
}
