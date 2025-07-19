package middleware

import (
	"context"
	"github.com/rookgm/gophermart/internal/service"
	"net/http"
)

type contextKey uint64

const (
	contextKeyUserID contextKey = iota
)

func Auth(ts service.TokenService) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, "can not get cookie", http.StatusUnauthorized)
				return
			}

			payload, err := ts.VerifyToken(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, payload.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(contextKeyUserID).(uint64)
	return id, ok
}
