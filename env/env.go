package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" || strings.TrimSpace(value) == "" {
		panic("Environment variable " + key + " is not set or empty")
	}
	return value
}

func GetEnviroment() string {
	env := os.Getenv(Enviroment)
	if env != "" {
		return env
	}
	return LocalEnvironment
}

func GetEnvAsDurationWithDefault(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		panic(fmt.Sprintf("Error during env '%s' conversion to duration: %s", key, value))
	}

	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	panic(fmt.Sprintf("Error during default value '%s' conversion to duration for key '%s'", defaultValue, key))
}

func GetEnvAsIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		panic(fmt.Sprintf("Error during env '%s' conversion to int", key))
	}
	return defaultValue
}
