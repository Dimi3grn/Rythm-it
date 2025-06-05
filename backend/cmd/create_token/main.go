// Créer le fichier: backend/cmd/create_token/main.go

package main

import (
	"fmt"
	"log"
	"time"

	"rythmitbackend/configs"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Charger la config pour récupérer JWT_SECRET
	cfg := configs.Load()

	// Claims compatibles avec votre middleware
	claims := jwt.MapClaims{
		"user_id":  float64(1), // JWT stocke en float64
		"username": "admin",
		"email":    "admin@rythmit.com",
		"is_admin": true,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Expire dans 24h
		"iat":      time.Now().Unix(),                     // Créé maintenant
		"nbf":      time.Now().Unix(),                     // Valide à partir de maintenant
	}

	// Créer le token avec HS256 (même méthode que votre middleware)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer avec la même clé secrète que votre .env
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		log.Fatalf("❌ Erreur création token: %v", err)
	}

	fmt.Println("🎯 TOKEN JWT VALIDE CRÉÉ!")
	fmt.Println("🔑 Clé utilisée:", cfg.JWT.Secret)
	fmt.Println("⏰ Expire dans 24h")
	fmt.Println()
	fmt.Println("📋 TOKEN À COPIER:")
	fmt.Println(tokenString)
	fmt.Println()
	fmt.Println("💻 COMMANDE POWERSHELL COMPLÈTE:")
	fmt.Printf("$token = \"%s\"\n", tokenString)
	fmt.Printf("$headers = @{ \"Authorization\" = \"Bearer $token\" }\n")
	fmt.Printf("Invoke-RestMethod -Uri \"http://localhost:8085/api/v1/profile\" -Method GET -Headers $headers\n")
	fmt.Println()
	fmt.Println("✅ Ce token est compatible avec votre middleware actuel!")
}
