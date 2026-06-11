package config

import (
	"os"
)

type Config struct {
	Port            string
	JWTSecret       []byte
	LNDHost         string
	LNDMacaroonHex  string
	LNDTLSCertPath  string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      []byte(getEnv("JWT_SECRET", "dev-secret-key-change-in-production")),
		LNDHost:        getEnv("LND_HOST", "localhost:8080"),
		LNDMacaroonHex: getEnv("LND_MACAROON_HEX", ""),
		LNDTLSCertPath: getEnv("LND_TLS_CERT_PATH", "./tls.cert"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}