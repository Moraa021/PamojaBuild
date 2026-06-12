package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var SQLDB *sql.DB

type DB struct {
	*sql.DB
}

func Init(dbPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	SQLDB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = SQLDB.Ping(); err != nil {
		return err
	}

	// Enable foreign keys
	_, err = SQLDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	log.Println("SQLite database connected")
	return nil
}

func Close() {
	if SQLDB != nil {
		SQLDB.Close()
	}
}

func RunMigrations() error {
	// Temporarily disable foreign keys during migrations
	_, err := SQLDB.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		return err
	}

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

		if _, err := SQLDB.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", path, err)
		}
	}

	// Re-enable foreign keys
	_, err = SQLDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func New(dataSourceName string) (*DB, error) {
	if err := Init(dataSourceName); err != nil {
		return nil, err
	}
	if err := RunMigrations(); err != nil {
		return nil, err
	}
	return &DB{DB: SQLDB}, nil
}

func (db *DB) Close() {
	Close()
}
