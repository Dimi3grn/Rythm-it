package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadImageHandler gère l'upload d'images pour les profils et threads
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📤 UploadImageHandler appelé - Method: %s", r.Method)
	
	// Vérifier que c'est une requête POST
	if r.Method != "POST" {
		log.Printf("❌ Méthode non autorisée: %s", r.Method)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		log.Printf("❌ Utilisateur non authentifié")
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	log.Printf("📤 UploadImageHandler appelé pour utilisateur: %s", user.Username)

	// Parser le formulaire multipart (limite à 10MB)
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		log.Printf("❌ Erreur parsing formulaire: %v", err)
		http.Error(w, "Erreur parsing formulaire", http.StatusBadRequest)
		return
	}

	// Récupérer le fichier
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		log.Printf("❌ Erreur récupération fichier: %v", err)
		http.Error(w, "Erreur récupération fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("📁 Fichier reçu: %s, Taille: %d bytes, Type: %s", 
		fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"))

	// Vérifier le type de fichier
	contentType := fileHeader.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		log.Printf("❌ Type de fichier non autorisé: %s", contentType)
		http.Error(w, "Type de fichier non autorisé. Seules les images PNG, JPG et JPEG sont acceptées.", http.StatusBadRequest)
		return
	}

	// Vérifier la taille du fichier (max 5MB)
	if fileHeader.Size > 5<<20 { // 5 MB
		log.Printf("❌ Fichier trop volumineux: %d bytes", fileHeader.Size)
		http.Error(w, "Fichier trop volumineux. Taille maximale: 5MB", http.StatusBadRequest)
		return
	}

	// Déterminer le type d'upload et le dossier de destination
	uploadType := r.FormValue("type") // "profile" ou "thread"
	var uploadDir string

	switch uploadType {
	case "profile":
		uploadDir = "uploads/profiles"
	case "thread":
		uploadDir = "uploads/threads"
	default:
		// Par défaut, considérer comme un upload de thread si pas spécifié
		uploadDir = "uploads/threads"
	}

	log.Printf("📁 Type d'upload: %s, Dossier: %s", uploadType, uploadDir)

	// Générer un nom de fichier unique
	fileName, err := generateUniqueFileName(fileHeader.Filename)
	if err != nil {
		log.Printf("❌ Erreur génération nom fichier: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	log.Printf("📝 Nom de fichier généré: %s", fileName)

	// Créer le répertoire s'il n'existe pas
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("❌ Erreur création dossier: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(uploadDir, fileName)

	// Debug: afficher le chemin complet
	absPath, _ := filepath.Abs(filePath)
	log.Printf("📁 Sauvegarde image vers: %s", absPath)

	// Créer le fichier de destination
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("❌ Erreur création fichier: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copier le contenu du fichier
	bytesWritten, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("❌ Erreur sauvegarde fichier: %v", err)
		http.Error(w, "Erreur sauvegarde", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Fichier sauvegardé: %d bytes écrits", bytesWritten)

	// URL publique de l'image
	imageURL := fmt.Sprintf("/%s/%s", uploadDir, fileName)

	log.Printf("✅ Image uploadée avec succès: %s", imageURL)

	// Retourner l'URL de l'image en JSON
	w.Header().Set("Content-Type", "application/json")
	response := fmt.Sprintf(`{"success": true, "imageUrl": "%s", "message": "Image uploadée avec succès"}`, imageURL)
	w.Write([]byte(response))
}

// isValidImageType vérifie si le type de fichier est une image valide
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// generateUniqueFileName génère un nom de fichier unique
func generateUniqueFileName(originalName string) (string, error) {
	// Récupérer l'extension
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".jpg" // Extension par défaut
	}

	// Générer un identifiant unique
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Créer le nom avec timestamp + random
	timestamp := time.Now().Unix()
	randomStr := hex.EncodeToString(bytes)

	fileName := fmt.Sprintf("%d_%s%s", timestamp, randomStr[:12], ext)
	return fileName, nil
}
