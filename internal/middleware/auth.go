package middleware

import (
	"context"
	"net/http"
	"strings"

	"library/internal/auth"
	"library/internal/models"
	"library/internal/services"
)

type contextKey string

const UserContextKey contextKey = "authenticated_user"

// Authenticate extracts JWT, loads user, and attaches user object to request context
func Authenticate(userService services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "authorization header missing"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "invalid header format, expected 'Bearer <token>'"}`, http.StatusUnauthorized)
				return
			}

			// 2. Parse & Validate JWT Token
			claims, err := auth.ParseToken(parts[1])
			if err != nil {
				http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// 3. Fetch full User object (or use data embedded in claims)
			user, err := userService.GetUserByID(r.Context(), claims.UserID)
			if err != nil {
				http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
				return
			}

			// 4. Attach User object to request context
			ctx := context.WithValue(r.Context(), UserContextKey, user)

			// 5. Pass request forward with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(allowedRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Grab user attached by Authenticate middleware
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// 2. Check if user's role matches any allowed role
			hasRole := false
			for _, role := range allowedRoles {
				if user.Role == string(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, `{"error": "forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			// 3. User has allowed role -> proceed to handler
			next.ServeHTTP(w, r)
		})
	}
}

// Helper function for handlers to easily get the authenticated user
func GetUserFromContext(ctx context.Context) (*models.BookMember, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.BookMember)
	return user, ok
}

// Helper to chain middleware onto a http.HandlerFunc
func Chain(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	var current http.Handler = h
	for i := len(middlewares) - 1; i >= 0; i-- {
		current = middlewares[i](current)
	}
	return current
}
