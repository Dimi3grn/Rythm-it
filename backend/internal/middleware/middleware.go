package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/internal/utils"
	"rythmitbackend/pkg/database"
	customjwt "rythmitbackend/pkg/jwt"

	"github.com/golang-jwt/jwt/v5"
)

// Logger middleware pour logger toutes les requêtes
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapped ResponseWriter pour capturer le status
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		log.Printf(
			"[%s] %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// Recovery middleware pour récupérer des panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("❌ PANIC: %v", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Une erreur interne est survenue", "status": 500}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// JSONMiddleware force le Content-Type JSON pour les routes API
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware vérifie l'authentification JWT et injecte l'utilisateur dans le contexte
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraire le token du header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("❌ Token manquant")
			utils.Unauthorized(w, "Token d'authentification requis")
			return
		}

		// Vérifier le format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Println("❌ Format token invalide")
			utils.Unauthorized(w, "Format du token invalide")
			return
		}

		tokenString := tokenParts[1]

		// Validation JWT simple avec github.com/golang-jwt/jwt/v5
		cfg := configs.Get()

		// Parser le token
		token, err := jwt.NewParser().Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Vérifier la méthode de signature
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("méthode de signature invalide")
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			log.Printf("❌ Erreur validation token: %v", err)
			utils.Unauthorized(w, "Token invalide ou expiré")
			return
		}

		// Vérifier que le token est valide
		if !token.Valid {
			log.Println("❌ Token non valide")
			utils.Unauthorized(w, "Token invalide")
			return
		}

		// Extraire les claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("❌ Claims invalides")
			utils.Unauthorized(w, "Token invalide")
			return
		}

		// Vérifier l'expiration manuellement
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				log.Println("❌ Token expiré")
				utils.Unauthorized(w, "Token expiré")
				return
			}
		}

		// Extraire les infos utilisateur
		userID, _ := claims["user_id"].(float64) // JWT stocke les nombres en float64
		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)
		isAdmin, _ := claims["is_admin"].(bool)

		log.Printf("✅ Auth réussie - User: %s (ID: %.0f)", username, userID)

		// Injecter dans le contexte
		ctx := context.WithValue(r.Context(), "user_id", uint(userID))
		ctx = context.WithValue(ctx, "username", username)
		ctx = context.WithValue(ctx, "email", email)
		ctx = context.WithValue(ctx, "is_admin", isAdmin)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware vérifie les droits administrateur
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est admin depuis le contexte
		isAdmin, ok := r.Context().Value("is_admin").(bool)
		if !ok || !isAdmin {
			utils.Forbidden(w, "Droits administrateur requis")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware vérifie l'auth mais n'échoue pas si pas de token
// Utile pour les endpoints qui fonctionnent avec ou sans auth
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraire le token du header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Pas de token, continuer sans utilisateur dans le contexte
			next.ServeHTTP(w, r)
			return
		}

		// Vérifier le format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Format invalide, continuer sans utilisateur
			next.ServeHTTP(w, r)
			return
		}

		tokenString := tokenParts[1]

		// Valider le token JWT
		cfg := configs.Get()
		tokenManager := customjwt.NewTokenManager(customjwt.Config{
			Secret:          cfg.JWT.Secret,
			ExpirationHours: cfg.JWT.ExpirationHours,
			Issuer:          "rythmit-api",
		})

		claims, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			// Token invalide, continuer sans utilisateur
			next.ServeHTTP(w, r)
			return
		}

		// Récupérer l'utilisateur depuis la base de données
		userRepo := repositories.NewUserRepository(database.DB)
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			// Utilisateur non trouvé, continuer sans utilisateur
			next.ServeHTTP(w, r)
			return
		}

		// Convertir en DTO de réponse
		userDTO := services.ToUserResponseDTO(user)

		// Injecter l'utilisateur dans le contexte
		ctx := context.WithValue(r.Context(), "user", userDTO)
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "is_admin", user.IsAdmin)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware gère les requêtes CORS (déjà géré par rs/cors mais on peut avoir un custom)
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Vérifier si l'origine est autorisée
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Gérer les requêtes preflight OPTIONS
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware limite le nombre de requêtes (simple implémentation en mémoire)
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Pour une vraie application, utiliser Redis ou une solution plus robuste
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implémentation simple - à améliorer en production
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wrapper pour capturer le status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper functions pour récupérer des données depuis le contexte

// GetUserFromContext récupère l'utilisateur depuis le contexte de la requête
func GetUserFromContext(r *http.Request) *services.UserResponseDTO {
	if user, ok := r.Context().Value("user").(*services.UserResponseDTO); ok {
		return user
	}
	return nil
}

// GetUserIDFromContext récupère l'ID utilisateur depuis le contexte
func GetUserIDFromContext(r *http.Request) (uint, bool) {
	if userID, ok := r.Context().Value("user_id").(uint); ok {
		return userID, true
	}
	return 0, false
}

// IsAdminFromContext vérifie si l'utilisateur est admin
func IsAdminFromContext(r *http.Request) bool {
	if isAdmin, ok := r.Context().Value("is_admin").(bool); ok {
		return isAdmin
	}
	return false
}
