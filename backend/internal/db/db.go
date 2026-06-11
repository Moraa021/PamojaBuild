package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(dataSourceName string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	db := &DB{DB: sqlDB}

	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func (db *DB) runMigrations() error {
	migrationsDir := "db/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, file.Name()))
		}
	}

	for _, path := range migrationFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", path, err)
		}
	}

	return nil
}