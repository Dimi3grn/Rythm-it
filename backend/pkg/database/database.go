package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"rythmitbackend/configs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DB instance globale de la base de données
	DB *sql.DB
)

// Connect initialise la connexion à la base de données MySQL
func Connect(cfg *configs.Config) error {
	var err error

	// Construction de la DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.Charset,
	)

	// Ouverture de la connexion
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("erreur ouverture connexion MySQL: %w", err)
	}

	// Configuration du pool de connexions
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test de la connexion
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("erreur ping MySQL: %w", err)
	}

	log.Printf("✅ Connexion MySQL établie: %s@%s:%s/%s",
		cfg.Database.User,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	return nil
}

// Close ferme la connexion à la base de données
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Health vérifie l'état de la connexion
func Health() error {
	if DB == nil {
		return fmt.Errorf("base de données non connectée")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("base de données non accessible: %w", err)
	}

	return nil
}

// BeginTransaction démarre une nouvelle transaction
func BeginTransaction() (*sql.Tx, error) {
	return DB.Begin()
}

// Query helper pour les requêtes SELECT
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return DB.Query(query, args...)
}

// QueryRow helper pour les requêtes SELECT qui retournent une seule ligne
func QueryRow(query string, args ...interface{}) *sql.Row {
	return DB.QueryRow(query, args...)
}

// Exec helper pour les requêtes INSERT, UPDATE, DELETE
func Exec(query string, args ...interface{}) (sql.Result, error) {
	return DB.Exec(query, args...)
}
