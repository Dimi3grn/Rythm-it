package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager interface pour la gestion des tokens JWT
type TokenManager interface {
	GenerateToken(claims Claims) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(oldToken string) (string, error)
	ExtractClaims(tokenString string) (*Claims, error)
}

// Claims structure personnalisée pour les claims JWT Rythmit
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Config configuration du token manager
type Config struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// tokenManager implémentation du TokenManager
type tokenManager struct {
	config Config
}

// NewTokenManager crée une nouvelle instance du gestionnaire de tokens
func NewTokenManager(config Config) TokenManager {
	// Valeurs par défaut
	if config.ExpirationHours <= 0 {
		config.ExpirationHours = 24
	}
	if config.Issuer == "" {
		config.Issuer = "rythmit-api"
	}
	if config.Secret == "" {
		config.Secret = "default-secret-change-in-production"
	}

	return &tokenManager{
		config: config,
	}
}

// GenerateToken génère un nouveau token JWT avec les claims spécifiés
func (tm *tokenManager) GenerateToken(claims Claims) (string, error) {
	// Définir l'expiration
	expirationTime := time.Now().Add(time.Duration(tm.config.ExpirationHours) * time.Hour)

	// Définir les claims standard
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    tm.config.Issuer,
		Subject:   fmt.Sprintf("%d", claims.UserID),
		Audience:  []string{"rythmit-web", "rythmit-mobile"},
	}

	// Créer le token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer le token
	tokenString, err := token.SignedString([]byte(tm.config.Secret))
	if err != nil {
		return "", fmt.Errorf("erreur signature token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken valide un token JWT et retourne les claims
func (tm *tokenManager) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token vide")
	}

	// Parser le token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(tm.config.Secret), nil
	})

	if err != nil {
		// Gestion spécifique des erreurs JWT
		if err.Error() == jwt.ErrTokenExpired.Error() {
			return nil, errors.New("token expiré")
		}
		if err.Error() == jwt.ErrTokenNotValidYet.Error() {
			return nil, errors.New("token pas encore valide")
		}
		if err.Error() == jwt.ErrTokenMalformed.Error() {
			return nil, errors.New("token malformé")
		}
		return nil, fmt.Errorf("token invalide: %w", err)
	}

	// Vérifier que le token est valide
	if !token.Valid {
		return nil, errors.New("token invalide")
	}

	// Extraire les claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("impossible d'extraire les claims")
	}

	// Validation supplémentaire des claims
	if err := tm.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("claims invalides: %w", err)
	}

	return claims, nil
}

// RefreshToken génère un nouveau token à partir d'un ancien (même si expiré récemment)
func (tm *tokenManager) RefreshToken(oldToken string) (string, error) {
	// Parser le token sans validation de l'expiration
	token, err := jwt.ParseWithClaims(oldToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(tm.config.Secret), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", fmt.Errorf("impossible de parser le token pour refresh: %w", err)
	}

	// Extraire les claims
	oldClaims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("impossible d'extraire les claims du token")
	}

	// Vérifier que le token n'est pas trop ancien (ex: max 7 jours)
	maxRefreshAge := 7 * 24 * time.Hour
	if time.Since(oldClaims.IssuedAt.Time) > maxRefreshAge {
		return "", errors.New("token trop ancien pour être rafraîchi")
	}

	// Créer de nouvelles claims avec les mêmes données utilisateur
	newClaims := Claims{
		UserID:   oldClaims.UserID,
		Username: oldClaims.Username,
		Email:    oldClaims.Email,
		IsAdmin:  oldClaims.IsAdmin,
	}

	// Générer un nouveau token
	return tm.GenerateToken(newClaims)
}

// ExtractClaims extrait les claims d'un token sans validation complète (utile pour debugging)
func (tm *tokenManager) ExtractClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tm.config.Secret), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("impossible d'extraire les claims: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("format de claims invalide")
	}

	return claims, nil
}

// validateClaims valide les claims personnalisées
func (tm *tokenManager) validateClaims(claims *Claims) error {
	// Vérifier les champs obligatoires
	if claims.UserID == 0 {
		return errors.New("UserID manquant")
	}

	if claims.Username == "" {
		return errors.New("Username manquant")
	}

	if claims.Email == "" {
		return errors.New("Email manquant")
	}

	// Vérifier le subject
	expectedSubject := fmt.Sprintf("%d", claims.UserID)
	if claims.Subject != expectedSubject {
		return fmt.Errorf("subject invalide: attendu %s, obtenu %s", expectedSubject, claims.Subject)
	}

	return nil
}

// Helper functions globales

// GenerateToken fonction helper globale
func GenerateToken(userID uint, username, email string, isAdmin bool, secret string, expirationHours int) (string, error) {
	config := Config{
		Secret:          secret,
		ExpirationHours: expirationHours,
		Issuer:          "rythmit-api",
	}

	tm := NewTokenManager(config)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		IsAdmin:  isAdmin,
	}

	return tm.GenerateToken(claims)
}

// ValidateToken fonction helper globale
func ValidateToken(tokenString, secret string) (*Claims, error) {
	config := Config{
		Secret: secret,
	}

	tm := NewTokenManager(config)
	return tm.ValidateToken(tokenString)
}

// GetTokenExpiration retourne l'expiration d'un token (utile pour les cookies)
func GetTokenExpiration(tokenString, secret string) (*time.Time, error) {
	config := Config{
		Secret: secret,
	}

	tm := NewTokenManager(config)
	claims, err := tm.ExtractClaims(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt == nil {
		return nil, errors.New("aucune date d'expiration dans le token")
	}

	expiration := claims.ExpiresAt.Time
	return &expiration, nil
}

// IsTokenExpired vérifie si un token est expiré
func IsTokenExpired(tokenString, secret string) (bool, error) {
	expiration, err := GetTokenExpiration(tokenString, secret)
	if err != nil {
		return true, err
	}

	return time.Now().After(*expiration), nil
}
