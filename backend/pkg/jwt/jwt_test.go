package jwt

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createTestConfig() Config {
	return Config{
		Secret:          "test-secret-key",
		ExpirationHours: 1,
		Issuer:          "test-rythmit",
	}
}

func createTestClaims() Claims {
	return Claims{
		UserID:   123,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}
}

func TestNewTokenManager(t *testing.T) {
	config := createTestConfig()
	tm := NewTokenManager(config)

	if tm == nil {
		t.Error("NewTokenManager ne devrait pas retourner nil")
	}

	// Test avec config vide
	emptyConfig := Config{}
	tm2 := NewTokenManager(emptyConfig)

	if tm2 == nil {
		t.Error("NewTokenManager devrait fonctionner avec config vide (valeurs par défaut)")
	}
}

func TestTokenManager_GenerateToken(t *testing.T) {
	tm := NewTokenManager(createTestConfig())
	claims := createTestClaims()

	token, err := tm.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Erreur génération token: %v", err)
	}

	if token == "" {
		t.Error("Le token généré ne devrait pas être vide")
	}

	// Vérifier que le token a 3 parties (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Token devrait avoir 3 parties, obtenu: %d", len(parts))
	}
}

func TestTokenManager_ValidateToken(t *testing.T) {
	tm := NewTokenManager(createTestConfig())
	originalClaims := createTestClaims()

	// Générer un token valide
	token, err := tm.GenerateToken(originalClaims)
	if err != nil {
		t.Fatalf("Erreur génération token: %v", err)
	}

	// Test validation token valide
	validatedClaims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("Erreur validation token: %v", err)
	}

	// Vérifications des claims
	if validatedClaims.UserID != originalClaims.UserID {
		t.Errorf("UserID attendu: %d, obtenu: %d", originalClaims.UserID, validatedClaims.UserID)
	}

	if validatedClaims.Username != originalClaims.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", originalClaims.Username, validatedClaims.Username)
	}

	if validatedClaims.Email != originalClaims.Email {
		t.Errorf("Email attendu: %s, obtenu: %s", originalClaims.Email, validatedClaims.Email)
	}

	if validatedClaims.IsAdmin != originalClaims.IsAdmin {
		t.Errorf("IsAdmin attendu: %t, obtenu: %t", originalClaims.IsAdmin, validatedClaims.IsAdmin)
	}

	// Vérifier les claims standard
	if validatedClaims.Issuer != "test-rythmit" {
		t.Errorf("Issuer attendu: test-rythmit, obtenu: %s", validatedClaims.Issuer)
	}

	expectedSubject := "123"
	if validatedClaims.Subject != expectedSubject {
		t.Errorf("Subject attendu: %s, obtenu: %s", expectedSubject, validatedClaims.Subject)
	}
}

func TestTokenManager_ValidateToken_Errors(t *testing.T) {
	tm := NewTokenManager(createTestConfig())

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Token vide",
			token:       "",
			expectError: true,
			errorMsg:    "token vide",
		},
		{
			name:        "Token malformé",
			token:       "invalid.token",
			expectError: true,
			errorMsg:    "token invalide",
		},
		{
			name:        "Token avec mauvaise signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectError: true,
			errorMsg:    "token invalide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm.ValidateToken(tt.token)

			if !tt.expectError {
				if err != nil {
					t.Errorf("Erreur inattendue: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("Devrait retourner une erreur")
				return
			}

			if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Erreur devrait contenir %q, obtenu: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestTokenManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Créer un config avec expiration très courte
	config := Config{
		Secret:          "test-secret",
		ExpirationHours: 1, // Durée normale
		Issuer:          "test",
	}

	tm := NewTokenManager(config)
	claims := createTestClaims()

	// Créer manuellement un token avec expiration passée
	expiredClaims := Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		IsAdmin:  claims.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expiré depuis 1h
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "test",
			Subject:   "123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Erreur création token expiré: %v", err)
	}

	// Le token devrait être rejeté car expiré
	_, err = tm.ValidateToken(tokenString)

	if err == nil {
		t.Error("Le token expiré devrait être rejeté")
	} else if !strings.Contains(err.Error(), "expired") {
		t.Errorf("L'erreur devrait mentionner l'expiration, obtenu: %v", err)
	}
}

func TestTokenManager_RefreshToken(t *testing.T) {
	tm := NewTokenManager(createTestConfig())
	originalClaims := createTestClaims()

	// Générer un token initial
	originalToken, err := tm.GenerateToken(originalClaims)
	if err != nil {
		t.Fatalf("Erreur génération token initial: %v", err)
	}

	// Attendre un peu pour s'assurer que les timestamps diffèrent
	time.Sleep(time.Second * 1)

	// Rafraîchir le token
	newToken, err := tm.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Erreur refresh token: %v", err)
	}

	if newToken == "" {
		t.Error("Le nouveau token ne devrait pas être vide")
	}

	if newToken == originalToken {
		t.Error("Le nouveau token devrait être différent de l'original")
	}

	// Valider le nouveau token
	newClaims, err := tm.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("Erreur validation nouveau token: %v", err)
	}

	// Les données utilisateur devraient être identiques
	if newClaims.UserID != originalClaims.UserID {
		t.Errorf("UserID devrait être conservé: %d vs %d", originalClaims.UserID, newClaims.UserID)
	}

	if newClaims.Username != originalClaims.Username {
		t.Errorf("Username devrait être conservé: %s vs %s", originalClaims.Username, newClaims.Username)
	}
}

