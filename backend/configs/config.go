package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config structure principale contenant toute la configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
	Security SecurityConfig
}

// AppConfig configuration de l'application
type AppConfig struct {
	Name        string
	Environment string
	Port        string
	Version     string
	URL         string
}

// DatabaseConfig configuration MySQL
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Charset  string
}

// JWTConfig configuration des tokens JWT
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	RefreshExpDays  int
}

// ServerConfig configuration du serveur HTTP
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// SecurityConfig configuration sécurité
type SecurityConfig struct {
	BcryptCost        int
	MinPasswordLength int
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
}

// instance unique de configuration (singleton)
var instance *Config

// Load charge la configuration depuis les variables d'environnement
func Load() *Config {
	if instance != nil {
		return instance
	}

	// Charger le fichier .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Pas de fichier .env trouvé, utilisation des valeurs par défaut")
	}

	instance = &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Rythmit"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8085"),
			Version:     getEnv("APP_VERSION", "0.1.0"),
			URL:         getEnv("APP_URL", "http://localhost:8085"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "rythmit_db"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-this-secret-key-in-production"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpDays:  getEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 7),
		},
		Server: ServerConfig{
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 15)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 15)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		},
		Security: SecurityConfig{
			BcryptCost:        getEnvAsInt("BCRYPT_COST", 12),
			MinPasswordLength: getEnvAsInt("MIN_PASSWORD_LENGTH", 12),
			MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:   time.Duration(getEnvAsInt("LOCKOUT_DURATION_MINUTES", 15)) * time.Minute,
		},
	}

	// Log de la configuration chargée (sans les secrets)
	log.Printf("✅ Configuration chargée pour l'environnement: %s", instance.App.Environment)
	log.Printf("   - App: %s v%s sur port %s", instance.App.Name, instance.App.Version, instance.App.Port)
	log.Printf("   - DB: %s@%s:%s/%s", instance.Database.User, instance.Database.Host, instance.Database.Port, instance.Database.Name)

	return instance
}

// Get retourne l'instance de configuration (doit être appelé après Load)
func Get() *Config {
	if instance == nil {
		log.Fatal("❌ Configuration non chargée. Appelez config.Load() d'abord")
	}
	return instance
}

// getEnv récupère une variable d'environnement ou retourne la valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt récupère une variable d'environnement comme int ou retourne la valeur par défaut
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// IsDevelopment vérifie si on est en environnement de développement
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction vérifie si on est en environnement de production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// GetDSN retourne la chaîne de connexion MySQL
func (c *Config) GetDSN() string {
	return c.Database.User + ":" + c.Database.Password + "@tcp(" +
		c.Database.Host + ":" + c.Database.Port + ")/" +
		c.Database.Name + "?charset=" + c.Database.Charset +
		"&parseTime=True&loc=Local"
}
