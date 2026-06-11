package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	ServerPort string

	// Database
	DBPath string

	// JWT
	JWTSecret string

	// LND (for Teammate 3)
	LNDHost        string
	LNDMacaroonHex string
	LNDTLSCertPath string
}

func Load() (*Config, error) {
	// Load .env file from backend directory
	godotenv.Load()

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "./data/pamoja.db"),
		JWTSecret:      getEnv("JWT_SECRET", "default-secret-key"),
		LNDHost:        getEnv("LND_HOST", "localhost:8080"),
		LNDMacaroonHex: getEnv("LND_MACAROON_HEX", ""),
		LNDTLSCertPath: getEnv("LND_TLS_CERT_PATH", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Counties list for Kenya
var KenyanCounties = []string{
	"Mombasa", "Kwale", "Kilifi", "Tana River", "Lamu", "Taita Taveta",
	"Garissa", "Wajir", "Mandera", "Marsabit", "Isiolo", "Meru",
	"Tharaka Nithi", "Embu", "Kitui", "Machakos", "Makueni", "Nyandarua",
	"Nyeri", "Kirinyaga", "Murang'a", "Kiambu", "Turkana", "West Pokot",
	"Samburu", "Trans Nzoia", "Uasin Gishu", "Elgeyo Marakwet", "Nandi",
	"Baringo", "Laikipia", "Nakuru", "Narok", "Kajiado", "Kericho",
	"Bomet", "Kakamega", "Vihiga", "Bungoma", "Busia", "Siaya", "Kisumu",
	"Homa Bay", "Migori", "Kisii", "Nyamira", "Nairobi",
}

func IsValidCounty(county string) bool {
	for _, c := range KenyanCounties {
		if c == county {
			return true
		}
	}
	return false
}