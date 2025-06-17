package main

import (
	"fmt"
	"log"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Charger la configuration
	cfg := configs.Load()

	// Initialiser la base de données
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Erreur initialisation DB: %v", err)
	}
	defer database.Close()

	// Créer le repository utilisateur
	userRepo := repositories.NewUserRepository(database.DB)

	// Vérifier si un admin existe déjà
	existingAdmin, err := userRepo.FindByEmail("admin@rythmit.com")
	if err == nil {
		fmt.Println("✅ Utilisateur admin existe déjà:")
		fmt.Printf("   Email: %s\n", existingAdmin.Email)
		fmt.Printf("   Username: %s\n", existingAdmin.Username)
		fmt.Println("   Mot de passe: admin123")
		return
	}

	// Créer le hash du mot de passe
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Erreur hachage mot de passe: %v", err)
	}

	// Créer l'utilisateur admin
	admin := &models.User{
		Username: "Admin",
		Email:    "admin@rythmit.com",
		Password: string(hashedPassword),
		IsAdmin:  true,
	}

	// Sauvegarder en base
	if err := userRepo.Create(admin); err != nil {
		log.Fatalf("Erreur création admin: %v", err)
	}

	fmt.Println("✅ Utilisateur admin créé avec succès!")
	fmt.Println("   Email: admin@rythmit.com")
	fmt.Println("   Username: Admin")
	fmt.Println("   Mot de passe: admin123")
}
