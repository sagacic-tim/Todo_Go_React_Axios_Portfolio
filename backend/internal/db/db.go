// internal/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// RunMigrations applies all the “.up.sql” files from your migrations/ folder.
func RunMigrations(sqlDB *sql.DB) error {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	// Prefer migrations relative to the executable (stable in Docker)
	// Fallback to CWD if that path does not exist.
	migrationsURL, err := buildMigrationsURL()
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func buildMigrationsURL() (string, error) {
	// 1) executable dir
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		// Common layouts:
		// - repo root: ./migrations
		// - binary in cmd/server: ../../migrations
		candidates := []string{
			filepath.Join(exeDir, "migrations"),
			filepath.Clean(filepath.Join(exeDir, "..", "migrations")),
			filepath.Clean(filepath.Join(exeDir, "..", "..", "migrations")),
		}
		for _, p := range candidates {
			if st, statErr := os.Stat(p); statErr == nil && st.IsDir() {
				return "file://" + p, nil
			}
		}
	}

	// 2) fallback to CWD
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine migrations path: %w", err)
	}
	p := filepath.Join(cwd, "migrations")
	if st, statErr := os.Stat(p); statErr == nil && st.IsDir() {
		return "file://" + p, nil
	}

	return "", fmt.Errorf("migrations directory not found (tried near executable and %s)", p)
}
