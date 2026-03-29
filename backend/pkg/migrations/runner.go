package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Run applies all pending SQL migrations from the given directory.
// It tracks applied migrations in a `schema_migrations` table so each
// file runs exactly once, even across restarts and redeployments.
func Run(db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("listing migration files: %w", err)
	}
	sort.Strings(files)

	applied, err := appliedMigrations(db)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)

		// Skip go-migrate up/down files and the bare CREATE DATABASE file
		if strings.HasSuffix(name, ".up.sql") ||
			strings.HasSuffix(name, ".down.sql") ||
			name == "000_create_database.sql" {
			continue
		}

		if applied[name] {
			log.Printf("✅ Migration déjà appliquée: %s", name)
			continue
		}

		log.Printf("🔄 Application de la migration: %s", name)
		if err := applyFile(db, file); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}

		if _, err := db.Exec(
			"INSERT INTO schema_migrations (name) VALUES (?)", name,
		); err != nil {
			return fmt.Errorf("recording migration %s: %w", name, err)
		}
		log.Printf("✅ Migration appliquée: %s", name)
	}
	return nil
}

// ensureMigrationsTable creates the tracking table if it doesn't exist.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	return err
}

// appliedMigrations returns a set of already-applied migration filenames.
func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

// applyFile executes every statement in a SQL file.
// Lines starting with USE or CREATE DATABASE are skipped since Railway
// already provides the target database via the connection string.
// Errors for duplicate columns/keys are tolerated so re-running is safe.
func applyFile(db *sql.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Remove lines that would conflict with Railway's DB context
	var filteredLines []string
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(strings.ToUpper(line))
		if strings.HasPrefix(trimmed, "USE ") ||
			strings.HasPrefix(trimmed, "CREATE DATABASE") ||
			strings.HasPrefix(trimmed, "DROP DATABASE") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	content := strings.Join(filteredLines, "\n")

	// Split on semicolons and run each statement individually
	stmts := strings.Split(content, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || stmt == "--" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			if isIdempotentError(err) {
				log.Printf("⚠️  Ignoré (déjà appliqué): %v", err)
				continue
			}
			return fmt.Errorf("executing statement: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

// isIdempotentError returns true for MySQL errors that mean the schema
// change was already applied (duplicate column, key already exists, etc.).
func isIdempotentError(err error) bool {
	msg := err.Error()
	idempotentPhrases := []string{
		"Duplicate column name",
		"Duplicate key name",
		"Can't DROP",
		"check that column/key exists",
		"already exists",
		"Unknown column",    // DROP COLUMN on missing column
		"duplicate column",
	}
	for _, phrase := range idempotentPhrases {
		if strings.Contains(msg, phrase) {
			return true
		}
	}
	return false
}
