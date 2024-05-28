package middlewares

import (
	"context"
	"net/http"
	"strings"

	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	jwt_helper "github.com/danzBraham/beli-mang/internal/helpers/jwt"
)

type ContextKey string

var (
	ContextUserIdKey  ContextKey = "userId"
	ContextIsAdminKey ContextKey = "isAdmin"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "Missing Authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "Invalid Authorization header")
			return
		}

		jwtPayload, err := jwt_helper.VerifyToken(tokenString)
		if err != nil {
			http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserIdKey, jwtPayload.UserId)
		ctx = context.WithValue(ctx, ContextIsAdminKey, jwtPayload.IsAdmin)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
