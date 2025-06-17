package middleware

import (
	"context"
	"net/http"
	"strings"

	"rythmitbackend/configs"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/internal/utils"
	"rythmitbackend/pkg/database"
	customjwt "rythmitbackend/pkg/jwt"
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

		// Validate the token using JWT directly
		cfg := configs.Get()
		tokenManager := customjwt.NewTokenManager(customjwt.Config{
			Secret:          cfg.JWT.Secret,
			ExpirationHours: cfg.JWT.ExpirationHours,
			Issuer:          "rythmit-api",
		})

		claims, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			if isWebRoute {
				// Clear the invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					Secure:   false, // Pour le développement local
					SameSite: http.SameSiteLaxMode,
				})
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			} else {
				utils.Unauthorized(w, "Token invalide ou expiré")
			}
			return
		}

		// Get user from database to verify it still exists
		userRepo := repositories.NewUserRepository(database.DB)
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			if isWebRoute {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			} else {
				utils.Unauthorized(w, "Utilisateur non trouvé")
			}
			return
		}

		// Convert to DTO
		userDTO := services.ToUserResponseDTO(user)

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user", userDTO)
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "username", user.Username)
		ctx = context.WithValue(ctx, "email", user.Email)
		ctx = context.WithValue(ctx, "is_admin", user.IsAdmin)

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
