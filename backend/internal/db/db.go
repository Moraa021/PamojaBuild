package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Enable foreign keys
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	log.Println("SQLite database connected")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func RunMigrations() error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Tasks table (for Teammate 2)
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			creator_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			region TEXT NOT NULL,
			location_detail TEXT,
			status TEXT DEFAULT 'open',
			goal_sats INTEGER,
			max_volunteers INTEGER NOT NULL,
			volunteer_mode TEXT NOT NULL,
			image_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(creator_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Volunteers table (for Teammate 2)
		`CREATE TABLE IF NOT EXISTS volunteers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(task_id, user_id),
			FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Donations table (for Teammate 3)
		`CREATE TABLE IF NOT EXISTS donations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			donor_id INTEGER NOT NULL,
			amount_sats INTEGER NOT NULL,
			payment_hash TEXT UNIQUE,
			status TEXT DEFAULT 'pending',
			is_anonymous INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY(donor_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Keyholders table (for Teammate 3)
		`CREATE TABLE IF NOT EXISTS keyholders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		// Payout requests table (for Teammate 3)
		`CREATE TABLE IF NOT EXISTS payout_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);`,

		// Payout signatures table (for Teammate 3)
		`CREATE TABLE IF NOT EXISTS payout_signatures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			payout_request_id INTEGER NOT NULL,
			keyholder_id INTEGER NOT NULL,
			signed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(payout_request_id, keyholder_id),
			FOREIGN KEY(payout_request_id) REFERENCES payout_requests(id) ON DELETE CASCADE,
			FOREIGN KEY(keyholder_id) REFERENCES keyholders(id) ON DELETE CASCADE
		);`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_region ON tasks(region);`,
		`CREATE INDEX IF NOT EXISTS idx_volunteers_task_id ON volunteers(task_id);`,
		`CREATE INDEX IF NOT EXISTS idx_volunteers_user_id ON volunteers(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_donations_task_id ON donations(task_id);`,
		`CREATE INDEX IF NOT EXISTS idx_donations_payment_hash ON donations(payment_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_payout_requests_task_id ON payout_requests(task_id);`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			log.Printf("Migration failed: %v\nSQL: %s", err, migration)
			return err
		}
	}

	log.Println("Database migrations completed")
	return nil
}
