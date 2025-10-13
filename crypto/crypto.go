package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	SaltLength = 16
	BCryptCost = 12
)

func GeneratePasswordSalt() (string, error) {
	bytes := make([]byte, SaltLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashPassword(password, salt string) (string, error) {
	saltedPassword := password + salt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), BCryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func ValidatePassword(password, salt, hash string) bool {
	saltedPassword := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}