func TestTokenManager_RefreshToken_TooOld(t *testing.T) {
	tm := NewTokenManager(createTestConfig())

	// Créer un token avec une date d'émission très ancienne
	oldClaims := Claims{
		UserID:   123,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now().Add(-8 * 24 * time.Hour)), // 8 jours dans le passé
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, oldClaims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Erreur création token ancien: %v", err)
	}

	// Le refresh devrait échouer
	_, err = tm.RefreshToken(tokenString)
	if err == nil {
		t.Error("Le refresh d'un token trop ancien devrait échouer")
	}

	if !strings.Contains(err.Error(), "trop ancien") {
		t.Errorf("L'erreur devrait mentionner que le token est trop ancien, obtenu: %v", err)
	}
}

func TestTokenManager_ExtractClaims(t *testing.T) {
	tm := NewTokenManager(createTestConfig())
	originalClaims := createTestClaims()

	token, err := tm.GenerateToken(originalClaims)
	if err != nil {
		t.Fatalf("Erreur génération token: %v", err)
	}

	// Extraire les claims
	extractedClaims, err := tm.ExtractClaims(token)
	if err != nil {
		t.Fatalf("Erreur extraction claims: %v", err)
	}

	if extractedClaims.UserID != originalClaims.UserID {
		t.Errorf("UserID attendu: %d, obtenu: %d", originalClaims.UserID, extractedClaims.UserID)
	}

	if extractedClaims.Username != originalClaims.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", originalClaims.Username, extractedClaims.Username)
	}
}

func TestHelperFunctions(t *testing.T) {
	secret := "test-secret"
	expirationHours := 1

	// Test GenerateToken helper
	token, err := GenerateToken(456, "helperuser", "helper@example.com", true, secret, expirationHours)
	if err != nil {
		t.Fatalf("Erreur GenerateToken helper: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken helper devrait retourner un token")
	}

	// Test ValidateToken helper
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Erreur ValidateToken helper: %v", err)
	}

	if claims.UserID != 456 {
		t.Errorf("UserID attendu: 456, obtenu: %d", claims.UserID)
	}

	if claims.Username != "helperuser" {
		t.Errorf("Username attendu: helperuser, obtenu: %s", claims.Username)
	}

	if !claims.IsAdmin {
		t.Error("IsAdmin devrait être true")
	}

	// Test GetTokenExpiration
	expiration, err := GetTokenExpiration(token, secret)
	if err != nil {
		t.Fatalf("Erreur GetTokenExpiration: %v", err)
	}

	if expiration == nil {
		t.Error("GetTokenExpiration ne devrait pas retourner nil")
	}

	if expiration.Before(time.Now()) {
		t.Error("Le token ne devrait pas être déjà expiré")
	}

	// Test IsTokenExpired
	isExpired, err := IsTokenExpired(token, secret)
	if err != nil {
		t.Fatalf("Erreur IsTokenExpired: %v", err)
	}

	if isExpired {
		t.Error("Le token ne devrait pas être expiré")
	}
}

func TestValidateClaims(t *testing.T) {
	tm := NewTokenManager(createTestConfig()).(*tokenManager)

	tests := []struct {
		name    string
		claims  Claims
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Claims valides",
			claims:  createTestClaims(),
			wantErr: false,
		},
		{
			name: "UserID manquant",
			claims: Claims{
				Username: "test",
				Email:    "test@example.com",
			},
			wantErr: true,
			errMsg:  "UserID manquant",
		},
		{
			name: "Username manquant",
			claims: Claims{
				UserID: 123,
				Email:  "test@example.com",
			},
			wantErr: true,
			errMsg:  "Username manquant",
		},
		{
			name: "Email manquant",
			claims: Claims{
				UserID:   123,
				Username: "test",
			},
			wantErr: true,
			errMsg:  "Email manquant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Définir le subject pour les tests
			if tt.claims.UserID != 0 {
				tt.claims.Subject = fmt.Sprintf("%d", tt.claims.UserID)
			}

			err := tm.validateClaims(&tt.claims)

			if tt.wantErr {
				if err == nil {
					t.Error("Devrait retourner une erreur")
					return
				}

				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Erreur devrait contenir %q, obtenu: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Erreur inattendue: %v", err)
			}
		})
	}
}

func BenchmarkTokenGeneration(b *testing.B) {
	tm := NewTokenManager(createTestConfig())
	claims := createTestClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tm.GenerateToken(claims)
		if err != nil {
			b.Fatalf("Erreur génération token: %v", err)
		}
	}
}

func BenchmarkTokenValidation(b *testing.B) {
	tm := NewTokenManager(createTestConfig())
	claims := createTestClaims()

	// Générer un token une fois
	token, err := tm.GenerateToken(claims)
	if err != nil {
		b.Fatalf("Erreur génération token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tm.ValidateToken(token)
		if err != nil {
			b.Fatalf("Erreur validation token: %v", err)
		}
	}
}
