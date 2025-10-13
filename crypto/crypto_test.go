package crypto

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestCrypto(t *testing.T) {
	password := "as2354@#$"
	salt, err := GeneratePasswordSalt()
	if err != nil {
		t.Fatalf("Error generating password salt: %v", err)
	}
	hash, err := HashPassword(password, salt)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	valid := ValidatePassword(password, salt, hash)
	if !valid {
		t.Fatalf("Error validating password: %v", err)
	}

	assert.Equal(t, true, valid)
}
