package handler

import (
	"context"
	"gotrade/order-service/pkg/utils"
	"net/http"
	"strings"
)

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
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(*utils.JWTClaims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), "userID", claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			}
		})
	}
}

