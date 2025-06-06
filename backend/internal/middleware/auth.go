package middleware

import (
	"context"
	"net/http"
	"strings"

	"rythmitbackend/internal/services"
	"rythmitbackend/internal/utils"
)

// AuthMiddlewareFunc is a middleware that handles authentication for both API and web routes
func AuthMiddlewareFunc(next http.Handler, isWebRoute bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Get token from either cookie (web) or Authorization header (API)
		if isWebRoute {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				if isWebRoute {
					http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				} else {
					utils.Unauthorized(w, "Token d'authentification requis")
				}
				return
			}
			tokenString = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.Unauthorized(w, "Token d'authentification requis")
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				utils.Unauthorized(w, "Format du token invalide")
				return
			}
			tokenString = tokenParts[1]
		}

		// Validate the token using the auth service
		authService := services.NewAuthService(nil, nil) // TODO: Pass proper dependencies
		claims, err := authService.ParseToken(tokenString)
		if err != nil {
			if isWebRoute {
				// Clear the invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				})
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			} else {
				utils.Unauthorized(w, "Token invalide ou expiré")
			}
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "is_admin", claims.IsAdmin)

		// Continue with the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is a middleware for web routes that requires authentication
func RequireAuth(next http.Handler) http.Handler {
	return AuthMiddlewareFunc(next, true)
}

// RequireAPIAuth is a middleware for API routes that requires authentication
func RequireAPIAuth(next http.Handler) http.Handler {
	return AuthMiddlewareFunc(next, false)
}
