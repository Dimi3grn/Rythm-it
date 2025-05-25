package middleware

import (
	"log"
	"net/http"
	"time"
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

// JSONMiddleware force le Content-Type JSON
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware vérifie l'authentification JWT (à implémenter)
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implémenter la vérification JWT
		// Pour l'instant, on laisse passer

		// token := r.Header.Get("Authorization")
		// if token == "" {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     w.Write([]byte(`{"error": "Token manquant"}`))
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}

// AdminMiddleware vérifie les droits admin
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Vérifier si l'utilisateur est admin
		// Pour l'instant, on bloque tout

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Accès réservé aux administrateurs"}`))
	})
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
