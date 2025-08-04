package env

import (
	"os"
	"testing"
	"time"
)

func Test_GetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectPanic   bool
		expectedValue string
	}{
		{
			name:          "Valid environment variable",
			key:           "TEST_KEY",
			value:         "test_value",
			expectPanic:   false,
			expectedValue: "test_value",
		},
		{
			name:        "Environment variable with whitespace",
			key:         "TEST_KEY_WHITESPACE",
			value:       "   ",
			expectPanic: true,
		},
		{
			name:        "Environment variable not set",
			key:         "TEST_KEY_NOT_SET",
			value:       "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic for key %s: %v", tt.key, r)
					}
				} else if tt.expectPanic {
					t.Errorf("Expected panic for key %s but did not panic", tt.key)
				}
			}()

			result := GetEnv(tt.key)
			if result != tt.expectedValue {
				t.Errorf("Expected value %s, got %s", tt.expectedValue, result)
			}
		})
	}
}

func Test_GetEnviroment(t *testing.T) {
	tests := []struct {
		name          string
		envVar        string
		envValue      string
		expectedValue string
	}{
		{
			name:          "Environment variable is set",
			envVar:        "ENVIRONMENT",
			envValue:      "production",
			expectedValue: "production",
		},
		{
			name:          "Environment variable is not set",
			envVar:        "ENVIRONMENT",
			envValue:      "",
			expectedValue: LocalEnvironment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			result := GetEnviroment()
			if result != tt.expectedValue {
				t.Errorf("Expected value %s, got %s", tt.expectedValue, result)
			}
		})
	}
}

func Test_GetEnvAsDurationWithDefault(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		defaultValue  string
		expectPanic   bool
		expectedValue time.Duration
	}{
		{
			name:          "Valid duration environment variable",
			key:           "TEST_DURATION_KEY",
			value:         "5m",
			defaultValue:  "1m",
			expectPanic:   false,
			expectedValue: 5 * time.Minute,
		},
		{
			name:         "Invalid duration environment variable",
			key:          "TEST_DURATION_KEY_INVALID",
			value:        "invalid",
			defaultValue: "1m",
			expectPanic:  true,
		},
		{
			name:          "Environment variable not set - uses default",
			key:           "TEST_DURATION_KEY_NOT_SET",
			value:         "",
			defaultValue:  "30s",
			expectPanic:   false,
			expectedValue: 30 * time.Second,
		},
		{
			name:         "Invalid default value",
			key:          "TEST_DURATION_KEY_NOT_SET",
			value:        "",
			defaultValue: "invalid",
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic for key %s: %v", tt.key, r)
					}
				} else if tt.expectPanic {
					t.Errorf("Expected panic for key %s but did not panic", tt.key)
				}
			}()

			result := GetEnvAsDurationWithDefault(tt.key, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("Expected value %v, got %v", tt.expectedValue, result)
			}
		})
	}
}

func Test_GetEnvAsIntWithDefault(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		defaultValue  int
		expectPanic   bool
		expectedValue int
	}{
		{
			name:          "Valid integer environment variable",
			key:           "TEST_INT_KEY",
			value:         "42",
			defaultValue:  0,
			expectPanic:   false,
			expectedValue: 42,
		},
		{
			name:         "Invalid integer environment variable",
			key:          "TEST_INT_KEY_INVALID",
			value:        "invalid",
			defaultValue: 0,
			expectPanic:  true,
		},
		{
			name:          "Environment variable not set",
			key:           "TEST_INT_KEY_NOT_SET",
			value:         "",
			defaultValue:  10,
			expectPanic:   false,
			expectedValue: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the environment variable
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic for key %s: %v", tt.key, r)
					}
				} else if tt.expectPanic {
					t.Errorf("Expected panic for key %s but did not panic", tt.key)
				}
			}()

			result := GetEnvAsIntWithDefault(tt.key, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("Expected value %d, got %d", tt.expectedValue, result)
			}
		})
	}
}
