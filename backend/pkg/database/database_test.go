package database

import (
	"testing"
	"time"

	"rythmitbackend/configs"
)

func TestDatabaseConnection(t *testing.T) {
	// Skip si on est en CI/CD sans MySQL
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	// Tester la connexion
	err := Connect(cfg)
	if err != nil {
		t.Fatalf("Impossible de se connecter à la base de données: %v", err)
	}
	defer Close()

	// Vérifier que la connexion est active
	if DB == nil {
		t.Fatal("La connexion DB est nil")
	}

	// Test ping
	err = DB.Ping()
	if err != nil {
		t.Fatalf("Ping de la base de données échoué: %v", err)
	}
}

func TestDatabaseHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	// Se connecter
	err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion échouée: %v", err)
	}
	defer Close()

	// Tester Health
	err = Health()
	if err != nil {
		t.Errorf("Health check échoué: %v", err)
	}
}

func TestDatabaseQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion échouée: %v", err)
	}
	defer Close()

	// Test simple query
	var result int
	err = DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("Query simple échouée: %v", err)
	}

	if result != 1 {
		t.Errorf("Résultat attendu: 1, obtenu: %d", result)
	}
}

func TestDatabaseTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion échouée: %v", err)
	}
	defer Close()

	// Vérifier que les tables existent
	tables := []string{
		"users",
		"threads",
		"messages",
		"tags",
		"battles",
		"battle_options",
	}

	for _, table := range tables {
		var tableName string
		query := "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_name = ?"
		err := DB.QueryRow(query, cfg.Database.Name, table).Scan(&tableName)

		if err != nil {
			t.Errorf("Table %s n'existe pas: %v", table, err)
		}
	}
}

func TestConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion échouée: %v", err)
	}
	defer Close()

	// Vérifier les paramètres du pool
	stats := DB.Stats()

	if stats.MaxOpenConnections != 25 {
		t.Errorf("MaxOpenConnections attendu: 25, obtenu: %d", stats.MaxOpenConnections)
	}

	// Test avec plusieurs connexions simultanées
	for i := 0; i < 10; i++ {
		go func() {
			var result int
			DB.QueryRow("SELECT 1").Scan(&result)
		}()
	}

	// Attendre un peu
	time.Sleep(100 * time.Millisecond)

	// Vérifier qu'on n'a pas d'erreurs
	if err := DB.Ping(); err != nil {
		t.Errorf("Ping échoué après connexions multiples: %v", err)
	}
}
