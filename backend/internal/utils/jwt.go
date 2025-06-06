package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims structure pour les claims du token JWT
type JWTClaims struct {
	UserID  uint `json:"user_id"`
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

var (
	jwtSecret       = []byte("your-secret-key") // TODO: Move to environment variable
	ErrInvalidToken = errors.New("token invalide")
	ErrExpiredToken = errors.New("token expiré")
)

// GenerateJWTToken génère un nouveau token JWT
func GenerateJWTToken(userID uint, isAdmin bool) (string, error) {
	claims := &JWTClaims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWTToken valide un token JWT et retourne ses claims
func ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// GetUserIDFromToken extrait l'ID utilisateur du token
func GetUserIDFromToken(tokenString string) (uint, error) {
	token, err := ValidateJWTToken(tokenString)
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}

// IsAdminFromToken vérifie si l'utilisateur est admin
func IsAdminFromToken(tokenString string) (bool, error) {
	token, err := ValidateJWTToken(tokenString)
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return false, ErrInvalidToken
	}

	return claims.IsAdmin, nil
}
