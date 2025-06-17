// Fichier: backend/internal/services/validation_service.go
package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationService gère la validation des données côté serveur
type ValidationService struct{}

// NewValidationService crée une nouvelle instance du service de validation
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidationError représente une erreur de validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationResult représente le résultat d'une validation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}

// ThreadValidationData structure pour valider les données de thread
type ThreadValidationData struct {
	Content  string   `json:"content"`
	Tags     []string `json:"tags,omitempty"`
	Genre    string   `json:"genre,omitempty"`
	ImageURL *string  `json:"image_url,omitempty"`
}

// UserValidationData structure pour valider les données utilisateur
type UserValidationData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// CommentValidationData structure pour valider les commentaires
type CommentValidationData struct {
	Content  string  `json:"content"`
	ImageURL *string `json:"image_url,omitempty"`
}

// ValidateThread valide les données d'un thread
func (vs *ValidationService) ValidateThread(data ThreadValidationData) ValidationResult {
	var errors []ValidationError

	// Validation du contenu
	if err := vs.validateContent(data.Content); err != nil {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: err.Error(),
			Code:    "INVALID_CONTENT",
		})
	}

	// Validation des tags
	if err := vs.validateTags(data.Tags); err != nil {
		errors = append(errors, ValidationError{
			Field:   "tags",
			Message: err.Error(),
			Code:    "INVALID_TAGS",
		})
	}

	// Validation du genre
	if err := vs.validateGenre(data.Genre); err != nil {
		errors = append(errors, ValidationError{
			Field:   "genre",
			Message: err.Error(),
			Code:    "INVALID_GENRE",
		})
	}

	// Validation de l'URL d'image
	if data.ImageURL != nil {
		if err := vs.validateImageURL(*data.ImageURL); err != nil {
			errors = append(errors, ValidationError{
				Field:   "image_url",
				Message: err.Error(),
				Code:    "INVALID_IMAGE_URL",
			})
		}
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateUser valide les données d'un utilisateur
func (vs *ValidationService) ValidateUser(data UserValidationData) ValidationResult {
	var errors []ValidationError

	// Validation du nom d'utilisateur
	if err := vs.validateUsername(data.Username); err != nil {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: err.Error(),
			Code:    "INVALID_USERNAME",
		})
	}

	// Validation de l'email
	if err := vs.validateEmail(data.Email); err != nil {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: err.Error(),
			Code:    "INVALID_EMAIL",
		})
	}

	// Validation du mot de passe (si fourni)
	if data.Password != "" {
		if err := vs.validatePassword(data.Password); err != nil {
			errors = append(errors, ValidationError{
				Field:   "password",
				Message: err.Error(),
				Code:    "INVALID_PASSWORD",
			})
		}
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateComment valide les données d'un commentaire
func (vs *ValidationService) ValidateComment(data CommentValidationData) ValidationResult {
	var errors []ValidationError

	// Validation du contenu
	if err := vs.validateContent(data.Content); err != nil {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: err.Error(),
			Code:    "INVALID_CONTENT",
		})
	}

	// Validation de l'URL d'image
	if data.ImageURL != nil {
		if err := vs.validateImageURL(*data.ImageURL); err != nil {
			errors = append(errors, ValidationError{
				Field:   "image_url",
				Message: err.Error(),
				Code:    "INVALID_IMAGE_URL",
			})
		}
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// Méthodes privées de validation

func (vs *ValidationService) validateContent(content string) error {
	content = strings.TrimSpace(content)

	if content == "" {
		return errors.New("Le contenu ne peut pas être vide")
	}

	if utf8.RuneCountInString(content) < 3 {
		return errors.New("Le contenu doit contenir au moins 3 caractères")
	}

	if utf8.RuneCountInString(content) > 5000 {
		return errors.New("Le contenu ne peut pas dépasser 5000 caractères")
	}

	// Vérifier les caractères interdits ou potentiellement dangereux
	if strings.Contains(content, "<script") || strings.Contains(content, "javascript:") {
		return errors.New("Le contenu contient des éléments non autorisés")
	}

	return nil
}

func (vs *ValidationService) validateTags(tags []string) error {
	if len(tags) > 10 {
		return errors.New("Vous ne pouvez pas ajouter plus de 10 tags")
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return errors.New("Les tags ne peuvent pas être vides")
		}
		if utf8.RuneCountInString(tag) > 50 {
			return errors.New("Les tags ne peuvent pas dépasser 50 caractères")
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`).MatchString(tag) {
			return fmt.Errorf("Le tag '%s' contient des caractères non autorisés", tag)
		}
	}

	return nil
}

func (vs *ValidationService) validateGenre(genre string) error {
	if genre == "" {
		return nil // Le genre est optionnel
	}

	validGenres := []string{
		"rock", "pop", "jazz", "classical", "electronic", "hip-hop",
		"country", "blues", "reggae", "folk", "metal", "punk",
		"r&b", "soul", "funk", "disco", "house", "techno",
		"ambient", "experimental", "world", "latin", "indie", "alternative",
	}

	genre = strings.ToLower(strings.TrimSpace(genre))
	for _, validGenre := range validGenres {
		if genre == validGenre {
			return nil
		}
	}

	return fmt.Errorf("Le genre '%s' n'est pas valide", genre)
}

func (vs *ValidationService) validateImageURL(imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)

	if imageURL == "" {
		return nil // URL optionnelle
	}

	// Vérifier que c'est une URL locale (sécurité)
	if !strings.HasPrefix(imageURL, "/uploads/") {
		return errors.New("L'URL de l'image doit pointer vers un fichier uploadé")
	}

	// Vérifier l'extension
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	hasValidExtension := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(imageURL), ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return errors.New("L'image doit être au format JPG, PNG, GIF ou WebP")
	}

	return nil
}

func (vs *ValidationService) validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.New("Le nom d'utilisateur ne peut pas être vide")
	}

	if utf8.RuneCountInString(username) < 3 {
		return errors.New("Le nom d'utilisateur doit contenir au moins 3 caractères")
	}

	if utf8.RuneCountInString(username) > 30 {
		return errors.New("Le nom d'utilisateur ne peut pas dépasser 30 caractères")
	}

	// Vérifier le format (lettres, chiffres, tirets et underscores uniquement)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(username) {
		return errors.New("Le nom d'utilisateur ne peut contenir que des lettres, chiffres, tirets et underscores")
	}

	// Vérifier que ça ne commence pas par un chiffre
	if regexp.MustCompile(`^[0-9]`).MatchString(username) {
		return errors.New("Le nom d'utilisateur ne peut pas commencer par un chiffre")
	}

	return nil
}

func (vs *ValidationService) validateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return errors.New("L'adresse email ne peut pas être vide")
	}

	// Regex simple pour validation email
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("L'adresse email n'est pas valide")
	}

	if utf8.RuneCountInString(email) > 255 {
		return errors.New("L'adresse email est trop longue")
	}

	return nil
}

func (vs *ValidationService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Le mot de passe doit contenir au moins 8 caractères")
	}

	if len(password) > 100 {
		return errors.New("Le mot de passe ne peut pas dépasser 100 caractères")
	}

	// Vérifier qu'il contient au moins une majuscule, une minuscule et un chiffre
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper {
		return errors.New("Le mot de passe doit contenir au moins une majuscule")
	}

	if !hasLower {
		return errors.New("Le mot de passe doit contenir au moins une minuscule")
	}

	if !hasDigit {
		return errors.New("Le mot de passe doit contenir au moins un chiffre")
	}

	return nil
}

// SanitizeInput nettoie et sécurise les entrées utilisateur
func (vs *ValidationService) SanitizeInput(input string) string {
	// Supprimer les espaces en début/fin
	input = strings.TrimSpace(input)

	// Supprimer les caractères de contrôle
	input = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(input, "")

	// Échapper les caractères HTML dangereux
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	input = strings.ReplaceAll(input, "&", "&amp;")

	return input
}

// ValidateAndSanitize valide et nettoie simultanément les données
func (vs *ValidationService) ValidateAndSanitize(data interface{}) (interface{}, ValidationResult) {
	switch v := data.(type) {
	case ThreadValidationData:
		v.Content = vs.SanitizeInput(v.Content)
		for i, tag := range v.Tags {
			v.Tags[i] = vs.SanitizeInput(tag)
		}
		v.Genre = vs.SanitizeInput(v.Genre)
		return v, vs.ValidateThread(v)

	case UserValidationData:
		v.Username = vs.SanitizeInput(v.Username)
		v.Email = vs.SanitizeInput(v.Email)
		// Ne pas sanitizer le mot de passe
		return v, vs.ValidateUser(v)

	case CommentValidationData:
		v.Content = vs.SanitizeInput(v.Content)
		return v, vs.ValidateComment(v)

	default:
		return data, ValidationResult{
			IsValid: false,
			Errors: []ValidationError{{
				Field:   "unknown",
				Message: "Type de données non supporté pour la validation",
				Code:    "UNSUPPORTED_DATA_TYPE",
			}},
		}
	}
}
