package main

import (
	"fmt"
	"log"
	"net/http"

	"rythmitbackend/configs"
	"rythmitbackend/internal/utils"
)

func main() {
	// Chargement de la configuration
	cfg := configs.Load()

	// Affichage bannière de démarrage
	displayBanner(cfg)

	// Configuration du serveur HTTP
	mux := http.NewServeMux()

	// Routes de base
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/", apiHandler)

	// Configuration du serveur avec timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Démarrage du serveur
	log.Printf("🚀 %s démarré sur http://localhost:%s\n", cfg.App.Name, cfg.App.Port)
	log.Printf("📝 Environment: %s\n", cfg.App.Environment)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("❌ Erreur démarrage serveur: %v", err)
	}
}

// homeHandler - Route racine
func homeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := configs.Get()

	data := map[string]string{
		"message": fmt.Sprintf("Bienvenue sur %s API", cfg.App.Name),
		"version": cfg.App.Version,
		"docs":    "/api/docs",
	}

	utils.Success(w, "API Rythmit opérationnelle", data)
}

// healthHandler - Endpoint de santé
func healthHandler(w http.ResponseWriter, r *http.Request) {
	cfg := configs.Get()

	health := map[string]interface{}{
		"status":      "healthy",
		"service":     cfg.App.Name,
		"version":     cfg.App.Version,
		"environment": cfg.App.Environment,
	}

	utils.Success(w, "Service en bonne santé", health)
}

// apiHandler - Handler temporaire pour toutes les routes /api/
func apiHandler(w http.ResponseWriter, r *http.Request) {
	// Pour l'instant, retourne la liste des endpoints disponibles
	endpoints := map[string][]string{
		"authentification": {
			"POST /api/register - Inscription",
			"POST /api/login - Connexion",
			"GET /api/profile - Profil utilisateur",
		},
		"threads": {
			"GET /api/threads - Liste des threads",
			"POST /api/threads - Créer un thread",
			"GET /api/threads/:id - Détails d'un thread",
			"PUT /api/threads/:id - Modifier un thread",
			"DELETE /api/threads/:id - Supprimer un thread",
		},
		"messages": {
			"GET /api/threads/:id/messages - Messages d'un thread",
			"POST /api/threads/:id/messages - Poster un message",
			"POST /api/messages/:id/fire - Fire un message",
			"POST /api/messages/:id/skip - Skip un message",
		},
		"battles": {
			"POST /api/battles - Créer une battle",
			"GET /api/battles/:id - Détails d'une battle",
			"POST /api/battles/:id/vote - Voter dans une battle",
			"GET /api/battles/active - Battles actives",
		},
	}

	utils.Success(w, "Endpoints API Rythmit", endpoints)
}

// displayBanner - Affiche la bannière ASCII au démarrage
func displayBanner(cfg *configs.Config) {
	banner := `
    ____        __  __            _ _ __  
   / __ \__  __/ /_/ /_  ____ _  ( ) / /_ 
  / /_/ / / / / __/ __ \/ __ ` + "`" + `/ |/ / __/
 / _, _/ /_/ / /_/ / / / / / /  / / /_  
/_/ |_|\__, /\__/_/ /_/_/ /_/  /_/\__/  
      /____/                            
	`
	fmt.Println(banner)
	fmt.Printf("🎵 %s v%s - Forum Musical Interactif\n", cfg.App.Name, cfg.App.Version)
	fmt.Println("================================================")
}
