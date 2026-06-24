package config

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	ServerPort      string // alias for Port for teammate compatibility
	DBPath          string
	JWTSecret       []byte
	LNDHost         string
	LNDMacaroonHex  string
	LNDTLSCertPath  string
	BTCRPCHost      string
	BTCRPCUser      string
	BTCRPCPass      string
	KeyholderKeys   []string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	keys := make([]string, 5)
	for i := 1; i <= 5; i++ {
		envKey := fmt.Sprintf("KEYHOLDER_%d_WIF", i)
		val := os.Getenv(envKey)
		if val == "" {
			// Generate deterministic keyholder WIF for dev/test fallback
			seed := fmt.Sprintf("keyholder-seed-%d-pamojabuild", i)
			hash := sha256.Sum256([]byte(seed))
			privKey, _ := btcec.PrivKeyFromBytes(hash[:])
			wif, _ := btcutil.NewWIF(privKey, &chaincfg.RegressionNetParams, true)
			val = wif.String()
		}
		keys[i-1] = val
	}

	port := getEnv("SERVER_PORT", "8080")
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	jwtSecretStr := getEnv("JWT_SECRET", "dev-secret-key-change-in-production")

	return &Config{
		Port:           port,
		ServerPort:     port,
		DBPath:         getEnv("DB_PATH", "./data/pamoja.db"),
		JWTSecret:      []byte(jwtSecretStr),
		LNDHost:        getEnv("LND_HOST", "localhost:8080"),
		LNDMacaroonHex: getEnv("LND_MACAROON_HEX", ""),
		LNDTLSCertPath: getEnv("LND_TLS_CERT_PATH", "./tls.cert"),
		BTCRPCHost:     getEnv("BTC_RPC_HOST", "localhost:18443"),
		BTCRPCUser:     getEnv("BTC_RPC_USER", "polaruser"),
		BTCRPCPass:     getEnv("BTC_RPC_PASS", "polarpass"),
		KeyholderKeys:  keys,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
