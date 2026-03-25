package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"log"
	"math/big"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	SupabaseURL       string
	SupabaseAnonKey   string
	Environment       string
	SupabasePublicKey *ecdsa.PublicKey
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	rawX := getEnv("SUPABASE_JWK_X", "")
	rawY := getEnv("SUPABASE_JWK_Y", "")
	var pubKey *ecdsa.PublicKey
	if rawX != "" && rawY != "" {
		pubKey = parseSupabaseKey(rawX, rawY)
	} else {
		log.Println("WARNING: SUPABASE_JWK_X or Y not found. JWT validation will fail.")
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:   getEnv("SUPABASE_ANON_KEY", ""),
		Environment:       getEnv("ENVIRONMENT", "development"),
		SupabasePublicKey: pubKey,
	}
}

func parseSupabaseKey(xStr, yStr string) *ecdsa.PublicKey {
	pubX, errX := base64.RawURLEncoding.DecodeString(xStr)
	pubY, errY := base64.RawURLEncoding.DecodeString(yStr)

	if errX != nil || errY != nil {
		log.Fatal("Error decoding Supabase JWK coordinates from Base64")
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pubX),
		Y:     new(big.Int).SetBytes(pubY),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
