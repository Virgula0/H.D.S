package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

// IsJWT checks if a string is in valid JWT format.
func IsJWT(t string) bool {
	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		// base64 decode should not return an error.
		_, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return false
		}
	}

	return true
}

func StringToBase64String(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// IsValidUsername
// At least 6 chars
func IsValidUsername(username string) bool {
	return len(username) > 6
}

// IsValidPassword checks if the password meets the following criteria:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one number
// - Contains at least one special character
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasDigit && hasSpecial
}

// GenerateToken generates a secure token of the specified length.
func GenerateToken(length int) string {
	// Calculate the required byte length for the token
	byteLength := length / 2

	// Generate random bytes
	bytes := make([]byte, byteLength)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Errorf("cannot create randon token %w", err))
	}

	// Convert the random bytes to a hexadecimal string
	token := hex.EncodeToString(bytes)

	return token
}
