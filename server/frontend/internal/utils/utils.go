package utils

import (
	"regexp"
	"strings"
)

// base64urlRegex ensures only valid base64url characters are allowed
var base64urlRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// IsJWT validates whether the given string is a valid JWT token format
func IsJWT(token string) bool {
	// JWT must have exactly 3 parts (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	// Validate each part for base64url encoding
	for _, part := range parts {
		if part == "" || !base64urlRegex.MatchString(part) {
			return false
		}
	}

	return true
}
