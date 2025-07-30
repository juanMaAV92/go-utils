package env

import (
	"fmt"
	"os"
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
