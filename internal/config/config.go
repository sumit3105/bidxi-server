package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	Port         string
	GinMode      string
	MongoURI     string
	MongoDBName  string
	JWTSecret    string
	JWTExpiry    time.Duration
	RateLimitRPM int
}

// global singleton
var AppConfig *Config

// Load reads environment variables (with .env fallback) and populates AppConfig.
func Load() {
	// Load .env file if it exists (ignored in production where env vars are set directly)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		expiryHours = 24
	}

	rateLimitRPM, err := strconv.Atoi(getEnv("RATE_LIMIT_RPM", "3000"))
	if err != nil {
		rateLimitRPM = 3000
	}

	AppConfig = &Config{
		Port:         getEnv("PORT", "8080"),
		GinMode:      getEnv("GIN_MODE", "debug"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:  getEnv("MONGO_DATABASE", "bidxi"),
		JWTSecret:    getEnv("JWT_SECRET", "bidxi-dev-secret-change-me"),
		JWTExpiry:    time.Duration(expiryHours) * time.Hour,
		RateLimitRPM: rateLimitRPM,
	}
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
