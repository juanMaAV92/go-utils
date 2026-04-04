package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// MustHave panics if any of the given environment variables are unset or blank.
// Lists all missing variables in a single panic message instead of failing on the first one.
func MustHave(keys ...string) {
	var missing []string
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		panic("missing required environment variables: " + strings.Join(missing, ", "))
	}
}

// GetEnv returns the value of key. Panics if the variable is unset or blank.
// Intended for required configuration — fail fast at startup.
func GetEnv(key string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		panic("environment variable " + key + " is not set or empty")
	}
	return value
}

// GetEnvWithDefault returns the value of key, or defaultValue if unset or blank.
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

// GetEnvironment returns the value of ENVIRONMENT, defaulting to "local".
func GetEnvironment() string {
	if env := os.Getenv(EnvironmentKey); env != "" {
		return env
	}
	return LocalEnvironment
}

// GetEnvAsIntWithDefault returns key parsed as int, or defaultValue if unset.
// Panics if the variable is set but not a valid integer.
func GetEnvAsIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("env %q is not a valid integer: %q", key, value))
	}
	return n
}

// GetEnvAsDurationWithDefault returns key parsed as time.Duration, or defaultValue if unset.
// Panics if the variable is set but not a valid duration string (e.g. "30s", "1m").
func GetEnvAsDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("env %q is not a valid duration: %q", key, value))
	}
	return d
}

// GetEnvAsBoolWithDefault returns key parsed as bool, or defaultValue if unset.
// Accepts the same values as strconv.ParseBool: "1", "t", "true", "0", "f", "false".
// Panics if the variable is set but not a valid bool string.
func GetEnvAsBoolWithDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf("env %q is not a valid bool: %q", key, value))
	}
	return b
}

// GetEnvAsSliceWithDefault returns key split by sep as a string slice, or defaultValue if unset or blank.
// Leading and trailing whitespace is trimmed from each element.
// Empty elements after trimming are excluded.
//
// Example: GetEnvAsSliceWithDefault("ORIGINS", ",", nil)
// with ORIGINS="http://a.com, http://b.com" → ["http://a.com", "http://b.com"]
func GetEnvAsSliceWithDefault(key, sep string, defaultValue []string) []string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	parts := strings.Split(value, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}
