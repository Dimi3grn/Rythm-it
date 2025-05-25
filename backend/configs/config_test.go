package configs

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Réinitialiser l'instance pour les tests
	instance = nil

	cfg := Load()

	if cfg == nil {
		t.Fatal("Configuration est nil")
	}

	// Vérifier les valeurs par défaut
	if cfg.App.Name != "Rythmit" {
		t.Errorf("App.Name attendu: Rythmit, obtenu: %s", cfg.App.Name)
	}

	if cfg.App.Port != "8085" {
		t.Errorf("App.Port attendu: 8085, obtenu: %s", cfg.App.Port)
	}
}

func TestGetConfig(t *testing.T) {
	// Réinitialiser et charger
	instance = nil
	Load()

	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() retourne nil après Load()")
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Environment: "development",
		},
	}

	if !cfg.IsDevelopment() {
		t.Error("IsDevelopment() devrait retourner true")
	}

	if cfg.IsProduction() {
		t.Error("IsProduction() devrait retourner false")
	}

	// Test production
	cfg.App.Environment = "production"

	if cfg.IsDevelopment() {
		t.Error("IsDevelopment() devrait retourner false en production")
	}

	if !cfg.IsProduction() {
		t.Error("IsProduction() devrait retourner true")
	}
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			User:     "testuser",
			Password: "testpass",
			Host:     "localhost",
			Port:     "3306",
			Name:     "testdb",
			Charset:  "utf8mb4",
		},
	}

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := cfg.GetDSN()

	if dsn != expected {
		t.Errorf("DSN incorrect.\nAttendu: %s\nObtenu:  %s", expected, dsn)
	}
}

func TestEnvOverride(t *testing.T) {
	// Sauvegarder et définir une variable d'environnement
	oldValue := os.Getenv("APP_NAME")
	os.Setenv("APP_NAME", "RythmitTest")
	defer os.Setenv("APP_NAME", oldValue)

	// Réinitialiser et recharger
	instance = nil
	cfg := Load()

	if cfg.App.Name != "RythmitTest" {
		t.Errorf("La variable d'environnement n'a pas été prise en compte. Obtenu: %s", cfg.App.Name)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		want        int
		wantDefault bool
	}{
		{"Valid int", "42", 42, false},
		{"Invalid int", "abc", 10, true},
		{"Empty string", "", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_INT", tt.envValue)
			defer os.Unsetenv("TEST_INT")

			got := getEnvAsInt("TEST_INT", 10)

			if tt.wantDefault && got != 10 {
				t.Errorf("getEnvAsInt() = %v, want default 10", got)
			} else if !tt.wantDefault && got != tt.want {
				t.Errorf("getEnvAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
