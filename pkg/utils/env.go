package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

// GetStringEnvOrDefault gets the value of the specified environment variable.
// If it does not exist, it returns the provided default value.
func GetStringEnvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return value
	}

	return defaultValue
}

// GetBoolEnvOrDefault gets the boolean value of the specified environment variable.
// If it does not exist or is not "true", it returns the provided default value.
func GetBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return value == "true"
	}

	return defaultValue
}

// GetIntEnvOrDefault gets the integer value of the specified environment variable.
// If it does not exist or is not an integer, it returns the provided default value.
// It may return an error if the string cannot be converted into an integer.
func GetIntEnvOrDefault(key string, defaultValue int) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return strconv.Atoi(value)
	}

	return defaultValue, nil
}

// GetDurationEnvOrDefault gets the time.Duration value of the specified environment variable.
// If it does not exist or is not a valid duration, it returns the provided default value.
// It may return an error if the string cannot be parsed into a duration.
func GetDurationEnvOrDefault(key string, defaultValue time.Duration) (time.Duration, error) {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Using %v from environment", key)

		return time.ParseDuration(value)
	}

	return defaultValue, nil
}
