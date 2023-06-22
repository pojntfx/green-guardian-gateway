package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

func GetStringEnvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return value
	}

	return defaultValue
}

func GetBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return value == "true"
	}

	return defaultValue
}

func GetIntEnvOrDefault(key string, defaultValue int) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return strconv.Atoi(value)
	}

	return defaultValue, nil
}

func GetDurationEnvOrDefault(key string, defaultValue time.Duration) (time.Duration, error) {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return time.ParseDuration(value)
	}

	return defaultValue, nil
}
