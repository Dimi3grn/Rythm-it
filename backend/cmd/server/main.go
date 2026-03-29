// Fichier: backend/cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"rythmitbackend/configs"
	"rythmitbackend/internal/router"
	"rythmitbackend/pkg/database"
	"rythmitbackend/pkg/migrations"
)

func main() {
	// Chargement de la configuration
	cfg := configs.Load()

	// Affichage bannière de démarrage
	displayBanner(cfg)

	// Connexion base de données
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("❌ Erreur connexion MySQL: %v", err)
	}
	defer database.Close()
	log.Println("✅ Base de données connectée")

	// Auto-migrations: apply all pending SQL files from migrations/
	if err := migrations.Run(database.DB, "migrations"); err != nil {
		log.Fatalf("❌ Erreur migrations: %v", err)
	}
	log.Println("✅ Migrations terminées")

	// Configuration du router avec support des templates
	handler := router.Init(cfg)

	// Configuration du serveur avec timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Démarrage du serveur
	log.Printf("🚀 %s démarré sur http://localhost:%s\n", cfg.App.Name, cfg.App.Port)
	log.Printf("📝 Environment: %s\n", cfg.App.Environment)
	log.Printf("🌐 Templates: Chargés depuis ../frontend/\n")
	log.Printf("📁 Fichiers statiques: /styles/ → ../frontend/styles/\n")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("❌ Erreur démarrage serveur: %v", err)
	}
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
	fmt.Printf("🔗 Templates Go + Frontend intégrés\n")
	fmt.Println("================================================")
}
