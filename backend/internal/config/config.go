package config

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type Config struct {
	Port            string
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

	return &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      []byte(getEnv("JWT_SECRET", "dev-secret-key-change-in-production")),
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