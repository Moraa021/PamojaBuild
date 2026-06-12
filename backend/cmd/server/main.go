package main

import (
	"log"
	"net/http"

	"PamojaBuild/internal/config"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg.DBPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed keyholders (for Teammate 3)
	if err := seedKeyholders(); err != nil {
		log.Println("Warning: Failed to seed keyholders:", err)
	}

	// Initialize lightning client
	lightningClient, err := lightning.NewClient(cfg)
	if err != nil {
		log.Fatal("Failed to create lightning client:", err)
	}

	// Setup router
	r := router.SetupRouter(string(cfg.JWTSecret), db.SQLDB, lightningClient)

	// Start server
	serverAddr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("Health check: http://localhost%s/health", serverAddr)
	log.Printf("Counties: http://localhost%s/config/counties", serverAddr)
	log.Printf("Register: POST http://localhost%s/auth/register", serverAddr)
	log.Printf("Login: POST http://localhost%s/auth/login", serverAddr)

	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func seedKeyholders() error {
	// Check if keyholders already exist
	var count int
	err := db.SQLDB.QueryRow("SELECT COUNT(*) FROM keyholders").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Keyholders already seeded, skipping...")
		return nil
	}

	// Seed 5 keyholders (using fixed phone numbers)
	keyholderPhones := []string{
		"0711000001",
		"0711000002",
		"0711000003",
		"0711000004",
		"0711000005",
	}

	for _, phone := range keyholderPhones {
		// Check if user exists
		var userID int64
		err := db.SQLDB.QueryRow("SELECT id FROM users WHERE phone = ?", phone).Scan(&userID)
		if err != nil {
			// Create user if doesn't exist
			result, err := db.SQLDB.Exec(
				"INSERT INTO users (phone, password_hash, role) VALUES (?, ?, ?)",
				phone,
				"$2a$10$tempHashForSeededKeyholders", // Temporary hash
				"keyholder",
			)
			if err != nil {
				log.Printf("Failed to create keyholder user %s: %v", phone, err)
				continue
			}
			userID, _ = result.LastInsertId()
		} else {
			// Update existing user to keyholder role
			_, err = db.SQLDB.Exec("UPDATE users SET role = 'keyholder' WHERE id = ?", userID)
			if err != nil {
				log.Printf("Failed to update user %d to keyholder: %v", userID, err)
				continue
			}
		}

		// Add to keyholders table
		_, err = db.SQLDB.Exec("INSERT INTO keyholders (user_id) VALUES (?)", userID)
		if err != nil {
			log.Printf("Failed to add keyholder %d: %v", userID, err)
		} else {
			log.Printf("Seeded keyholder: %s (user_id: %d)", phone, userID)
		}
	}

	return nil
}