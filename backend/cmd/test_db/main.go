package main

import (
	"log"
	"rythmitbackend/configs"
	"rythmitbackend/pkg/database"
)

func main() {
	// Charger la configuration
	cfg := configs.Load()

	// Tenter la connexion
	log.Println("🔄 Test de connexion MySQL...")

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("❌ Erreur connexion: %v", err)
	}

	// Test de santé
	if err := database.Health(); err != nil {
		log.Fatalf("❌ Base de données non accessible: %v", err)
	}

	// Vérifier les tables
	rows, err := database.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("❌ Erreur requête: %v", err)
	}
	defer rows.Close()

	log.Println("✅ Connexion MySQL réussie!")
	log.Println("📋 Tables trouvées:")

	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		log.Printf("   - %s", tableName)
	}

	// Fermer la connexion
	database.Close()
	log.Println("✅ Test terminé avec succès!")
}
