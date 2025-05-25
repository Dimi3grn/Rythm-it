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

func init() {
	validate = validator.New()

	// Enregistrer des validations personnalisées
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
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
			message = "Le mot de passe doit contenir au moins 12 caractères, une majuscule et un caractère spécial"
		case "username":
			message = "Le nom d'utilisateur ne doit contenir que des lettres, chiffres et underscores"
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

	// Au moins 12 caractères
	if len(password) < 12 {
		return false
	}

	var (
		hasUpper   bool
		hasSpecial bool
	)

	// Vérifier la présence d'une majuscule et d'un caractère spécial
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case !unicode.IsLetter(char) && !unicode.IsNumber(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasSpecial
}

// validateUsername valide un nom d'utilisateur
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Regex pour lettres, chiffres et underscore uniquement
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	return matched
}

// ValidateEmail vérifie si un email est valide
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

// ValidatePassword vérifie si un mot de passe respecte les critères
func ValidatePassword(password string) (bool, string) {
	if len(password) < 12 {
		return false, "Le mot de passe doit contenir au moins 12 caractères"
	}

	var (
		hasUpper   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case !unicode.IsLetter(char) && !unicode.IsNumber(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Le mot de passe doit contenir au moins une majuscule"
	}

	if !hasSpecial {
		return false, "Le mot de passe doit contenir au moins un caractère spécial"
	}

	return true, ""
}
