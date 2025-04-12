package common

import (
	"log"
	"os"
)

// GetEnvOrDefault looks up an environment variable or returns a fallback.
func GetEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using default for env var %s: %s", key, fallback)
	return fallback
}
