package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	WorkerCount int
	JobBuffer   int
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found; using default environment settings")
	}

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgresql://herald:heraldpassword@localhost:5432/herald?sslmode=disable")
	workers, _ := strconv.Atoi(getEnv("WORKER_COUNT", "5"))
	buffer, _ := strconv.Atoi(getEnv("JOB_BUFFER_SIZE", "100"))

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		WorkerCount: workers,
		JobBuffer:   buffer,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
