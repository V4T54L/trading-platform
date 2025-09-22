package handler

import (
	"context"
	"gotrade/account-service/pkg/utils"
	"net/http"
	"strings"
)

// JWTAuthMiddleware validates the JWT token from the Authorization header.
func JWTAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			token, err := utils.ValidateJWT(tokenString, secret)
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*utils.JWTClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

