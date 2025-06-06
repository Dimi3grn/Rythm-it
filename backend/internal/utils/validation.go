package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Validator instance globale
var validate *validator.Validate

// Patterns de validation
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlRegex      = regexp.MustCompile(`^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`)
)

func init() {
	validate = validator.New()

	// Enregistrer des validations personnalisées
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("email", validateEmail)
	validate.RegisterValidation("url", validateURL)
	validate.RegisterValidation("nohtml", validateNoHTML)
	validate.RegisterValidation("alphanumspace", validateAlphaNumSpace)
}

// ValidateStruct valide une structure selon ses tags
func ValidateStruct(s interface{}) []ValidationError {
	var errors []ValidationError

	err := validate.Struct(s)
	if err == nil {
		return errors
	}

	// Convertir les erreurs validator en nos ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		var message string

		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s est requis", err.Field())
		case "email":
			message = "Email invalide"
		case "min":
			message = fmt.Sprintf("%s doit contenir au moins %s caractères", err.Field(), err.Param())
		case "max":
			message = fmt.Sprintf("%s ne doit pas dépasser %s caractères", err.Field(), err.Param())
		case "password":
			message = "Le mot de passe doit contenir au moins 12 caractères, une majuscule, une minuscule, un chiffre et un caractère spécial"
		case "username":
			message = "Le nom d'utilisateur doit contenir entre 3 et 30 caractères (lettres, chiffres et underscores uniquement)"
		case "url":
			message = "URL invalide"
		case "nohtml":
			message = "Le HTML n'est pas autorisé"
		case "alphanumspace":
			message = "Seuls les lettres, chiffres et espaces sont autorisés"
		case "oneof":
			message = fmt.Sprintf("%s doit être l'une des valeurs suivantes: %s", err.Field(), err.Param())
		default:
			message = fmt.Sprintf("%s est invalide", err.Field())
		}

		errors = append(errors, ValidationError{
			Field:   strings.ToLower(err.Field()),
			Message: message,
		})
	}

	return errors
}

// validatePassword valide un mot de passe selon les critères du projet
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 12 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsername valide un nom d'utilisateur
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

// validateEmail valide une adresse email
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return emailRegex.MatchString(email)
}

// validateURL valide une URL
func validateURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	return urlRegex.MatchString(url)
}

// validateNoHTML vérifie qu'une chaîne ne contient pas de HTML
func validateNoHTML(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	return !strings.ContainsAny(str, "<>")
}

// validateAlphaNumSpace vérifie qu'une chaîne ne contient que des lettres, chiffres et espaces
func validateAlphaNumSpace(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	for _, char := range str {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && !unicode.IsSpace(char) {
			return false
		}
	}
	return true
}

// SanitizeString nettoie une chaîne de caractères
func SanitizeString(s string) string {
	// Supprimer les espaces en début et fin
	s = strings.TrimSpace(s)

	// Remplacer les caractères HTML par leurs entités
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")

	// Supprimer les caractères de contrôle
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)

	return s
}

// ValidateAndSanitize valide et nettoie une structure
func ValidateAndSanitize(s interface{}) ([]ValidationError, error) {
	// Valider d'abord
	errors := ValidateStruct(s)
	if len(errors) > 0 {
		return errors, nil
	}

	// TODO: Implémenter la sanitization automatique des champs string
	// Cela nécessiterait de la réflexion pour parcourir la structure
	// et nettoyer tous les champs string

	return nil, nil
}

// ValidateEmail validates an email address using a regex pattern
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}
