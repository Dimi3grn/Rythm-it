package main

import (
	"io/ioutil"
	"log"
	"rythmitbackend/configs"
	"rythmitbackend/pkg/database"
)

func main() {
	// Charger la configuration
	cfg := configs.Load()

	// Se connecter à la base de données
	log.Println("🔄 Connexion à la base de données...")
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("❌ Erreur connexion: %v", err)
	}
	defer database.Close()

	// Lire le fichier de migration
	log.Println("📄 Lecture du fichier de migration...")
	migrationSQL, err := ioutil.ReadFile("migrations/003_add_avatar_to_profiles.sql")
	if err != nil {
		log.Fatalf("❌ Erreur lecture migration: %v", err)
	}

	// Exécuter la migration
	log.Println("🔄 Exécution de la migration...")
	_, err = database.DB.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("❌ Erreur exécution migration: %v", err)
	}

	log.Println("✅ Migration exécutée avec succès!")
	log.Println("📋 Table user_profiles créée")
}
