package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHasher(t *testing.T) {
	// Test avec coût valide
	hasher := NewPasswordHasher(12)
	if hasher == nil {
		t.Error("NewPasswordHasher ne devrait pas retourner nil")
	}

	// Test avec coût invalide (trop bas)
	hasher2 := NewPasswordHasher(2)
	if hasher2 == nil {
		t.Error("NewPasswordHasher devrait créer un hasher même avec un coût invalide")
	}

	// Test avec coût invalide (trop haut)
	hasher3 := NewPasswordHasher(50)
	if hasher3 == nil {
		t.Error("NewPasswordHasher devrait créer un hasher même avec un coût invalide")
	}
}

func TestHashPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost) // Coût minimum pour les tests

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Mot de passe valide",
			password: "ValidPassword123!",
			wantErr:  false,
		},
		{
			name:     "Mot de passe complexe",
			password: "SuperSecure456@Complex",
			wantErr:  false,
		},
		{
			name:     "Mot de passe vide",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Mot de passe trop court",
			password: "Short1!",
			wantErr:  true,
		},
		{
			name:     "Mot de passe sans majuscule",
			password: "nouppercase123!",
			wantErr:  true,
		},
		{
			name:     "Mot de passe sans caractère spécial",
			password: "NoSpecialChar123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() devrait retourner une erreur pour %q", tt.password)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() erreur inattendue: %v", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() devrait retourner un hash non-vide")
			}

			// Vérifier que le hash est différent du mot de passe original
			if hash == tt.password {
				t.Error("Le hash ne devrait pas être identique au mot de passe")
			}

			// Vérifier que le hash peut être vérifié
			if !hasher.CheckPassword(tt.password, hash) {
				t.Error("Le hash généré ne peut pas être vérifié")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)
	password := "TestPassword123!"

	// Générer un hash
	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("Erreur génération hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "Mot de passe correct",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "Mot de passe incorrect",
			password: "WrongPassword123!",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Mot de passe vide",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Hash vide",
			password: password,
			hash:     "",
			want:     false,
		},
		{
			name:     "Hash invalide",
			password: password,
			hash:     "invalid-hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.CheckPassword(tt.password, tt.hash)
			if result != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Mot de passe valide complet",
			password: "ValidPassword123!",
			wantErr:  false,
		},
		{
			name:     "Mot de passe très complexe",
			password: "Super$ecure!Passw0rd2024",
			wantErr:  false,
		},
		{
			name:     "Trop court",
			password: "Short1!",
			wantErr:  true,
			errMsg:   "12 caractères",
		},
		{
			name:     "Sans majuscule",
			password: "nouppercase123!",
			wantErr:  true,
			errMsg:   "majuscule",
		},
		{
			name:     "Sans caractère spécial",
			password: "NoSpecialChar123",
			wantErr:  true,
			errMsg:   "caractère spécial",
		},
		{
			name:     "Sans minuscule",
			password: "NOLOWERCASE123!",
			wantErr:  true,
			errMsg:   "minuscule",
		},
		{
			name:     "Sans chiffre",
			password: "NoNumbersHere!",
			wantErr:  true,
			errMsg:   "chiffre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.ValidatePasswordStrength(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordStrength() devrait retourner une erreur pour %q", tt.password)
					return
				}

				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Erreur devrait contenir %q, obtenu: %v", tt.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidatePasswordStrength() erreur inattendue pour %q: %v", tt.password, err)
			}
		})
	}
}

func TestEstimatePasswordStrength(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.DefaultCost).(*bcryptHasher)

	tests := []struct {
		name     string
		password string
		minScore int
		maxScore int
	}{
		{
			name:     "Mot de passe vide",
			password: "",
			minScore: 0,
			maxScore: 0,
		},
		{
			name:     "Mot de passe faible",
			password: "password",
			minScore: 15,
			maxScore: 40,
		},
		{
			name:     "Mot de passe moyen",
			password: "Password123",
			minScore: 50,
			maxScore: 75,
		},
		{
			name:     "Mot de passe fort",
			password: "SecurePassword123!",
			minScore: 85,
			maxScore: 100,
		},
		{
			name:     "Mot de passe très fort",
			password: "VerySecure!Password2024@Complex",
			minScore: 90,
			maxScore: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := hasher.EstimatePasswordStrength(tt.password)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("EstimatePasswordStrength(%q) = %d, want between %d and %d",
					tt.password, score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Longueur par défaut",
			length: 12,
		},
		{
			name:   "Longueur personnalisée",
			length: 16,
		},
		{
			name:   "Longueur trop petite",
			length: 8, // Devrait être ajusté à 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateSecurePassword(tt.length)
			if err != nil {
				t.Errorf("GenerateSecurePassword() erreur: %v", err)
				return
			}

			if password == "" {
				t.Error("GenerateSecurePassword() ne devrait pas retourner un mot de passe vide")
			}

			// Vérifier que le mot de passe généré respecte les critères
			hasher := NewPasswordHasher(bcrypt.DefaultCost)
			if err := hasher.ValidatePasswordStrength(password); err != nil {
				t.Errorf("Le mot de passe généré ne respecte pas les critères: %v", err)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	password := "TestPassword123!"

	// Test HashPassword helper
	hash, err := HashPassword(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword() helper erreur: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() helper devrait retourner un hash")
	}

	// Test CheckPassword helper
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword() helper devrait valider le mot de passe")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Error("CheckPassword() helper ne devrait pas valider un mauvais mot de passe")
	}

	// Test ValidatePasswordStrength helper
	if err := ValidatePasswordStrength(password); err != nil {
		t.Errorf("ValidatePasswordStrength() helper erreur: %v", err)
	}

	if err := ValidatePasswordStrength("weak"); err == nil {
		t.Error("ValidatePasswordStrength() helper devrait rejeter un mot de passe faible")
	}
}

func TestPasswordHashingConsistency(t *testing.T) {
	hasher := NewPasswordHasher(bcrypt.MinCost)
	password := "ConsistentTest123!"

	// Générer plusieurs hash du même mot de passe
	hash1, err1 := hasher.HashPassword(password)
	hash2, err2 := hasher.HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("Erreur génération hash: %v, %v", err1, err2)
	}

	// Les hash devraient être différents (salt différent)
	if hash1 == hash2 {
		t.Error("Les hash du même mot de passe devraient être différents grâce au salt")
	}

	// Mais les deux devraient valider le même mot de passe
	if !hasher.CheckPassword(password, hash1) {
		t.Error("Premier hash ne valide pas le mot de passe")
	}

	if !hasher.CheckPassword(password, hash2) {
		t.Error("Deuxième hash ne valide pas le mot de passe")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.HashPassword(password)
		if err != nil {
			b.Fatalf("Erreur hachage: %v", err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	hasher := NewPasswordHasher(bcrypt.DefaultCost)
	password := "BenchmarkPassword123!"

	// Générer un hash une fois
	hash, err := hasher.HashPassword(password)
	if err != nil {
		b.Fatalf("Erreur génération hash: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.CheckPassword(password, hash)
	}
}
