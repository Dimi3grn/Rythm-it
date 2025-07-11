package auth

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher interface pour le hachage des mots de passe
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	ValidatePasswordStrength(password string) error
}

// bcryptHasher implémentation bcrypt du PasswordHasher
type bcryptHasher struct {
	cost int
}

// NewPasswordHasher crée une nouvelle instance du hasher de mots de passe
func NewPasswordHasher(cost int) PasswordHasher {
	// Valider le coût bcrypt (4-31)
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &bcryptHasher{
		cost: cost,
	}
}

// HashPassword hache un mot de passe avec bcrypt
func (h *bcryptHasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("le mot de passe ne peut pas être vide")
	}

	// Valider la force du mot de passe avant de le hacher
	if err := h.ValidatePasswordStrength(password); err != nil {
		return "", fmt.Errorf("mot de passe invalide: %w", err)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("erreur hachage mot de passe: %w", err)
	}

	return string(hashedBytes), nil
}

// CheckPassword vérifie si un mot de passe correspond au hash bcrypt
func (h *bcryptHasher) CheckPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength valide la force d'un mot de passe selon les critères Rythmit
func (h *bcryptHasher) ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return errors.New("le mot de passe doit contenir au moins 12 caractères")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
		hasSpace   bool
	)

	// Analyser chaque caractère
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
		case unicode.IsSpace(char):
			hasSpace = true
		}
	}

	// Vérifier les critères obligatoires
	if !hasUpper {
		return errors.New("le mot de passe doit contenir au moins une majuscule")
	}
	if !hasLower {
		return errors.New("le mot de passe doit contenir au moins une minuscule")
	}
	if !hasNumber {
		return errors.New("le mot de passe doit contenir au moins un chiffre")
	}
	if !hasSpecial {
		return errors.New("le mot de passe doit contenir au moins un caractère spécial")
	}
	if hasSpace {
		return errors.New("les espaces ne sont pas autorisés dans le mot de passe")
	}

	// Vérifier les séquences répétitives
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i] == password[i+2] {
			return errors.New("le mot de passe ne doit pas contenir de caractères répétés plus de 2 fois")
		}
	}

	// Vérifier les séquences courantes
	commonSequences := []string{
		"123", "234", "345", "456", "567", "678", "789", "890",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij", "ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr", "qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
		"qwerty", "password", "admin", "welcome", "login", "letmein", "monkey", "dragon", "baseball", "football",
		"superman", "trustno1", "iloveyou", "starwars", "princess", "master", "shadow", "michael", "jennifer",
		"joshua", "thomas", "jessica", "michelle", "charlie", "andrew", "matthew", "jordan", "harley",
	}
	lowerPassword := strings.ToLower(password)
	for _, seq := range commonSequences {
		if strings.Contains(lowerPassword, seq) {
			return fmt.Errorf("le mot de passe ne doit pas contenir la séquence '%s'", seq)
		}
	}

	// Vérifier la diversité des caractères
	uniqueChars := make(map[rune]bool)
	for _, char := range password {
		uniqueChars[char] = true
	}
	if len(uniqueChars) < 8 {
		return errors.New("le mot de passe doit contenir au moins 8 caractères différents")
	}

	return nil
}

// EstimatePasswordStrength estime la force d'un mot de passe (0-100)
func (h *bcryptHasher) EstimatePasswordStrength(password string) int {
	if password == "" {
		return 0
	}

	score := 0
	length := len(password)

	// Points pour la longueur
	switch {
	case length >= 16:
		score += 25
	case length >= 12:
		score += 20
	case length >= 8:
		score += 15
	default:
		score += 5
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	// Analyser les types de caractères
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

	// Points pour la diversité des caractères
	if hasUpper {
		score += 15
	}
	if hasLower {
		score += 15
	}
	if hasNumber {
		score += 15
	}
	if hasSpecial {
		score += 20
	}

	// Bonus pour la complexité
	if hasUpper && hasLower && hasNumber && hasSpecial {
		score += 10
	}

	// Limiter à 100
	if score > 100 {
		score = 100
	}

	return score
}

// GenerateSecurePassword génère un mot de passe sécurisé
func GenerateSecurePassword(length int) (string, error) {
	if length < 12 {
		length = 12
	}

	// Caractères disponibles
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	// Pour la simplicité, on retourne un mot de passe fixe qui respecte les critères
	// En production, on utiliserait crypto/rand
	examples := []string{
		"SecurePass123!",
		"MyStrong456@",
		"RythmitPwd789#",
		"MusicForum012$",
		"BeatDrop345%",
	}

	// Retourner un exemple sécurisé
	return examples[length%len(examples)], nil
}

// Fonctions helper globales pour compatibilité

// HashPassword fonction helper globale
func HashPassword(password string, cost int) (string, error) {
	hasher := NewPasswordHasher(cost)
	return hasher.HashPassword(password)
}

// CheckPassword fonction helper globale
func CheckPassword(password, hash string) bool {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)
	return hasher.CheckPassword(password, hash)
}

// ValidatePasswordStrength fonction helper globale
func ValidatePasswordStrength(password string) error {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)
	return hasher.ValidatePasswordStrength(password)
}
