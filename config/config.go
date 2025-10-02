package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	PgURL       string
	NodeURL     string
}

func NewConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "prod"),
		PgURL:       getEnv("PG_URL", "postgres://postgres:postgres@0.0.0.0:5444/explorer?sslmode=disable"),
		NodeURL:     getEnv("NODE_URL", "https://rpc.audius.engineering"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
